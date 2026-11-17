package config

import (
	"strings"
	"testing"
)

// 基线的合法环境变量。
func baseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYSQL_DSN", "user:pass@tcp(127.0.0.1:3306)/teachingopen")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("APP_ENV", "production")
}

func TestLoadValid(t *testing.T) {
	baseEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	if cfg.Server.Addr != ":8080" {
		t.Fatalf("默认监听地址错误: %s", cfg.Server.Addr)
	}
	if cfg.Security.LoginFailLimit != 5 {
		t.Fatalf("默认失败上限错误: %d", cfg.Security.LoginFailLimit)
	}
}

// 缺少 JWT 密钥必须拒绝启动。
func TestMissingJWTSecretRejected(t *testing.T) {
	baseEnv(t)
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("缺少 JWT_SECRET 应拒绝启动")
	} else if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("错误信息应指明 JWT_SECRET，实际: %v", err)
	}
}

// 短密钥必须拒绝。
func TestShortJWTSecretRejected(t *testing.T) {
	baseEnv(t)
	t.Setenv("JWT_SECRET", "tooshort")
	if _, err := Load(); err == nil {
		t.Fatal("短 JWT_SECRET 应拒绝启动")
	}
}

// 缺少 DSN 必须拒绝。
func TestMissingDSNRejected(t *testing.T) {
	baseEnv(t)
	t.Setenv("MYSQL_DSN", "")
	if _, err := Load(); err == nil {
		t.Fatal("缺少 MYSQL_DSN 应拒绝启动")
	}
}

// 生产环境禁止开启 Swagger。
func TestSwaggerForbiddenInProd(t *testing.T) {
	baseEnv(t)
	t.Setenv("ENABLE_SWAGGER", "true")
	if _, err := Load(); err == nil {
		t.Fatal("生产环境开启 Swagger 应拒绝启动")
	}
}

// 非生产环境允许开启 Swagger。
func TestSwaggerAllowedInDev(t *testing.T) {
	baseEnv(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("ENABLE_SWAGGER", "true")
	if _, err := Load(); err != nil {
		t.Fatalf("开发环境应允许 Swagger: %v", err)
	}
}

// CORS 通配符必须被拒绝。
func TestWildcardCORSRejected(t *testing.T) {
	baseEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example.com,*")
	if _, err := Load(); err == nil {
		t.Fatal("CORS 通配符应被拒绝")
	}
}

// 生产环境禁止 pprof。
func TestPprofForbiddenInProd(t *testing.T) {
	baseEnv(t)
	t.Setenv("ENABLE_PPROF", "true")
	if _, err := Load(); err == nil {
		t.Fatal("生产环境开启 pprof 应拒绝启动")
	}
}
