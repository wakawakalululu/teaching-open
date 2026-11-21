// register_p2_8.go P2-8 收尾域路由装配：SysUserAgent（syscrud 通用引擎）
// 与 ng-alain（getAppData/getDictItems——Angular 前端辅助端点）。
// ThirdLogin render/callback 依赖第三方 OAuth 凭据（本部署未配置），文档化跳过。
package cmd

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/syscrud"
	svcSC "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/syscrud"
)

// wireUserAgentAndNgAlain 装配 SysUserAgent CRUD 与 ng-alain 辅助端点。
func wireUserAgentAndNgAlain(group *ghttp.RouterGroup, guard func(method, path string, requireAuth bool),
) {
	ua := syscrud.NewDomain(svcSC.TableDef{
		Table: "sys_user_agent",
		Cols: []string{"id", "user_name", "agent_user_name", "start_time", "end_time",
			"status", "create_name", "create_by", "create_time", "update_name", "update_by", "update_time"},
	})
	group.GET("/sys/sysUserAgent/list", func(r *ghttp.Request) { ua.List(r, "") })
	group.POST("/sys/sysUserAgent/add", ua.Add)
	group.PUT("/sys/sysUserAgent/edit", ua.Edit)
	group.DELETE("/sys/sysUserAgent/delete", ua.Delete)
	group.DELETE("/sys/sysUserAgent/deleteBatch", ua.DeleteBatch)
	group.GET("/sys/sysUserAgent/queryById", ua.QueryByID)
	group.GET("/sys/sysUserAgent/queryByUserName", ua.QueryByUserName)

	group.GET("/sys/ng-alain/getAppData", func(r *ghttp.Request) { syscrud.NgAlainAppData(r) })
	group.GET("/sys/ng-alain/getDictItems/{dictCode}", func(r *ghttp.Request) {
		syscrud.NgAlainDictItems(r, r.Get("dictCode").String())
	})

	guard("GET", "/api/sys/sysUserAgent/list", true)
	guard("POST", "/api/sys/sysUserAgent/add", true)
	guard("PUT", "/api/sys/sysUserAgent/edit", true)
	guard("DELETE", "/api/sys/sysUserAgent/delete", true)
	guard("DELETE", "/api/sys/sysUserAgent/deleteBatch", true)
	guard("GET", "/api/sys/sysUserAgent/queryById", true)
	guard("GET", "/api/sys/sysUserAgent/queryByUserName", true)
	guard("GET", "/api/sys/ng-alain/getAppData", true)
	guard("GET", "/api/sys/ng-alain/getDictItems/{dictCode}", true)
}
