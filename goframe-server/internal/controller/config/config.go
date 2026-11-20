// Package config 提供站点配置的 HTTP 处理器（契约蓝本：SysConfigController）。
package config

import (
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcConfig "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/config"
)

// Controller 站点配置控制器。
type Controller struct {
	svc *svcConfig.Service
}

// New 构造控制器。
func New(svc *svcConfig.Service) *Controller {
	return &Controller{svc: svc}
}

// GetConfig 按配置键取值；键不存在时返回 success:false（同 Java setSuccess(false)）。
func (c *Controller) GetConfig(r *ghttp.Request) {
	data, err := c.svc.GetConfig(r.Context(), r.Get("key").String())
	if err != nil {
		g.Log().Errorf(r.Context(), "查询配置异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
		return
	}
	if data == nil {
		common.Fail(r, "操作成功！")
		return
	}
	common.OK(r, data.ConfigValue)
}

// SaveTenantConfig 保存租户配置（admin 专属）。
//
// 契约对应前端 views/system/TenantConfig.vue：
//
//	POST /sys/config/saveTenantConfig  {配置键: 配置值, ...}
func (c *Controller) SaveTenantConfig(r *ghttp.Request) {
	var kv map[string]interface{}
	if err := r.Parse(&kv); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	operatorID := r.GetCtxVar(middleware.CtxKeyUserID).String()
	if err := c.svc.SaveTenantConfig(r.Context(), operatorID, kv); err != nil {
		switch {
		case errors.Is(err, svcConfig.ErrForbidden):
			common.Fail(r, err.Error())
		case errors.Is(err, svcConfig.ErrKeyRequired):
			common.Fail(r, err.Error())
		default:
			g.Log().Errorf(r.Context(), "保存租户配置异常 err=%s", err.Error())
			common.Fail(r, "操作失败")
		}
		return
	}
	common.OKMsg(r, "修改成功", nil)
}

// GetAllConfigList 全部生效配置列表。
func (c *Controller) GetAllConfigList(r *ghttp.Request) {
	res, err := c.svc.GetAllConfigList(r.Context())
	if err != nil {
		g.Log().Errorf(r.Context(), "查询配置列表异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}

// GetAllConfigMap 全部生效配置映射。
func (c *Controller) GetAllConfigMap(r *ghttp.Request) {
	res, err := c.svc.GetAllConfigMap(r.Context())
	if err != nil {
		g.Log().Errorf(r.Context(), "查询配置映射异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}

// GetCurrentConfig 对外公开配置（库内配置 + 服务端站点参数）。
func (c *Controller) GetCurrentConfig(r *ghttp.Request) {
	res, err := c.svc.GetCurrentConfig(r.Context())
	if err != nil {
		g.Log().Errorf(r.Context(), "查询公开配置异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}
