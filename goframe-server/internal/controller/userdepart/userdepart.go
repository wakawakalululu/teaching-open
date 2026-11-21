// Package userdepart 提供用户-部门关联域的 HTTP 处理器
// （契约蓝本：SysUserController L413/L457/L951/L1001/L1016/L1062/L1091/L1127/L1157）。
package userdepart

import (
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcUserDepart "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/userdepart"
)

// Controller 用户-部门关联域控制器。
type Controller struct {
	svc *svcUserDepart.Service
}

// New 构造控制器。
func New(svc *svcUserDepart.Service) *Controller {
	return &Controller{svc: svc}
}

// ok 内部便捷成功响应。
func ok(r *ghttp.Request, data interface{}) {
	common.OK(r, data)
}

// writeError 统一错误映射：业务错误返回原话术，未知错误记录日志并返回通用话术。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcUserDepart.ErrPermissionDenied),
		errors.Is(err, svcUserDepart.ErrNoRelation),
		errors.Is(err, svcUserDepart.ErrDepartNotFound),
		errors.Is(err, svcUserDepart.ErrParamRequired):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "用户部门操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// parsePage 解析分页参数（默认 1/10，同 Java @RequestParam 默认值；非法值回退默认）。
func parsePage(r *ghttp.Request) (pageNo, pageSize int) {
	pageNo = r.Get("pageNo", 1).Int()
	pageSize = r.Get("pageSize", 10).Int()
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return pageNo, pageSize
}

// UserDepartList 查询指定用户和部门关联的数据。
//
// 契约：GET /sys/user/userDepartList?userId=（SysUserController L413）；
// result：[{key,value,title,children:[]}]；有数据 success=true+「查找成功」，
// 无数据 success=false+「查找失败」（前端 UserModal.vue 仅在 success 时消费）。
func (c *Controller) UserDepartList(r *ghttp.Request) {
	list, err := c.svc.QueryDepartIdsOfUser(r.Context(), r.Get("userId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	if len(list) > 0 {
		// Java：message 固定为「查找成功」
		common.OKMsg(r, "查找成功", list)
		return
	}
	common.Fail(r, "查找失败")
}

// QueryUserByDepId 根据部门 id 查询用户信息（含下级部门）。
//
// 契约：GET /sys/user/queryUserByDepId?id=&realname=（SysUserController L457）；
// result：List<SysUser>（password/salt 不输出；orgCode 借位存部门名）；
// 不分页（前端 JSelectUserByDepModal 本地分页）。
func (c *Controller) QueryUserByDepId(r *ghttp.Request) {
	list, err := c.svc.QueryUserByDepId(r.Context(),
		r.Get("id").String(), r.Get("realname").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, list)
}

// DepartUserList 部门用户列表（分页）。
//
// 契约：GET /sys/user/departUserList?pageNo=&pageSize=&depId=（SysUserController L951），
// 另支持 username/realname 过滤（QueryGenerator 字符串规则）与 column/order 排序；
// result：IPage{records,total,size,current,pages}；dep 圈定为空时 result=null。
func (c *Controller) DepartUserList(r *ghttp.Request) {
	pageNo, pageSize := parsePage(r)
	res, err := c.svc.DepartUserList(r.Context(),
		r.GetCtxVar(middleware.CtxKeyUserID).String(),
		svcUserDepart.DepartUserListInput{
			PageNo:      pageNo,
			PageSize:    pageSize,
			DepId:       r.Get("depId").String(),
			Username:    r.Get("username").String(),
			Realname:    r.Get("realname").String(),
			OrderColumn: r.Get("column").String(),
			OrderType:   r.Get("order").String(),
		})
	if err != nil {
		c.writeError(r, err)
		return
	}
	if res == nil {
		ok(r, nil)
		return
	}
	ok(r, res)
}

// QueryByOrgCode 根据 orgCode 查询用户（含子部门，分页）。
//
// 契约：GET /sys/user/queryByOrgCode?pageNo=&pageSize=&orgCode=
// （SysUserController L1001，orgCode 必填）；result：
// IPage<SysUserSysDepartModel>，records 字段
// {id,realname,workNo,post,telephone,email,phone,departId,departName}。
func (c *Controller) QueryByOrgCode(r *ghttp.Request) {
	pageNo, pageSize := parsePage(r)
	res, err := c.svc.QueryByOrgCode(r.Context(), svcUserDepart.QueryUserByOrgCodeInput{
		PageNo:   pageNo,
		PageSize: pageSize,
		OrgCode:  r.Get("orgCode").String(),
		Realname: r.Get("realname").String(),
		WorkNo:   r.Get("workNo").String(),
	})
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// QueryByOrgCodeForAddressList 通讯录查询（同部门树用户合并为一条记录）。
//
// 契约：GET /sys/user/queryByOrgCodeForAddressList?pageNo=&pageSize=&orgCode=
// （SysUserController L1016，orgCode 可选，为空全查）；result：
// IPage，records 字段 {userId,realname,workNo,post,telephone,email,phone,
// departId,departName}，多部门用户 departName 以 " | " 拼接。
func (c *Controller) QueryByOrgCodeForAddressList(r *ghttp.Request) {
	pageNo, pageSize := parsePage(r)
	res, err := c.svc.QueryByOrgCodeForAddressList(r.Context(), svcUserDepart.QueryUserByOrgCodeInput{
		PageNo:   pageNo,
		PageSize: pageSize,
		OrgCode:  r.Get("orgCode").String(),
		Realname: r.Get("realname").String(),
		WorkNo:   r.Get("workNo").String(),
	})
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// editSysDepartWithUserReq 编辑用户与部门关系的请求体（Java SysDepartUsersVO）。
type editSysDepartWithUserReq struct {
	DepId      string   `json:"depId"`
	UserIdList []string `json:"userIdList"`
}

// EditSysDepartWithUser 给指定部门添加对应的用户。
//
// 契约：POST /sys/user/editSysDepartWithUser，请求体
// {depId, userIdList:[]}（SysUserController L1062）；增量补挂语义
// （已存在的关系跳过，不删除），成功话术「添加成功!」。
func (c *Controller) EditSysDepartWithUser(r *ghttp.Request) {
	var req editSysDepartWithUserReq
	if err := r.Parse(&req); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.EditSysDepartWithUser(r.Context(), req.DepId, req.UserIdList); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功!", nil)
}

// DeleteUserInDepart 删除指定机构的用户关系。
//
// 契约：DELETE /sys/user/deleteUserInDepart?depId=&userId=
// （SysUserController L1091，查询串传参）；含角色等级校验与部门角色用户
// 关系级联清理；成功话术「删除成功!」，无关联报
// 「当前选中部门与用户无关联关系!」。
func (c *Controller) DeleteUserInDepart(r *ghttp.Request) {
	if err := c.svc.DeleteUserInDepart(r.Context(),
		r.Get("depId").String(), r.Get("userId").String(),
		r.GetCtxVar(middleware.CtxKeyUserID).String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteUserInDepartBatch 批量删除指定机构的用户关系。
//
// 契约：DELETE /sys/user/deleteUserInDepartBatch?depId=&userIds=
// （SysUserController L1127，userIds 逗号分隔）；任一目标用户等级高于
// 当前用户即整体拒绝；无异常一律返回「删除成功!」。
func (c *Controller) DeleteUserInDepartBatch(r *ghttp.Request) {
	userIds := svcUserDepart.SplitIDs(r.Get("userIds").String())
	if err := c.svc.DeleteUserInDepartBatch(r.Context(),
		r.Get("depId").String(), userIds,
		r.GetCtxVar(middleware.CtxKeyUserID).String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// GetCurrentUserDeparts 查询当前用户的所有部门/当前部门编码。
//
// 契约：GET /sys/user/getCurrentUserDeparts（SysUserController L1157，
// 无参数，取当前登录人）；result：{list:[SysDepart...], orgCode}。
func (c *Controller) GetCurrentUserDeparts(r *ghttp.Request) {
	list, orgCode, err := c.svc.GetCurrentUserDeparts(r.Context(),
		r.GetCtxVar(middleware.CtxKeyUserID).String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, g.Map{
		"list":    list,
		"orgCode": orgCode,
	})
}
