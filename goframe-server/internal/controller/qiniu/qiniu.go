// Package qiniu 提供七牛上传令牌的 HTTP 处理器
// （契约蓝本：QiniuController 的 2 个端点，/common/qiniu/getToken[ByKey]）。
package qiniu

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	svcQ "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/qiniu"
)

// Controller 七牛令牌控制器。
type Controller struct {
	cfg svcQ.Config
}

// New 构造控制器。
func New(cfg svcQ.Config) *Controller {
	return &Controller{cfg: cfg}
}

// GetToken 上传令牌（key 不限定）。
// 契约：ANY /common/qiniu/getToken → result 为令牌字符串（code=200）。
func (c *Controller) GetToken(r *ghttp.Request) {
	common.OK(r, c.cfg.UploadToken(""))
}

// GetTokenByKey 限定 key 的覆盖上传令牌。
// 契约：ANY /common/qiniu/getTokenByKey?key= → result 为令牌字符串。
func (c *Controller) GetTokenByKey(r *ghttp.Request) {
	common.OK(r, c.cfg.UploadToken(r.Get("key").String()))
}
