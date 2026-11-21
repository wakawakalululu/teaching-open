// Package teachingmenu 提供门户菜单域的 HTTP 处理器
// （契约蓝本：TeachingMenuController 的 11 个端点）。
package teachingmenu

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	svcTM "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingmenu"
)

// Controller 门户菜单控制器。
type Controller struct {
	svc *svcTM.Service
}

// New 构造控制器。
func New(svc *svcTM.Service) *Controller {
	return &Controller{svc: svc}
}

// menuReq POST/PUT 请求体（TeachingMenu 可编辑字段）。
type menuReq struct {
	Id                 string   `json:"id"`
	ParentId           *string  `json:"parentId"`
	Name               *string  `json:"name"`
	Icon               *string  `json:"icon"`
	Url                *string  `json:"url"`
	SortNo             *float64 `json:"sortNo"`
	MenuType           *int     `json:"menuType"`
	Leaf               *bool    `json:"leaf"`
	Route              *bool    `json:"route"`
	Description        *string  `json:"description"`
	Hidden             *bool    `json:"hidden"`
	NeedLogin          *int     `json:"needLogin"`
	InternalOrExternal *bool    `json:"internalOrExternal"`
	DelFlag            *int     `json:"delFlag"`
}

// toMenuInput 请求体 → 服务层输入。
func toMenuInput(in menuReq) svcTM.MenuInput {
	return svcTM.MenuInput{
		Id: in.Id, ParentId: in.ParentId, Name: in.Name, Icon: in.Icon,
		Url: in.Url, SortNo: in.SortNo, MenuType: in.MenuType, Leaf: in.Leaf,
		Route: in.Route, Description: in.Description, Hidden: in.Hidden,
		NeedLogin: in.NeedLogin, InternalOrExternal: in.InternalOrExternal,
		DelFlag: in.DelFlag,
	}
}

// writeError 统一错误映射（delete 异常分支返回「删除成功!」的缺陷话术未照搬）。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcTM.ErrMenuNotFound), errors.Is(err, svcTM.ErrNoMenuInfo):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// GetUserMenu 门户路由菜单（anon）。
// 契约：GET /teaching/menu/getUserMenu（未登录可调）；话术「查询成功」。
func (c *Controller) GetUserMenu(r *ghttp.Request) {
	res, err := c.svc.GetUserMenu(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "查询成功", res)
}

// List 全量菜单树。
// 契约：GET /teaching/menu/list。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetSystemMenuList 一级菜单扁平列表。
// 契约：GET /teaching/menu/getSystemMenuList。
func (c *Controller) GetSystemMenuList(r *ghttp.Request) {
	res, err := c.svc.GetSystemMenuList(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetSystemSubmenu 子菜单扁平列表。
// 契约：GET /teaching/menu/getSystemSubmenu?parentId=必填。
func (c *Controller) GetSystemSubmenu(r *ghttp.Request) {
	res, err := c.svc.GetSystemSubmenu(r.Context(), r.Get("parentId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetSystemSubmenuBatch 批量子菜单。
// 契约：GET /teaching/menu/getSystemSubmenuBatch?parentIds=逗号分隔。
func (c *Controller) GetSystemSubmenuBatch(r *ghttp.Request) {
	res, err := c.svc.GetSystemSubmenuBatch(r.Context(), splitIDs(r.Get("parentIds").String()))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 添加菜单。
// 契约：POST /teaching/menu/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in menuReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toMenuInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑菜单。
// 契约：PUT/POST /teaching/menu/edit；成功话术「修改成功！」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in menuReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toMenuInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "修改成功！", nil)
}

// Delete 删除菜单（物理级联递归）。
// 契约：DELETE /teaching/menu/delete?id=；成功话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除菜单。
// 契约：DELETE /teaching/menu/deleteBatch?ids=；话术「删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// QueryTreeList 菜单树 + 全部 ids。
// 契约：GET /teaching/menu/queryTreeList → result {treeList, ids}。
func (c *Controller) QueryTreeList(r *ghttp.Request) {
	tree, ids, err := c.svc.QueryTreeList(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, g.Map{"treeList": tree, "ids": ids})
}

// QueryListAsync 异步子菜单。
// 契约：GET /teaching/menu/queryListAsync?pid=可选；空 →「未找到角色信息」。
func (c *Controller) QueryListAsync(r *ghttp.Request) {
	res, err := c.svc.QueryListAsync(r.Context(), r.Get("pid").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// splitIDs 逗号拆分批量 ids 参数。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
