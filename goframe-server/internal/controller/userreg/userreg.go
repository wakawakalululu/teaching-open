// Package userreg 提供注册/找回密码自服务链路 HTTP 处理器。
//
// 契约蓝本：SysUserController.java#userRegister/phoneVerification/passwordChange。
// 公开依据：ShiroConfig.java L83（register）/L85（phoneVerification）/
// L86（passwordChange）三条 anon 名单。
//
// 前端调用方（web/src 全量核对）：
//   - Register.vue L244：POST /sys/user/register，body {username,password,
//     email,phone,smscode}（email/smscode 当前因输入框被注释而不传）；
//   - Step2.vue L113：POST /sys/user/phoneVerification，body {phone,smscode}，
//     成功后取 res.result（即提交的验证码）传给 Step3；
//   - Step3.vue L81：GET /sys/user/passwordChange?username=&password=&
//     smscode=&phone=，读 res.success/res.message。
//
// 业务失败一律 HTTP 200 + success:false（Java Result.error500 语义）。
package userreg

import (
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	svcSms "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sms"
	svcUserreg "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/userreg"
)

// Controller 注册/找回密码控制器。
type Controller struct {
	svc *svcUserreg.Service
}

// New 构造控制器。
func New(svc *svcUserreg.Service) *Controller {
	return &Controller{svc: svc}
}

// Register 用户注册。
//
// 契约：POST /sys/user/register，body JSON（见 svcUserreg.RegisterInput）。
// 成功 {success:true, message:"注册成功", result:null}（Java result.success
// 只设 message，result 为 null）；前端 Register.vue 只读 res.success/res.message。
func (c *Controller) Register(r *ghttp.Request) {
	var in svcUserreg.RegisterInput
	if err := r.Parse(&in); err != nil {
		common.Fail(r, svcUserreg.ErrRegisterFailed.Error())
		return
	}
	if err := c.svc.Register(r.Context(), in, r.GetClientIp()); err != nil {
		c.writeRegisterError(r, err)
		return
	}
	common.OKMsg(r, "注册成功", nil)
}

// writeRegisterError 注册链路错误映射。
func (c *Controller) writeRegisterError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcUserreg.ErrUsernameTaken),
		errors.Is(err, svcUserreg.ErrPhoneTaken),
		errors.Is(err, svcUserreg.ErrEmailTaken),
		errors.Is(err, svcUserreg.ErrRegClosed),
		errors.Is(err, svcUserreg.ErrPasswordTooLong),
		errors.Is(err, svcSms.ErrCodeWrong),
		errors.Is(err, svcSms.ErrTooManyRequests):
		common.Fail(r, err.Error())
	case errors.Is(err, svcUserreg.ErrRegisterFailed):
		// 服务层已记录内部日志；话术与 Java catch-all 一致
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "注册链路处理异常 err=%s", err.Error())
		common.Fail(r, svcUserreg.ErrRegisterFailed.Error())
	}
}

// PhoneVerification 找回密码第二步：手机号+验证码校验。
//
// 契约：POST /sys/user/phoneVerification，body {"phone","smscode"}。
// 成功 {success:true, result:提交的 smscode}——Java setResult(smscode) 契约，
// Step2.vue L119 取 res.result 传给 Step3，零改动对接必须保留该回显。
// 安全说明：回显的是用户本轮刚提交、600s 内有效的一次性验证码，属 Java
// 原有契约行为（不是凭空可得的秘密）。
func (c *Controller) PhoneVerification(r *ghttp.Request) {
	var req struct {
		Phone   string `json:"phone"`
		Smscode string `json:"smscode"`
	}
	if err := r.Parse(&req); err != nil {
		// Java：字段缺失时 smscode.equals(null) 判不匹配 → 「手机验证码错误」，
		// Go 版收敛为同一话术
		common.Fail(r, svcSms.ErrCodeWrong.Error())
		return
	}
	if err := c.svc.PhoneVerification(r.Context(), req.Phone, req.Smscode); err != nil {
		if errors.Is(err, svcSms.ErrCodeWrong) {
			common.Fail(r, err.Error())
			return
		}
		g.Log().Errorf(r.Context(), "手机号验证处理异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
		return
	}
	common.OKMsg(r, "", req.Smscode)
}

// PasswordChange 找回密码第三步：改密。
//
// 契约：GET /sys/user/passwordChange?username=&password=&smscode=&phone=
// （Java 四个 @RequestParam 均必填）。成功 {success:true,
// message:"密码重置完成！", result:null}（Java 不 setResult）。
func (c *Controller) PasswordChange(r *ghttp.Request) {
	err := c.svc.PasswordChange(r.Context(),
		r.Get("username").String(),
		r.Get("password").String(),
		r.Get("smscode").String(),
		r.Get("phone").String())
	if err != nil {
		c.writeChangeError(r, err)
		return
	}
	common.OKMsg(r, "密码重置完成！", nil)
}

// writeChangeError 改密链路错误映射。
func (c *Controller) writeChangeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcUserreg.ErrParamMissing),
		errors.Is(err, svcUserreg.ErrUserNotFound),
		errors.Is(err, svcUserreg.ErrPasswordTooLong),
		errors.Is(err, svcSms.ErrCodeExpired),
		errors.Is(err, svcSms.ErrCodeMismatch):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "找回密码改密处理异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}
