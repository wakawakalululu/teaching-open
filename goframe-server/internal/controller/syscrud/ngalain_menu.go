// ngalain_menu.go ng-alain 菜单数据填充
// （契约蓝本：NgAlainServiceImpl#getJeecgMenu——queryByUser 权限数据 →
// 路由树 → parseNgAlain 转换为 ng-alain 菜单格式 [{text,group,children}]）。
package syscrud

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	svcPerm "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/permission"
)

// NgAlainAppDataFilled getAppData 完整响应（菜单数据填充版）。
// permSvc 由 cmd.go 注入（复用权限域 queryByUser 查询）。
func NgAlainAppDataFilled(r *ghttp.Request, permSvc *svcPerm.Service) {
	userID := r.GetCtxVar(middleware.CtxKeyUserID).String()
	username := r.GetCtxVar(middleware.CtxKeyUsername).String()

	perms, err := permSvc.QueryByUser(r.Context(), username)
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}

	menuList := buildNgAlainMenu(perms, nil)

	user := g.Map{}
	ures, uerr := g.DB().Model("sys_user").
		Fields("username", "realname", "avatar", "email").
		Where("id", userID).Limit(1).One()
	if uerr == nil && !ures.IsEmpty() {
		user = ures.Map()
	}

	common.OK(r, g.Map{
		"user": g.Map{
			"name":   username,
			"avatar": user["avatar"],
			"email":  user["email"],
			"token":  r.Header.Get("X-Access-Token"),
		},
		"menu": []interface{}{
			g.Map{
				"text":     "jeecg菜单",
				"group":    true,
				"children": menuList,
			},
		},
		"app": g.Map{
			"name":        "jeecg-boot-angular",
			"description": "jeecg+ng-alain整合版本",
		},
	})
}

// ngAlainMenu ng-alain 菜单节点。
type ngAlainMenu struct {
	Text     string        `json:"text"`
	Reuse    bool          `json:"reuse"`
	Link     string        `json:"link,omitempty"`
	Icon     *ngAlainIcon  `json:"icon,omitempty"`
	Children []ngAlainMenu `json:"children,omitempty"`
}

// ngAlainIcon 图标。
type ngAlainIcon struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// buildNgAlainMenu 构建菜单树（递归，parentId 空为根，非叶节点挂子；
// 叶子 → link 字段；非叶 → icon + children；对齐 Java parseNgAlain）。
func buildNgAlainMenu(perms []entity.SysPermission, parent *entity.SysPermission) []ngAlainMenu {
	var out []ngAlainMenu
	for i := range perms {
		p := perms[i]
		if p.MenuType != 0 && p.MenuType != 1 {
			continue
		}
		node := ngAlainMenu{
			Text:  p.Name,
			Reuse: true,
		}
		if parent == nil {
			if p.ParentId != "" {
				continue
			}
		} else if p.ParentId != parent.Id {
			continue
		}
		// 叶子节点 → link
		if p.IsLeaf == 1 {
			node.Link = p.Url
		} else {
			// 非叶 → 图标 + 递归子
			if p.Icon != "" {
				node.Icon = &ngAlainIcon{Type: "icon", Value: p.Icon}
			}
			node.Children = buildNgAlainMenu(perms, &perms[i])
		}
		out = append(out, node)
	}
	return out
}
