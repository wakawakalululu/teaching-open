// Package userrole 提供用户管理之「角色分配域」的 HTTP 处理器
// （契约蓝本：SysUserController 的 queryUserRole/userRoleList/addSysUserRole/
// deleteUserRole/deleteUserRoleBatch 五个端点）。
package userrole

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcUserRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/userrole"
)

// Controller 用户角色分配控制器。
type Controller struct {
	svc *svcUserRole.Service
}

// New 构造控制器。
func New(svc *svcUserRole.Service) *Controller {
	return &Controller{svc: svc}
}

// addSysUserRoleReq POST /sys/user/addSysUserRole 请求体（Java SysUserRoleVO）。
type addSysUserRoleReq struct {
	RoleId     string   `json:"roleId"`
	UserIdList []string `json:"userIdList"`
}

// ok 内部便捷成功响应。
func ok(r *ghttp.Request, data interface{}) {
	common.OK(r, data)
}

// writeError 统一错误映射：业务错误返回原话术，未知错误记录日志并返回通用话术
// （Java 版会把异常消息拼进响应，Go 版脱敏为通用话术）。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcUserRole.ErrPermissionDenied),
		errors.Is(err, svcUserRole.ErrNoUserRole),
		errors.Is(err, svcUserRole.ErrUserRequired):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "用户角色操作异常 err=%s", err.Error())
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

// splitIDs 逗号拆分批量 ids 参数（去除空白项与空段，解析健壮化）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// QueryUserRole 查询某用户已分配的角色 ID 列表。
// 契约：GET /sys/user/queryUserRole?userid=必填 → result 为角色 ID 字符串数组；
// 用户无角色时 HTTP 200 + success:false（话术「未找到用户相关角色信息」）。
func (c *Controller) QueryUserRole(r *ghttp.Request) {
	res, err := c.svc.QueryUserRole(r.Context(), r.Get("userid").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// UserRoleList 角色下用户分页列表。
// 契约：GET /sys/user/userRoleList?pageNo=&pageSize=&roleId=&username=
// result 为 IPage {records,total,size,current,pages}，records 为白名单字段。
func (c *Controller) UserRoleList(r *ghttp.Request) {
	res, err := c.svc.UserRoleList(r.Context(), parsePage(r),
		r.Get("roleId").String(), r.Get("username").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// AddSysUserRole 给指定角色添加用户（幂等，已存在关系跳过）。
// 契约：POST /sys/user/addSysUserRole，请求体 {"roleId","userIdList"}。
func (c *Controller) AddSysUserRole(r *ghttp.Request) {
	var in addSysUserRoleReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.AddSysUserRole(r.Context(), operatorID(r), in.RoleId, in.UserIdList); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功!", nil)
}

// DeleteUserRole 删除指定角色的单个用户关系。
// 契约：DELETE /sys/user/deleteUserRole?roleId=必填&userId=必填（query string）。
func (c *Controller) DeleteUserRole(r *ghttp.Request) {
	if err := c.svc.DeleteUserRole(r.Context(), operatorID(r),
		r.Get("roleId").String(), r.Get("userId").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteUserRoleBatch 批量删除指定角色的用户关系。
// 契约：DELETE /sys/user/deleteUserRoleBatch?roleId=必填&userIds=逗号分隔（query string）。
func (c *Controller) DeleteUserRoleBatch(r *ghttp.Request) {
	if err := c.svc.DeleteUserRoleBatch(r.Context(), operatorID(r),
		r.Get("roleId").String(), splitIDs(r.Get("userIds").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}
