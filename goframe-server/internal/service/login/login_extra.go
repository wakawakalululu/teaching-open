// login_extra.go LoginController 剩余端点业务用例：
// phoneLogin / selectDepart / loginfo / visitInfo（契约出处见各方法注释）。
package login

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	svcDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/dict"
)

// ErrPhoneInvalid 手机号不存在（对齐 Java checkUserIsEffective 的失败话术族）。
var ErrPhoneInvalid = errors.New("该用户不存在")

// ErrSmsCode 手机验证码错误（Java phoneLogin 原话术）。
var ErrSmsCode = errors.New("手机验证码错误")

// PhoneLogin 手机验证码登录。
//
// 契约蓝本 LoginController#phoneLogin（POST /sys/phoneLogin，anon）：
//   - mobile 查 sys_user → 校验有效（status=1、del_flag=0）；
//   - 短信验证码由调用方（sms 域）先行校验后传入；
//   - 签发令牌并随全量字典下发（与账号登录同一响应形状）。
//
// 注：Java 版对登录日志的 redis token 登记依赖 RedisUtil，Go 版由
// auth.Manager 吊销/刷新体系承载（令牌一次性刷新对）。
func (s *Service) PhoneLogin(ctx context.Context, phone, clientIP string) (*LoginResult, error) {
	if phone == "" {
		return nil, ErrPhoneInvalid
	}
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Phone, phone).
		Where(dao.SysUser.Columns().DelFlag, 0).
		Limit(1).Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user.Id == "" {
		return nil, ErrPhoneInvalid
	}
	if user.Status != 1 {
		return nil, ErrAccountDisabled
	}
	roles, err := s.ListRoles(ctx, user.Id)
	if err != nil {
		roles = nil
	}
	tokens, err := s.tokens.Issue(user.Id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("签发令牌失败: %w", err)
	}
	var dictItems map[string][]svcDict.DictModel
	if s.dict != nil {
		if items, derr := s.dict.QueryAllDictItems(ctx); derr == nil {
			dictItems = items
		}
	}
	return &LoginResult{Tokens: tokens, User: &user, Roles: roles, DictItems: dictItems}, nil
}

// SelectDepart 登录后选择当前部门。
//
// 契约蓝本 LoginController#selectDepart（PUT /sys/selectDepart）：
// 更新 sys_user.org_code = orgCode（按 username 定位；body 未带 username
// 时取当前登录人），返回 {userInfo: 实体}。
func (s *Service) SelectDepart(ctx context.Context, username, orgCode string) (*entity.SysUser, error) {
	cols := dao.SysUser.Columns()
	if username == "" {
		return nil, errors.New("用户名为空")
	}
	if orgCode == "" {
		return nil, errors.New("部门编码为空")
	}
	if _, err := dao.SysUser.Ctx(ctx).
		Where(cols.Username, username).
		Data(map[string]interface{}{"org_code": orgCode, "update_time": gtimeNowStr()}).
		Update(); err != nil {
		return nil, fmt.Errorf("更新部门失败: %w", err)
	}
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Where(cols.Username, username).Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// LogInfo 首页登录统计（今日/总量/今日IP）。
//
// 契约蓝本 LoginController#loginfo + SysLogMapper 统计 SQL：
//   - totalVisitCount: count(1) where log_type=1；
//   - todayVisitCount: 同上 + create_time ∈ [今日0点, 明日0点)；
//   - todayIp: count(distinct ip) 同区间。
func (s *Service) LogInfo(ctx context.Context) (map[string]interface{}, error) {
	dayStart, dayEnd := todayRange()
	obj := map[string]interface{}{}
	for _, item := range []struct {
		key   string
		query string
		args  []interface{}
	}{
		{"totalVisitCount", "SELECT COUNT(1) AS n FROM sys_log WHERE log_type = 1", nil},
		{"todayVisitCount", "SELECT COUNT(1) AS n FROM sys_log WHERE log_type = 1 AND create_time >= ? AND create_time < ?", []interface{}{dayStart, dayEnd}},
		{"todayIp", "SELECT COUNT(DISTINCT ip) AS n FROM sys_log WHERE log_type = 1 AND create_time >= ? AND create_time < ?", []interface{}{dayStart, dayEnd}},
	} {
		rec, err := g.DB().GetOne(ctx, item.query, item.args...)
		if err != nil {
			return nil, fmt.Errorf("查询登录统计失败: %w", err)
		}
		obj[item.key] = rec["n"].Int()
	}
	return obj, nil
}

// VisitInfo 近 7 日访问趋势（按天聚合 visit/ip，键小写化对齐 Java
// toLowerCasePageList）。
func (s *Service) VisitInfo(ctx context.Context) ([]map[string]interface{}, error) {
	now := time.Now()
	dayEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayStart := dayEnd.AddDate(0, 0, -7)
	rows, err := g.DB().GetAll(ctx,
		`SELECT COUNT(*) AS visit, COUNT(DISTINCT ip) AS ip,
 DATE_FORMAT(create_time, '%Y-%m-%d') AS tian,
 DATE_FORMAT(create_time, '%m-%d') AS type
 FROM sys_log WHERE log_type = 1 AND create_time >= ? AND create_time < ?
 GROUP BY tian, type`, dayStart.Format("2006-01-02 15:04:05"), dayEnd.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("查询访问趋势失败: %w", err)
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]interface{}{
			"visit": r["visit"].Int(),
			"ip":    r["ip"].Int(),
			"tian":  r["tian"].String(),
			"type":  r["type"].String(),
		})
	}
	return out, nil
}

// todayRange 今日 [0点, 明日0点) 字符串区间。
func todayRange() (string, string) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return dayStart.Format("2006-01-02 15:04:05"),
		dayStart.AddDate(0, 0, 1).Format("2006-01-02 15:04:05")
}

// gtimeNowStr 当前时间字符串。
func gtimeNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// trimSpace helper（避免 strings 重复导入告警）。
func trimSpaceHelp(s string) string { return strings.TrimSpace(s) }
