// Package usercore 提供用户管理之「核心 CRUD 域」的 HTTP 处理器
// （契约蓝本：SysUserController 的 list/add/edit/delete/deleteBatch/frozenBatch
// 六个端点）。
//
// 路由归属（需在 internal/cmd/cmd.go 合并注册，见本域交付报告）：
//   - 全部为受保护路由（Java Shiro /** = jwt，list 端点的 @PermissionData
//     是数据权限切面而非匿名放行）；
//   - GET /sys/user/list、POST /sys/user/add、PUT /sys/user/edit、
//     DELETE /sys/user/delete、DELETE /sys/user/deleteBatch、
//     PUT /sys/user/frozenBatch。
package usercore

import (
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcUsercore "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/usercore"
)

// Controller 用户核心 CRUD 控制器。
type Controller struct {
	svc *svcUsercore.Service
}

// New 构造控制器。
func New(svc *svcUsercore.Service) *Controller {
	return &Controller{svc: svc}
}

// writeError 统一错误映射：业务错误返回 Java 契约原话术，未知错误记录日志并
// 返回通用话术（Java add/edit/frozenBatch 的 catch 分支即「操作失败」；
// 业务失败一律 HTTP 200 + success:false）。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcUsercore.ErrPermissionDenied),
		errors.Is(err, svcUsercore.ErrEntityNotFound),
		errors.Is(err, svcUsercore.ErrNoDeptScope),
		errors.Is(err, svcUsercore.ErrRoleForbidden),
		errors.Is(err, svcUsercore.ErrPasswordRequired):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "用户核心操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// operatorID 当前登录用户 ID（中间件 RequireAuth 注入，等价 Java getCurrentUser）。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// operatorName 当前登录用户名（写 create_by/update_by 用，等价
// JwtUtil.getUserNameByToken）。
func operatorName(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUsername).String()
}

// parseListInput 解析 list 查询参数（QueryGenerator bean 过滤字段保留原始规则串）。
// 分页默认 1/10 同 Java @RequestParam；0/负值防御性归一（前端翻页控件不会发 0，
// MyBatis-Plus 对 size<=0 走全量查询语义，Go 版统一按默认值处理）。
func parseListInput(r *ghttp.Request) svcUsercore.ListInput {
	pageNo := r.Get("pageNo", 1).Int()
	pageSize := r.Get("pageSize", 10).Int()
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return svcUsercore.ListInput{
		PageNo:          pageNo,
		PageSize:        pageSize,
		Id:              r.Get("id").String(),
		Username:        r.Get("username").String(),
		Realname:        r.Get("realname").String(),
		Phone:           r.Get("phone").String(),
		Email:           r.Get("email").String(),
		Sex:             r.Get("sex").String(),
		Status:          r.Get("status").String(),
		DepartId:        r.Get("departId").String(),
		DepartIds:       r.Get("departIds").String(),
		RoleId:          r.Get("roleId").String(),
		RoleCode:        r.Get("roleCode").String(),
		CreateTimeBegin: r.Get("createTime_begin").String(),
		CreateTimeEnd:   r.Get("createTime_end").String(),
		OrderColumn:     r.Get("column").String(),
		OrderType:       r.Get("order").String(),
	}
}

// List 用户分页列表。
//
// 契约：GET /sys/user/list?pageNo=&pageSize=&[departId=&roleId=&roleCode=
// &username=&realname=&phone=&email=&sex=&status=&createTime_begin=&
// createTime_end=&column=&order=]（SysUserController L92-167）。
// 响应 result：IPage {records,total,size,current,pages}，records 为
// sys_user 全字段白名单 + sex_dictText/status_dictText/departIds_dictText +
// orgCodeTxt/roleTxt；数据范围外（无负责部门）返回 success:false
// 「您没有负责的班级」。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), operatorID(r), parseListInput(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增用户。
//
// 契约：POST /sys/user/add，请求体 = SysUser 字段（前端 UserModal 提交
// id/username/password/realname/birthday/sex/email/phone/telephone/workNo/
// post/school/avatar/activitiSync/userIdentity/departIds）+ selectedroles
// + selecteddeparts（SysUserController L169-210）。
// 响应：result=null，话术「添加成功！」；角色等级不足返回
// 「权限不足，无法分配所选角色」。
func (c *Controller) Add(r *ghttp.Request) {
	var in svcUsercore.UserPayload
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), operatorID(r), operatorName(r), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑用户。
//
// 契约：PUT /sys/user/edit，请求体同 Add（必带 id），三表级联见服务层
// （SysUserController L212-262）。响应：result=null，话术「修改成功!」；
// 目标不存在「未找到对应实体」、等级不足「权限不足」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in svcUsercore.UserPayload
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), operatorID(r), operatorName(r), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "修改成功!", nil)
}

// Delete 删除单个用户（del_flag 逻辑删）。
//
// 契约：DELETE /sys/user/delete?id=必填（query string，JeecgListMixin 的
// deleteAction 传参，SysUserController L268-276）。响应：result=null，
// 话术「删除用户成功」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), operatorID(r), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除用户成功", nil)
}

// DeleteBatch 批量删除用户。
//
// 契约：DELETE /sys/user/deleteBatch?ids=id1,id2（逗号分隔，JeecgListMixin
// batchDel 传参，SysUserController L282-292）。响应：result=null，
// 话术「批量删除用户成功」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), operatorID(r),
		svcUsercore.SplitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除用户成功", nil)
}

// FrozenBatch 冻结/解冻用户。
//
// 契约：PUT /sys/user/frozenBatch，请求体 {"ids":"id1,id2","status":1|2}
// （UserList.handleFrozen/batchFrozen 提交，SysUserController L300-324）。
// status 为 JSON 数字或字符串均可（前端两种写法并存），经 gvar 归一为字符串。
// 响应：result=null，话术「操作成功!」。
func (c *Controller) FrozenBatch(r *ghttp.Request) {
	if err := c.svc.FrozenBatch(r.Context(), operatorID(r),
		svcUsercore.SplitIDs(r.Get("ids").String()), r.Get("status").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "操作成功!", nil)
}
