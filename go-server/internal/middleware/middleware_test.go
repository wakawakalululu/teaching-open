package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wakawakalululu/teaching-open/go-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/go-server/internal/config"
)

func newTestTokens(t *testing.T) *auth.Manager {
	t.Helper()
	m, err := auth.NewManager(config.JWTConfig{
		Secret:     "0123456789abcdef0123456789abcdef",
		Issuer:     "test",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func testRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.New(slog.NewTextHandler(io_Discard{}, nil))
	m := newTestTokens(t)
	g := r.Group("")
	g.Use(RequireAuth(m, logger))
	g.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"uid": c.GetString(CtxKeyUserID)})
	})
	return r
}

// io_Discard 丢弃日志输出，避免测试噪声。
type io_Discard struct{}

func (io_Discard) Write(p []byte) (int, error) { return len(p), nil }

// 无令牌应返回 401。
func TestRequireAuthNoToken(t *testing.T) {
	r := testRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际 %d", w.Code)
	}
	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["success"] != false {
		t.Fatalf("响应体不符: %s", w.Body.String())
	}
}

// 非法令牌应返回 401。
func TestRequireAuthBadToken(t *testing.T) {
	r := testRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Access-Token", "not-a-jwt")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际 %d", w.Code)
	}
}

// 有效令牌应放行，并注入用户上下文。
func TestRequireAuthValidToken(t *testing.T) {
	r := testRouter(t)
	m := newTestTokens(t)
	pair, _ := m.Issue("u-9", "bob")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Access-Token", pair.AccessToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "u-9") {
		t.Fatalf("未注入用户上下文: %s", w.Body.String())
	}
}

// Authorization: Bearer 也应被接受（前端渐进迁移用）。
func TestRequireAuthBearerHeader(t *testing.T) {
	r := testRouter(t)
	m := newTestTokens(t)
	pair, _ := m.Issue("u-9", "bob")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d", w.Code)
	}
}

// CORS：白名单内的 Origin 才能获得 CORS 头。
func TestCORSAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"https://ok.example.com"}))
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	// 白名单内
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://ok.example.com")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://ok.example.com" {
		t.Fatalf("白名单来源应放行，实际 Allow-Origin=%q", got)
	}

	// 白名单外：不得回显
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	req2.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("非白名单来源不应回显，实际 %q", got)
	}
}

// CORS：非白名单的预检请求应被拒绝。
func TestCORSPreflightRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"https://ok.example.com"}))
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("期望 403，实际 %d", w.Code)
	}
}

// 安全响应头必须存在。
func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	expect := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "SAMEORIGIN",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for k, v := range expect {
		if got := w.Header().Get(k); got != v {
			t.Fatalf("%s 期望 %q，实际 %q", k, v, got)
		}
	}
}

// panic 不应泄露堆栈，且返回 500 + 追踪号。
func TestRecoverHidesPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io_Discard{}, nil))
	r := gin.New()
	r.Use(RequestID())
	r.Use(Recover(logger))
	r.GET("/boom", func(c *gin.Context) { panic("secret internal detail") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("期望 500，实际 %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "secret internal detail") {
		t.Fatalf("响应体泄露了 panic 内容: %s", w.Body.String())
	}
}
