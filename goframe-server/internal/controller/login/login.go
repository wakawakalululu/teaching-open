// Package login 提供登录、刷新、登出相关的 HTTP 处理器。
//
// 响应结构与 Java 版 Result<T> 对齐（success/code/message/result），
// 前端 axios 拦截器无需改动即可继续工作。
package login

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/service/login"
)

// Controller 登录控制器。
type Controller struct {
	svc *login.Service
}

// New 构造控制器。
func New(svc *login.Service) *Controller {
	return &Controller{svc: svc}
}

// ok 统一成功响应。
func ok(r *ghttp.Request, data interface{}) {
	r.Response.WriteJson(g.Map{
		"success": true,
		"code":    http.StatusOK,
		"result":  data,
	})
}

// fail 统一失败响应。
//
// 用 WriteHeader 而非 WriteStatus：后者会在未写内容时把状态文本
// （如 "Unauthorized"）写入响应体，污染 JSON。
func fail(r *ghttp.Request, status int, msg string) {
	r.Response.WriteHeader(status)
	r.Response.WriteJson(g.Map{
		"success": false,
		"code":    status,
		"message": msg,
	})
}

// Login 账号密码登录。
func (c *Controller) Login(r *ghttp.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := r.Parse(&req); err != nil || req.Username == "" || req.Password == "" {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}
	if len(req.Username) > 64 || len(req.Password) > 128 {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}

	result, err := c.svc.Login(r.Context(), req.Username, req.Password, r.GetClientIp())
	if err != nil {
		switch {
		case errors.Is(err, login.ErrAccountLocked):
			fail(r, http.StatusTooManyRequests, "密码错误次数过多，账号已临时锁定，请稍后再试")
		case errors.Is(err, login.ErrAccountDisabled):
			fail(r, http.StatusForbidden, "账号已被锁定，请联系管理员")
		case errors.Is(err, login.ErrInvalidCredential):
			fail(r, http.StatusUnauthorized, "用户名或密码错误")
		default:
			g.Log().Errorf(r.Context(), "登录处理异常 err=%s", err.Error())
			fail(r, http.StatusInternalServerError, "登录失败，请稍后重试")
		}
		return
	}

	ok(r, g.Map{
		"token":        result.Tokens.AccessToken,
		"refreshToken": result.Tokens.RefreshToken,
		"expiresIn":    result.Tokens.ExpiresIn,
		"userInfo":     publicUser(result.User, result.Roles),
		// Java 登录响应契约字段：前端缓存全量字典，免二次请求
		"sysAllDictItems": result.DictItems,
	})
}

// Refresh 用刷新令牌换新令牌（旧刷新令牌即时失效）。
func (c *Controller) Refresh(r *ghttp.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := r.Parse(&req); err != nil || req.RefreshToken == "" {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}
	result, err := c.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		fail(r, http.StatusUnauthorized, "登录状态已失效，请重新登录")
		return
	}
	ok(r, g.Map{
		"token":        result.Tokens.AccessToken,
		"refreshToken": result.Tokens.RefreshToken,
		"expiresIn":    result.Tokens.ExpiresIn,
	})
}

// Logout 吊销当前令牌。
func (c *Controller) Logout(r *ghttp.Request) {
	raw := r.Header.Get("X-Access-Token")
	if raw == "" {
		authz := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			raw = strings.TrimSpace(authz[7:])
		}
	}
	if err := c.svc.Logout(r.Context(), raw); err != nil {
		g.Log().Warning(r.Context(), "登出吊销令牌失败", err.Error())
	}
	ok(r, "退出登录成功")
}

// CurrentUser 返回当前登录用户的基本信息。
func (c *Controller) CurrentUser(r *ghttp.Request) {
	ok(r, g.Map{
		"id":       r.GetCtxVar(middleware.CtxKeyUserID).String(),
		"username": r.GetCtxVar(middleware.CtxKeyUsername).String(),
	})
}

// UpdatePassword 用户自助改密（需登录，校验旧密码）。
//
// 契约对应前端 components/tools/UserPassword.vue：
//
//	PUT /sys/user/updatePassword  {username, oldpassword, password, confirmpassword}
func (c *Controller) UpdatePassword(r *ghttp.Request) {
	var req struct {
		Username        string `json:"username"`
		OldPassword     string `json:"oldpassword"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmpassword"`
	}
	if err := r.Parse(&req); err != nil {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}
	// 防越权：只允许修改本人密码
	current := r.GetCtxVar(middleware.CtxKeyUsername).String()
	if current != "" && req.Username != "" && current != req.Username {
		fail(r, http.StatusForbidden, "只能修改本人密码")
		return
	}
	if req.Username == "" {
		req.Username = current
	}
	if len(req.Password) > 128 || len(req.OldPassword) > 128 {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}

	if err := c.svc.ChangePassword(r.Context(), req.Username, req.OldPassword, req.Password, req.ConfirmPassword); err != nil {
		c.writePasswordError(r, err)
		return
	}
	ok(r, "密码修改成功!")
}

// ChangePassword 管理员重置他人密码（不校验旧密码，需 admin 角色）。
//
// 契约对应前端 views/system/modules/PasswordModal.vue：
//
//	PUT /sys/user/changePassword  {username, password, confirmpassword}
func (c *Controller) ChangePassword(r *ghttp.Request) {
	var req struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmpassword"`
	}
	if err := r.Parse(&req); err != nil {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}
	if len(req.Password) > 128 {
		fail(r, http.StatusBadRequest, "请求参数不合法")
		return
	}
	if err := c.svc.ResetPassword(r.Context(), r.GetCtxVar(middleware.CtxKeyUserID).String(), req.Username, req.Password, req.ConfirmPassword); err != nil {
		c.writePasswordError(r, err)
		return
	}
	ok(r, "密码重置成功!")
}

// writePasswordError 将改密错误映射为统一响应。
func (c *Controller) writePasswordError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, login.ErrForbidden):
		fail(r, http.StatusForbidden, err.Error())
	case errors.Is(err, login.ErrUserNotFound):
		fail(r, http.StatusNotFound, err.Error())
	case errors.Is(err, login.ErrOldPasswordWrong):
		fail(r, http.StatusBadRequest, err.Error())
	case errors.Is(err, login.ErrEmptyNewPassword),
		errors.Is(err, login.ErrConfirmMismatch),
		errors.Is(err, login.ErrWeakPassword):
		fail(r, http.StatusBadRequest, err.Error())
	default:
		g.Log().Errorf(r.Context(), "改密处理异常 err=%s", err.Error())
		fail(r, http.StatusInternalServerError, "操作失败，请稍后重试")
	}
}

// publicUser 组装对外用户信息（白名单 DTO，新增字段默认不输出）。
//
// 字段选取依据：Java 登录响应为完整 SysUser 序列化，前端实际消费
// username/realname/avatar/orgCode/id/birthday/userIdentity（web/src 全量
// 扫描确认），此处输出白名单超集；password/salt 永不输出。
func publicUser(u *entity.SysUser, roles []string) g.Map {
	if u == nil {
		return g.Map{}
	}
	resp := g.Map{
		"id":           u.Id,
		"username":     u.Username,
		"realname":     u.Realname,
		"avatar":       u.Avatar,
		"orgCode":      u.OrgCode,
		"roles":        roles,
		"birthday":     u.Birthday,
		"sex":          u.Sex,
		"userIdentity": u.UserIdentity,
		"status":       u.Status,
	}
	if u.Phone != "" {
		resp["phone"] = maskPhone(u.Phone)
	}
	if u.Email != "" {
		resp["email"] = maskEmail(u.Email)
	}
	return resp
}

// maskPhone 手机号脱敏：保留前 3 位与后 4 位。
func maskPhone(p string) string {
	r := []rune(p)
	if len(r) < 7 {
		return strings.Repeat("*", len(r))
	}
	return string(r[:3]) + "****" + string(r[len(r)-4:])
}

// maskEmail 邮箱脱敏：保留首字符与域名。
func maskEmail(e string) string {
	at := strings.LastIndex(e, "@")
	if at <= 1 {
		return "***"
	}
	return e[:1] + "***" + e[at:]
}
