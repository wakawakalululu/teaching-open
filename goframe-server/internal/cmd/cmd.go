// Package cmd 启动编排：配置校验 → 依赖装配 → 路由守卫自检 → 启动服务。
package cmd

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/genv"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/config"
	ctrlLogin "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/login"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/notify"
	svcLogin "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/login"
)

var (
	// Main 主命令。
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "teaching-open GoFrame 服务端",
		Func:  mainFunc,
	}
)

// 公开路径白名单（唯一允许绕过认证的入口）。
var publicPaths = map[string]struct{}{
	"/api/actuator/health":  {},
	"/api/sys/login":        {},
	"/api/sys/mLogin":       {},
	"/api/sys/refreshToken": {},
}

// routeGuard 路由守卫登记项。
type routeGuard struct {
	method      string
	path        string
	requireAuth bool
}

// guardedRoutes 由 registerRoutes 填充，供 verifyRouteGuards 校验。
var guardedRoutes []routeGuard

func mainFunc(ctx context.Context, parser *gcmd.Parser) (err error) {
	// 1. 业务配置强校验（缺密钥即拒绝启动）
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	g.Log().Infof(ctx, "配置加载完成 env=%s version=%s", cfg.Logistics.Env, cfg.Logistics.Version)

	// 1.1 数据库连接串：优先环境变量 MYSQL_LINK，否则用 config.yaml 兜底
	if link := genv.Get("MYSQL_LINK").String(); link != "" {
		gdb.SetConfig(gdb.Config{
			"default": gdb.ConfigGroup{
				gdb.ConfigNode{Link: link, Type: "mysql", Debug: false},
			},
		})
		g.Log().Infof(ctx, "数据库连接串已从环境变量 MYSQL_LINK 装载")
	}

	// 2. 令牌、限流、业务服务装配
	revoker := svcLogin.NewRedisRevoker()
	tokens, err := auth.NewManager(cfg.JWT, revoker)
	if err != nil {
		return err
	}
	limiter := auth.NewLoginLimiter(cfg.Security.LoginFailLimit, cfg.Security.LoginFailWindow)
	feishu := notify.NewFeishu(cfg.Security.FeishuWebhook)
	loginSvc := svcLogin.New(tokens, limiter, feishu)
	loginCtrl := ctrlLogin.New(loginSvc)

	// 2.1 数据库连接自检（带病启动则直接失败）
	if err := g.DB().PingMaster(); err != nil {
		g.Log().Warningf(ctx, "数据库连接不可用: %s", err.Error())
		feishu.SendAsync(ctx, "teaching-open 启动失败：数据库连接不可用 "+err.Error())
		return err
	}
	g.Log().Infof(ctx, "数据库连接正常")

	// 3. HTTP 服务与路由
	s := g.Server()
	s.Use(middleware.Recover)
	s.Use(middleware.SecurityHeaders)
	s.Use(middleware.CORS(cfg.Security.AllowedOrigins))

	s.BindHandler("/api/actuator/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{"status": "UP"})
	})

	// 公开路由
	s.Group("/api/sys", func(group *ghttp.RouterGroup) {
		group.POST("/login", loginCtrl.Login)
		group.POST("/mLogin", loginCtrl.Login)
		group.POST("/refreshToken", loginCtrl.Refresh)
	})

	// 受保护路由
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.RequireAuth(tokens))
		group.GET("/sys/user/info", loginCtrl.CurrentUser)
		group.POST("/sys/logout", loginCtrl.Logout)
	})

	verifyRouteGuards()

	s.Run()
	return nil
}

// registerGuard 登记路由守卫要求（供启动自检使用）。
func registerGuard(method, path string, requireAuth bool) {
	guardedRoutes = append(guardedRoutes, routeGuard{method: method, path: path, requireAuth: requireAuth})
}

// verifyRouteGuards 启动自检：所有写操作路由必须声明认证要求。
//
// 这是针对 Java 版结构性问题的修复：Java 版用「Shiro anon 名单」定义公开接口，
// 新增接口默认受 /** = jwt 保护，但只要有人往 anon 名单追加一条，该接口就永久
// 失去鉴权。Go 版反转为「默认受保护 + 显式声明公开」，未声明认证的写路由
// 若不在最小公开白名单内，直接 panic 拒绝启动（fail-fast）。
func verifyRouteGuards() {
	// 公开写路由
	registerGuard("POST", "/api/sys/login", false)
	registerGuard("POST", "/api/sys/mLogin", false)
	registerGuard("POST", "/api/sys/refreshToken", false)
	// 受保护写路由
	registerGuard("POST", "/api/sys/logout", true)

	for _, rt := range guardedRoutes {
		if !isMutating(rt.method) {
			continue
		}
		if rt.requireAuth {
			continue
		}
		if _, allowed := publicPaths[rt.path]; !allowed {
			g.Log().Errorf(gctxCtx(), "写操作路由未声明认证且不在公开白名单，拒绝启动 method=%s path=%s", rt.method, rt.path)
			panic("路由守卫校验失败: " + rt.method + " " + rt.path)
		}
	}
	if len(guardedRoutes) == 0 {
		panic("路由守卫校验失败: 未登记任何路由")
	}
}

func isMutating(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func gctxCtx() context.Context {
	return context.Background()
}
