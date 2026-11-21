// Package sms 提供 /sys/sms 短信验证码 HTTP 处理器。
//
// 契约蓝本：LoginController#sms（POST /sys/sms，body JSON）。
// 公开依据：ShiroConfig.java L80 filterChainDefinitionMap.put("/sys/sms", "anon")。
//
// 前端调用方（web/src 全量核对）：
//   - Register.vue L273：{mobile, smsmode:"1"}（注册；UI 中按钮被注释但端点契约保留）；
//   - Step2.vue L156：{mobile, smsmode:"2"}（忘记密码）；
//   - Login.vue L332：{mobile, smsmode:"0"}（手机登录）。
//
// 响应：成功 {success:true, message:"", result:null}（Java 仅 setSuccess(true)，
// result 的赋值在 Java 源码中被注释掉，恒为 null；验证码不随响应下发）；
// 失败 HTTP 200 + success:false + message。
package sms

import (
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	svcSms "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sms"
)

// Controller 短信验证码控制器。
type Controller struct {
	svc *svcSms.Service
}

// New 构造控制器。
func New(svc *svcSms.Service) *Controller {
	return &Controller{svc: svc}
}

// Send 发送短信验证码。
//
// 契约：POST /sys/sms，body {"mobile":"...","smsmode":"0|1|2"}。
// smsmode 语义：0=登录模板、1=注册模板、2=忘记密码模板；其余值 Java 落
// 「短信验证码发送失败,请稍后重试」（sendSms 未被调用），Go 版一致。
func (c *Controller) Send(r *ghttp.Request) {
	var req struct {
		Mobile  string `json:"mobile"`
		Smsmode string `json:"smsmode"`
	}
	if err := r.Parse(&req); err != nil {
		// Java：body 缺 mobile 时 jsonObject.get("mobile").toString() 抛 NPE →
		// 全局 500；Go 版收敛为「手机号为空」话术（前端调用方均传合法 JSON）。
		common.Fail(r, svcSms.ErrMobileRequired.Error())
		return
	}
	if err := c.svc.Send(r.Context(), req.Mobile, req.Smsmode, r.GetClientIp()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "", nil)
}

// writeError 业务错误映射：契约错误返回 Java 原话术（HTTP 200 + success:false），
// 未知错误（存储/数据库基础设施故障）记录日志并返回通用话术，不泄露细节。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcSms.ErrMobileRequired),
		errors.Is(err, svcSms.ErrCodeStillValid),
		errors.Is(err, svcSms.ErrPhoneRegistered),
		errors.Is(err, svcSms.ErrUserNotExist),
		errors.Is(err, svcSms.ErrUserFrozen),
		errors.Is(err, svcSms.ErrSendFailed),
		errors.Is(err, svcSms.ErrTooManyRequests):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "短信验证码发送异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}
