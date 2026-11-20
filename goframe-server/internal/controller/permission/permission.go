// Package permission 提供菜单/权限的 HTTP 处理器（契约蓝本：SysPermissionController）。
package permission

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcPermission "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/permission"
)

// Controller 菜单权限控制器。
type Controller struct {
	svc *svcPermission.Service
}

// New 构造控制器。
func New(svc *svcPermission.Service) *Controller {
	return &Controller{svc: svc}
}

// GetUserPermissionByToken 当前用户的路由菜单与按钮权限。
//
// 契约对应前端 store/modules/user.js GetPermissionList：
//
//	GET /sys/permission/getUserPermissionByToken
//	→ result.{menu, auth, allAuth}
//
// Java 版从 query 参数的 token 里解出用户名；Go 版统一走认证中间件注入的
// 请求上下文（前端 axios 拦截器同时携带 X-Access-Token 头），query 参数
// 忽略，避免重复解析与伪造。
func (c *Controller) GetUserPermissionByToken(r *ghttp.Request) {
	username := r.GetCtxVar(middleware.CtxKeyUsername).String()
	if username == "" {
		common.Fail(r, "TOKEN不允许为空！")
		return
	}
	res, err := c.svc.GetUserPermissionByToken(r.Context(), username)
	if err != nil {
		g.Log().Errorf(r.Context(), "查询用户菜单异常 err=%s", err.Error())
		common.Fail(r, "查询失败")
		return
	}
	common.OKMsg(r, "查询成功", res)
}
