package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wakawakalululu/teaching-open/go-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/go-server/internal/config"
)

// buildTestServer 构造一个仅含路由骨架、无数据库依赖的测试服务。
func buildTestServer(t *testing.T, origins []string) (*Server, *auth.Manager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mgr, err := auth.NewManager(config.JWTConfig{
		Secret:     "0123456789abcdef0123456789abcdef",
		Issuer:     "test",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Logistics.Env = "test"
	cfg.Security.AllowedOrigins = origins
	cfg.Server.TrustedProxies = nil

	guardedRoutes = nil
	s := &Server{cfg: cfg, logger: testLogger(), tokens: mgr}
	s.engine = s.buildEngine()
	return s, mgr
}

// 健康检查应公开可访问且不泄露内部信息。
func TestHealthEndpoint(t *testing.T) {
	s, _ := buildTestServer(t, nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/actuator/health", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "UP") {
		t.Fatalf("响应不符: %s", body)
	}
	// 不得暴露版本号、数据源等细节
	for _, leak := range []string{"dev", "mysql", "dsn", "version"} {
		if strings.Contains(strings.ToLower(body), leak) {
			t.Fatalf("健康检查泄露内部信息 %q: %s", leak, body)
		}
	}
}

// 未认证访问受保护接口应 401。
func TestProtectedEndpointRequiresAuth(t *testing.T) {
	s, _ := buildTestServer(t, nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/sys/user/info", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际 %d", w.Code)
	}
}

// 携带有效令牌访问受保护接口应 200。
func TestProtectedEndpointWithToken(t *testing.T) {
	s, mgr := buildTestServer(t, nil)
	pair, _ := mgr.Issue("u-1", "alice")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sys/user/info", nil)
	req.Header.Set("X-Access-Token", pair.AccessToken)
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d body=%s", w.Code, w.Body.String())
	}
	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Success {
		t.Fatalf("响应 success 应为 true: %s", w.Body.String())
	}
}

// 登录接口参数缺失应 400，且不因内部错误返回 500。
func TestLoginValidation(t *testing.T) {
	s, _ := buildTestServer(t, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sys/login", strings.NewReader(`{"username":"a"}`))
	req.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际 %d", w.Code)
	}
}

// 超长输入应被前置拒绝。
func TestLoginRejectsOversizedInput(t *testing.T) {
	s, _ := buildTestServer(t, nil)
	payload := `{"username":"` + strings.Repeat("a", 200) + `","password":"x"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sys/login", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望 400，实际 %d", w.Code)
	}
}

// 每个响应都应带安全头与请求追踪号。
func TestResponseSecurityHeaders(t *testing.T) {
	s, _ := buildTestServer(t, []string{"https://ok.example.com"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/actuator/health", nil)
	req.Header.Set("Origin", "https://ok.example.com")
	s.Handler().ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("缺少 X-Content-Type-Options")
	}
	if w.Header().Get("X-Request-Id") == "" {
		t.Fatal("缺少 X-Request-Id")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://ok.example.com" {
		t.Fatal("白名单来源未获 CORS 头")
	}
}

// 非白名单来源不得获得 CORS 授权。
func TestCORSNotGrantedForUnknownOrigin(t *testing.T) {
	s, _ := buildTestServer(t, []string{"https://ok.example.com"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/actuator/health", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	s.Handler().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("非白名单来源不应获得授权，实际 %q", got)
	}
}

// 登录响应不应包含 password / salt 字段。
func TestLoginResponseHasNoSecrets(t *testing.T) {
	// 该断言针对 publicUser 的输出结构，直接校验字段集合
	u := testUser()
	out := publicUser(u, []string{"admin"})
	for _, banned := range []string{"password", "salt"} {
		if _, exists := out[banned]; exists {
			t.Fatalf("响应包含敏感字段 %q", banned)
		}
	}
}
