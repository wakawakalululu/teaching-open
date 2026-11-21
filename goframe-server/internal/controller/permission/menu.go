// menu.go 提供菜单权限 CRUD / 角色授权 / 数据规则 / 部门授权的 HTTP 处理器
// （契约蓝本：SysPermissionController 的 19 个端点；与 P0-2 的
// getUserPermissionByToken 同包）。
package permission

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	svcPermission "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/permission"
)

// menuAddEditReq POST/PUT /sys/permission/add|edit 请求体。
// 指针字段对齐 MyBatis-Plus NOT_NULL 字段策略（未提供不参与写库）。
type menuAddEditReq struct {
	Id                 string   `json:"id"`
	ParentId           *string  `json:"parentId"`
	Name               *string  `json:"name"`
	Url                *string  `json:"url"`
	Component          *string  `json:"component"`
	ComponentName      *string  `json:"componentName"`
	Redirect           *string  `json:"redirect"`
	MenuType           *int     `json:"menuType"`
	Perms              *string  `json:"perms"`
	PermsType          *string  `json:"permsType"`
	SortNo             *float64 `json:"sortNo"`
	AlwaysShow         *int     `json:"alwaysShow"`
	Icon               *string  `json:"icon"`
	IsRoute            *int     `json:"isRoute"`
	KeepAlive          *int     `json:"keepAlive"`
	Hidden             *int     `json:"hidden"`
	Description        *string  `json:"description"`
	InternalOrExternal *int     `json:"internalOrExternal"`
	Status             *string  `json:"status"`
}

// dataRuleAddEditReq 数据规则 add/edit 请求体。
type dataRuleAddEditReq struct {
	Id             string  `json:"id"`
	PermissionId   *string `json:"permissionId"`
	RuleName       *string `json:"ruleName"`
	RuleColumn     *string `json:"ruleColumn"`
	RuleConditions *string `json:"ruleConditions"`
	RuleValue      *string `json:"ruleValue"`
	Status         *string `json:"status"`
}

// rolePermSaveReq POST /sys/permission/saveRolePermission 请求体。
// 注意 lastpermissionIds 为 Java 原文小写 p（前端 UserRoleModal 同名键）。
type rolePermSaveReq struct {
	RoleId            string `json:"roleId"`
	PermissionIds     string `json:"permissionIds"`
	LastPermissionIds string `json:"lastpermissionIds"`
}

// departPermSaveReq POST /sys/permission/saveDepartPermission 请求体。
// 此处为 lastPermissionIds 大写 P（Java saveDepartPermission 的键名）。
type departPermSaveReq struct {
	DepartId          string `json:"departId"`
	PermissionIds     string `json:"permissionIds"`
	LastPermissionIds string `json:"lastPermissionIds"`
}

// writeMenuError 统一错误映射：业务错误返回 Java 原话术，未知错误记录日志
// 并返回通用话术（Java 版把异常消息拼进响应，Go 版脱敏）。
func writeMenuError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcPermission.ErrMenuNotFound),
		errors.Is(err, svcPermission.ErrNoPermInfo):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "菜单权限操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// splitIDs 逗号拆分批量 ids 参数（去除空白项与空段，解析健壮化）。
func splitMenuIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// toMenuInput 请求体 → 服务层输入。
func toMenuInput(in menuAddEditReq) svcPermission.MenuInput {
	return svcPermission.MenuInput{
		Id:                 in.Id,
		ParentId:           in.ParentId,
		Name:               in.Name,
		Url:                in.Url,
		Component:          in.Component,
		ComponentName:      in.ComponentName,
		Redirect:           in.Redirect,
		MenuType:           in.MenuType,
		Perms:              in.Perms,
		PermsType:          in.PermsType,
		SortNo:             in.SortNo,
		AlwaysShow:         in.AlwaysShow,
		Icon:               in.Icon,
		IsRoute:            in.IsRoute,
		KeepAlive:          in.KeepAlive,
		Hidden:             in.Hidden,
		Description:        in.Description,
		InternalOrExternal: in.InternalOrExternal,
		Status:             in.Status,
	}
}

// toDataRuleInput 请求体 → 服务层输入。
func toDataRuleInput(in dataRuleAddEditReq) svcPermission.DataRuleInput {
	return svcPermission.DataRuleInput{
		Id:             in.Id,
		PermissionId:   in.PermissionId,
		RuleName:       in.RuleName,
		RuleColumn:     in.RuleColumn,
		RuleConditions: in.RuleConditions,
		RuleValue:      in.RuleValue,
		Status:         in.Status,
	}
}

// MenuList 菜单管理页全量树。
// 契约：GET /sys/permission/list → result 为 SysPermissionTree 树形列表。
func (c *Controller) MenuList(r *ghttp.Request) {
	res, err := c.svc.MenuTreeList(r.Context())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// SystemMenuList 一级菜单列表（懒加载根层）。
// 契约：GET /sys/permission/getSystemMenuList。
func (c *Controller) SystemMenuList(r *ghttp.Request) {
	res, err := c.svc.SystemMenuList(r.Context())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// SystemSubmenu 查询子菜单（懒加载展开）。
// 契约：GET /sys/permission/getSystemSubmenu?parentId=必填。
func (c *Controller) SystemSubmenu(r *ghttp.Request) {
	res, err := c.svc.SystemSubmenu(r.Context(), r.Get("parentId").String())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// SystemSubmenuBatch 批量查询子菜单。
// 契约：GET /sys/permission/getSystemSubmenuBatch?parentIds=逗号分隔
// → result 为 {parentId: [子菜单]} 映射。
func (c *Controller) SystemSubmenuBatch(r *ghttp.Request) {
	res, err := c.svc.SystemSubmenuBatch(r.Context(), splitMenuIDs(r.Get("parentIds").String()))
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// MenuAdd 添加菜单。
// 契约：POST /sys/permission/add；成功话术「添加成功！」；服务端做
// intelligentProcessData 预处理 + is_leaf/del_flag 维护。
func (c *Controller) MenuAdd(r *ghttp.Request) {
	var in menuAddEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.AddMenu(r.Context(), toMenuInput(in)); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// MenuEdit 编辑菜单。
// 契约：PUT/POST /sys/permission/edit（Java 两种方法均接受，路由两条登记）；
// 成功话术「修改成功！」；服务端做父节点变化时的 is_leaf 双向维护。
func (c *Controller) MenuEdit(r *ghttp.Request) {
	var in menuAddEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.EditMenu(r.Context(), toMenuInput(in)); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "修改成功！", nil)
}

// MenuDelete 删除菜单（物理级联递归）。
// 契约：DELETE /sys/permission/delete?id=必填；成功话术「删除成功!」。
func (c *Controller) MenuDelete(r *ghttp.Request) {
	if err := c.svc.DeleteMenu(r.Context(), r.Get("id").String()); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// MenuDeleteBatch 批量删除菜单。
// 契约：DELETE /sys/permission/deleteBatch?ids=逗号分隔；成功话术「删除成功!」。
func (c *Controller) MenuDeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteMenuBatch(r.Context(), splitMenuIDs(r.Get("ids").String())); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// PermissionTreeList 菜单权限树（角色/部门授权弹窗数据源）。
// 契约：GET /sys/permission/queryTreeList → result {treeList, ids}。
func (c *Controller) PermissionTreeList(r *ghttp.Request) {
	tree, ids, err := c.svc.PermissionTreeList(r.Context())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, g.Map{"treeList": tree, "ids": ids})
}

// QueryListAsync 异步加载子菜单。
// 契约：GET /sys/permission/queryListAsync?pid=可选；空列表 →「未找到角色信息」
// （Java 原话术）；Java 版 SQL 引用不存在列必然报错的缺陷已修正（见服务层文档）。
func (c *Controller) QueryListAsync(r *ghttp.Request) {
	res, err := c.svc.QueryListAsync(r.Context(), r.Get("pid").String())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// QueryRolePermission 角色已授权菜单 ID 列表。
// 契约：GET /sys/permission/queryRolePermission?roleId=必填。
func (c *Controller) QueryRolePermission(r *ghttp.Request) {
	res, err := c.svc.QueryRolePermission(r.Context(), r.Get("roleId").String())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// SaveRolePermission 保存角色授权（diff 增量）。
// 契约：POST /sys/permission/saveRolePermission；成功话术「保存成功！」。
func (c *Controller) SaveRolePermission(r *ghttp.Request) {
	var in rolePermSaveReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SaveRolePermission(r.Context(), in.RoleId, in.PermissionIds, in.LastPermissionIds); err != nil {
		g.Log().Errorf(r.Context(), "保存角色授权失败 err=%s", err.Error())
		common.Fail(r, "授权失败！")
		return
	}
	common.OKMsg(r, "保存成功！", nil)
}

// GetPermRuleListByPermId 菜单的数据规则列表。
// 契约：GET /sys/permission/getPermRuleListByPermId?permissionId=。
func (c *Controller) GetPermRuleListByPermId(r *ghttp.Request) {
	res, err := c.svc.GetPermRuleListByPermId(r.Context(), r.Get("permissionId").String())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// AddPermissionRule 添加数据规则。
// 契约：POST /sys/permission/addPermissionRule；成功话术「添加成功！」；
// 服务端写 create_time 并维护菜单 rule_flag。
func (c *Controller) AddPermissionRule(r *ghttp.Request) {
	var in dataRuleAddEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.AddPermissionRule(r.Context(), toDataRuleInput(in)); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// EditPermissionRule 编辑数据规则。
// 契约：PUT/POST /sys/permission/editPermissionRule；成功话术「更新成功！」。
func (c *Controller) EditPermissionRule(r *ghttp.Request) {
	var in dataRuleAddEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.EditPermissionRule(r.Context(), toDataRuleInput(in)); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "更新成功！", nil)
}

// DeletePermissionRule 删除数据规则。
// 契约：DELETE /sys/permission/deletePermissionRule?id=必填；成功话术「删除成功！」。
func (c *Controller) DeletePermissionRule(r *ghttp.Request) {
	if err := c.svc.DeletePermissionRule(r.Context(), r.Get("id").String()); err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "删除成功！", nil)
}

// QueryPermissionRule 数据规则条件查询。
// 契约：GET /sys/permission/queryPermissionRule（QueryGenerator 消费面）；
// 成功话术「查询成功！」。
func (c *Controller) QueryPermissionRule(r *ghttp.Request) {
	res, err := c.svc.QueryPermissionRule(r.Context(), svcPermission.DataRuleFilter{
		PermissionId:   r.Get("permissionId").String(),
		RuleName:       r.Get("ruleName").String(),
		RuleColumn:     r.Get("ruleColumn").String(),
		RuleConditions: r.Get("ruleConditions").String(),
		RuleValue:      r.Get("ruleValue").String(),
		Status:         r.Get("status").String(),
	})
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OKMsg(r, "查询成功！", res)
}

// QueryDepartPermission 部门已授权菜单 ID 列表。
// 契约：GET /sys/permission/queryDepartPermission?departId=必填。
func (c *Controller) QueryDepartPermission(r *ghttp.Request) {
	res, err := c.svc.QueryDepartPermission(r.Context(), r.Get("departId").String())
	if err != nil {
		writeMenuError(r, err)
		return
	}
	common.OK(r, res)
}

// SaveDepartPermission 保存部门授权（diff 增量）。
// 契约：POST /sys/permission/saveDepartPermission；成功话术「保存成功！」。
func (c *Controller) SaveDepartPermission(r *ghttp.Request) {
	var in departPermSaveReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SaveDepartPermission(r.Context(), in.DepartId, in.PermissionIds, in.LastPermissionIds); err != nil {
		g.Log().Errorf(r.Context(), "保存部门授权失败 err=%s", err.Error())
		common.Fail(r, "授权失败！")
		return
	}
	common.OKMsg(r, "保存成功！", nil)
}
