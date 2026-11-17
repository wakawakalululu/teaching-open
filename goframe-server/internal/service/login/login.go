// Package login 承载登录、刷新、登出业务用例。
//
// 使用 GoFrame gdb ORM（通过 dao.SysUser），全部占位符绑定，杜绝 SQL 注入。
package login

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/notify"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
)

// 登录错误（对外统一话术，不区分账号不存在与密码错误）。
var (
	// ErrInvalidCredential 账号或密码错误。
	ErrInvalidCredential = errors.New("用户名或密码错误")
	// ErrAccountLocked 账号被锁定。
	ErrAccountLocked = errors.New("密码错误次数过多，账号已临时锁定，请稍后再试")
	// ErrAccountDisabled 账号被禁用/冻结。
	ErrAccountDisabled = errors.New("账号已被锁定，请联系管理员")
)

// LoginResult 登录成功后的返回。
type LoginResult struct {
	Tokens *auth.TokenPair
	User   *entity.SysUser
	Roles  []string
}

// Service 处理登录用例。
type Service struct {
	tokens  *auth.Manager
	limiter *auth.LoginLimiter
	legacy  *passwd.LegacyPBE
	feishu  *notify.FeishuNotifier
}

// New 构造服务。
func New(tokens *auth.Manager, limiter *auth.LoginLimiter, feishu *notify.FeishuNotifier) *Service {
	return &Service{
		tokens:  tokens,
		limiter: limiter,
		legacy:  passwd.NewLegacyPBE(),
		feishu:  feishu,
	}
}

// Login 执行账号密码登录。
//
// 安全要点：
//  1. 限流前置，避免账号存在性探测；
//  2. 统一错误话术，防止账号枚举；
//  3. 时间恒定：账号不存在时也执行一次 bcrypt 校验；
//  4. rehash-on-login：存量 PBE 密文校验通过后立即升级为 bcrypt。
func (s *Service) Login(ctx context.Context, username, password, clientIP string) (*LoginResult, error) {
	limitKey := s.limiter.Key(hashSubject(username), clientIP)

	blocked, count, err := s.limiter.Blocked(ctx, limitKey)
	if err != nil {
		// Redis 不可用时 fail-open，仅告警
		g.Log().Warning(ctx, "登录限流检查失败，按放行处理", err.Error())
	}
	if blocked {
		g.Log().Warningf(ctx, "账号因失败次数过多被锁定 user=%s ip=%s fail_count=%d",
			maskText(username), clientIP, count)
		s.feishu.SendAsync(ctx, "【安全告警】账号登录失败次数过多被临时锁定\n账号: "+maskText(username)+"\nIP: "+clientIP+"\n失败次数: "+fmt.Sprint(count))
		return nil, ErrAccountLocked
	}

	user, err := s.getByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, errNotFound) {
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
		// 账号不存在：仍执行一次等价散列运算，抹平时间差
		_, _ = passwd.Hash(password)
		s.recordFailure(ctx, limitKey, username, clientIP)
		return nil, ErrInvalidCredential
	}

	if user.Status != 1 {
		return nil, ErrAccountDisabled
	}

	if err := s.verifyPassword(ctx, user, password); err != nil {
		s.recordFailure(ctx, limitKey, username, clientIP)
		return nil, ErrInvalidCredential
	}

	if err := s.limiter.Reset(ctx, limitKey); err != nil {
		g.Log().Warning(ctx, "清除登录失败计数失败", err.Error())
	}

	roles, err := s.ListRoles(ctx, user.Id)
	if err != nil {
		g.Log().Warningf(ctx, "查询用户角色失败，按无角色处理 user_id=%s err=%s", user.Id, err.Error())
		roles = nil
	}

	tokens, err := s.tokens.Issue(user.Id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("签发令牌失败: %w", err)
	}

	return &LoginResult{Tokens: tokens, User: user, Roles: roles}, nil
}

// verifyPassword 按存储方案校验密码，并在存量格式命中时触发升级。
func (s *Service) verifyPassword(ctx context.Context, user *entity.SysUser, plain string) error {
	switch passwd.Detect(user.Password) {
	case passwd.SchemeBcrypt:
		return passwd.VerifyBcrypt(plain, user.Password)

	case passwd.SchemeLegacyPBE:
		if err := s.legacy.Verify(user.Username, plain, user.Password, user.Salt); err != nil {
			return err
		}
		// 校验通过，立即升级为 bcrypt
		newHash, hashErr := passwd.Hash(plain)
		if hashErr != nil {
			g.Log().Warningf(ctx, "密码散列升级失败，本次登录仍放行 user_id=%s err=%s", user.Id, hashErr.Error())
			return nil
		}
		if _, updErr := dao.SysUser.Ctx(ctx).
			Where(do.SysUser{Id: user.Id}).
			Data(do.SysUser{Password: newHash, Salt: ""}).
			Update(); updErr != nil {
			g.Log().Warningf(ctx, "回写升级后密码失败 user_id=%s err=%s", user.Id, updErr.Error())
			return nil
		}
		g.Log().Infof(ctx, "密码散列已升级为 bcrypt user_id=%s", user.Id)
		return nil

	default:
		return passwd.ErrMismatch
	}
}

func (s *Service) recordFailure(ctx context.Context, key, username, clientIP string) {
	n, err := s.limiter.Fail(ctx, key)
	if err != nil {
		g.Log().Warning(ctx, "记录登录失败计数失败", err.Error())
		return
	}
	g.Log().Infof(ctx, "登录失败 user=%s ip=%s fail_count=%d", maskText(username), clientIP, n)
}

// Refresh 用刷新令牌换取新的令牌对，并吊销旧刷新令牌（一次性使用）。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	claims, err := s.tokens.Parse(ctx, refreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidCredential
	}
	user, err := s.getByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidCredential
	}
	if user.Status != 1 {
		return nil, ErrAccountDisabled
	}
	if revokeErr := s.tokens.Revoke(ctx, refreshToken); revokeErr != nil {
		g.Log().Warning(ctx, "吊销旧刷新令牌失败", revokeErr.Error())
	}
	roles, _ := s.ListRoles(ctx, user.Id)
	tokens, err := s.tokens.Issue(user.Id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("签发令牌失败: %w", err)
	}
	return &LoginResult{Tokens: tokens, User: user, Roles: roles}, nil
}

// Logout 吊销当前访问令牌。
func (s *Service) Logout(ctx context.Context, accessToken string) error {
	return s.tokens.Revoke(ctx, accessToken)
}

// ---- 数据访问（全部走 gdb，占位符绑定） ----

var errNotFound = errors.New("记录不存在")

func (s *Service) getByUsername(ctx context.Context, username string) (*entity.SysUser, error) {
	var user *entity.SysUser
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Username: username, DelFlag: 0}).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errNotFound
	}
	return user, nil
}

func (s *Service) getByID(ctx context.Context, id string) (*entity.SysUser, error) {
	var user *entity.SysUser
	err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: id, DelFlag: 0}).
		Scan(&user)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errNotFound
	}
	return user, nil
}

// ListRoles 查询用户的角色编码。
//
// 注意：sys_role 表没有 del_flag 字段（已核对 teachingopen2.6.sql 结构），
// 不能加 r.del_flag 过滤，否则会因未知列直接 SQL 报错。
func (s *Service) ListRoles(ctx context.Context, userID string) ([]string, error) {
	var (
		codes []string
		err   error
	)
	err = dao.SysUser.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Fields("r.role_code").
		Where("ur.user_id", userID).
		Ctx(ctx).
		Scan(&codes)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// hashSubject 对用户名做不可逆摘要，避免明文写入 Redis 键名。
func hashSubject(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
}

// maskText 对日志中的账号做脱敏（保留首尾各 1 字符）。
func maskText(s string) string {
	r := []rune(s)
	if len(r) <= 2 {
		return "**"
	}
	return string(r[:1]) + "***" + string(r[len(r)-1:])
}
