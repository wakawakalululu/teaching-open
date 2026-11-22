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
	ctrlAdditionalWork "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/additionalwork"
	ctrlAnnouncement "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/announcement"
	ctrlCaptcha "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/captcha"
	ctrlCategory "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/category"
	ctrlConfig "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/config"
	ctrlCourse "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/course"
	ctrlCourseDept "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/coursedept"
	ctrlCourseUnit "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/courseunit"
	ctrlDepart "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/depart"
	ctrlDepartPerm "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/departpermission"
	ctrlDepartRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/departrole"
	ctrlDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/dict"
	ctrlExcelx "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/excelx"
	ctrlLogin "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/login"
	ctrlMessage "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/message"
	ctrlOSS "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/oss"
	ctrlPermission "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/permission"
	ctrlQiniu "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/qiniu"
	ctrlRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/role"
	ctrlScratch "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/scratch"
	ctrlSms "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/sms"
	ctrlSysFile "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/sysfile"
	ctrlTeachingDayLog "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/teachingdaylog"
	ctrlTeachingMenu "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/teachingmenu"
	ctrlTeachingNews "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/teachingnews"
	ctrlTeachingUser "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/teachinguser"
	ctrlTeachingWork "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/teachingwork"
	ctrlUpload "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/upload"
	ctrlUseraux "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/useraux"
	ctrlUsercore "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/usercore"
	ctrlUserDepart "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/userdepart"
	ctrlUserReg "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/userreg"
	ctrlUserRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/userrole"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/notify"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
	svcAdditionalWork "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/additionalwork"
	svcAnnouncement "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/announcement"
	svcCaptcha "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/captcha"
	svcCategory "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/category"
	svcConfig "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/config"
	svcCourse "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/course"
	svcCourseDept "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/coursedept"
	svcCourseUnit "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/courseunit"
	svcDepart "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/depart"
	svcDepartPerm "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/departpermission"
	svcDepartRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/departrole"
	svcDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/dict"
	svcLogin "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/login"
	svcMessage "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/message"
	svcOSS "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/oss"
	svcPermission "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/permission"
	svcRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/role"
	svcScratch "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/scratch"
	svcSms "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sms"
	svcSysFile "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sysfile"
	svcSSPkg "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sysspecial"
	svcTeachingDayLog "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingdaylog"
	svcTeachingMenu "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingmenu"
	svcTeachingNews "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingnews"
	svcTeachingUser "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachinguser"
	svcTeachingWork "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingwork"
	svcUseraux "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/useraux"
	svcUsercore "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/usercore"
	svcUserDepart "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/userdepart"
	svcUserReg "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/userreg"
	svcUserRole "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/userrole"
	svcWS "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/ws"
	// svcQiniu 以值配置注入控制器（无状态服务）
	svcQiniu "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/qiniu"
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
	"/api/actuator/health":                       {},
	"/api/sys/login":                             {},
	"/api/sys/mLogin":                            {},
	"/api/sys/refreshToken":                      {},
	"/api/sys/sms":                               {},
	"/api/sys/user/checkOnlyUser":                {},
	"/api/sys/user/querySysUser":                 {},
	"/api/sys/user/register":                     {},
	"/api/sys/user/phoneVerification":            {},
	"/api/sys/user/passwordChange":               {},
	"/api/teaching/teachingCourse/getHomeCourse": {},
	"/api/teaching/menu/getUserMenu":             {},
	"/api/sys/phoneLogin":                        {},
	"/api/sys/randomImage/{key}":                 {},
	"/api/sys/checkCaptcha":                      {},
	"/api/sys/getEncryptedString":                {},
	"/api/websocket/{userId}":                    {},
	"/api/sys/ng-alain/getDictItems/{dictCode}":  {},
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
	// P0-1 业务域：字典 / 分类 / 站点配置
	dictSvc := svcDict.New()
	dictCtrl := ctrlDict.New(dictSvc)
	categoryCtrl := ctrlCategory.New(svcCategory.New())
	smsSvc := svcSms.New(cfg.Security.RedisEnabled)
	// 登录链路（依赖字典服务随响应下发全量字典）
	loginSvc := svcLogin.New(tokens, limiter, feishu, dictSvc)
	loginCtrl := ctrlLogin.NewWithSms(loginSvc, smsSvc)
	// P0-2 业务域：菜单/权限
	permissionCtrl := ctrlPermission.New(svcPermission.New())
	// P0-4 业务域：角色管理（add/edit/delete 域内做 admin 角色校验）
	roleCtrl := ctrlRole.New(svcRole.New())
	// P0-4 业务域：部门管理 / 部门角色 / 部门权限
	departCtrl := ctrlDepart.New(svcDepart.New())
	departRoleCtrl := ctrlDepartRole.New(svcDepartRole.New())
	departPermCtrl := ctrlDepartPerm.New(svcDepartPerm.New())
	// P1 业务域：课程 / 门户用户上下文（_url 注入依赖站点静态域名与上传类型）
	courseSvc := svcCourse.New(cfg.Site.StaticDomain, cfg.Site.QiniuDomain, cfg.Site.UploadType)
	courseCtrl := ctrlCourse.New(courseSvc)
	courseUnitCtrl := ctrlCourseUnit.New(svcCourseUnit.New(cfg.Site.StaticDomain,
		cfg.Site.QiniuDomain, cfg.Site.UploadType, cfg.Site.FilePreview,
		cfg.Site.OwKey, cfg.Site.OwIv))
	courseDeptCtrl := ctrlCourseDept.New(svcCourseDept.New())
	teachingMenuCtrl := ctrlTeachingMenu.New(svcTeachingMenu.New())
	teachingUserCtrl := ctrlTeachingUser.New(svcTeachingUser.New())
	// P1-4 业务域：文件记录 / 七牛令牌 / 本地上传与静态预览
	sysFileCtrl := ctrlSysFile.New(svcSysFile.New(cfg.Storage.LocalRoot))
	qiniuCtrl := ctrlQiniu.New(svcQiniu.Config{
		AccessKey: cfg.Qiniu.AccessKey,
		SecretKey: cfg.Qiniu.SecretKey,
		Bucket:    cfg.Qiniu.Bucket,
	})
	uploadCtrl := ctrlUpload.New(cfg.Site.UploadType, cfg.Storage.LocalRoot)
	// P0 收尾：OSS 文件域（本地存储后端）
	ossCtrl := ctrlOSS.New(svcOSS.New(cfg.Storage.LocalRoot, cfg.Site.StaticDomain))
	// P2 业务域：作业（标签/点赞/班级日志计数走 KV 存储，键格式兼容灰度双写）
	teachingWorkCtrl := ctrlTeachingWork.New(svcTeachingWork.New(cfg.Security.RedisEnabled,
		cfg.Site.StaticDomain, cfg.Site.QiniuDomain, cfg.Site.UploadType))
	// P2 业务域：资讯（anon 列表/详情）/附加作业
	teachingNewsCtrl := ctrlTeachingNews.New(svcTeachingNews.New())
	additionalWorkCtrl := ctrlAdditionalWork.New(svcAdditionalWork.New(cfg.Security.RedisEnabled))
	// P2 业务域：班级日志（计数行 upsert 与聚合报表）
	// WebSocket 实时推送中心（anon 端点 + doReleaseData 广播）
	wsHub := svcWS.New()
	teachingDayLogCtrl := ctrlTeachingDayLog.New(svcTeachingDayLog.New(cfg.Security.RedisEnabled))
	// P2 业务域：登录验证码（anon，Redis/内存双态 600s TTL）
	captchaCtrl := ctrlCaptcha.New(svcCaptcha.New(cfg.Security.RedisEnabled))
	// P2 业务域：消息模板/发送记录（表存在于 V2.8 DDL L821/L843）
	messageCtrl := ctrlMessage.New(svcMessage.New())
	// Excel 能力：全域 exportXls/importExcel + importStudent（excelize 引擎）
	excelxCtrl := ctrlExcelx.New(func(plain string) string {
		h, herr := passwd.Hash(plain)
		if herr != nil {
			return ""
		}
		return h
	})
	// P2 业务域：Scratch 素材库/背包（背包内容本地存储，七牛上传为文档化待补齐）
	scratchCtrl := ctrlScratch.New(svcScratch.New(cfg.Security.RedisEnabled,
		cfg.Site.StaticDomain, cfg.Site.QiniuDomain, cfg.Site.UploadType,
		cfg.Storage.LocalRoot))
	// P2 业务域：系统通告/发送记录
	announcementCtrl := ctrlAnnouncement.New(svcAnnouncement.New())
	svcAnnouncement.SetBroadcaster(wsHub)
	// P0-3 业务域：用户角色分配 / 用户-部门关联 / 用户辅助（查询/回收站）
	userRoleCtrl := ctrlUserRole.New(svcUserRole.New())
	userDepartCtrl := ctrlUserDepart.New(svcUserDepart.New())
	userauxCtrl := ctrlUseraux.New(svcUseraux.New())
	// 站点配置服务：config 域与注册链路共用同一实例（注册读取 allowReg/_defaultRole/_defaultDepart）
	configSvc := svcConfig.New(svcConfig.SiteOptions{
		UploadType:   cfg.Site.UploadType,
		StaticDomain: cfg.Site.StaticDomain,
		FilePreview:  cfg.Site.FilePreview,
		OwID:         cfg.Site.OwID,
	})
	configCtrl := ctrlConfig.New(configSvc)
	// P0-3 业务域：用户核心 CRUD；注册/找回密码自服务（短信验证码 fail-closed 存储，
	// REDIS_ENABLED=false 时走进程内存储兜底）
	usercoreCtrl := ctrlUsercore.New(svcUsercore.New())
	userRegCtrl := ctrlUserReg.New(svcUserReg.New(smsSvc, configSvc))
	smsCtrl := ctrlSms.New(smsSvc)

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
		// 用户辅助域：与 Java Shiro anon 名单对齐（ShiroConfig.java L82/L84），
		// 注册页/忘记密码页在未登录态调用
		group.GET("/user/checkOnlyUser", userauxCtrl.CheckOnlyUser)
		group.GET("/user/querySysUser", userauxCtrl.QuerySysUser)
		// 注册/找回密码自服务链路（P0-3，ShiroConfig.java L80/L83/L85/L86 anon）
		group.POST("/sms", smsCtrl.Send)
		group.POST("/user/register", userRegCtrl.Register)
		group.POST("/user/phoneVerification", userRegCtrl.PhoneVerification)
		group.GET("/user/passwordChange", userRegCtrl.PasswordChange)
		// 登录验证码域 anon（登录页未登录可调）——注意组前缀已是 /api/sys，
		// 路由串不得再带 /sys（Java 契约 /sys/randomImage → /api/sys/randomImage）
		group.GET("/randomImage/{key}", captchaCtrl.RandomImage)
		group.POST("/checkCaptcha", captchaCtrl.CheckCaptcha)
		// 手机登录/加密参数 anon
		group.POST("/phoneLogin", loginCtrl.PhoneLogin)
		group.GET("/getEncryptedString", loginCtrl.GetEncryptedString)
		// 本地静态预览 anon（ShiroConfig.java:88 图片预览与下载不限 token）
		group.GET("/common/static/*path", uploadCtrl.Static)
	})

	// 教学/门户域 anon：这些控制器在 Java 中是顶层 @RequestMapping("/teaching/...")
	// 与 WebSocket /websocket/{userId}，前端以 baseURL=/api 调用 → 真实路径
	// /api/teaching/... 与 /api/websocket/...；不得放进 /api/sys 组
	// （否则实际路径多出 /sys 前缀，前端 404 且路由守卫启动即失败）。
	s.Group("/api", func(group *ghttp.RouterGroup) {
		// WebSocket 实时推送 anon（ShiroConfig.java:129）——Java 消息 JSON 形状
		group.GET("/websocket/{userId}", wsHub.Handler)
		// 课程域 anon（ShiroConfig.java:158 首页未登录可调）
		group.GET("/teaching/teachingCourse/getHomeCourse", courseCtrl.GetHomeCourse)
		// 门户菜单 anon（ShiroConfig.java:145）
		group.GET("/teaching/menu/getUserMenu", teachingMenuCtrl.GetUserMenu)
		// 作业域 anon（ShiroConfig.java:152-157）
		group.GET("/teaching/teachingWork/userInfo", teachingWorkCtrl.GetUserInfo)
		group.GET("/teaching/teachingWork/studentWorkInfo", teachingWorkCtrl.GetStudentWorkInfo)
		group.GET("/teaching/teachingWork/greatWork", teachingWorkCtrl.GreatWorkList)
		group.GET("/teaching/teachingWork/starWork", teachingWorkCtrl.StarWork)
		group.GET("/teaching/teachingWork/leaderboard", teachingWorkCtrl.Leaderboard)
		group.GET("/teaching/teachingWork/getWorkComments", teachingWorkCtrl.GetWorkComments)
		// 资讯域 anon（ShiroConfig.java:160-161）
		group.GET("/teaching/teachingNews/newsList", teachingNewsCtrl.NewsList)
		group.GET("/teaching/teachingNews/newsDetail", teachingNewsCtrl.NewsDetail)
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

		// 菜单/权限域（P0-2）
		group.GET("/sys/permission/getUserPermissionByToken", permissionCtrl.GetUserPermissionByToken)

		// 菜单权限 CRUD / 角色授权 / 数据规则 / 部门授权域（P0-4b）
		group.GET("/sys/permission/list", permissionCtrl.MenuList)
		group.GET("/sys/permission/getSystemMenuList", permissionCtrl.SystemMenuList)
		group.GET("/sys/permission/getSystemSubmenu", permissionCtrl.SystemSubmenu)
		group.GET("/sys/permission/getSystemSubmenuBatch", permissionCtrl.SystemSubmenuBatch)
		group.POST("/sys/permission/add", permissionCtrl.MenuAdd)
		group.PUT("/sys/permission/edit", permissionCtrl.MenuEdit)
		group.POST("/sys/permission/edit", permissionCtrl.MenuEdit)
		// Java 契约为 GET（SysPermissionController#edit）——补方法别名，防 405/404
		group.GET("/sys/permission/edit", permissionCtrl.MenuEdit)
		group.DELETE("/sys/permission/delete", permissionCtrl.MenuDelete)
		group.DELETE("/sys/permission/deleteBatch", permissionCtrl.MenuDeleteBatch)
		group.GET("/sys/permission/queryTreeList", permissionCtrl.PermissionTreeList)
		group.GET("/sys/permission/queryListAsync", permissionCtrl.QueryListAsync)
		group.GET("/sys/permission/queryRolePermission", permissionCtrl.QueryRolePermission)
		group.POST("/sys/permission/saveRolePermission", permissionCtrl.SaveRolePermission)
		group.GET("/sys/permission/getPermRuleListByPermId", permissionCtrl.GetPermRuleListByPermId)
		group.POST("/sys/permission/addPermissionRule", permissionCtrl.AddPermissionRule)
		group.PUT("/sys/permission/editPermissionRule", permissionCtrl.EditPermissionRule)
		group.POST("/sys/permission/editPermissionRule", permissionCtrl.EditPermissionRule)
		// Java 契约为 GET（SysPermissionController#editPermissionRule）
		group.GET("/sys/permission/editPermissionRule", permissionCtrl.EditPermissionRule)
		group.DELETE("/sys/permission/deletePermissionRule", permissionCtrl.DeletePermissionRule)
		group.GET("/sys/permission/queryPermissionRule", permissionCtrl.QueryPermissionRule)
		group.GET("/sys/permission/queryDepartPermission", permissionCtrl.QueryDepartPermission)
		group.POST("/sys/permission/saveDepartPermission", permissionCtrl.SaveDepartPermission)

		// 角色管理域（P0-4a）
		group.GET("/sys/role/list", roleCtrl.List)
		group.POST("/sys/role/add", roleCtrl.Add)
		group.PUT("/sys/role/edit", roleCtrl.Edit)
		group.POST("/sys/role/edit", roleCtrl.Edit)
		group.DELETE("/sys/role/delete", roleCtrl.Delete)
		group.DELETE("/sys/role/deleteBatch", roleCtrl.DeleteBatch)
		group.GET("/sys/role/queryById", roleCtrl.QueryById)
		group.GET("/sys/role/queryall", roleCtrl.QueryAll)
		group.GET("/sys/role/queryMySubRole", roleCtrl.QueryMySubRole)
		group.GET("/sys/role/checkRoleCode", roleCtrl.CheckRoleCode)
		group.GET("/sys/role/queryTreeList", roleCtrl.QueryTreeList)
		group.GET("/sys/role/datarule/:permissionId/:roleId", roleCtrl.DataruleList)
		group.POST("/sys/role/datarule", roleCtrl.DataruleSave)

		// 部门管理域（P0-4c）
		group.GET("/sys/sysDepart/queryTreeList", departCtrl.QueryTreeList)
		group.GET("/sys/sysDepart/queryMyDeptTreeList", departCtrl.QueryMyDeptTreeList)
		group.GET("/sys/sysDepart/queryIdTree", departCtrl.QueryIdTree)
		group.POST("/sys/sysDepart/add", departCtrl.Add)
		group.PUT("/sys/sysDepart/edit", departCtrl.Edit)
		group.DELETE("/sys/sysDepart/delete", departCtrl.Delete)
		group.DELETE("/sys/sysDepart/deleteBatch", departCtrl.DeleteBatch)
		group.GET("/sys/sysDepart/searchBy", departCtrl.SearchBy)
		group.GET("/sys/sysDepart/listAll", departCtrl.ListAll)
		group.GET("/sys/sysDepart/removeAll", departCtrl.RemoveAll)

		// 部门角色域（P0-4c）
		group.GET("/sys/sysDepartRole/list", departRoleCtrl.List)
		group.POST("/sys/sysDepartRole/add", departRoleCtrl.Add)
		group.PUT("/sys/sysDepartRole/edit", departRoleCtrl.Edit)
		group.DELETE("/sys/sysDepartRole/delete", departRoleCtrl.Delete)
		group.DELETE("/sys/sysDepartRole/deleteBatch", departRoleCtrl.DeleteBatch)
		group.GET("/sys/sysDepartRole/queryById", departRoleCtrl.QueryById)
		group.GET("/sys/sysDepartRole/getDeptRoleList", departRoleCtrl.GetDeptRoleList)
		group.POST("/sys/sysDepartRole/deptRoleUserAdd", departRoleCtrl.DeptRoleUserAdd)
		group.GET("/sys/sysDepartRole/getDeptRoleByUserId", departRoleCtrl.GetDeptRoleByUserID)
		group.GET("/sys/sysDepartRole/datarule/:permissionId/:departId/:roleId", departRoleCtrl.DataruleList)
		group.POST("/sys/sysDepartRole/datarule", departRoleCtrl.DataruleSave)

		// 部门权限域（P0-4c）
		group.GET("/sys/sysDepartPermission/list", departPermCtrl.List)
		group.POST("/sys/sysDepartPermission/add", departPermCtrl.Add)
		group.PUT("/sys/sysDepartPermission/edit", departPermCtrl.Edit)
		group.DELETE("/sys/sysDepartPermission/delete", departPermCtrl.Delete)
		group.DELETE("/sys/sysDepartPermission/deleteBatch", departPermCtrl.DeleteBatch)
		group.GET("/sys/sysDepartPermission/queryById", departPermCtrl.QueryById)
		group.GET("/sys/sysDepartPermission/datarule/:permissionId/:departId", departPermCtrl.DataruleList)
		group.POST("/sys/sysDepartPermission/datarule", departPermCtrl.DataruleSave)
		group.GET("/sys/sysDepartPermission/queryDeptRolePermission", departPermCtrl.QueryDeptRolePermission)
		group.POST("/sys/sysDepartPermission/saveDeptRolePermission", departPermCtrl.SaveDeptRolePermission)
		group.GET("/sys/sysDepartPermission/queryTreeListForDeptRole", departPermCtrl.QueryTreeListForDeptRole)

		// 课程域（P1-1）
		group.GET("/teaching/teachingCourse/mineCourse", courseCtrl.MineCourse)
		group.GET("/teaching/teachingCourse/list", courseCtrl.List)
		group.POST("/teaching/teachingCourse/add", courseCtrl.Add)
		group.PUT("/teaching/teachingCourse/edit", courseCtrl.Edit)
		group.DELETE("/teaching/teachingCourse/delete", courseCtrl.Delete)
		group.DELETE("/teaching/teachingCourse/deleteBatch", courseCtrl.DeleteBatch)
		group.GET("/teaching/teachingCourse/queryById", courseCtrl.QueryById)

		// 门户用户上下文域（P1-1；@RequestMapping 无 method → GET/POST 双登记）
		group.GET("/teaching/user/info", teachingUserCtrl.Info)
		group.POST("/teaching/user/info", teachingUserCtrl.Info)
		group.GET("/teaching/user/getUserRootDept", teachingUserCtrl.UserRootDept)
		group.POST("/teaching/user/getUserRootDept", teachingUserCtrl.UserRootDept)
		group.PUT("/teaching/user/edit", teachingUserCtrl.Edit)
		group.GET("/teaching/user/appUserList", teachingUserCtrl.AppUserList)
		group.PUT("/sys/user/appEdit", teachingUserCtrl.AppEdit)

		// 课程单元域（P1-2）
		group.GET("/teaching/teachingCourseUnit/mineUnit", courseUnitCtrl.MineUnit)
		group.GET("/teaching/teachingCourseUnit/list", courseUnitCtrl.List)
		group.GET("/teaching/teachingCourseUnit/getUnitWorkInfo", courseUnitCtrl.GetUnitWorkInfo)
		group.POST("/teaching/teachingCourseUnit/add", courseUnitCtrl.Add)
		group.PUT("/teaching/teachingCourseUnit/edit", courseUnitCtrl.Edit)
		group.PUT("/teaching/teachingCourseUnit/editBatch", courseUnitCtrl.EditBatch)
		group.DELETE("/teaching/teachingCourseUnit/delete", courseUnitCtrl.Delete)
		group.DELETE("/teaching/teachingCourseUnit/deleteBatch", courseUnitCtrl.DeleteBatch)
		group.GET("/teaching/teachingCourseUnit/queryById", courseUnitCtrl.QueryById)

		// 班级课程授权域（P1-3）
		group.GET("/teaching/teachingCourseDept/list", courseDeptCtrl.List)
		group.POST("/teaching/teachingCourseDept/addOrUpdate", courseDeptCtrl.AddOrUpdate)
		group.POST("/teaching/teachingCourseDept/add", courseDeptCtrl.Add)
		group.PUT("/teaching/teachingCourseDept/edit", courseDeptCtrl.Edit)
		group.DELETE("/teaching/teachingCourseDept/delete", courseDeptCtrl.Delete)
		group.DELETE("/teaching/teachingCourseDept/deleteBatch", courseDeptCtrl.DeleteBatch)
		group.GET("/teaching/teachingCourseDept/queryById", courseDeptCtrl.QueryById)

		// 门户菜单域（P1-3；getUserMenu 为 anon 走公开组）
		group.GET("/teaching/menu/list", teachingMenuCtrl.List)
		group.GET("/teaching/menu/getSystemMenuList", teachingMenuCtrl.GetSystemMenuList)
		group.GET("/teaching/menu/getSystemSubmenu", teachingMenuCtrl.GetSystemSubmenu)
		group.GET("/teaching/menu/getSystemSubmenuBatch", teachingMenuCtrl.GetSystemSubmenuBatch)
		group.POST("/teaching/menu/add", teachingMenuCtrl.Add)
		group.PUT("/teaching/menu/edit", teachingMenuCtrl.Edit)
		group.POST("/teaching/menu/edit", teachingMenuCtrl.Edit)
		// Java 契约为 GET（TeachingMenuController#edit）
		group.GET("/teaching/menu/edit", teachingMenuCtrl.Edit)
		group.DELETE("/teaching/menu/delete", teachingMenuCtrl.Delete)
		group.DELETE("/teaching/menu/deleteBatch", teachingMenuCtrl.DeleteBatch)
		group.GET("/teaching/menu/queryTreeList", teachingMenuCtrl.QueryTreeList)
		group.GET("/teaching/menu/queryListAsync", teachingMenuCtrl.QueryListAsync)

		// 文件记录域（P1-4，基路径 /system/sysFile 对齐 Java）
		group.GET("/system/sysFile/list", sysFileCtrl.List)
		group.POST("/system/sysFile/add", sysFileCtrl.Add)
		group.PUT("/system/sysFile/edit", sysFileCtrl.Edit)
		group.DELETE("/system/sysFile/delete", sysFileCtrl.Delete)
		group.DELETE("/system/sysFile/deleteByPath", sysFileCtrl.DeleteByPath)
		group.DELETE("/system/sysFile/deleteBatch", sysFileCtrl.DeleteBatch)
		group.GET("/system/sysFile/queryById", sysFileCtrl.QueryById)

		// 七牛令牌域（P1-4；@RequestMapping 无 method → GET/POST 双登记）
		group.GET("/common/qiniu/getToken", qiniuCtrl.GetToken)
		group.POST("/common/qiniu/getToken", qiniuCtrl.GetToken)
		group.GET("/common/qiniu/getTokenByKey", qiniuCtrl.GetTokenByKey)
		group.POST("/common/qiniu/getTokenByKey", qiniuCtrl.GetTokenByKey)

		// 本地上传（P1-4）
		group.POST("/sys/common/upload", uploadCtrl.Upload)

		// 作业域（P2-1）
		group.GET("/teaching/teachingWork/mine", teachingWorkCtrl.Mine)
		group.GET("/teaching/teachingWork/mineAdditionalWork", teachingWorkCtrl.MineAdditionalWork)
		group.POST("/teaching/teachingWork/submit", teachingWorkCtrl.Submit)
		group.POST("/teaching/teachingWork/sendWork", teachingWorkCtrl.SendWork)
		group.GET("/teaching/teachingWork/list", teachingWorkCtrl.WorkList)
		group.POST("/teaching/teachingWork/add", teachingWorkCtrl.WorkAdd)
		group.PUT("/teaching/teachingWork/edit", teachingWorkCtrl.WorkEdit)
		group.DELETE("/teaching/teachingWork/delete", teachingWorkCtrl.WorkDelete)
		group.DELETE("/teaching/teachingWork/deleteBatch", teachingWorkCtrl.WorkDeleteBatch)
		group.GET("/teaching/teachingWork/queryById", teachingWorkCtrl.WorkQueryById)
		group.GET("/teaching/teachingWork/queryTeachingWorkCorrectByMainId", teachingWorkCtrl.QueryCorrectsByMainID)
		group.GET("/teaching/teachingWork/queryTeachingWorkCommentByMainId", teachingWorkCtrl.QueryCommentsByMainID)

		// 资讯域（P2-2）
		group.GET("/teaching/teachingNews/list", teachingNewsCtrl.List)
		group.POST("/teaching/teachingNews/add", teachingNewsCtrl.Add)
		group.PUT("/teaching/teachingNews/edit", teachingNewsCtrl.Edit)
		group.DELETE("/teaching/teachingNews/delete", teachingNewsCtrl.Delete)
		group.DELETE("/teaching/teachingNews/deleteBatch", teachingNewsCtrl.DeleteBatch)
		group.GET("/teaching/teachingNews/queryById", teachingNewsCtrl.QueryByID)

		// 附加作业域（P2-2）
		group.GET("/teaching/teachingAdditionalWork/getWorkInfo", additionalWorkCtrl.GetWorkInfo)
		group.GET("/teaching/teachingAdditionalWork/list", additionalWorkCtrl.List)
		group.GET("/teaching/teachingAdditionalWork/teacherList", additionalWorkCtrl.TeacherList)
		group.POST("/teaching/teachingAdditionalWork/add", additionalWorkCtrl.Add)
		group.PUT("/teaching/teachingAdditionalWork/edit", additionalWorkCtrl.Edit)
		group.DELETE("/teaching/teachingAdditionalWork/delete", additionalWorkCtrl.Delete)
		group.DELETE("/teaching/teachingAdditionalWork/deleteBatch", additionalWorkCtrl.DeleteBatch)
		group.GET("/teaching/teachingAdditionalWork/queryById", additionalWorkCtrl.QueryByID)

		// 班级日志域（P2-3）
		group.GET("/teaching/teachingDepartDayLog/getReport", teachingDayLogCtrl.GetReport)
		group.GET("/teaching/teachingDepartDayLog/getReportGroupByDepart", teachingDayLogCtrl.GetReportGroupByDepart)
		group.GET("/teaching/teachingDepartDayLog/getReportGroupByMonth", teachingDayLogCtrl.GetReportGroupByMonth)
		group.GET("/teaching/teachingDepartDayLog/list", teachingDayLogCtrl.List)
		group.POST("/teaching/teachingDepartDayLog/add", teachingDayLogCtrl.Add)
		group.PUT("/teaching/teachingDepartDayLog/edit", teachingDayLogCtrl.Edit)
		group.DELETE("/teaching/teachingDepartDayLog/delete", teachingDayLogCtrl.Delete)
		group.DELETE("/teaching/teachingDepartDayLog/deleteBatch", teachingDayLogCtrl.DeleteBatch)
		group.GET("/teaching/teachingDepartDayLog/queryById", teachingDayLogCtrl.QueryByID)
		group.GET("/teaching/teachingDepartDayLog/unitViewLog", teachingDayLogCtrl.UnitViewLog)

		// Scratch 素材库域（P2-4）
		group.GET("/teaching/teachingScratchAssets/list", scratchCtrl.List)
		group.POST("/teaching/teachingScratchAssets/add", scratchCtrl.AssetAdd)
		group.PUT("/teaching/teachingScratchAssets/edit", scratchCtrl.AssetEdit)
		group.DELETE("/teaching/teachingScratchAssets/delete", scratchCtrl.AssetDelete)
		group.DELETE("/teaching/teachingScratchAssets/deleteBatch", scratchCtrl.AssetDeleteBatch)
		group.GET("/teaching/teachingScratchAssets/queryById", scratchCtrl.AssetQueryByID)
		group.GET("/teaching/teachingScratchAssets/getScratchAssets", scratchCtrl.GetScratchAssets)

		// Scratch 背包域（P2-4）
		group.GET("/teaching/scratch/backpack/content", scratchCtrl.BackpackContent)
		group.POST("/teaching/scratch/backpack/save", scratchCtrl.BackpackSave)
		group.DELETE("/teaching/scratch/backpack/delete", scratchCtrl.BackpackDelete)
		group.GET("/teaching/scratch/backpack/getFile/*filename", scratchCtrl.BackpackGetFile)

		// 系统通告域（P2-5）
		group.GET("/sys/annountCement/list", announcementCtrl.List)
		group.POST("/sys/annountCement/add", announcementCtrl.Add)
		group.PUT("/sys/annountCement/edit", announcementCtrl.Edit)
		group.DELETE("/sys/annountCement/delete", announcementCtrl.Delete)
		group.DELETE("/sys/annountCement/deleteBatch", announcementCtrl.DeleteBatch)
		group.GET("/sys/annountCement/queryById", announcementCtrl.QueryByID)
		group.GET("/sys/annountCement/doReleaseData", announcementCtrl.DoReleaseData)
		group.GET("/sys/annountCement/doReovkeData", announcementCtrl.DoReovkeData)
		group.GET("/sys/annountCement/listByUser", announcementCtrl.ListByUser)
		group.GET("/sys/annountCement/syncNotic", announcementCtrl.SyncNotic)

		// 通告发送记录域（P2-5）
		group.GET("/sys/sysAnnouncementSend/list", announcementCtrl.SendList)
		group.POST("/sys/sysAnnouncementSend/add", announcementCtrl.SendAdd)
		group.PUT("/sys/sysAnnouncementSend/edit", announcementCtrl.SendEdit)
		group.DELETE("/sys/sysAnnouncementSend/delete", announcementCtrl.SendDelete)
		group.DELETE("/sys/sysAnnouncementSend/deleteBatch", announcementCtrl.SendDeleteBatch)
		group.GET("/sys/sysAnnouncementSend/queryById", announcementCtrl.SendQueryByID)
		group.PUT("/sys/sysAnnouncementSend/editByAnntIdAndUserId", announcementCtrl.EditByAnntIdAndUserID)
		group.GET("/sys/sysAnnouncementSend/getMyAnnouncementSend", announcementCtrl.GetMyAnnouncementSend)
		group.PUT("/sys/sysAnnouncementSend/readAll", announcementCtrl.ReadAll)

		// 消息模板域（P2-7）
		group.GET("/message/sysMessageTemplate/list", messageCtrl.TemplateList)
		group.POST("/message/sysMessageTemplate/add", messageCtrl.TemplateAdd)
		group.PUT("/message/sysMessageTemplate/edit", messageCtrl.TemplateEdit)
		group.DELETE("/message/sysMessageTemplate/delete", messageCtrl.TemplateDelete)
		group.DELETE("/message/sysMessageTemplate/deleteBatch", messageCtrl.TemplateDeleteBatch)
		group.GET("/message/sysMessageTemplate/queryById", messageCtrl.TemplateQueryByID)
		group.POST("/message/sysMessageTemplate/sendMsg", messageCtrl.SendMsg)

		// 消息发送记录域（P2-7）
		group.GET("/message/sysMessage/list", messageCtrl.SmsList)
		group.POST("/message/sysMessage/add", messageCtrl.SmsAdd)
		group.PUT("/message/sysMessage/edit", messageCtrl.SmsEdit)
		group.DELETE("/message/sysMessage/delete", messageCtrl.SmsDelete)
		group.DELETE("/message/sysMessage/deleteBatch", messageCtrl.SmsDeleteBatch)
		group.GET("/message/sysMessage/queryById", messageCtrl.SmsQueryByID)

		// P2-6 泛型域与特殊端点（在闭包内以 group/guard 装配）
		wireP2_6Domains(group, registerGuard, svcSSPkg.New())

		// P2-8 收尾域：SysUserAgent（syscrud）+ ng-alain（Angular 辅助）
		wireUserAgentAndNgAlain(group, registerGuard, svcPermission.New())

		// 契约补齐：方法别名 + 遗漏端点（Java 411 方法级对账后逐条对齐）
		wireContractGaps(group, registerGuard, dictCtrl, loginCtrl, teachingUserCtrl, uploadCtrl)

		// P0 收尾：字典表分页/403/transitRESTful/OSS/工作台真实数据
		wireP0Gaps(group, registerGuard, dictCtrl, ossCtrl)

		// Excel 能力：全域 exportXls/importExcel + importStudent（excelize 引擎）
		for _, exBase := range excelxBases {
			group.GET(exBase+"/exportXls", excelxCtrl.ExportXls(exBase))
			group.POST(exBase+"/exportXls", excelxCtrl.ExportXls(exBase))
			group.POST(exBase+"/importExcel", excelxCtrl.ImportExcel(exBase))
			registerGuard("GET", "/api"+exBase+"/exportXls", true)
			registerGuard("POST", "/api"+exBase+"/exportXls", true)
			registerGuard("POST", "/api"+exBase+"/importExcel", true)
		}
		// 学生批量导入（student 角色 + 默认班级绑定）
		group.POST("/sys/user/importStudent", excelxCtrl.ImportStudent)
		registerGuard("POST", "/api/sys/user/importStudent", true)
		group.POST("/teaching/teachingWork/saveComment", teachingWorkCtrl.SaveComment)
		group.GET("/teaching/teachingWork/getWorkTags", teachingWorkCtrl.GetWorkTags)
		group.GET("/teaching/teachingWork/setWorkTag", teachingWorkCtrl.SetWorkTag)
		group.DELETE("/teaching/teachingWork/delWorkTag", teachingWorkCtrl.DelWorkTag)

		// 用户角色分配域（P0-3）
		group.GET("/sys/user/queryUserRole", userRoleCtrl.QueryUserRole)
		group.GET("/sys/user/userRoleList", userRoleCtrl.UserRoleList)
		group.POST("/sys/user/addSysUserRole", userRoleCtrl.AddSysUserRole)
		group.DELETE("/sys/user/deleteUserRole", userRoleCtrl.DeleteUserRole)
		group.DELETE("/sys/user/deleteUserRoleBatch", userRoleCtrl.DeleteUserRoleBatch)

		// 用户-部门关联域（P0-3）
		group.GET("/sys/user/userDepartList", userDepartCtrl.UserDepartList)
		group.GET("/sys/user/queryUserByDepId", userDepartCtrl.QueryUserByDepId)
		group.GET("/sys/user/departUserList", userDepartCtrl.DepartUserList)
		group.GET("/sys/user/queryByOrgCode", userDepartCtrl.QueryByOrgCode)
		group.GET("/sys/user/queryByOrgCodeForAddressList", userDepartCtrl.QueryByOrgCodeForAddressList)
		group.POST("/sys/user/editSysDepartWithUser", userDepartCtrl.EditSysDepartWithUser)
		group.DELETE("/sys/user/deleteUserInDepart", userDepartCtrl.DeleteUserInDepart)
		group.DELETE("/sys/user/deleteUserInDepartBatch", userDepartCtrl.DeleteUserInDepartBatch)
		group.GET("/sys/user/getCurrentUserDeparts", userDepartCtrl.GetCurrentUserDeparts)

		// 用户辅助域（P0-3）
		group.GET("/sys/user/queryById", userauxCtrl.QueryById)
		group.GET("/sys/user/generateUserId", userauxCtrl.GenerateUserId)
		group.GET("/sys/user/queryByIds", userauxCtrl.QueryByIds)
		group.GET("/sys/user/recycleBin", userauxCtrl.RecycleBin)
		group.PUT("/sys/user/putRecycleBin", userauxCtrl.PutRecycleBin)
		group.DELETE("/sys/user/deleteRecycleBin", userauxCtrl.DeleteRecycleBin)

		// 用户核心 CRUD 域（P0-3）
		group.GET("/sys/user/list", usercoreCtrl.List)
		group.POST("/sys/user/add", usercoreCtrl.Add)
		group.PUT("/sys/user/edit", usercoreCtrl.Edit)
		group.DELETE("/sys/user/delete", usercoreCtrl.Delete)
		group.DELETE("/sys/user/deleteBatch", usercoreCtrl.DeleteBatch)
		group.PUT("/sys/user/frozenBatch", usercoreCtrl.FrozenBatch)
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
	// 菜单/权限域（P0-2）
	registerGuard("GET", "/api/sys/permission/getUserPermissionByToken", true)

	// 菜单权限 CRUD / 角色授权 / 数据规则 / 部门授权域（P0-4b）
	registerGuard("GET", "/api/sys/permission/list", true)
	registerGuard("GET", "/api/sys/permission/getSystemMenuList", true)
	registerGuard("GET", "/api/sys/permission/getSystemSubmenu", true)
	registerGuard("GET", "/api/sys/permission/getSystemSubmenuBatch", true)
	registerGuard("POST", "/api/sys/permission/add", true)
	registerGuard("PUT", "/api/sys/permission/edit", true)
	registerGuard("POST", "/api/sys/permission/edit", true)
	registerGuard("DELETE", "/api/sys/permission/delete", true)
	registerGuard("DELETE", "/api/sys/permission/deleteBatch", true)
	registerGuard("GET", "/api/sys/permission/queryTreeList", true)
	registerGuard("GET", "/api/sys/permission/queryListAsync", true)
	registerGuard("GET", "/api/sys/permission/queryRolePermission", true)
	registerGuard("POST", "/api/sys/permission/saveRolePermission", true)
	registerGuard("GET", "/api/sys/permission/getPermRuleListByPermId", true)
	registerGuard("POST", "/api/sys/permission/addPermissionRule", true)
	registerGuard("PUT", "/api/sys/permission/editPermissionRule", true)
	registerGuard("POST", "/api/sys/permission/editPermissionRule", true)
	registerGuard("DELETE", "/api/sys/permission/deletePermissionRule", true)
	registerGuard("GET", "/api/sys/permission/queryPermissionRule", true)
	registerGuard("GET", "/api/sys/permission/queryDepartPermission", true)
	registerGuard("POST", "/api/sys/permission/saveDepartPermission", true)

	// 角色管理域（P0-4a；add/edit/delete/deleteBatch 服务内含 admin 校验）
	registerGuard("GET", "/api/sys/role/list", true)
	registerGuard("POST", "/api/sys/role/add", true)
	registerGuard("PUT", "/api/sys/role/edit", true)
	registerGuard("POST", "/api/sys/role/edit", true)
	registerGuard("DELETE", "/api/sys/role/delete", true)
	registerGuard("DELETE", "/api/sys/role/deleteBatch", true)
	registerGuard("GET", "/api/sys/role/queryById", true)
	registerGuard("GET", "/api/sys/role/queryall", true)
	registerGuard("GET", "/api/sys/role/queryMySubRole", true)
	registerGuard("GET", "/api/sys/role/checkRoleCode", true)
	registerGuard("GET", "/api/sys/role/queryTreeList", true)
	registerGuard("GET", "/api/sys/role/datarule/:permissionId/:roleId", true)
	registerGuard("POST", "/api/sys/role/datarule", true)

	// 部门管理域（P0-4c）
	registerGuard("GET", "/api/sys/sysDepart/queryTreeList", true)
	registerGuard("GET", "/api/sys/sysDepart/queryMyDeptTreeList", true)
	registerGuard("GET", "/api/sys/sysDepart/queryIdTree", true)
	registerGuard("POST", "/api/sys/sysDepart/add", true)
	registerGuard("PUT", "/api/sys/sysDepart/edit", true)
	registerGuard("DELETE", "/api/sys/sysDepart/delete", true)
	registerGuard("DELETE", "/api/sys/sysDepart/deleteBatch", true)
	registerGuard("GET", "/api/sys/sysDepart/searchBy", true)
	registerGuard("GET", "/api/sys/sysDepart/listAll", true)
	registerGuard("GET", "/api/sys/sysDepart/removeAll", true)

	// 部门角色域（P0-4c）
	registerGuard("GET", "/api/sys/sysDepartRole/list", true)
	registerGuard("POST", "/api/sys/sysDepartRole/add", true)
	registerGuard("PUT", "/api/sys/sysDepartRole/edit", true)
	registerGuard("DELETE", "/api/sys/sysDepartRole/delete", true)
	registerGuard("DELETE", "/api/sys/sysDepartRole/deleteBatch", true)
	registerGuard("GET", "/api/sys/sysDepartRole/queryById", true)
	registerGuard("GET", "/api/sys/sysDepartRole/getDeptRoleList", true)
	registerGuard("POST", "/api/sys/sysDepartRole/deptRoleUserAdd", true)
	registerGuard("GET", "/api/sys/sysDepartRole/getDeptRoleByUserId", true)
	registerGuard("GET", "/api/sys/sysDepartRole/datarule/:permissionId/:departId/:roleId", true)
	registerGuard("POST", "/api/sys/sysDepartRole/datarule", true)

	// 部门权限域（P0-4c）
	registerGuard("GET", "/api/sys/sysDepartPermission/list", true)
	registerGuard("POST", "/api/sys/sysDepartPermission/add", true)
	registerGuard("PUT", "/api/sys/sysDepartPermission/edit", true)
	registerGuard("DELETE", "/api/sys/sysDepartPermission/delete", true)
	registerGuard("DELETE", "/api/sys/sysDepartPermission/deleteBatch", true)
	registerGuard("GET", "/api/sys/sysDepartPermission/queryById", true)
	registerGuard("GET", "/api/sys/sysDepartPermission/datarule/:permissionId/:departId", true)
	registerGuard("POST", "/api/sys/sysDepartPermission/datarule", true)
	registerGuard("GET", "/api/sys/sysDepartPermission/queryDeptRolePermission", true)
	registerGuard("POST", "/api/sys/sysDepartPermission/saveDeptRolePermission", true)
	registerGuard("GET", "/api/sys/sysDepartPermission/queryTreeListForDeptRole", true)

	// 课程域（P1-1）
	registerGuard("GET", "/api/teaching/teachingCourse/mineCourse", true)
	registerGuard("GET", "/api/teaching/teachingCourse/getHomeCourse", false)
	registerGuard("GET", "/api/teaching/teachingCourse/list", true)
	registerGuard("POST", "/api/teaching/teachingCourse/add", true)
	registerGuard("PUT", "/api/teaching/teachingCourse/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingCourse/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingCourse/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingCourse/queryById", true)

	// 门户用户上下文域（P1-1）
	registerGuard("GET", "/api/teaching/user/info", true)
	registerGuard("POST", "/api/teaching/user/info", true)
	registerGuard("GET", "/api/teaching/user/getUserRootDept", true)
	registerGuard("POST", "/api/teaching/user/getUserRootDept", true)
	registerGuard("PUT", "/api/teaching/user/edit", true)
	registerGuard("GET", "/api/teaching/user/appUserList", true)
	registerGuard("PUT", "/api/sys/selectDepart", true)
	registerGuard("GET", "/api/sys/loginfo", true)
	registerGuard("GET", "/api/sys/visitInfo", true)

	// 消息域（P2-7）
	registerGuard("GET", "/api/message/sysMessageTemplate/list", true)
	registerGuard("POST", "/api/message/sysMessageTemplate/add", true)
	registerGuard("PUT", "/api/message/sysMessageTemplate/edit", true)
	registerGuard("DELETE", "/api/message/sysMessageTemplate/delete", true)
	registerGuard("DELETE", "/api/message/sysMessageTemplate/deleteBatch", true)
	registerGuard("GET", "/api/message/sysMessageTemplate/queryById", true)
	registerGuard("POST", "/api/message/sysMessageTemplate/sendMsg", true)
	registerGuard("GET", "/api/message/sysMessage/list", true)
	registerGuard("POST", "/api/message/sysMessage/add", true)
	registerGuard("PUT", "/api/message/sysMessage/edit", true)
	registerGuard("DELETE", "/api/message/sysMessage/delete", true)
	registerGuard("DELETE", "/api/message/sysMessage/deleteBatch", true)
	registerGuard("GET", "/api/message/sysMessage/queryById", true)

	// 消息域（P2-7）
	registerGuard("GET", "/api/message/sysMessageTemplate/list", true)
	registerGuard("POST", "/api/message/sysMessageTemplate/add", true)
	registerGuard("PUT", "/api/message/sysMessageTemplate/edit", true)
	registerGuard("DELETE", "/api/message/sysMessageTemplate/delete", true)
	registerGuard("DELETE", "/api/message/sysMessageTemplate/deleteBatch", true)
	registerGuard("GET", "/api/message/sysMessageTemplate/queryById", true)
	registerGuard("POST", "/api/message/sysMessageTemplate/sendMsg", true)
	registerGuard("GET", "/api/message/sysMessage/list", true)
	registerGuard("POST", "/api/message/sysMessage/add", true)
	registerGuard("PUT", "/api/message/sysMessage/edit", true)
	registerGuard("DELETE", "/api/message/sysMessage/delete", true)
	registerGuard("DELETE", "/api/message/sysMessage/deleteBatch", true)
	registerGuard("GET", "/api/message/sysMessage/queryById", true)
	registerGuard("PUT", "/api/sys/user/appEdit", true)

	// 课程单元域（P1-2）
	registerGuard("GET", "/api/teaching/teachingCourseUnit/mineUnit", true)
	registerGuard("GET", "/api/teaching/teachingCourseUnit/list", true)
	registerGuard("GET", "/api/teaching/teachingCourseUnit/getUnitWorkInfo", true)
	registerGuard("POST", "/api/teaching/teachingCourseUnit/add", true)
	registerGuard("PUT", "/api/teaching/teachingCourseUnit/edit", true)
	registerGuard("PUT", "/api/teaching/teachingCourseUnit/editBatch", true)
	registerGuard("DELETE", "/api/teaching/teachingCourseUnit/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingCourseUnit/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingCourseUnit/queryById", true)

	// 班级课程授权域（P1-3）
	registerGuard("GET", "/api/teaching/teachingCourseDept/list", true)
	registerGuard("POST", "/api/teaching/teachingCourseDept/addOrUpdate", true)
	registerGuard("POST", "/api/teaching/teachingCourseDept/add", true)
	registerGuard("PUT", "/api/teaching/teachingCourseDept/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingCourseDept/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingCourseDept/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingCourseDept/queryById", true)

	// 门户菜单域（P1-3；getUserMenu anon）
	registerGuard("GET", "/api/teaching/menu/list", true)
	registerGuard("GET", "/api/teaching/menu/getSystemMenuList", true)
	registerGuard("GET", "/api/teaching/menu/getSystemSubmenu", true)
	registerGuard("GET", "/api/teaching/menu/getSystemSubmenuBatch", true)
	registerGuard("POST", "/api/teaching/menu/add", true)
	registerGuard("PUT", "/api/teaching/menu/edit", true)
	registerGuard("POST", "/api/teaching/menu/edit", true)
	registerGuard("DELETE", "/api/teaching/menu/delete", true)
	registerGuard("DELETE", "/api/teaching/menu/deleteBatch", true)
	registerGuard("GET", "/api/teaching/menu/queryTreeList", true)
	registerGuard("GET", "/api/teaching/menu/queryListAsync", true)
	registerGuard("GET", "/api/teaching/menu/getUserMenu", false)

	// 文件记录域（P1-4）
	registerGuard("GET", "/api/system/sysFile/list", true)
	registerGuard("POST", "/api/system/sysFile/add", true)
	registerGuard("PUT", "/api/system/sysFile/edit", true)
	registerGuard("DELETE", "/api/system/sysFile/delete", true)
	registerGuard("DELETE", "/api/system/sysFile/deleteByPath", true)
	registerGuard("DELETE", "/api/system/sysFile/deleteBatch", true)
	registerGuard("GET", "/api/system/sysFile/queryById", true)

	// 七牛令牌域（P1-4，需登录）
	registerGuard("GET", "/api/common/qiniu/getToken", true)
	registerGuard("POST", "/api/common/qiniu/getToken", true)
	registerGuard("GET", "/api/common/qiniu/getTokenByKey", true)
	registerGuard("POST", "/api/common/qiniu/getTokenByKey", true)

	// 本地上传（P1-4）
	registerGuard("POST", "/api/sys/common/upload", true)
	// 本地静态预览为 anon（ShiroConfig.java:88），非写操作不强制登记
	registerGuard("GET", "/api/sys/common/static/*path", false)

	// 作业域（P2-1）
	registerGuard("GET", "/api/teaching/teachingWork/mine", true)
	registerGuard("GET", "/api/teaching/teachingWork/mineAdditionalWork", true)
	registerGuard("POST", "/api/teaching/teachingWork/submit", true)
	registerGuard("POST", "/api/teaching/teachingWork/sendWork", true)
	registerGuard("GET", "/api/teaching/teachingWork/list", true)
	registerGuard("POST", "/api/teaching/teachingWork/add", true)
	registerGuard("PUT", "/api/teaching/teachingWork/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingWork/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingWork/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingWork/queryById", true)
	registerGuard("GET", "/api/teaching/teachingWork/queryTeachingWorkCorrectByMainId", true)
	registerGuard("GET", "/api/teaching/teachingWork/queryTeachingWorkCommentByMainId", true)
	registerGuard("POST", "/api/teaching/teachingWork/saveComment", true)
	registerGuard("GET", "/api/teaching/teachingWork/getWorkTags", true)
	registerGuard("GET", "/api/teaching/teachingWork/setWorkTag", true)
	registerGuard("DELETE", "/api/teaching/teachingWork/delWorkTag", true)

	// 作业域 anon（ShiroConfig.java:152-157）
	registerGuard("GET", "/api/teaching/teachingWork/userInfo", false)
	registerGuard("GET", "/api/teaching/teachingWork/studentWorkInfo", false)
	registerGuard("GET", "/api/teaching/teachingWork/greatWork", false)
	registerGuard("GET", "/api/teaching/teachingWork/starWork", false)
	registerGuard("GET", "/api/teaching/teachingWork/leaderboard", false)
	registerGuard("GET", "/api/teaching/teachingWork/getWorkComments", false)

	// 资讯域（P2-2）
	registerGuard("GET", "/api/teaching/teachingNews/list", true)
	registerGuard("POST", "/api/teaching/teachingNews/add", true)
	registerGuard("PUT", "/api/teaching/teachingNews/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingNews/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingNews/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingNews/queryById", true)
	registerGuard("GET", "/api/teaching/teachingNews/newsList", false)
	registerGuard("GET", "/api/teaching/teachingNews/newsDetail", false)

	// 附加作业域（P2-2）
	registerGuard("GET", "/api/teaching/teachingAdditionalWork/getWorkInfo", true)
	registerGuard("GET", "/api/teaching/teachingAdditionalWork/list", true)
	registerGuard("GET", "/api/teaching/teachingAdditionalWork/teacherList", true)
	registerGuard("POST", "/api/teaching/teachingAdditionalWork/add", true)
	registerGuard("PUT", "/api/teaching/teachingAdditionalWork/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingAdditionalWork/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingAdditionalWork/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingAdditionalWork/queryById", true)

	// 班级日志域（P2-3）
	registerGuard("GET", "/api/teaching/teachingDepartDayLog/getReport", true)
	registerGuard("GET", "/api/teaching/teachingDepartDayLog/getReportGroupByDepart", true)
	registerGuard("GET", "/api/teaching/teachingDepartDayLog/getReportGroupByMonth", true)
	registerGuard("GET", "/api/teaching/teachingDepartDayLog/list", true)
	registerGuard("POST", "/api/teaching/teachingDepartDayLog/add", true)
	registerGuard("PUT", "/api/teaching/teachingDepartDayLog/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingDepartDayLog/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingDepartDayLog/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingDepartDayLog/queryById", true)
	registerGuard("GET", "/api/teaching/teachingDepartDayLog/unitViewLog", true)

	// Scratch 素材库域（P2-4）
	registerGuard("GET", "/api/teaching/teachingScratchAssets/list", true)
	registerGuard("POST", "/api/teaching/teachingScratchAssets/add", true)
	registerGuard("PUT", "/api/teaching/teachingScratchAssets/edit", true)
	registerGuard("DELETE", "/api/teaching/teachingScratchAssets/delete", true)
	registerGuard("DELETE", "/api/teaching/teachingScratchAssets/deleteBatch", true)
	registerGuard("GET", "/api/teaching/teachingScratchAssets/queryById", true)
	registerGuard("GET", "/api/teaching/teachingScratchAssets/getScratchAssets", true)

	// Scratch 背包域（P2-4）
	registerGuard("GET", "/api/teaching/scratch/backpack/content", true)
	registerGuard("POST", "/api/teaching/scratch/backpack/save", true)
	registerGuard("DELETE", "/api/teaching/scratch/backpack/delete", true)
	registerGuard("GET", "/api/teaching/scratch/backpack/getFile/*filename", true)

	// 系统通告域（P2-5）
	registerGuard("GET", "/api/sys/annountCement/list", true)
	registerGuard("POST", "/api/sys/annountCement/add", true)
	registerGuard("PUT", "/api/sys/annountCement/edit", true)
	registerGuard("DELETE", "/api/sys/annountCement/delete", true)
	registerGuard("DELETE", "/api/sys/annountCement/deleteBatch", true)
	registerGuard("GET", "/api/sys/annountCement/queryById", true)
	registerGuard("GET", "/api/sys/annountCement/doReleaseData", true)
	registerGuard("GET", "/api/sys/annountCement/doReovkeData", true)
	registerGuard("GET", "/api/sys/annountCement/listByUser", true)
	registerGuard("GET", "/api/sys/annountCement/syncNotic", true)

	// 通告发送记录域（P2-5）
	registerGuard("GET", "/api/sys/sysAnnouncementSend/list", true)
	registerGuard("POST", "/api/sys/sysAnnouncementSend/add", true)
	registerGuard("PUT", "/api/sys/sysAnnouncementSend/edit", true)
	registerGuard("DELETE", "/api/sys/sysAnnouncementSend/delete", true)
	registerGuard("DELETE", "/api/sys/sysAnnouncementSend/deleteBatch", true)
	registerGuard("GET", "/api/sys/sysAnnouncementSend/queryById", true)
	registerGuard("PUT", "/api/sys/sysAnnouncementSend/editByAnntIdAndUserId", true)
	registerGuard("GET", "/api/sys/sysAnnouncementSend/getMyAnnouncementSend", true)
	registerGuard("PUT", "/api/sys/sysAnnouncementSend/readAll", true)
	// 用户辅助域公开读接口（Shiro anon 依据 ShiroConfig.java L82/L84）
	registerGuard("GET", "/api/sys/user/checkOnlyUser", false)
	registerGuard("GET", "/api/sys/user/querySysUser", false)

	// 用户角色分配域（P0-3）
	registerGuard("GET", "/api/sys/user/queryUserRole", true)
	registerGuard("GET", "/api/sys/user/userRoleList", true)
	registerGuard("POST", "/api/sys/user/addSysUserRole", true)
	registerGuard("DELETE", "/api/sys/user/deleteUserRole", true)
	registerGuard("DELETE", "/api/sys/user/deleteUserRoleBatch", true)

	// 用户-部门关联域（P0-3）
	registerGuard("GET", "/api/sys/user/userDepartList", true)
	registerGuard("GET", "/api/sys/user/queryUserByDepId", true)
	registerGuard("GET", "/api/sys/user/departUserList", true)
	registerGuard("GET", "/api/sys/user/queryByOrgCode", true)
	registerGuard("GET", "/api/sys/user/queryByOrgCodeForAddressList", true)
	registerGuard("POST", "/api/sys/user/editSysDepartWithUser", true)
	registerGuard("DELETE", "/api/sys/user/deleteUserInDepart", true)
	registerGuard("DELETE", "/api/sys/user/deleteUserInDepartBatch", true)
	registerGuard("GET", "/api/sys/user/getCurrentUserDeparts", true)

	// 用户辅助域（P0-3）
	registerGuard("GET", "/api/sys/user/queryById", true)
	registerGuard("GET", "/api/sys/user/generateUserId", true)
	registerGuard("GET", "/api/sys/user/queryByIds", true)
	registerGuard("GET", "/api/sys/user/recycleBin", true)
	registerGuard("PUT", "/api/sys/user/putRecycleBin", true)
	registerGuard("DELETE", "/api/sys/user/deleteRecycleBin", true)

	// 注册/找回密码自服务链路（P0-3，ShiroConfig.java L80/L83/L85/L86 anon）
	registerGuard("POST", "/api/sys/sms", false)
	registerGuard("POST", "/api/sys/user/register", false)
	registerGuard("POST", "/api/sys/user/phoneVerification", false)
	registerGuard("GET", "/api/sys/user/passwordChange", false)
	registerGuard("GET", "/api/sys/randomImage/{key}", false)
	registerGuard("POST", "/api/sys/checkCaptcha", false)
	registerGuard("GET", "/api/websocket/{userId}", false)

	// 用户核心 CRUD 域（P0-3）
	registerGuard("GET", "/api/sys/user/list", true)
	registerGuard("POST", "/api/sys/user/add", true)
	registerGuard("PUT", "/api/sys/user/edit", true)
	registerGuard("DELETE", "/api/sys/user/delete", true)
	registerGuard("DELETE", "/api/sys/user/deleteBatch", true)
	registerGuard("PUT", "/api/sys/user/frozenBatch", true)

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
