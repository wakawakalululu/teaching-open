// Package role 提供系统角色管理域的 HTTP 处理器
// （契约蓝本：SysRoleController 的 12 个端点，exportXls/importExcel 有意未迁移，
// 与 P0-1/P0-3 的 Excel 域决策一致）。
package role

import (
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/role"
)

// Controller 角色管理控制器。
type Controller struct {
	svc *svcRole.Service
}

// New 构造控制器。
func New(svc *svcRole.Service) *Controller {
	return &Controller{svc: svc}
}

// addEditReq POST/PUT /sys/role/add|edit 请求体。
//
// 指针字段对齐 MyBatis-Plus NOT_NULL 字段策略：JSON 未提供的字段不参与写库。
type addEditReq struct {
	Id          string  `json:"id"`
	RoleName    *string `json:"roleName"`
	RoleCode    *string `json:"roleCode"`
	Description *string `json:"description"`
	RoleLevel   *int    `json:"roleLevel"`
}

// dataruleSaveReq POST /sys/role/datarule 请求体。
type dataruleSaveReq struct {
	PermissionId string  `json:"permissionId"`
	RoleId       string  `json:"roleId"`
	DataRuleIds  *string `json:"dataRuleIds"`
}

// okDefault Java new Result<>() 默认形状（success=true, code=0, message=操作成功！）。
// 仓库统一约定成功响应 code=200（与 P0-1/P0-2 先例一致，前端仅消费 success），故
// 默认形状端点同样走 common.OK；此处注释仅作契约溯源，不再单独实现 code=0。
func okDefault(r *ghttp.Request, data interface{}) {
	common.OK(r, data)
}

// writeCheckRoleCode checkRoleCode 的精确契约响应：
// Java 语义为 result 恒为 true，冲突时 success=false + code 保持默认 0 +
// message「角色编码已存在」（不走 error500，code 不是 500），前端校验器只读
// success/message，此处按原形状逐字段对齐。
func writeCheckRoleCode(r *ghttp.Request, conflict bool) {
	result := g.Map{
		"success":   true,
		"message":   "操作成功！",
		"code":      0,
		"result":    true,
		"timestamp": time.Now().UnixMilli(),
	}
	if conflict {
		result["success"] = false
		result["message"] = "角色编码已存在"
	}
	r.Response.WriteJson(result)
}

// writeNoAuth 管理操作鉴权失败响应（Java @RequiresRoles 违规 →
// JeecgBootExceptionHandler → Result.noauth：code=510）。
func writeNoAuth(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"success":   false,
		"message":   "没有权限，请联系管理员授权",
		"code":      510,
		"result":    nil,
		"timestamp": time.Now().UnixMilli(),
	})
}

// writeError 统一错误映射：业务错误返回 Java 原话术，未知错误记录日志并返回
// 通用话术（Java 版把异常消息拼进响应，Go 版脱敏）。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcRole.ErrNoAuth):
		writeNoAuth(r)
	case errors.Is(err, svcRole.ErrNotFound),
		errors.Is(err, svcRole.ErrNoRoleInfo),
		errors.Is(err, svcRole.ErrRoleNotSelected),
		errors.Is(err, svcRole.ErrNoPermConfig),
		errors.Is(err, svcRole.ErrSaveRolePermFirst),
		errors.Is(err, svcRole.ErrSaveFailed):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "角色域操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// operatorID 当前登录用户 ID（中间件 RequireAuth 注入，等价 Java getCurrentUser）。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// parsePage 解析分页参数（默认 1/10，同 Java @RequestParam 默认值）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// splitIDs 逗号拆分批量 ids 参数（去除空白项与空段，解析健壮化；
// 解析为空集时由服务层返回「未选中角色！」，对齐 Java isEmpty 分支）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// List 角色分页列表。
// 契约：GET /sys/role/list?pageNo=&pageSize=&roleName=&createTime_begin=&createTime_end=&column=&order=
func (c *Controller) List(r *ghttp.Request) {
	f := svcRole.ListFilter{
		RoleName:        r.Get("roleName").String(),
		RoleCode:        r.Get("roleCode").String(),
		Description:     r.Get("description").String(),
		CreateTimeBegin: r.Get("createTime_begin").String(),
		CreateTimeEnd:   r.Get("createTime_end").String(),
		OrderColumn:     r.Get("column").String(),
		OrderType:       r.Get("order").String(),
	}
	if v := r.Get("roleLevel"); v != nil && !v.IsNil() {
		lv := v.Int()
		f.RoleLevel = &lv
	}
	res, err := c.svc.List(r.Context(), parsePage(r), f)
	if err != nil {
		c.writeError(r, err)
		return
	}
	okDefault(r, res)
}

// Add 新增角色（admin）。
// 契约：POST /sys/role/add，body 为 SysRole 字段；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	svcIn := svcRole.SaveInput{Id: in.Id, RoleName: in.RoleName, RoleCode: in.RoleCode,
		Description: in.Description, RoleLevel: in.RoleLevel}
	if err := c.svc.Add(r.Context(), operatorID(r), svcIn); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑角色（admin）。
// 契约：PUT /sys/role/edit（Java 同时接受 POST，Go 路由两条均登记），
// 先查存在性，成功话术「修改成功！」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	svcIn := svcRole.SaveInput{Id: in.Id, RoleName: in.RoleName, RoleCode: in.RoleCode,
		Description: in.Description, RoleLevel: in.RoleLevel}
	if err := c.svc.Edit(r.Context(), operatorID(r), svcIn); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "修改成功！", nil)
}

// Delete 删除角色（admin，物理级联）。
// 契约：DELETE /sys/role/delete?id=必填；角色不存在也返回「删除角色成功」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), operatorID(r), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除角色成功", nil)
}

// DeleteBatch 批量删除角色（admin，物理级联）。
// 契约：DELETE /sys/role/deleteBatch?ids=逗号分隔；空 ids →「未选中角色！」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), operatorID(r), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除角色成功!", nil)
}

// QueryById 角色详情。
// 契约：GET /sys/role/queryById?id=必填；不存在 →「未找到对应实体」。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	okDefault(r, res)
}

// QueryAll 全量角色。
// 契约：GET /sys/role/queryall；空列表 →「未找到角色信息」。
func (c *Controller) QueryAll(r *ghttp.Request) {
	res, err := c.svc.QueryAll(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	okDefault(r, res)
}

// QueryMySubRole 当前用户可分配的角色。
// 契约：GET /sys/role/queryMySubRole；admin/dev 看全部，否则 role_level 严格小于自身等级。
func (c *Controller) QueryMySubRole(r *ghttp.Request) {
	res, err := c.svc.QueryMySubRole(r.Context(), operatorID(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	okDefault(r, res)
}

// CheckRoleCode 角色编码唯一性校验。
// 契约：GET /sys/role/checkRoleCode?id=&roleCode=；result 恒为 true，
// 冲突时 success=false + code=0 + message「角色编码已存在」。
func (c *Controller) CheckRoleCode(r *ghttp.Request) {
	conflict, err := c.svc.CheckRoleCode(r.Context(), r.Get("id").String(), r.Get("roleCode").String())
	if err != nil {
		// Java 异常分支：success=false + result=false + message=异常消息（脱敏为通用话术）
		g.Log().Errorf(r.Context(), "角色编码校验异常 err=%s", err.Error())
		r.Response.WriteJson(g.Map{
			"success":   false,
			"message":   "操作失败",
			"code":      0,
			"result":    false,
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}
	writeCheckRoleCode(r, conflict)
}

// QueryTreeList 菜单权限树（角色授权弹窗数据源）。
// 契约：GET /sys/role/queryTreeList → result {treeList, ids}。
func (c *Controller) QueryTreeList(r *ghttp.Request) {
	tree, ids, err := c.svc.QueryPermissionTreeList(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	okDefault(r, g.Map{"treeList": tree, "ids": ids})
}

// DataruleList 角色的菜单数据规则与已勾选项。
// 契约：GET /sys/role/datarule/:permissionId/:roleId → {datarule, drChecked?}；
// 菜单无数据规则 →「未找到权限配置信息」。
func (c *Controller) DataruleList(r *ghttp.Request) {
	res, err := c.svc.DataruleList(r.Context(),
		r.Get("permissionId").String(), r.Get("roleId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	// Result.ok(map)：code=200、message 保持默认「操作成功！」（Result.java:79-85
	// 不设 message）；drChecked 缺失时整键不出现（对齐 Java）
	payload := g.Map{"datarule": res.Rules}
	if res.HasChecked {
		payload["drChecked"] = res.DrChecked
	}
	common.OKMsg(r, "操作成功！", payload)
}

// DataruleSave 保存角色的菜单数据规则勾选。
// 契约：POST /sys/role/datarule，body {permissionId, roleId, dataRuleIds}；
// 关系行缺失 →「请先保存角色菜单权限!」。
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
