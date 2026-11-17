package config

import (
	"strings"
	"testing"

	"github.com/gogf/gf/v2/os/genv"
)

func TestValidateRequiresSecret(t *testing.T) {
	genv.Set("JWT_SECRET", "")
	genv.Set("JWT_PREV_SECRET", "")
	genv.Set("CORS_ALLOWED_ORIGINS", "")
	defer func() {
		genv.Remove("JWT_SECRET")
		genv.Remove("JWT_PREV_SECRET")
		genv.Remove("CORS_ALLOWED_ORIGINS")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("缺密钥应拒绝启动")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("错误信息应包含 JWT_SECRET，得到: %v", err)
	}
}

func TestValidateRejectsWildcardOrigin(t *testing.T) {
	genv.Set("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	genv.Set("CORS_ALLOWED_ORIGINS", "*")
	defer func() {
		genv.Remove("JWT_SECRET")
		genv.Remove("CORS_ALLOWED_ORIGINS")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("通配符 Origin 应拒绝启动")
	}
	if !strings.Contains(err.Error(), "通配符") {
		t.Errorf("错误信息应包含通配符提示，得到: %v", err)
	}
}

func TestValidateOK(t *testing.T) {
	genv.Set("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	genv.Set("CORS_ALLOWED_ORIGINS", "https://example.com")
	defer func() {
		genv.Remove("JWT_SECRET")
		genv.Remove("CORS_ALLOWED_ORIGINS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("合法配置应通过，得到: %v", err)
	}
	if cfg.JWT.Secret != "0123456789abcdef0123456789abcdef" {
		t.Error("密钥应正确装载")
	}
	if len(cfg.Security.AllowedOrigins) != 1 {
		t.Error("应有一个白名单 Origin")
	}
}
