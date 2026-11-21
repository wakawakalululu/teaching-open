// ngalain.go ng-alain（Angular 前端框架）辅助端点
// （契约蓝本：NgAlainController#getAppData / #getDictItems/{dictCode}，
// 服务 Angular 管理模板的用户数据与字典取值；本部署前端为 Vue，端点
// 保留契约完整性）。
package syscrud

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
)

// ngAlainAppData 用户数据（{user:{name,avatar,email,token}, menu:[], app:{...}}）。
func NgAlainAppData(r *ghttp.Request) {
	userID := r.GetCtxVar(middleware.CtxKeyUserID).String()
	user, err := g.DB().Model("sys_user").
		Fields("username", "realname", "avatar", "email").
		Where("id", userID).Limit(1).One()
	if err != nil || user.IsEmpty() {
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, g.Map{
		"user": g.Map{
			"name":   user["username"].String(),
			"avatar": user["avatar"].String(),
			"email":  user["email"].String(),
			"token":  r.Header.Get("X-Access-Token"),
		},
		"menu": []interface{}{},
		"app":  g.Map{"name": "jeecg-boot-angular", "description": "jeecg+ng-alain整合版本"},
	})
}

// ngAlainDictItems 字典取值（[{value,label}]；数值型 value 转 int 对齐 Java）。
func NgAlainDictItems(r *ghttp.Request, dictCode string) {
	rows, err := g.DB().GetAll(r.Context(),
		`SELECT s.item_value AS value, s.item_text AS label
 FROM sys_dict_item s
 WHERE s.dict_id = (SELECT id FROM sys_dict WHERE dict_code = ?)
 ORDER BY s.sort_order ASC`, dictCode)
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"value": row["value"].Interface(),
			"label": row["label"].String(),
		})
	}
	r.Response.WriteJson(out)
}
