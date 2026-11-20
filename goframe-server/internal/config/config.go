// Package config 负责加载与校验服务端运行配置。
//
// 相对 Java 版的关键改进：
//  1. 密钥不落配置文件，全部由环境变量注入，缺失即启动失败（fail-fast）；
//  2. 跨域白名单必须显式声明，禁止通配符；
//  3. 配置项集中定义，避免 prod/dev/test 多份 yml 漂移。
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/genv"
)

// Config 服务端完整运行配置（业务侧扩展项）。
// 网络/数据库/日志等框架配置由 GoFrame 的 config.yaml 与 manifest 统一承载。
type Config struct {
	JWT       JWTConfig
	Security  SecurityConfig
	Storage   StorageConfig
	Site      SiteConfig
	Logistics LogisticsConfig
}

// JWTConfig 令牌参数。
type JWTConfig struct {
	// Secret 必须 ≥32 字节，由环境变量 JWT_SECRET 注入。
	Secret string
	// Issuer 签发方标识。
	Issuer string
	// AccessTTL 访问令牌有效期。
	AccessTTL time.Duration
	// RefreshTTL 刷新令牌有效期。
	RefreshTTL time.Duration
	// PrevSecret 用于密钥轮换期间的旧密钥验签（只验不签）。
	PrevSecret string
}

// SecurityConfig 安全相关开关。
type SecurityConfig struct {
	// AllowedOrigins 跨域白名单，禁止 "*"。
	AllowedOrigins []string
	// LoginFailLimit 同一账号+IP 在窗口内的失败次数上限。
	LoginFailLimit int
	// LoginFailWindow 失败计数窗口。
	LoginFailWindow time.Duration
	// MaxUploadBytes 单文件上传上限。
	MaxUploadBytes int64
	// FeishuWebhook 飞书机器人告警地址（为空则禁用告警）。
	FeishuWebhook string
	// RedisEnabled 是否启用 Redis（限流 + 令牌吊销）。
	// 关闭时这两项能力自动降级：限流 fail-open、吊销 fail-open。
	RedisEnabled bool
}

// StorageConfig 对象存储参数。
type StorageConfig struct {
	Driver     string
	LocalRoot  string
	PublicBase string
}

// SiteConfig 前端站点展示参数（/sys/config/getCurrentConfig 契约字段，
// 来源对齐 Java application.yml 的 jeecg.uploadType/jeecg.path.staticDomain 等配置）。
type SiteConfig struct {
	// UploadType 上传方式（local/qiniu/...）。
	UploadType string
	// StaticDomain 静态资源域名。
	StaticDomain string
	// FilePreview 文件预览方式（空或 ow365）。
	FilePreview string
	// OwID ow365 预览应用 id（FilePreview=ow365 时输出到前端）。
	OwID string
}

// LogisticsConfig 运行环境标识。
type LogisticsConfig struct {
	Env     string
	Version string
}

func env(key, def string) string {
	v := genv.Get(key)
	if v == nil || strings.TrimSpace(v.String()) == "" {
		return def
	}
	return v.String()
}

// Load 从环境变量装载业务配置并做强校验。
// 数据库/HTTP 等框架配置由 gcfg 自动加载 manifest/config/config.yaml。
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Logistics.Env = env("APP_ENV", "production")
	cfg.Logistics.Version = env("APP_VERSION", "dev")

	cfg.JWT.Secret = env("JWT_SECRET", "")
	cfg.JWT.PrevSecret = env("JWT_PREV_SECRET", "")
	cfg.JWT.Issuer = env("JWT_ISSUER", "teaching-open")
	cfg.JWT.AccessTTL = 2 * time.Hour
	cfg.JWT.RefreshTTL = 7 * 24 * time.Hour

	rawOrigins := env("CORS_ALLOWED_ORIGINS", "")
	if rawOrigins != "" {
		for _, item := range strings.Split(rawOrigins, ",") {
			if v := strings.TrimSpace(item); v != "" {
				cfg.Security.AllowedOrigins = append(cfg.Security.AllowedOrigins, v)
			}
		}
	}
	cfg.Security.LoginFailLimit = 5
	cfg.Security.LoginFailWindow = 15 * time.Minute
	cfg.Security.MaxUploadBytes = 100 << 20 // 100MB
	cfg.Security.FeishuWebhook = env("FEISHU_WEBHOOK_URL", "")
	// 默认关闭 Redis：无 Redis 时登录限流与令牌吊销走 fail-open 降级，
	// 避免开发/演示环境受保护接口全部 401。生产环境应显式设置 REDIS_ENABLED=true。
	cfg.Security.RedisEnabled = env("REDIS_ENABLED", "false") == "true"

	cfg.Storage.Driver = env("STORAGE_DRIVER", "local")
	cfg.Storage.LocalRoot = env("STORAGE_LOCAL_ROOT", "./data/upload")
	cfg.Storage.PublicBase = env("STORAGE_PUBLIC_BASE", "/api/sys/common/static")

	cfg.Site.UploadType = env("UPLOAD_TYPE", "local")
	cfg.Site.StaticDomain = env("STATIC_DOMAIN", "")
	cfg.Site.FilePreview = env("FILE_PREVIEW", "")
	cfg.Site.OwID = env("OW365_ID", "")

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate 启动前强校验，任一不满足即拒绝启动。
func (c *Config) Validate() error {
	var problems []string

	if len(c.JWT.Secret) < 32 {
		problems = append(problems, fmt.Sprintf("JWT_SECRET 长度不足 32 字节（当前 %d）", len(c.JWT.Secret)))
	}
	for _, origin := range c.Security.AllowedOrigins {
		if origin == "*" {
			problems = append(problems, "CORS_ALLOWED_ORIGINS 不允许使用通配符 *")
		}
	}
	if c.Security.LoginFailLimit < 1 {
		problems = append(problems, "LOGIN_FAIL_LIMIT 必须 ≥ 1")
	}
	if c.Storage.Driver != "local" {
		problems = append(problems, fmt.Sprintf("STORAGE_DRIVER=%s 暂未实现，当前仅支持 local", c.Storage.Driver))
	}

	if len(problems) > 0 {
		return errors.New("配置校验失败:\n  - " + strings.Join(problems, "\n  - "))
	}
	return nil
}

// Ctx 返回一个带默认上下文的快捷入口（用于无需请求上下文的场景）。
var Ctx = gctx.GetInitCtx
