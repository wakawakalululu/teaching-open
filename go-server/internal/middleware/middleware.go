// Package middleware 提供 HTTP 中间件：安全响应头、跨域、认证、访问日志、panic 兜底。
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 注入安全响应头。
//
// 与 Java 版对比：Java 版完全依赖 Nginx 配置注入响应头，一旦绕过网关
// （如直接访问 8080）就全部失效。Go 版在应用层先注入一份，网关层再叠加，
// 形成纵深防御。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Permitted-Cross-Domain-Policies", "none")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// 仅对 HTTPS 站点下发 HSTS；内网 HTTP 部署下发会被浏览器忽略，无副作用但无收益
		if c.Request.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// CORS 按显式白名单放行跨域请求。
//
// 关键点：
//   - 不回显任意 Origin，只放行配置中列出的值；
//   - 允许携带凭据时禁止使用通配符（浏览器规范要求）；
//   - 预检请求不进入业务 Handler。
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimRight(strings.ToLower(o), "/")] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		normalized := strings.TrimRight(strings.ToLower(origin), "/")
		if _, ok := allowed[normalized]; !ok {
			// 不在白名单：明确不返回任何 CORS 头，由浏览器拦截
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Vary", "Origin")
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		reqHeaders := c.GetHeader("Access-Control-Request-Headers")
		if reqHeaders == "" {
			reqHeaders = "Authorization,Content-Type,X-Access-Token,X-Requested-With"
		}
		h.Set("Access-Control-Allow-Headers", reqHeaders)
		h.Set("Access-Control-Expose-Headers", "X-Request-Id")
		h.Set("Access-Control-Max-Age", "3600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Recover 捕获 panic，返回统一错误结构，且不把堆栈暴露给客户端。
func Recover(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID := RequestIDFrom(c)
				logger.Error("请求处理发生 panic",
					slog.String("request_id", reqID),
					slog.String("path", c.FullPath()),
					slog.Any("panic", r),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"code":    http.StatusInternalServerError,
					"message": "服务内部错误，请稍后重试（追踪号：" + reqID + "）",
				})
			}
		}()
		c.Next()
	}
}

// RequestID 为每个请求生成/透传追踪号，便于日志关联。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" || len(reqID) > 64 {
			reqID = newRequestID()
		}
		c.Set(ctxKeyRequestID, reqID)
		c.Writer.Header().Set("X-Request-Id", reqID)
		c.Next()
	}
}

const ctxKeyRequestID = "request_id"

// RequestIDFrom 取出当前请求的追踪号。
func RequestIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// AccessLog 记录结构化访问日志；不记录请求体与敏感头。
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("http_request",
			slog.String("request_id", RequestIDFrom(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
