// Package departrole 提供部门角色域的 HTTP 处理器
// （契约蓝本：SysDepartRoleController 的 11 个核心端点）。
package departrole

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcDepartRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/departrole"
)

// Controller 部门角色控制器。
type Controller struct {
	svc *svcDepartRole.Service
}

// New 构造控制器。
func New(svc *svcDepartRole.Service) *Controller {
	return &Controller{svc: svc}
}

// addEditReq POST/PUT /sys/sysDepartRole/add|edit 请求体。
type addEditReq struct {
	Id          string  `json:"id"`
	DepartId    *string `json:"departId"`
	RoleName    *string `json:"roleName"`
	RoleCode    *string `json:"roleCode"`
	Description *string `json:"description"`
}

// deptRoleUserAddReq POST /sys/sysDepartRole/deptRoleUserAdd 请求体
// （键名对齐 Java JSONObject 读取）。
type deptRoleUserAddReq struct {
	NewRoleId string `json:"newRoleId"`
	OldRoleId string `json:"oldRoleId"`
	UserId    string `json:"userId"`
}

// dataruleSaveReq POST /sys/sysDepartRole/datarule 请求体。
type dataruleSaveReq struct {
	PermissionId string  `json:"permissionId"`
	RoleId       string  `json:"roleId"`
	DataRuleIds  *string `json:"dataRuleIds"`
}

// writeError 统一错误映射。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcDepartRole.ErrNoPermConfig),
		errors.Is(err, svcDepartRole.ErrSaveRolePermFirst),
		errors.Is(err, svcDepartRole.ErrSaveFailed):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "部门角色操作异常 err=%s", err.Error())
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

// List 部门角色分页列表。
// 契约：GET /sys/sysDepartRole/list?pageNo=&pageSize=&deptId=&roleName=...
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("deptId").String(), r.Get("roleName").String(),
		r.Get("roleCode").String(), r.Get("description").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增部门角色。
// 契约：POST /sys/sysDepartRole/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), svcDepartRole.DepartRoleInput{
		Id: in.Id, DepartId: in.DepartId, RoleName: in.RoleName,
		RoleCode: in.RoleCode, Description: in.Description,
	}); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑部门角色。
// 契约：PUT /sys/sysDepartRole/edit；成功话术「编辑成功!」（无存在性校验）。
func (c *Controller) Edit(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), svcDepartRole.DepartRoleInput{
		Id: in.Id, DepartId: in.DepartId, RoleName: in.RoleName,
		RoleCode: in.RoleCode, Description: in.Description,
	}); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除部门角色（无级联）。
// 契约：DELETE /sys/sysDepartRole/delete?id=必填；成功话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除部门角色（无级联）。
// 契约：DELETE /sys/sysDepartRole/deleteBatch?ids=逗号分隔；话术「批量删除成功！」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	ids := splitIDs(r.Get("ids").String())
	if err := c.svc.DeleteBatch(r.Context(), ids); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功！", nil)
}

// QueryById 部门角色详情。
// 契约：GET /sys/sysDepartRole/queryById?id=必填；不存在时 result=null 仍 success。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetDeptRoleList 用户在指定部门（含前缀子部门）可用的部门角色。
// 契约：GET /sys/sysDepartRole/getDeptRoleList?departId=&userId=（均必填）。
func (c *Controller) GetDeptRoleList(r *ghttp.Request) {
	res, err := c.svc.GetDeptRoleList(r.Context(),
		r.Get("departId").String(), r.Get("userId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// DeptRoleUserAdd 保存用户与部门角色绑定（diff）。
// 契约：POST /sys/sysDepartRole/deptRoleUserAdd，body {newRoleId, oldRoleId, userId}。
func (c *Controller) DeptRoleUserAdd(r *ghttp.Request) {
	var in deptRoleUserAddReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.DeptRoleUserAdd(r.Context(), in.UserId, in.NewRoleId, in.OldRoleId); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// GetDeptRoleByUserID 用户的部门角色绑定列表。
// 契约：GET /sys/sysDepartRole/getDeptRoleByUserId?userId=必填。
func (c *Controller) GetDeptRoleByUserID(r *ghttp.Request) {
	res, err := c.svc.GetDeptRoleByUserID(r.Context(), r.Get("userId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// DataruleList 部门角色的菜单数据规则与已勾选项。
// 契约：GET /sys/sysDepartRole/datarule/:permissionId/:departId/:roleId；
// 部门未授权或菜单无规则 →「未找到权限配置信息」。
func (c *Controller) DataruleList(r *ghttp.Request) {
	res, err := c.svc.DataruleList(r.Context(),
		r.Get("permissionId").String(), r.Get("departId").String(), r.Get("roleId").String())
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

// DataruleSave 保存部门角色的菜单数据规则勾选。
// 契约：POST /sys/sysDepartRole/datarule；关系缺失 →「请先保存角色菜单权限!」。
func (c *Controller) DataruleSave(r *ghttp.Request) {
	var in dataruleSaveReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.DataruleSave(r.Context(), in.PermissionId, in.RoleId, in.DataRuleIds); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "保存成功!", nil)
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
