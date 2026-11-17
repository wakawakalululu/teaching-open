// Package middleware 提供 GoFrame HTTP 中间件：安全响应头、跨域白名单、认证、panic 兜底。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/auth"
)

// 认证失败原因（对外统一话术）。
const msgUnauthorized = "登录状态已失效，请重新登录"

// SecurityHeaders 注入安全响应头，形成纵深防御（应用层 + 网关层）。
func SecurityHeaders(r *ghttp.Request) {
	resp := r.Response
	resp.Header().Set("X-Content-Type-Options", "nosniff")
	resp.Header().Set("X-Frame-Options", "SAMEORIGIN")
	resp.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	resp.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	resp.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	resp.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	if r.TLS != nil {
		resp.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
	r.Middleware.Next()
}

// CORS 按显式白名单放行跨域请求，禁止通配符。
func CORS(allowedOrigins []string) ghttp.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimRight(strings.ToLower(o), "/")] = struct{}{}
	}

	return func(r *ghttp.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			r.Middleware.Next()
			return
		}
		normalized := strings.TrimRight(strings.ToLower(origin), "/")
		if _, ok := allowed[normalized]; !ok {
			if r.Method == http.MethodOptions {
				r.Response.WriteStatus(http.StatusForbidden)
				return
			}
			r.Middleware.Next()
			return
		}

		resp := r.Response
		resp.Header().Set("Access-Control-Allow-Origin", origin)
		resp.Header().Set("Vary", "Origin")
		resp.Header().Set("Access-Control-Allow-Credentials", "true")
		resp.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		reqHeaders := r.Header.Get("Access-Control-Request-Headers")
		if reqHeaders == "" {
			reqHeaders = "Authorization,Content-Type,X-Access-Token,X-Requested-With"
		}
		resp.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		resp.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
		resp.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			r.Response.WriteStatus(http.StatusNoContent)
			return
		}
		r.Middleware.Next()
	}
}

// RequireAuth 强制 JWT 认证，兼容 X-Access-Token 与 Authorization: Bearer。
func RequireAuth(mgr *auth.Manager) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		raw := extractToken(r)
		if raw == "" {
			writeUnauthorized(r)
			return
		}
		claims, err := mgr.Parse(r.Context(), raw, auth.TokenTypeAccess)
		if err != nil {
			g.Log().Infof(r.Context(), "鉴权失败 path=%s reason=%s", r.URL.Path, err.Error())
			writeUnauthorized(r)
			return
		}
		r.SetCtxVar(CtxKeyUserID, claims.UserID)
		r.SetCtxVar(CtxKeyUsername, claims.Username)
		r.SetCtxVar(CtxKeyClaims, claims)
		r.Middleware.Next()
	}
}

// 上下文键。
const (
	// CtxKeyUserID 上下文中的用户 ID 键。
	CtxKeyUserID = "uid"
	// CtxKeyUsername 上下文中的用户名键。
	CtxKeyUsername = "uname"
	// CtxKeyClaims 上下文中的 JWT 载荷键。
	CtxKeyClaims = "claims"
)

// extractToken 从请求中提取令牌，兼容两种头。
func extractToken(r *ghttp.Request) string {
	if v := r.Header.Get("X-Access-Token"); v != "" {
		return strings.TrimSpace(v)
	}
	authz := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	return ""
}

func writeUnauthorized(r *ghttp.Request) {
	r.Response.WriteStatus(http.StatusUnauthorized)
	r.Response.WriteJson(g.Map{
		"success": false,
		"code":    http.StatusUnauthorized,
		"message": msgUnauthorized,
	})
}

// Recover 捕获 panic，返回统一错误结构，不泄露堆栈。
func Recover(r *ghttp.Request) {
	defer func() {
		if err := recover(); err != nil {
			g.Log().Errorf(r.Context(), "请求处理发生 panic path=%s panic=%v", r.URL.Path, err)
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.WriteJson(g.Map{
				"success": false,
				"code":    http.StatusInternalServerError,
				"message": "服务内部错误，请稍后重试",
			})
		}
	}()
	r.Middleware.Next()
}
