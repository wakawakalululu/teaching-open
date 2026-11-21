// Package captcha 提供登录验证码域的 HTTP 处理器
// （契约蓝本：LoginController#randomImage / #checkCaptcha，anon 端点）。
package captcha

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	svcCP "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/captcha"
)

// Controller 验证码控制器。
type Controller struct {
	svc *svcCP.Service
}

// New 构造控制器。
func New(svc *svcCP.Service) *Controller {
	return &Controller{svc: svc}
}

// RandomImage 生成验证码图片（anon）。
// 契约：GET /sys/randomImage/{key} → result 为 data:image/png;base64 串。
func (c *Controller) RandomImage(r *ghttp.Request) {
	res, err := c.svc.Generate(r.Context(), r.Get("key").String())
	if err != nil {
		common.Fail(r, "获取验证码出错")
		return
	}
	common.OK(r, res)
}

// checkCaptchaReq POST /checkCaptcha 请求体（SysLoginModel 消费字段）。
type checkCaptchaReq struct {
	Captcha  string `json:"captcha"`
	CheckKey string `json:"checkKey"`
}

// CheckCaptcha 校验验证码（anon）。
// 契约：POST /sys/checkCaptcha；错误话术「验证码无效」/「验证码错误」。
func (c *Controller) CheckCaptcha(r *ghttp.Request) {
	var in checkCaptchaReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "验证码无效")
		return
	}
	if err := c.svc.Check(r.Context(), in.Captcha, in.CheckKey); err != nil {
		common.Fail(r, err.Error())
		return
	}
	common.OK(r, nil)
}
