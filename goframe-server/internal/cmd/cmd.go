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
	ctrlCategory "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/category"
	ctrlConfig "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/config"
	ctrlDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/dict"
	ctrlLogin "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/login"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/notify"
	svcCategory "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/category"
	svcConfig "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/config"
	svcDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/dict"
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
	var (
		revoker   auth.Revoker
		tokenOpts []auth.Option
	)
	if cfg.Security.RedisEnabled {
		revoker = svcLogin.NewRedisRevoker()
		g.Log().Info(ctx, "Redis 已启用：登录限流与令牌吊销生效")
	} else {
		// 无 Redis：吊销不可用，令牌校验 fail-open，保证受保护接口可用
		tokenOpts = append(tokenOpts, auth.WithRevokerFailOpen(true))
		g.Log().Warning(ctx, "Redis 未启用：登录限流与令牌吊销降级为 fail-open（仅适用于开发/演示）")
	}
	tokens, err := auth.NewManager(cfg.JWT, revoker, tokenOpts...)
	if err != nil {
		return err
	}
	limiter := auth.NewLoginLimiter(cfg.Security.LoginFailLimit, cfg.Security.LoginFailWindow)
	feishu := notify.NewFeishu(cfg.Security.FeishuWebhook)
	loginSvc := svcLogin.New(tokens, limiter, feishu)
	loginCtrl := ctrlLogin.New(loginSvc)
	// P0-1 业务域：字典 / 分类 / 站点配置
	dictCtrl := ctrlDict.New(svcDict.New())
	categoryCtrl := ctrlCategory.New(svcCategory.New())
	configCtrl := ctrlConfig.New(svcConfig.New(svcConfig.SiteOptions{
		UploadType:   cfg.Site.UploadType,
		StaticDomain: cfg.Site.StaticDomain,
		FilePreview:  cfg.Site.FilePreview,
		OwID:         cfg.Site.OwID,
	}))

	// 2.1 数据库连接自检（带病启动则直接失败）
	if err := g.DB().PingMaster(); err != nil {
		g.Log().Warningf(ctx, "数据库连接不可用: %s", err.Error())
		feishu.SendAsync(ctx, "teaching-open 启动失败：数据库连接不可用 "+err.Error())
		return err
	}
	g.Log().Infof(ctx, "数据库连接正常")

	// 2.2 存量密码迁移进度统计（只读，用于评估 weakest-link 规模）
	if total, bcryptCount, statErr := reportLegacyAccounts(ctx); statErr != nil {
		g.Log().Warningf(ctx, "存量密码统计失败: %s", statErr.Error())
	} else {
		legacy := total - bcryptCount
		g.Log().Infof(ctx, "密码散列迁移进度 total_users=%d bcrypt_users=%d legacy_users=%d",
			total, bcryptCount, legacy)
		if legacy > 0 {
			g.Log().Warningf(ctx, "仍有 %d 个账号使用可逆 DES 密文，需触发登录完成 rehash 迁移", legacy)
		}
	}

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
		// 与 Java Shiro anon 名单对齐的公开读接口（登录页/未登录页面也会调用）：
		// /sys/dict/getDictItems/**、/sys/config/getCurrentConfig、/sys/config/getConfig
		group.GET("/dict/getDictItems/:dictCode", dictCtrl.GetDictItems)
		group.GET("/config/getCurrentConfig", configCtrl.GetCurrentConfig)
		group.GET("/config/getConfig", configCtrl.GetConfig)
	})

	// 受保护路由
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.RequireAuth(tokens))
		group.GET("/sys/user/info", loginCtrl.CurrentUser)
		group.POST("/sys/logout", loginCtrl.Logout)
		// 密码管理：自助改密（校验旧密码）与管理员重置（服务层校验 admin 角色）
		group.PUT("/sys/user/updatePassword", loginCtrl.UpdatePassword)
		group.PUT("/sys/user/changePassword", loginCtrl.ChangePassword)

		// 字典域（P0-1）
		group.GET("/sys/dict/queryAllDictItems", dictCtrl.QueryAllDictItems)
		group.GET("/sys/dict/getDictText/:dictCode/:key", dictCtrl.GetDictText)
		group.GET("/sys/dict/loadDict/:dictCode", dictCtrl.LoadDict)
		group.GET("/sys/dict/loadDictItem/:dictCode", dictCtrl.LoadDictItem)
		group.GET("/sys/dict/list", dictCtrl.List)
		group.GET("/sys/dict/treeList", dictCtrl.TreeList)
		group.GET("/sys/dict/getByCode", dictCtrl.GetByCode)
		group.GET("/sys/dict/deleteList", dictCtrl.DeleteList)
		group.GET("/sys/dict/refleshCache", dictCtrl.RefleshCache)
		group.POST("/sys/dict/add", dictCtrl.Add)
		group.PUT("/sys/dict/edit", dictCtrl.Edit)
		group.DELETE("/sys/dict/delete", dictCtrl.Delete)
		group.DELETE("/sys/dict/deleteBatch", dictCtrl.DeleteBatch)
		group.DELETE("/sys/dict/deletePhysic/:id", dictCtrl.DeletePhysic)
		group.PUT("/sys/dict/back/:id", dictCtrl.Back)

		// 字典项域（P0-1）
		group.GET("/sys/dictItem/list", dictCtrl.ItemList)
		group.POST("/sys/dictItem/add", dictCtrl.ItemAdd)
		group.PUT("/sys/dictItem/edit", dictCtrl.ItemEdit)
		group.DELETE("/sys/dictItem/delete", dictCtrl.ItemDelete)
		group.DELETE("/sys/dictItem/deleteBatch", dictCtrl.ItemDeleteBatch)

		// 分类字典域（P0-1）
		group.GET("/sys/category/rootList", categoryCtrl.RootList)
		group.GET("/sys/category/childList", categoryCtrl.ChildList)
		group.GET("/sys/category/queryById", categoryCtrl.QueryById)
		group.GET("/sys/category/loadOne", categoryCtrl.LoadOne)
		group.GET("/sys/category/loadTreeChildren", categoryCtrl.LoadTreeChildren)
		group.GET("/sys/category/loadTreeRoot", categoryCtrl.LoadTreeRoot)
		group.GET("/sys/category/loadDictItem", categoryCtrl.LoadDictItem)
		group.GET("/sys/category/loadAllData", categoryCtrl.LoadAllData)
		group.GET("/sys/category/checkCode", categoryCtrl.CheckCode)
		group.GET("/sys/category/loadTreeData", categoryCtrl.LoadTreeData)
		group.POST("/sys/category/add", categoryCtrl.Add)
		group.PUT("/sys/category/edit", categoryCtrl.Edit)
		group.DELETE("/sys/category/delete", categoryCtrl.Delete)
		group.DELETE("/sys/category/deleteBatch", categoryCtrl.DeleteBatch)

		// 站点配置域（P0-1）
		group.GET("/sys/config/getAllConfigList", configCtrl.GetAllConfigList)
		group.GET("/sys/config/getAllConfigMap", configCtrl.GetAllConfigMap)
		group.POST("/sys/config/saveTenantConfig", configCtrl.SaveTenantConfig)
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
	registerGuard("PUT", "/api/sys/user/updatePassword", true)
	registerGuard("PUT", "/api/sys/user/changePassword", true)
	// 字典域（P0-1）
	registerGuard("POST", "/api/sys/dict/add", true)
	registerGuard("PUT", "/api/sys/dict/edit", true)
	registerGuard("DELETE", "/api/sys/dict/delete", true)
	registerGuard("DELETE", "/api/sys/dict/deleteBatch", true)
	registerGuard("DELETE", "/api/sys/dict/deletePhysic/:id", true)
	registerGuard("PUT", "/api/sys/dict/back/:id", true)
	// 字典项域（P0-1）
	registerGuard("POST", "/api/sys/dictItem/add", true)
	registerGuard("PUT", "/api/sys/dictItem/edit", true)
	registerGuard("DELETE", "/api/sys/dictItem/delete", true)
	registerGuard("DELETE", "/api/sys/dictItem/deleteBatch", true)
	// 分类字典域（P0-1）
	registerGuard("POST", "/api/sys/category/add", true)
	registerGuard("PUT", "/api/sys/category/edit", true)
	registerGuard("DELETE", "/api/sys/category/delete", true)
	registerGuard("DELETE", "/api/sys/category/deleteBatch", true)
	// 站点配置域（P0-1）
	registerGuard("POST", "/api/sys/config/saveTenantConfig", true)

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

// reportLegacyAccounts 统计仍使用 Java 版 PBE 密文的账号数量，用于评估迁移进度。
//
// 该查询只读，表名/列名为代码内常量；BCRYPT 判定与 passwd.Detect 保持一致
// （$2a$/$2b$/$2y$ 前缀），避免两处逻辑漂移。
func reportLegacyAccounts(ctx context.Context) (total, bcryptCount int, err error) {
	const query = `SELECT
			COUNT(*) AS total,
			SUM(CASE
				WHEN password LIKE '$2a$%' OR password LIKE '$2b$%' OR password LIKE '$2y$%'
				THEN 1 ELSE 0 END) AS bcrypt_count
		FROM sys_user
		WHERE del_flag = 0`

	one, err := g.DB().GetOne(ctx, query)
	if err != nil {
		return 0, 0, err
	}
	if one.IsEmpty() {
		return 0, 0, nil
	}
	total = one["total"].Int()
	bcryptCount = one["bcrypt_count"].Int()
	// SUM 在空表时返回 NULL，Int() 会得到 0，语义正确
	return total, bcryptCount, nil
}

func gctxCtx() context.Context {
	return context.Background()
}
