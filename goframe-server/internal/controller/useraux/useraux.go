// Package useraux 提供用户管理辅助域的 HTTP 处理器（契约蓝本：SysUserController.java）。
//
// 路由归属（需在 internal/cmd/cmd.go 合并注册，见本域交付报告）：
//   - 公开路由（Java Shiro anon 名单依据 ShiroConfig.java L82/L84）：
//     GET /sys/user/checkOnlyUser、GET /sys/user/querySysUser
//     （注册页 Register.vue 与找回密码页 Step1.vue 在未登录态调用）；
//   - 受保护路由（Java Shiro /** = jwt）：queryById/generateUserId/
//     queryByIds/recycleBin/putRecycleBin/deleteRecycleBin。
package useraux

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcUseraux "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/useraux"
)

// Controller 用户辅助域控制器。
type Controller struct {
	svc *svcUseraux.Service
}

// New 构造控制器。
func New(svc *svcUseraux.Service) *Controller {
	return &Controller{svc: svc}
}

// writeError 统一错误映射：业务错误返回 Java 契约原话术，未知错误记录日志
// 并返回通用话术（业务失败一律 HTTP 200 + success:false）。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcUseraux.ErrEntityNotFound):
		common.Fail(r, err.Error())
	case errors.Is(err, svcUseraux.ErrUserExists):
		common.Fail(r, err.Error())
	case errors.Is(err, svcUseraux.ErrVerifyFailed):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "用户辅助域操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// splitIDs 逗号拆分 ids/userIds 参数（去除空白项）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// QueryById 按 id 查用户详情。
//
// 契约：GET /sys/user/queryById?id=...（SysUserController.java L326-337）。
// 响应 result：entity.SysUser 全字段（password/salt 不序列化，等价 Java
// @JsonProperty(WRITE_ONLY)）；未找到 → success:false "未找到对应实体"。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// CheckOnlyUser 用户名/手机号/邮箱唯一性校验。
//
// 契约：GET /sys/user/checkOnlyUser?username=|email=|phone=
// （SysUserController.java L364-388）。
// 响应 result：恒为 true（Java setResult(true) 后不改写），前端只读
// res.success：存在 → success:false "用户账号已存在"；不存在 → success:true。
// 本端点为公开接口（ShiroConfig.java L82 anon），注册页/找回密码页未登录调用。
func (c *Controller) CheckOnlyUser(r *ghttp.Request) {
	exists, err := c.svc.CheckOnlyUser(r.Context(),
		r.Get("username").String(), r.Get("email").String(), r.Get("phone").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	if exists {
		common.Fail(r, svcUseraux.ErrUserExists.Error())
		return
	}
	common.OK(r, true)
}

// GenerateUserId 生成用户 ID（32 位无连字符小写十六进制 UUID v4）。
//
// 契约：GET /sys/user/generateUserId（SysUserController.java L441-449）。
// 响应 result：字符串 id；前端 UserModal.vue/DepartWindow.vue 取
// res.result 作为新增用户主键。
func (c *Controller) GenerateUserId(r *ghttp.Request) {
	id, err := c.svc.GenerateUserId()
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, id)
}

// QueryByIds 按 ids 批量查用户。
//
// 契约：GET /sys/user/queryByIds?userIds=id1,id2,...（L813-822）。
// 响应 result：entity.SysUser 数组（del_flag=0 过滤）；userIds 缺省/空白时
// 返回空数组（Java split 后按空 id 拼接 IN 列表查不出行，等价）。
func (c *Controller) QueryByIds(r *ghttp.Request) {
	res, err := c.svc.QueryByIds(r.Context(), splitIDs(r.Get("userIds").String()))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// QuerySysUser 按用户名或手机号查询用户（改密/找回密码前校验）。
//
// 契约：GET /sys/user/querySysUser?phone=... 或 ?username=...
// （SysUserController.java L1274-1303）。
// 响应 result：{username, phone(脱敏)} 白名单两字段；未命中 →
// success:false "验证失败"。本端点为公开接口（ShiroConfig.java L84 anon）。
func (c *Controller) QuerySysUser(r *ghttp.Request) {
	res, err := c.svc.QuerySysUser(r.Context(),
		r.Get("username").String(), r.Get("phone").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// RecycleBin 回收站用户列表（del_flag=1，无分页）。
//
// 契约：GET /sys/user/recycleBin（SysUserController.java L1454-1466）。
// 响应 result：entity.SysUser 数组，orgCode 已按 Java 契约覆写为部门名
// （联表 sys_user_depart + sys_depart）；Java Result.ok(list) 话术为「成功」，
// 此处用 OKMsg 精确对齐。前端 UserRecycleBinModal.vue 取 result 为表格
// 数据源（列：username/realname/avatar/orgCode，rowKey=id）。
func (c *Controller) RecycleBin(r *ghttp.Request) {
	res, err := c.svc.RecycleBin(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "成功", res)
}

// PutRecycleBin 从回收站还原用户（del_flag 1 → 0）。
//
// 契约：PUT /sys/user/putRecycleBin，body JSON {"userIds":"id1,id2"}
// （SysUserController.java L1474-1484）。端点名 put 但语义为「还原」
// （前端 handleRevert 调用），话术「还原成功」，result=null。
// update_by 写当前登录用户名（取自 JWT 中间件上下文）。
func (c *Controller) PutRecycleBin(r *ghttp.Request) {
	updateBy := r.GetCtxVar(middleware.CtxKeyUsername).String()
	if err := c.svc.PutRecycleBin(r.Context(),
		splitIDs(r.Get("userIds").String()), updateBy); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "还原成功", nil)
}

// DeleteRecycleBin 从回收站彻底删除用户（物理删除 del_flag=1 行）。
//
// 契约：DELETE /sys/user/deleteRecycleBin?userIds=id1,id2
// （SysUserController.java L1492-1498），话术「删除成功」，result=null。
// 按 Java removeLogicDeleted 级联物理删除 sys_user_depart/sys_user_role
// 关联行（单事务，见 service 注释）。
func (c *Controller) DeleteRecycleBin(r *ghttp.Request) {
	if err := c.svc.DeleteRecycleBin(r.Context(),
		splitIDs(r.Get("userIds").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功", nil)
}
