package httpapi

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// maskPhone / maskEmail 脱敏行为。
func TestMaskPhone(t *testing.T) {
	cases := []struct{ in, want string }{
		{"13812345678", "138****5678"},
		{"1234567", "123****4567"},
		{"123456", "******"},
		{"", ""},
	}
	for _, c := range cases {
		if got := maskPhone(c.in); got != c.want {
			t.Fatalf("maskPhone(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct{ in, want string }{
		{"alice@example.com", "a***@example.com"},
		{"a@b.com", "***"},
		{"", "***"},
	}
	for _, c := range cases {
		if got := maskEmail(c.in); got != c.want {
			t.Fatalf("maskEmail(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

// 路由守卫：写路由未声明认证且不在白名单时必须 panic（拒绝启动）。
func TestVerifyRouteGuardsRejectsUnguardedWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	guardedRoutes = nil
	defer func() { guardedRoutes = nil }()

	s := &Server{logger: testLogger()}
	g := gin.New().Group("/api")
	// 故意注册一条未声明认证的写路由
	s.register(g, "POST", "/admin/danger", false, func(c *gin.Context) {})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("未声明认证的写路由应导致启动失败")
		}
	}()
	s.verifyRouteGuards()
}

// 白名单内的公开写路由应通过。
func TestVerifyRouteGuardsAllowsPublicLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	guardedRoutes = nil
	defer func() { guardedRoutes = nil }()

	s := &Server{logger: testLogger()}
	g := gin.New().Group("/api")
	s.register(g, "POST", "/sys/login", false, func(c *gin.Context) {})
	s.register(g, "POST", "/sys/logout", true, func(c *gin.Context) {})

	// 不应 panic
	s.verifyRouteGuards()
}

// 空路由表应拒绝启动（防止误配置导致无保护启动）。
func TestVerifyRouteGuardsRejectsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	guardedRoutes = nil
	defer func() { guardedRoutes = nil }()

	s := &Server{logger: testLogger()}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("空路由表应导致启动失败")
		}
	}()
	s.verifyRouteGuards()
}
