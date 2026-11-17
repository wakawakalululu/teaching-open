package auth

import (
	"context"
	"testing"
	"time"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/config"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	cfg := config.JWTConfig{
		Secret:     "0123456789abcdef0123456789abcdef", // 32 字节
		Issuer:     "teaching-open",
		AccessTTL:  2 * time.Hour,
		RefreshTTL: 7 * 24 * time.Hour,
	}
	m, err := NewManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewManager 失败: %v", err)
	}
	return m
}

// TestIssueAndParse 签发后能正常解析。
func TestIssueAndParse(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	pair, err := m.Issue("u1", "alice")
	if err != nil {
		t.Fatalf("Issue 失败: %v", err)
	}
	claims, err := m.Parse(ctx, pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}
	if claims.UserID != "u1" || claims.Username != "alice" {
		t.Errorf("载荷不符: %+v", claims)
	}
}

// TestRefreshTokenNotUsableAsAccess 刷新令牌不能当访问令牌用。
func TestRefreshTokenNotUsableAsAccess(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	pair, _ := m.Issue("u1", "alice")
	if _, err := m.Parse(ctx, pair.RefreshToken, TokenTypeAccess); err == nil {
		t.Error("刷新令牌不应通过访问令牌类型校验")
	}
}

// TestTamperedTokenRejected 篡改令牌应被拒绝。
func TestTamperedTokenRejected(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	pair, _ := m.Issue("u1", "alice")
	tampered := pair.AccessToken + "x"
	if _, err := m.Parse(ctx, tampered, TokenTypeAccess); err == nil {
		t.Error("篡改令牌应被拒绝")
	}
}

// TestWrongSecretRejected 错误密钥应被拒绝。
func TestWrongSecretRejected(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	pair, _ := m.Issue("u1", "alice")

	wrong := &Manager{secret: []byte("fedcba9876543210fedcba9876543210"), issuer: "teaching-open"}
	if _, err := wrong.Parse(ctx, pair.AccessToken, TokenTypeAccess); err == nil {
		t.Error("错误密钥应被拒绝")
	}
}

// TestShortSecretRejected 短密钥应拒绝构造。
func TestShortSecretRejected(t *testing.T) {
	cfg := config.JWTConfig{Secret: "short"}
	if _, err := NewManager(cfg, nil); err == nil {
		t.Error("短密钥应拒绝构造")
	}
}

// TestRevokerRevoked 吊销后令牌不可用。
func TestRevokerRevoked(t *testing.T) {
	ctx := context.Background()
	var revoked bool
	revoker := &memRevoker{revoked: &revoked}
	cfg := config.JWTConfig{
		Secret:     "0123456789abcdef0123456789abcdef",
		Issuer:     "teaching-open",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}
	m, err := NewManager(cfg, revoker)
	if err != nil {
		t.Fatalf("NewManager 失败: %v", err)
	}
	pair, _ := m.Issue("u1", "alice")
	revoked = true
	if _, err := m.Parse(ctx, pair.AccessToken, TokenTypeAccess); err != ErrTokenRevoked {
		t.Errorf("吊销后应返回 ErrTokenRevoked，得到: %v", err)
	}
}

type memRevoker struct {
	revoked *bool
}

func (r *memRevoker) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	*r.revoked = true
	return nil
}

func (r *memRevoker) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	return *r.revoked, nil
}
