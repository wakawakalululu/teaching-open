// Package sms 承载短信验证码的生成、存储、校验与「发送」编排。
//
// 契约蓝本：LoginController#sms（POST /sys/sms，ShiroConfig.java L80 anon 公开）。
// register（SysUserController L1185）、phoneVerification（L1320）、
// passwordChange（L1340）共用同一套验证码存储，本包为其提供存储原语。
//
// Java 端语义（逐条对齐）：
//   - 验证码 6 位纯数字（RandomUtil.randomNumbers(6)）；
//   - 存储 key = 手机号原串（Java redisUtil 直接用裸手机号做 key，无前缀）；
//     Go 版内部统一加命名空间前缀 sms:code:，属进程内实现细节，不影响外部契约；
//   - TTL 600 秒（redisUtil.set(mobile, captcha, 600)）；
//   - 重发控制：Redis 中已存在未过期验证码 → 拒绝重发，话术
//     「验证码10分钟内，仍然有效！」。Java 没有独立重发间隔，600s 的
//     「存活即拒发」即其事实上的重发限速（>60s 下限），Go 对齐，无需另加 60s；
//   - smsmode：0=登录模板、1=注册模板、2=忘记密码模板，其余模板 Java 走
//     b=false → 「短信验证码发送失败,请稍后重试」。
//
// 有意偏差（文档化）：Java 经 DySmsHelper 调阿里云 dysmsapi 发真实短信；
// Go 版不接任何外部短信服务商（外部依赖），验证码以 INFO 级日志输出
// （日志注明仅限开发环境）并照常入库存储，发送动作标记为未接入。
// 因此 Java 中「阿里云未配置/发送失败」的两条错误话术在本版不会出现。
package sms

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
)

// 业务错误（HTTP 层映射为 Result{success:false}，话术与 Java 契约逐字一致）。
var (
	// ErrMobileRequired 手机号为空（LoginController L247 原话术）。
	ErrMobileRequired = errors.New("手机号不允许为空！")
	// ErrCodeStillValid 旧验证码仍存活，拒绝重发（L253 原话术）。
	ErrCodeStillValid = errors.New("验证码10分钟内，仍然有效！")
	// ErrPhoneRegistered 注册模板下手机号已注册（L268 原话术，含 Java 源码
	// 原有的前导空格，前端原样展示，保留以精确对齐）。
	ErrPhoneRegistered = errors.New(" 手机号已经注册，请直接登录！")
	// ErrUserNotExist 登录/忘记密码模板下手机号未注册
	// （checkUserIsEffective 情况1 原话术；已删除用户因查询过滤 del_flag=0
	// 同样落到本分支，与 Java 行为一致）。
	ErrUserNotExist = errors.New("该用户不存在，请注册")
	// ErrUserFrozen 用户被冻结（checkUserIsEffective 情况3 原话术）。
	ErrUserFrozen = errors.New("该用户已冻结")
	// ErrSendFailed 未知 smsmode（Java 中 sendSms 未被调用、b=false 的兜底话术）。
	ErrSendFailed = errors.New("短信验证码发送失败,请稍后重试")
	// ErrCodeWrong 验证码错误（register/phoneVerification 契约话术；
	// 存储缺失、不匹配、存储不可用三种情形统一落此话术，fail-closed）。
	ErrCodeWrong = errors.New("手机验证码错误")
	// ErrCodeExpired 找回密码改密时验证码已失效（passwordChange 原话术）。
	ErrCodeExpired = errors.New("短信验证码失效！")
	// ErrCodeMismatch 找回密码改密时验证码不匹配（passwordChange 原话术）。
	ErrCodeMismatch = errors.New("短信验证码不匹配！")
	// ErrTooManyRequests 行为限流触发（Go 版加固项，Java 无对应话术）。
	ErrTooManyRequests = errors.New("操作过于频繁，请稍后再试")
)

// 契约常量（与 Java CommonConstant.SMS_TPL_TYPE_* 对齐）。
const (
	// ModeLogin 登录模板（Login.vue L332）。
	ModeLogin = "0"
	// ModeRegister 注册模板（Register.vue L271）。
	ModeRegister = "1"
	// ModeForgetPassword 忘记密码模板（Step2.vue L154）。
	ModeForgetPassword = "2"
	// CodeTTL 验证码有效期 600 秒（Java redisUtil.set(mobile, captcha, 600)）。
	CodeTTL = 600 * time.Second
	// codeLength 验证码位数（Java RandomUtil.randomNumbers(6)）。
	codeLength = 6
)

// 行为限流参数（加固项：Java 无 /sys/sms 的 IP 维度限流；数值为人为设定，
// 取值宽松以保证教室/办公网 NAT 场景下正常使用不受影响）。
var (
	smsSendIPLimit  = 30
	smsSendIPWindow = 10 * time.Minute
)

// Service 短信验证码业务服务。
type Service struct {
	store    CodeStore
	sendGate *RateWindow
}

// New 构造服务。
//
// redisEnabled 来自 REDIS_ENABLED 配置（写法对齐 login 域 NewRedisRevoker 的
// 装配方式）：true 走 Redis 存储；false 走进程内存储兜底——校验链路始终
// fail-closed（缺失/不匹配/存储异常一律判失败），绝不出现「验证码永远通过」。
func New(redisEnabled bool) *Service {
	var store CodeStore
	if redisEnabled {
		store = newRedisCodeStore()
	} else {
		store = newMemCodeStore()
	}
	return &Service{
		store:    store,
		sendGate: NewRateWindow(smsSendIPLimit, smsSendIPWindow),
	}
}

// Send 编排一次验证码发送（契约蓝本 LoginController#sms L239-311）。
//
// 流程与 Java 逐条对齐：
//  1. mobile 为空 → ErrMobileRequired；
//  2. 同 IP 行为限流（Go 加固项）；
//  3. 已存在未过期验证码 → ErrCodeStillValid（Java 的重发限速）；
//  4. 生成 6 位数字验证码；
//  5. 按 smsmode 做前置校验（涉及 sys_user 查询）；
//  6. 「发送」（Go 版：日志输出验证码，通道未接入）；
//  7. 验证码写入存储，TTL 600s。
func (s *Service) Send(ctx context.Context, mobile, smsmode, clientIP string) error {
	if mobile == "" {
		return ErrMobileRequired
	}
	if !s.sendGate.Allow("ip:" + clientIP) {
		g.Log().Warningf(ctx, "短信验证码发送触发同 IP 行为限流 ip=%s phone=%s", clientIP, maskPhone(mobile))
		return ErrTooManyRequests
	}
	// Java 重发控制：旧码存活期内一律拒发（键与模式无关，Java 即如此）
	if _, found, err := s.store.Load(ctx, mobile); err != nil {
		return fmt.Errorf("读取验证码存储失败: %w", err)
	} else if found {
		return ErrCodeStillValid
	}

	captcha, err := genCode(codeLength)
	if err != nil {
		return err
	}

	switch smsmode {
	case ModeRegister:
		// 注册模板：手机号已注册则拒绝（Java L265-271）
		user, err := s.getUserByPhone(ctx, mobile)
		if err != nil {
			return err
		}
		if user != nil {
			return ErrPhoneRegistered
		}
	case ModeLogin, ModeForgetPassword:
		// 登录/忘记密码模板：校验用户有效性（Java checkUserIsEffective）
		user, err := s.getUserByPhone(ctx, mobile)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrUserNotExist
		}
		if user.Status == 2 {
			// CommonConstant.USER_FREEZE = 2
			return ErrUserFrozen
		}
	default:
		// Java：未知模板不会调用 sendSms，b=false → 发送失败话术
		return ErrSendFailed
	}

	// 发送动作（有意偏差，见包注释）：真实短信通道未接入，验证码落日志。
	// 日志含明文验证码，仅限开发/演示环境使用；生产环境必须接入真实通道后
	// 移除本日志（生产日志不应出现验证码）。
	g.Log().Infof(ctx, "【短信验证码-仅限开发环境】phone=%s code=%s（真实短信通道未接入，发送动作仅为日志占位）",
		maskPhone(mobile), captcha)

	if err := s.store.Save(ctx, mobile, captcha, CodeTTL); err != nil {
		return fmt.Errorf("写入验证码存储失败: %w", err)
	}
	return nil
}

// Verify 通用验证码校验（register / phoneVerification 语义）。
//
// Java 语义：redis 取不到（null）与取到但不匹配都落到同一句
// 「手机验证码错误」。Go 版维持该语义并显式 fail-closed：存储不可用
// （Redis 故障）同样判失败，绝不放行。空串 code 恒不通过。
func (s *Service) Verify(ctx context.Context, phone, code string) error {
	stored, found, err := s.store.Load(ctx, phone)
	if err != nil {
		g.Log().Warningf(ctx, "验证码存储不可用，校验按失败处理（fail-closed） phone=%s err=%s",
			maskPhone(phone), err.Error())
		return ErrCodeWrong
	}
	if !found || !passwd.ConstantTimeCompare(stored, code) {
		return ErrCodeWrong
	}
	return nil
}

// VerifyReset 找回密码改密前的校验（passwordChange 语义）。
//
// Java 区分两种失败：Redis 无值 → 「短信验证码失效！」；有值不匹配 →
// 「短信验证码不匹配！」。存储不可用时归入「失效」并记录告警（fail-closed）。
func (s *Service) VerifyReset(ctx context.Context, phone, code string) error {
	stored, found, err := s.store.Load(ctx, phone)
	if err != nil {
		g.Log().Warningf(ctx, "验证码存储不可用，改密校验按失效处理（fail-closed） phone=%s err=%s",
			maskPhone(phone), err.Error())
		return ErrCodeExpired
	}
	if !found {
		return ErrCodeExpired
	}
	if !passwd.ConstantTimeCompare(stored, code) {
		return ErrCodeMismatch
	}
	return nil
}

// ReArm 找回密码中间步：phoneVerification 通过后重存验证码，供 passwordChange 复核。
//
// Java L1331 redisUtil.set(phone, smscode) 不带 TTL——验证码从此永不过期，
// 属存量缺陷。Go 版加固为与发送时一致的 600s TTL：前端 Step2 → Step3 间隔
// 仅数分钟，行为不受影响；改密后验证码即被 Consume 作废。
func (s *Service) ReArm(ctx context.Context, phone, code string) error {
	if err := s.store.Save(ctx, phone, code, CodeTTL); err != nil {
		return fmt.Errorf("验证码续期失败: %w", err)
	}
	return nil
}

// Consume 作废指定手机号的验证码（passwordChange 成功后调用）。
//
// 加固项：Java 改密成功后不删除验证码，10 分钟窗口内可重复改密（重放）；
// Go 版一次性消费，前端流程无感知（Step3 为终点步骤）。
func (s *Service) Consume(ctx context.Context, phone string) error {
	return s.store.Delete(ctx, phone)
}

// getUserByPhone 按手机号查用户（对齐 Java SysUserMapper#getUserByPhone：
// WHERE phone=? AND del_flag=0，删除用户查不到即视为不存在）。
func (s *Service) getUserByPhone(ctx context.Context, phone string) (*entity.SysUser, error) {
	cols := dao.SysUser.Columns()
	var one entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(cols.Phone, phone).
		Where(cols.DelFlag, 0).
		Scan(&one); err != nil {
		return nil, fmt.Errorf("按手机号查询用户失败: %w", err)
	}
	if one.Id == "" {
		return nil, nil
	}
	return &one, nil
}

// genCode 生成 n 位纯数字验证码（crypto/rand，等价 Java RandomUtil.randomNumbers(n)，
// 允许前导 0）。
func genCode(n int) (string, error) {
	const digits = "0123456789"
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("生成验证码失败: %w", err)
		}
		out[i] = digits[idx.Int64()]
	}
	return string(out), nil
}

// maskPhone 手机号脱敏：保留前 3 位与后 4 位，中间以 **** 替代（长度不足 7 原样返回）。
func maskPhone(phone string) string {
	if len(phone) >= 7 {
		return phone[:3] + "****" + phone[len(phone)-4:]
	}
	return phone
}
