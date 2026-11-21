// Package departpermission 提供部门权限域的 HTTP 处理器
// （契约蓝本：SysDepartPermissionController 的 11 个核心端点）。
package departpermission

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcDepartPerm "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/departpermission"
)

// Controller 部门权限控制器。
type Controller struct {
	svc *svcDepartPerm.Service
}

// New 构造控制器。
func New(svc *svcDepartPerm.Service) *Controller {
	return &Controller{svc: svc}
}

// addEditReq POST/PUT /sys/sysDepartPermission/add|edit 请求体。
type addEditReq struct {
	Id           string  `json:"id"`
	DepartId     *string `json:"departId"`
	PermissionId *string `json:"permissionId"`
	DataRuleIds  *string `json:"dataRuleIds"`
}

// dataruleSaveReq POST /sys/sysDepartPermission/datarule 请求体。
type dataruleSaveReq struct {
	PermissionId string  `json:"permissionId"`
	DepartId     string  `json:"departId"`
	DataRuleIds  *string `json:"dataRuleIds"`
}

// deptRolePermSaveReq POST /sys/sysDepartPermission/saveDeptRolePermission 请求体。
// 注意 lastpermissionIds 为 Java 原文小写 p。
type deptRolePermSaveReq struct {
	RoleId            string `json:"roleId"`
	PermissionIds     string `json:"permissionIds"`
	LastPermissionIds string `json:"lastpermissionIds"`
}

// writeError 统一错误映射。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcDepartPerm.ErrNoPermConfig),
		errors.Is(err, svcDepartPerm.ErrSaveDeptPermFirst),
		errors.Is(err, svcDepartPerm.ErrSaveFailed):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "部门权限操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// parsePage 解析分页参数（默认 1/10）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// List 部门授权分页列表。
// 契约：GET /sys/sysDepartPermission/list?pageNo=&pageSize=&departId=&permissionId=。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("departId").String(), r.Get("permissionId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增部门授权。
// 契约：POST /sys/sysDepartPermission/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), svcDepartPerm.DepartPermissionInput{
		Id: in.Id, DepartId: in.DepartId, PermissionId: in.PermissionId, DataRuleIds: in.DataRuleIds,
	}); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑部门授权。
// 契约：PUT /sys/sysDepartPermission/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), svcDepartPerm.DepartPermissionInput{
		Id: in.Id, DepartId: in.DepartId, PermissionId: in.PermissionId, DataRuleIds: in.DataRuleIds,
	}); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除部门授权（无级联）。
// 契约：DELETE /sys/sysDepartPermission/delete?id=必填；成功话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除部门授权（无级联）。
// 契约：DELETE /sys/sysDepartPermission/deleteBatch?ids=；话术「批量删除成功！」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功！", nil)
}

// QueryById 部门授权详情。
// 契约：GET /sys/sysDepartPermission/queryById?id=必填；不存在 result=null 仍 success。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// DataruleList 部门的菜单数据规则与已勾选项。
// 契约：GET /sys/sysDepartPermission/datarule/:permissionId/:departId；
// 菜单无规则 →「未找到权限配置信息」。
func (c *Controller) DataruleList(r *ghttp.Request) {
	res, err := c.svc.DataruleList(r.Context(),
		r.Get("permissionId").String(), r.Get("departId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	payload := g.Map{"datarule": res.Rules}
	if res.HasChecked {
		payload["drChecked"] = res.DrChecked
	}
	common.OK(r, payload)
}

// DataruleSave 保存部门的菜单数据规则勾选。
// 契约：POST /sys/sysDepartPermission/datarule；关系缺失 →「请先保存部门菜单权限!」。
func (c *Controller) DataruleSave(r *ghttp.Request) {
	var in dataruleSaveReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.DataruleSave(r.Context(), in.PermissionId, in.DepartId, in.DataRuleIds); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "保存成功!", nil)
}

// QueryDeptRolePermission 部门角色已授权菜单 ID 列表。
// 契约：GET /sys/sysDepartPermission/queryDeptRolePermission?roleId=必填。
func (c *Controller) QueryDeptRolePermission(r *ghttp.Request) {
	res, err := c.svc.QueryDeptRolePermission(r.Context(), r.Get("roleId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// SaveDeptRolePermission 保存部门角色的菜单授权（diff）。
// 契约：POST /sys/sysDepartPermission/saveDeptRolePermission，body
// {roleId, permissionIds, lastpermissionIds}；成功话术「保存成功！」。
func (c *Controller) SaveDeptRolePermission(r *ghttp.Request) {
	var in deptRolePermSaveReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SaveDeptRolePermission(r.Context(), in.RoleId, in.PermissionIds, in.LastPermissionIds); err != nil {
		g.Log().Errorf(r.Context(), "保存部门角色授权失败 err=%s", err.Error())
		common.Fail(r, "授权失败！")
		return
	}
	common.OKMsg(r, "保存成功！", nil)
}

// QueryTreeListForDeptRole 部门角色授权弹窗的菜单树。
// 契约：GET /sys/sysDepartPermission/queryTreeListForDeptRole?departId=必填
// → result {treeList, ids}。
func (c *Controller) QueryTreeListForDeptRole(r *ghttp.Request) {
	tree, ids, err := c.svc.QueryTreeListForDeptRole(r.Context(), r.Get("departId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, g.Map{"treeList": tree, "ids": ids})
}

// splitIDs 逗号拆分批量 ids 参数（去除空白项与空段）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
