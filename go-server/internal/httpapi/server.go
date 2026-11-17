// Package httpapi 组装 HTTP 路由与处理器。
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wakawakalululu/teaching-open/go-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/go-server/internal/config"
	"github.com/wakawakalululu/teaching-open/go-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/go-server/internal/service"
	"github.com/wakawakalululu/teaching-open/go-server/internal/store"
)

// APIResponse 统一响应结构。
//
// 字段与 Java 版 Result 对齐（success/code/message/result），
// 前端 axios 拦截器无需改动即可继续工作。
type APIResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Code: http.StatusOK, Result: data})
}

func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, APIResponse{Success: false, Code: status, Message: msg})
}

// Server 聚合 HTTP 层依赖。
type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	login  *service.LoginService
	tokens *auth.Manager
	engine *gin.Engine
}

// New 构造 HTTP 服务。
func New(cfg *config.Config, logger *slog.Logger, loginSvc *service.LoginService, tokens *auth.Manager) *Server {
	if cfg.Logistics.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	s := &Server{cfg: cfg, logger: logger, login: loginSvc, tokens: tokens}
	s.engine = s.buildEngine()
	return s
}

// Handler 返回底层 http.Handler。
func (s *Server) Handler() http.Handler { return s.engine }

// buildEngine 注册中间件与路由。
func (s *Server) buildEngine() *gin.Engine {
	r := gin.New()
	if err := r.SetTrustedProxies(s.cfg.Server.TrustedProxies); err != nil {
		s.logger.Warn("设置受信代理失败，按不信任处理", slog.String("error", err.Error()))
		_ = r.SetTrustedProxies(nil)
	}

	r.Use(middleware.RequestID())
	r.Use(middleware.Recover(s.logger))
	r.Use(middleware.AccessLog(s.logger))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS(s.cfg.Security.AllowedOrigins))
	r.MaxMultipartMemory = 8 << 20

	// 健康检查：仅返回存活状态，不暴露版本号与依赖详情
	r.GET("/api/actuator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	api := r.Group("/api")

	// 公开路由：仅登录、刷新令牌、健康检查
	pub := api.Group("/sys")
	{
		s.register(pub, http.MethodPost, "/login", false, s.handleLogin)
		s.register(pub, http.MethodPost, "/mLogin", false, s.handleLogin)
		s.register(pub, http.MethodPost, "/refreshToken", false, s.handleRefresh)
	}

	// 受保护路由：默认分组，需 JWT
	authed := api.Group("")
	authed.Use(middleware.RequireAuth(s.tokens, s.logger))
	{
		authed.GET("/sys/user/info", s.handleCurrentUser)
		s.register(authed, http.MethodPost, "/sys/logout", true, s.handleLogout)
	}

	s.verifyRouteGuards()
	return r
}

// 公开路径白名单（唯一允许绕过认证的入口）。
var publicPaths = map[string]struct{}{
	"/api/actuator/health":  {},
	"/api/sys/login":        {},
	"/api/sys/mLogin":       {},
	"/api/sys/refreshToken": {},
}

// 路由登记表：记录每条写操作路由是否要求认证。
// gin.RouteInfo 不暴露 handler 链，因此改为在注册时显式登记，
// 既能做启动断言，也能在测试中直接校验，不依赖运行时反射。
type routeGuard struct {
	method      string
	path        string
	requireAuth bool
}

// guardedRoutes 由 registerRoutes 填充，供 verifyRouteGuards 校验。
var guardedRoutes []routeGuard

// register 包装路由注册，同时登记守卫要求。
func (s *Server) register(g *gin.RouterGroup, method, path string, requireAuth bool, handlers ...gin.HandlerFunc) {
	g.Handle(method, path, handlers...)
	guardedRoutes = append(guardedRoutes, routeGuard{
		method:      method,
		path:        g.BasePath() + path,
		requireAuth: requireAuth,
	})
}

// verifyRouteGuards 启动自检：所有写操作路由必须声明认证要求。
//
// 这是针对 Java 版结构性问题的修复：
// Java 版用「Shiro 过滤链 anon 名单」定义公开接口，新增接口默认受 /** = jwt 保护，
// 但只要有人往 anon 名单追加一条（原代码已有 30+ 条），该接口就永久失去鉴权，
// 且极难在代码评审中发现。Go 版反转为「默认受保护 + 显式声明公开」，
// 一旦有写路由未声明 requireAuth 即拒绝启动（fail-fast）。
func (s *Server) verifyRouteGuards() {
	for _, rt := range guardedRoutes {
		if !isMutating(rt.method) {
			continue
		}
		if rt.requireAuth {
			continue
		}
		// 未要求认证的写路由必须落在最小公开白名单内
		if _, allowed := publicPaths[rt.path]; !allowed {
			s.logger.Error("写操作路由未声明认证且不在公开白名单，拒绝启动",
				slog.String("method", rt.method),
				slog.String("path", rt.path),
			)
			panic("路由守卫校验失败: " + rt.method + " " + rt.path)
		}
	}
	if len(guardedRoutes) == 0 {
		panic("路由守卫校验失败: 未登记任何路由")
	}
}

func isMutating(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// ---- 处理器 ----

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// handleLogin 处理账号密码登录。
func (s *Server) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数不合法")
		return
	}
	if len(req.Username) > 64 || len(req.Password) > 128 {
		fail(c, http.StatusBadRequest, "请求参数不合法")
		return
	}

	result, err := s.login.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAccountLocked):
			fail(c, http.StatusTooManyRequests, "密码错误次数过多，账号已临时锁定，请稍后再试")
		case errors.Is(err, service.ErrAccountDisabled):
			fail(c, http.StatusForbidden, "账号已被锁定，请联系管理员")
		case errors.Is(err, service.ErrInvalidCredential):
			fail(c, http.StatusUnauthorized, "用户名或密码错误")
		default:
			s.logger.Error("登录处理异常", slog.String("error", err.Error()))
			fail(c, http.StatusInternalServerError, "登录失败，请稍后重试")
		}
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Code:    http.StatusOK,
		Result: gin.H{
			"token":        result.Tokens.AccessToken,
			"refreshToken": result.Tokens.RefreshToken,
			"expiresIn":    result.Tokens.ExpiresIn,
			"userInfo":     publicUser(result.User, result.Roles),
		},
	})
}

// handleRefresh 用刷新令牌换新令牌（旧刷新令牌即时失效）。
func (s *Server) handleRefresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数不合法")
		return
	}
	result, err := s.login.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		fail(c, http.StatusUnauthorized, "登录状态已失效，请重新登录")
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true, Code: http.StatusOK,
		Result: gin.H{
			"token":        result.Tokens.AccessToken,
			"refreshToken": result.Tokens.RefreshToken,
			"expiresIn":    result.Tokens.ExpiresIn,
		},
	})
}

// handleLogout 吊销当前令牌。
func (s *Server) handleLogout(c *gin.Context) {
	raw := c.GetHeader("X-Access-Token")
	if raw == "" {
		authz := c.GetHeader("Authorization")
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			raw = strings.TrimSpace(authz[7:])
		}
	}
	if err := s.login.Logout(c.Request.Context(), raw); err != nil {
		s.logger.Warn("登出吊销令牌失败", slog.String("error", err.Error()))
	}
	ok(c, "退出登录成功")
}

// handleCurrentUser 返回当前登录用户的基本信息。
func (s *Server) handleCurrentUser(c *gin.Context) {
	ok(c, gin.H{
		"id":       c.GetString(middleware.CtxKeyUserID),
		"username": c.GetString(middleware.CtxKeyUsername),
	})
}

// publicUser 组装对外用户信息。
//
// 采用显式 DTO 而非直接序列化实体：Java 版多个接口直接返回 SysUser 实体，
// 安全性依赖字段上的 @JsonProperty(WRITE_ONLY)。一旦新增字段忘记加注解
// （如手机号、身份证、家庭住址），就会随接口一起泄露。
// Go 版用白名单 DTO，新增字段默认不输出。
func publicUser(u *store.User, roles []string) gin.H {
	if u == nil {
		return gin.H{}
	}
	resp := gin.H{
		"id":       u.ID,
		"username": u.Username,
		"realname": u.Realname,
		"avatar":   u.Avatar,
		"orgCode":  u.OrgCode.String,
		"roles":    roles,
	}
	// 手机号与邮箱按需脱敏输出（此处仅回显脱敏值，完整值走专门的资料接口）
	if u.Phone.Valid {
		resp["phone"] = maskPhone(u.Phone.String)
	}
	if u.Email.Valid {
		resp["email"] = maskEmail(u.Email.String)
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
