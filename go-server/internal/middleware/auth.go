package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wakawakalululu/teaching-open/go-server/internal/auth"
)

// 认证失败原因（对外统一话术，避免区分「用户不存在」与「密码错误」）。
const (
	msgUnauthorized = "登录状态已失效，请重新登录"
)

// RequireAuth 强制 JWT 认证。
//
// 兼容策略：Java 版前端把令牌放在 X-Access-Token 头，Go 版保留该头，
// 同时支持标准的 Authorization: Bearer，便于前端渐进迁移。
func RequireAuth(mgr *auth.Manager, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractToken(c)
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false, "code": http.StatusUnauthorized, "message": msgUnauthorized,
			})
			return
		}
		claims, err := mgr.Parse(c.Request.Context(), raw, auth.TokenTypeAccess)
		if err != nil {
			// 只记录原因，不回显给客户端
			logger.Info("鉴权失败",
				slog.String("request_id", RequestIDFrom(c)),
				slog.String("path", c.FullPath()),
				slog.String("reason", err.Error()),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false, "code": http.StatusUnauthorized, "message": msgUnauthorized,
			})
			return
		}
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyClaims, claims)
		c.Next()
	}
}

// RequireRole 校验角色，需在 RequireAuth 之后使用。
//
// 与 Java 版对比：Java 版权限校验散落在 @RequiresPermissions/@RequiresRoles 注解，
// 且 ShiroConfig 里存在大量 anon 例外，容易出现「加了注解但被 anon 短路」的情况。
// Go 版把路由分组与权限要求显式绑定，并在启动时校验：凡是写操作路由，
// 必须挂载 RequireAuth，缺失即启动失败（见 router 的 self-check）。
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		userRoles, ok := c.Get(CtxKeyRoles)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false, "code": http.StatusForbidden, "message": "没有权限，请联系管理员授权",
			})
			return
		}
		list, _ := userRoles.([]string)
		for _, r := range list {
			if _, hit := allowed[r]; hit {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false, "code": http.StatusForbidden, "message": "没有权限，请联系管理员授权",
		})
	}
}

// 上下文键。
const (
	CtxKeyUserID   = "uid"
	CtxKeyUsername = "uname"
	CtxKeyClaims   = "claims"
	CtxKeyRoles    = "roles"
)

// extractToken 从请求中提取令牌，兼容两种头。
func extractToken(c *gin.Context) string {
	if v := c.GetHeader("X-Access-Token"); v != "" {
		return strings.TrimSpace(v)
	}
	authz := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	return ""
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf)
}
