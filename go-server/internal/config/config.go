// Package config 负责加载与校验服务端运行配置。
//
// 设计要点（相对 Java 版的关键改进）：
//  1. 密钥不落配置文件，全部由环境变量注入，缺失即启动失败（fail-fast）；
//  2. 所有对外地址、跨域白名单必须显式声明，不存在"默认放开"的兜底分支；
//  3. 配置项集中定义，避免 Java 版存在 prod/dev/test 三份 yml 漂移的问题。
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 是服务端的完整运行配置。
type Config struct {
	Server    ServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Security  SecurityConfig
	Storage   StorageConfig
	Logistics LogisticsConfig
}

// ServerConfig HTTP 服务参数。
type ServerConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	// TrustedProxies 为空表示不信任任何代理头，直接使用连接源地址，
	// 避免 Java 版直接采信 X-Forwarded-For 导致的 IP 伪造。
	TrustedProxies []string
}

// MySQLConfig 数据库参数。
type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig 缓存参数。
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// JWTConfig 令牌参数。
type JWTConfig struct {
	// Secret 必须 ≥32 字节，由环境变量注入。
	Secret string
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
	// LoginFailLimit 同一账号+IP 在 LoginFailWindow 内允许的失败次数。
	LoginFailLimit int
	// LoginFailWindow 失败计数窗口。
	LoginFailWindow time.Duration
	// EnableSwagger 仅非生产环境可开启。
	EnableSwagger bool
	// EnablePprof 仅本地调试可开启。
	EnablePprof bool
	// MaxUploadBytes 单文件上传上限。
	MaxUploadBytes int64
}

// StorageConfig 对象存储参数（对应 Java 版 local/qiniu/alioss/minio）。
type StorageConfig struct {
	Driver string
	// 本地存储
	LocalRoot string
	// 公共访问前缀
	PublicBase string
}

// LogisticsConfig 运行环境标识。
type LogisticsConfig struct {
	Env     string
	Version string
}

// Env 读取环境变量，缺失时返回默认值。
func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func envInt(key string, def int) (int, error) {
	raw := env(key, "")
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("环境变量 %s 不是合法整数: %q", key, raw)
	}
	return v, nil
}

func envBool(key string, def bool) (bool, error) {
	raw := env(key, "")
	if raw == "" {
		return def, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("环境变量 %s 不是合法布尔值: %q", key, raw)
	}
	return v, nil
}

// Load 从环境变量装载配置并做完整校验。
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Logistics.Env = env("APP_ENV", "production")
	cfg.Logistics.Version = env("APP_VERSION", "dev")

	cfg.Server.Addr = env("HTTP_ADDR", ":8080")
	cfg.Server.ReadTimeout = 30 * time.Second
	cfg.Server.WriteTimeout = 60 * time.Second
	cfg.Server.ShutdownTimeout = 15 * time.Second
	if raw := env("TRUSTED_PROXIES", ""); raw != "" {
		for _, item := range strings.Split(raw, ",") {
			if v := strings.TrimSpace(item); v != "" {
				cfg.Server.TrustedProxies = append(cfg.Server.TrustedProxies, v)
			}
		}
	}

	cfg.MySQL.DSN = env("MYSQL_DSN", "")
	cfg.MySQL.MaxOpenConns, _ = envInt("MYSQL_MAX_OPEN", 50)
	cfg.MySQL.MaxIdleConns, _ = envInt("MYSQL_MAX_IDLE", 10)
	cfg.MySQL.ConnMaxLifetime = 30 * time.Minute

	cfg.Redis.Addr = env("REDIS_ADDR", "127.0.0.1:6379")
	cfg.Redis.Password = env("REDIS_PASSWORD", "")
	cfg.Redis.DB, _ = envInt("REDIS_DB", 0)
	cfg.Redis.PoolSize, _ = envInt("REDIS_POOL_SIZE", 20)

	cfg.JWT.Secret = env("JWT_SECRET", "")
	cfg.JWT.PrevSecret = env("JWT_PREV_SECRET", "")
	cfg.JWT.Issuer = env("JWT_ISSUER", "teaching-open")
	cfg.JWT.AccessTTL = 2 * time.Hour
	cfg.JWT.RefreshTTL = 7 * 24 * time.Hour

	rawOrigins := env("CORS_ALLOWED_ORIGINS", "")
	if rawOrigins != "" {
		for _, item := range strings.Split(rawOrigins, ",") {
			v := strings.TrimSpace(item)
			if v == "" {
				continue
			}
			cfg.Security.AllowedOrigins = append(cfg.Security.AllowedOrigins, v)
		}
	}
	cfg.Security.LoginFailLimit, _ = envInt("LOGIN_FAIL_LIMIT", 5)
	cfg.Security.LoginFailWindow = 15 * time.Minute
	cfg.Security.EnableSwagger, _ = envBool("ENABLE_SWAGGER", false)
	cfg.Security.EnablePprof, _ = envBool("ENABLE_PPROF", false)
	cfg.Security.MaxUploadBytes = 100 << 20 // 100MB

	cfg.Storage.Driver = env("STORAGE_DRIVER", "local")
	cfg.Storage.LocalRoot = env("STORAGE_LOCAL_ROOT", "./data/upload")
	cfg.Storage.PublicBase = env("STORAGE_PUBLIC_BASE", "/api/sys/common/static")

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate 做启动前的强校验，任一不满足即拒绝启动。
func (c *Config) Validate() error {
	var problems []string

	if c.MySQL.DSN == "" {
		problems = append(problems, "MYSQL_DSN 未配置")
	}
	if len(c.JWT.Secret) < 32 {
		problems = append(problems, fmt.Sprintf("JWT_SECRET 长度不足 32 字节（当前 %d）", len(c.JWT.Secret)))
	}
	for _, origin := range c.Security.AllowedOrigins {
		if origin == "*" {
			problems = append(problems, "CORS_ALLOWED_ORIGINS 不允许使用通配符 *")
		}
	}
	if c.Security.EnableSwagger && c.Logistics.Env == "production" {
		problems = append(problems, "生产环境不允许开启 ENABLE_SWAGGER")
	}
	if c.Security.EnablePprof && c.Logistics.Env == "production" {
		problems = append(problems, "生产环境不允许开启 ENABLE_PPROF")
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
