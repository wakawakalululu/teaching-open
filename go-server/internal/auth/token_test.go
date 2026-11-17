package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wakawakalululu/teaching-open/go-server/internal/config"
)

const testSecret = "0123456789abcdef0123456789abcdef" // 32 字节

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m, err := NewManager(config.JWTConfig{
		Secret:     testSecret,
		Issuer:     "test-issuer",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("构造 Manager 失败: %v", err)
	}
	return m
}

func TestIssueAndParse(t *testing.T) {
	m := newTestManager(t)
	pair, err := m.Issue("u-1", "alice")
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("令牌为空")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("访问令牌与刷新令牌不应相同")
	}

	claims, err := m.Parse(context.Background(), pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("解析访问令牌失败: %v", err)
	}
	if claims.UserID != "u-1" || claims.Username != "alice" {
		t.Fatalf("载荷不符: %+v", claims)
	}
}

// 刷新令牌不能被当作访问令牌使用（防止权限提升）。
func TestRefreshTokenRejectedAsAccess(t *testing.T) {
	m := newTestManager(t)
	pair, _ := m.Issue("u-1", "alice")

	if _, err := m.Parse(context.Background(), pair.RefreshToken, TokenTypeAccess); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("期望拒绝刷新令牌作为访问令牌，实际 err=%v", err)
	}
}

// 篡改载荷后签名应失效。
func TestTamperedTokenRejected(t *testing.T) {
	m := newTestManager(t)
	pair, _ := m.Issue("u-1", "alice")

	tampered := pair.AccessToken[:len(pair.AccessToken)-4] + "AAAA"
	if _, err := m.Parse(context.Background(), tampered, TokenTypeAccess); err == nil {
		t.Fatal("篡改后的令牌应被拒绝")
	}
}

// 其他密钥签发的令牌应被拒绝。
func TestWrongSecretRejected(t *testing.T) {
	m := newTestManager(t)
	other, err := NewManager(config.JWTConfig{
		Secret:    "ffffffffffffffffffffffffffffffff",
		Issuer:    "test-issuer",
		AccessTTL: time.Minute,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	pair, _ := other.Issue("u-1", "alice")

	if _, err := m.Parse(context.Background(), pair.AccessToken, TokenTypeAccess); err == nil {
		t.Fatal("不同密钥签发的令牌应被拒绝")
	}
}

// 过短密钥应在构造时被拒绝。
func TestShortSecretRejected(t *testing.T) {
	if _, err := NewManager(config.JWTConfig{Secret: "short", Issuer: "x"}, nil); err == nil {
		t.Fatal("短密钥应被拒绝")
	}
}

type memRevoker struct {
	revoked map[string]bool
	failNext bool
}

func (r *memRevoker) Revoke(_ context.Context, id string, _ time.Duration) error {
	if r.revoked == nil {
		r.revoked = map[string]bool{}
	}
	r.revoked[id] = true
	return nil
}

func (r *memRevoker) IsRevoked(_ context.Context, id string) (bool, error) {
	if r.failNext {
		return false, errors.New("存储不可用")
	}
	return r.revoked[id], nil
}

// 已吊销令牌必须失效。
func TestRevokedTokenRejected(t *testing.T) {
	rev := &memRevoker{}
	m, err := NewManager(config.JWTConfig{
		Secret: testSecret, Issuer: "test-issuer",
		AccessTTL: time.Minute, RefreshTTL: time.Hour,
	}, rev)
	if err != nil {
		t.Fatal(err)
	}
	pair, _ := m.Issue("u-1", "alice")

	if _, err := m.Parse(context.Background(), pair.AccessToken, TokenTypeAccess); err != nil {
		t.Fatalf("首次解析应成功: %v", err)
	}
	if err := m.Revoke(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("吊销失败: %v", err)
	}
	if _, err := m.Parse(context.Background(), pair.AccessToken, TokenTypeAccess); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("期望 ErrTokenRevoked，实际 %v", err)
	}
}

// 吊销存储故障时应 fail-closed，不放行任何令牌。
func TestRevokerFailureIsFailClosed(t *testing.T) {
	rev := &memRevoker{failNext: true}
	m, _ := NewManager(config.JWTConfig{
		Secret: testSecret, Issuer: "test-issuer",
		AccessTTL: time.Minute, RefreshTTL: time.Hour,
	}, rev)
	pair, _ := m.Issue("u-1", "alice")

	if _, err := m.Parse(context.Background(), pair.AccessToken, TokenTypeAccess); err == nil {
		t.Fatal("吊销状态不可判定时应拒绝令牌")
	}
}

// 密钥轮换：旧密钥签发的令牌在轮换窗口内仍可通过验证。
func TestKeyRotation(t *testing.T) {
	oldSecret := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	oldMgr, _ := NewManager(config.JWTConfig{
		Secret: oldSecret, Issuer: "test-issuer",
		AccessTTL: time.Minute, RefreshTTL: time.Hour,
	}, nil)
	pair, _ := oldMgr.Issue("u-1", "alice")

	newMgr, err := NewManager(config.JWTConfig{
		Secret:     testSecret,
		PrevSecret: oldSecret,
		Issuer:     "test-issuer",
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newMgr.Parse(context.Background(), pair.AccessToken, TokenTypeAccess); err != nil {
		t.Fatalf("轮换窗口内旧令牌应可用: %v", err)
	}
}
