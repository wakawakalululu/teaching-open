// Package userreg 承载用户自助注册与找回密码业务用例。
//
// 契约蓝本：SysUserController.java
//   - POST /sys/user/register        L1185-1266（ShiroConfig.java L83 anon）；
//   - POST /sys/user/phoneVerification L1320-1335（L85 anon）；
//   - GET  /sys/user/passwordChange    L1340-1379（L86 anon）。
//
// 验证码存储/校验统一走 internal/service/sms（fail-closed）；站点配置
// （allowReg/_defaultRole/_defaultDepart）复用 internal/service/config 的
// GetConfig；数据库操作全部走 dao/do 占位符绑定。
//
// 有意偏差（文档化）：Java 注册与改密用 PasswordUtil.encrypt 的
// PBEWithMD5AndDES 可逆密文 + 8 字节盐；Go 版按仓库既定策略改用
// internal/passwd 的 bcrypt（salt 落空串），与登录链路 rehash-on-login 的
// 目标格式一致。存量用户不受影响（校验按 Detect 分派）。
package userreg

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
	svcConfig "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/config"
	svcSms "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sms"
)

// 业务错误（HTTP 层映射为 Result{success:false}，话术与 Java 契约逐字一致）。
var (
	// ErrUsernameTaken 用户名已注册（Java L1205 原话术）。
	ErrUsernameTaken = errors.New("用户名已注册")
	// ErrPhoneTaken 手机号已注册（Java L1211 原话术）。
	ErrPhoneTaken = errors.New("该手机号已注册")
	// ErrEmailTaken 邮箱已被注册（Java L1219 原话术）。
	ErrEmailTaken = errors.New("邮箱已被注册")
	// ErrRegClosed 未开放注册（Java L1235 原话术；allowReg 配置 != "1" 即关闭）。
	ErrRegClosed = errors.New("未开放注册")
	// ErrRegisterFailed 注册失败（Java L1263 catch-all 原话术；内部错误统一话术）。
	ErrRegisterFailed = errors.New("注册失败")
	// ErrParamMissing 改密必填参数缺失（Java L1347 原话术）。
	ErrParamMissing = errors.New("重置密码失败！")
	// ErrUserNotFound 用户不存在（Java L1366 原话术）。Java 语义即明示
	// 「未找到用户」，按任务约定照 Java 走，不做通用化隐藏。
	ErrUserNotFound = errors.New("未找到用户！")
	// ErrPasswordTooLong 密码超过 bcrypt 72 字节上限（Go 版新增校验；
	// Java DES 无此限制，前端表单本就限制长度，常规路径不会触发）。
	ErrPasswordTooLong = errors.New("密码长度超过 72 字节上限")
)

// 行为限流参数（加固项：Java register 无任何限流；数值为人为设定，
// 取值宽松以保证教室/办公网 NAT 场景下正常使用不受影响）。
const (
	// regIPLimit 同 IP 注册尝试上限（10 分钟窗口）。
	regIPLimit = 20
	// regIPLimitWindow 同 IP 限流窗口。
	regIPLimitWindow = 10 * time.Minute
	// regPhoneLimit 同手机号注册尝试上限（10 分钟窗口）。
	regPhoneLimit = 10
	// regPhoneLimitWindow 同手机号限流窗口。
	regPhoneLimitWindow = 10 * time.Minute
	// maxPasswordBytes bcrypt 明文长度上限。
	maxPasswordBytes = 72
)

// RegisterInput register 请求体（Java @RequestBody JSONObject 的键集合 + 前端
// Register.vue L237 实际提交字段：username/password/email/phone/smscode，
// realname 为 Java 支持的可选字段）。
type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Smscode  string `json:"smscode"`
	Realname string `json:"realname"`
}

// Service 注册/找回密码业务服务。
type Service struct {
	sms       *svcSms.Service
	cfg       *svcConfig.Service
	ipGate    *svcSms.RateWindow
	phoneGate *svcSms.RateWindow
}

// New 构造服务。cfgSvc 传入复用的站点配置服务实例（可为 nil，仅测试场景）。
func New(smsSvc *svcSms.Service, cfgSvc *svcConfig.Service) *Service {
	return &Service{
		sms:       smsSvc,
		cfg:       cfgSvc,
		ipGate:    svcSms.NewRateWindow(regIPLimit, regIPLimitWindow),
		phoneGate: svcSms.NewRateWindow(regPhoneLimit, regPhoneLimitWindow),
	}
}

// Register 用户注册（契约蓝本 SysUserController#userRegister L1185-1266）。
//
// 与 Java 逐条对齐的校验顺序：
//  1. 同 IP / 同手机号行为限流（Go 加固项，Java 无；限流先行以防唯一性
//     查询被滥用为存在性探测）；
//  2. username 为空 → 用手机号兜底（Java L1193-1196）；password 为空 →
//     随机 8 位小写字母+数字（Java L1197-1201 RandomUtil.randomString(8)）；
//  3. 唯一性校验：username → phone → email（Java L1203-1223 顺序；
//     getUserByName/getUserByPhone/getUserByEmail 均带 del_flag=0 过滤）；
//  4. 验证码校验（Java L1224-1230：phone 与 smscode 均非空才校验，缺任一
//     直接跳过。注意：当前前端 Register.vue 的短信验证码输入框被注释、
//     不传 smscode，该跳过语义是 Java 原有行为，照实迁移并在前端开放
//     验证码输入后自然生效）；
//  5. 注册开关（Java L1232-1237：getConfigItem("allowReg") != "1" →
//     「未开放注册」；种子数据 allowReg=0 即默认关闭，与 Java 一致）；
//  6. 事务写入用户行 + 默认角色/部门关联（Java L1241-1260 无事务，中途
//     失败会留下半截数据；Go 版收敛为原子写入，成功话术一致）。
func (s *Service) Register(ctx context.Context, in RegisterInput, clientIP string) error {
	if !s.ipGate.Allow("ip:" + clientIP) {
		g.Log().Warningf(ctx, "注册触发同 IP 行为限流 ip=%s phone=%s", clientIP, maskPhone(in.Phone))
		return svcSms.ErrTooManyRequests
	}
	if in.Phone != "" && !s.phoneGate.Allow("phone:"+in.Phone) {
		g.Log().Warningf(ctx, "注册触发同手机号行为限流 phone=%s", maskPhone(in.Phone))
		return svcSms.ErrTooManyRequests
	}

	// Java：未设置用户名则用手机号；未设置密码则随机生成
	username := in.Username
	if username == "" {
		username = in.Phone
	}
	password := in.Password
	if password == "" {
		var err error
		if password, err = randomPassword(8); err != nil {
			return err
		}
	}
	if len(password) > maxPasswordBytes {
		return ErrPasswordTooLong
	}

	// 唯一性校验（顺序与 Java 一致；phone 为空串时 Java 的 eq(phone,'') 同样
	// 不会命中（列值为 NULL），Go 版 Where(Phone, "") 语义相同）
	cols := dao.SysUser.Columns()
	n, err := dao.SysUser.Ctx(ctx).
		Where(cols.Username, username).
		Where(cols.DelFlag, 0).
		Count()
	if err != nil {
		return fmt.Errorf("查询用户名占用失败: %w", err)
	}
	if n > 0 {
		return ErrUsernameTaken
	}
	n, err = dao.SysUser.Ctx(ctx).
		Where(cols.Phone, in.Phone).
		Where(cols.DelFlag, 0).
		Count()
	if err != nil {
		return fmt.Errorf("查询手机号占用失败: %w", err)
	}
	if n > 0 {
		return ErrPhoneTaken
	}
	if in.Email != "" {
		n, err = dao.SysUser.Ctx(ctx).
			Where(cols.Email, in.Email).
			Where(cols.DelFlag, 0).
			Count()
		if err != nil {
			return fmt.Errorf("查询邮箱占用失败: %w", err)
		}
		if n > 0 {
			return ErrEmailTaken
		}
	}

	// 验证码校验：仅 phone 与 smscode 均非空时执行（Java 原语义，fail-closed
	// 见 sms.Verify）
	if in.Phone != "" && in.Smscode != "" {
		if err := s.sms.Verify(ctx, in.Phone, in.Smscode); err != nil {
			return err
		}
	}

	// 注册开关（Java 顺序：验证码之后、落库之前）
	if s.cfg != nil {
		conf, cfgErr := s.cfg.GetConfig(ctx, "allowReg")
		if cfgErr != nil {
			return fmt.Errorf("查询注册开关配置失败: %w", cfgErr)
		}
		if conf == nil || conf.ConfigValue != "1" {
			return ErrRegClosed
		}
	}

	// 密码散列（有意偏差：Java DES → bcrypt，见包注释）
	hash, hashErr := passwd.Hash(password)
	if hashErr != nil {
		return fmt.Errorf("生成密码散列失败: %w", hashErr)
	}

	// 默认角色/部门（Java _defaultRole/_defaultDepart 为逗号分隔 ID，
	// 空/缺失时 addUserWithRole/addUserWithDepart 直接跳过）
	roleIDs, cfgErr := s.configIDList(ctx, "_defaultRole")
	if cfgErr != nil {
		return cfgErr
	}
	departIDs, cfgErr := s.configIDList(ctx, "_defaultDepart")
	if cfgErr != nil {
		return cfgErr
	}

	if txErr := s.insertUserTx(ctx, username, hash, in, roleIDs, departIDs); txErr != nil {
		g.Log().Errorf(ctx, "注册写入失败 username=%s err=%s", maskText(username), txErr.Error())
		return ErrRegisterFailed
	}
	g.Log().Infof(ctx, "用户注册成功 username=%s phone=%s", maskText(username), maskPhone(in.Phone))
	return nil
}

// insertUserTx 事务写入用户与默认角色/部门关联。
//
// 字段对齐 Java L1241-1254：status=1（USER_UNFREEZE）、del_flag=0、
// activiti_sync=0（ACT_SYNC_0）；create_time/update_time 服务端写
// gtime.Now()（Java createTime 显式赋值，updateTime 由 MyBatis-Plus 填充）。
// 主键：Java ID_WORKER_STR 雪花串，Go 版沿用仓库先例 guid.S()（32 位 hex），
// 列宽 varchar(32) 兼容，对前端无契约影响。
func (s *Service) insertUserTx(ctx context.Context, username, hash string, in RegisterInput, roleIDs, departIDs []string) error {
	return dao.SysUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		userID := guid.S()
		data := do.SysUser{
			Id:           userID,
			Username:     username,
			Password:     hash,
			Salt:         "", // bcrypt 自带盐，列落空串（与 login 域 rehash 写法一致）
			Status:       1,
			DelFlag:      0,
			ActivitiSync: 0,
			CreateTime:   gtime.Now(),
			UpdateTime:   gtime.Now(),
		}
		// 空/缺省字段保持 NULL（对齐 Java entity 字段为 null 时的落库形态）
		if in.Realname != "" {
			data.Realname = in.Realname
		}
		if in.Email != "" {
			data.Email = in.Email
		}
		if in.Phone != "" {
			data.Phone = in.Phone
		}
		if _, insErr := dao.SysUser.Ctx(ctx).Data(data).Insert(); insErr != nil {
			return fmt.Errorf("写入用户失败: %w", insErr)
		}
		for _, rid := range roleIDs {
			if _, insErr := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
				Id:     guid.S(),
				UserId: userID,
				RoleId: rid,
			}).Insert(); insErr != nil {
				return fmt.Errorf("写入默认角色关联失败: %w", insErr)
			}
		}
		for _, did := range departIDs {
			if _, insErr := dao.SysUserDepart.Ctx(ctx).Data(do.SysUserDepart{
				Id:     guid.S(),
				UserId: userID,
				DepId:  did,
			}).Insert(); insErr != nil {
				return fmt.Errorf("写入默认部门关联失败: %w", insErr)
			}
		}
		return nil
	})
}

// PhoneVerification 找回密码第二步：手机号+验证码校验
// （契约蓝本 SysUserController#phoneVerification L1320-1335）。
//
// Java 语义：校验失败统一「手机验证码错误」（含 Redis 无值）；通过后把
// 验证码原样重存（Java 无 TTL，属缺陷；Go 版 ReArm 加固为 600s TTL），
// 供第三步 passwordChange 复核。
func (s *Service) PhoneVerification(ctx context.Context, phone, smscode string) error {
	if err := s.sms.Verify(ctx, phone, smscode); err != nil {
		return err
	}
	return s.sms.ReArm(ctx, phone, smscode)
}

// PasswordChange 找回密码第三步：改密
// （契约蓝本 SysUserController#passwordChange L1340-1379，GET query 参数）。
//
// 与 Java 逐条对齐：
//  1. username/password/smscode/phone 任一为空 → 「重置密码失败！」；
//     （Java @RequestParam required=true 缺参时 Spring 直接 400，前端恒传
//     全部参数，Go 版统一走业务失败话术，前端展示语义一致）
//  2. 验证码：Redis 无值 → 「短信验证码失效！」；不匹配 → 「短信验证码
//     不匹配！」——passwordChange 确实要求 smscode（Java L1343 必填参数）；
//  3. 按 username + phone + del_flag=0 查用户（Java getOne 联合条件；
//     @TableLogic 自动追加 del_flag=0），未命中 → 「未找到用户！」；
//  4. bcrypt 重散列并更新 password/salt/update_time（Java 为新盐+DES 密文）；
//  5. 加固项：改密成功后立即作废验证码（Java 不删除，窗口内可重放）；
//     作废失败仅告警，不影响已完成的改密。
func (s *Service) PasswordChange(ctx context.Context, username, password, smscode, phone string) error {
	if username == "" || password == "" || smscode == "" || phone == "" {
		return ErrParamMissing
	}
	if len(password) > maxPasswordBytes {
		return ErrPasswordTooLong
	}
	if err := s.sms.VerifyReset(ctx, phone, smscode); err != nil {
		return err
	}

	cols := dao.SysUser.Columns()
	var one entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(cols.Username, username).
		Where(cols.Phone, phone).
		Where(cols.DelFlag, 0).
		Scan(&one); err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if one.Id == "" {
		return ErrUserNotFound
	}

	hash, hashErr := passwd.Hash(password)
	if hashErr != nil {
		return fmt.Errorf("生成密码散列失败: %w", hashErr)
	}

	if _, err := dao.SysUser.Ctx(ctx).
		Where(cols.Id, one.Id).
		Data(do.SysUser{
			Password:   hash,
			Salt:       "",
			UpdateTime: gtime.Now(),
		}).Update(); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 一次性消费验证码（加固项，见函数注释）
	if err := s.sms.Consume(ctx, phone); err != nil {
		g.Log().Warningf(ctx, "改密后作废验证码失败 phone=%s err=%s", maskPhone(phone), err.Error())
	}
	g.Log().Infof(ctx, "找回密码改密成功 username=%s phone=%s", maskText(username), maskPhone(phone))
	return nil
}

// configIDList 读取逗号分隔 ID 列表型站点配置（_defaultRole/_defaultDepart）。
// 配置缺失或空串返回 nil（Java oConvertUtils.isNotEmpty 判空跳过）。
// Java split 不去空白，此处 Trim 仅更健壮（合法 ID 不含空白，语义一致）。
func (s *Service) configIDList(ctx context.Context, key string) ([]string, error) {
	if s.cfg == nil {
		return nil, nil
	}
	conf, err := s.cfg.GetConfig(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("查询配置 %s 失败: %w", key, err)
	}
	if conf == nil || conf.ConfigValue == "" {
		return nil, nil
	}
	var out []string
	for _, part := range strings.Split(conf.ConfigValue, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out, nil
}

// randomPassword 生成 n 位随机密码（小写字母+数字，对齐 hutool
// RandomUtil.randomString 的 BASE_CHAR_NUMBER 字符集）。
func randomPassword(n int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("生成随机密码失败: %w", err)
		}
		out[i] = alphabet[idx.Int64()]
	}
	return string(out), nil
}

// maskPhone 手机号脱敏（保留前 3 位与后 4 位）。
func maskPhone(phone string) string {
	if len(phone) >= 7 {
		return phone[:3] + "****" + phone[len(phone)-4:]
	}
	return phone
}

// maskText 账号脱敏（保留首尾各 1 字符，与 login 域先例一致）。
func maskText(s string) string {
	r := []rune(s)
	if len(r) <= 2 {
		return "**"
	}
	return string(r[:1]) + "***" + string(r[len(r)-1:])
}
