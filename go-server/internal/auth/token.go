// Package auth 实现令牌签发、校验与登录失败限流。
//
// 与 Java 版的关键差异：
//  1. Java 版用「用户密码的 MD5」作为 HMAC 密钥（JwtUtil.sign(username, syspassword)），
//     密钥强度取决于密码且可被数据库泄露直接推导。Go 版改为独立的高熵服务端密钥，
//     通过环境变量注入，并支持密钥轮换（PrevSecret 只验不签）。
//  2. Java 版 token 有效期 30 小时，且 Redis 缓存有效期拉到 2 倍，实际在线窗口可达 60 小时。
//     Go 版拆分为短效 AccessToken + 长效 RefreshToken，并支持主动吊销。
//  3. Java 版的限流维度是「用户名 + IP」，Go 版沿用该维度但计数与判定原子化。
package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/wakawakalululu/teaching-open/go-server/internal/config"
)

// 令牌类型。
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	// ErrInvalidToken 令牌无效（签名错误、格式错误、类型不符）。
	ErrInvalidToken = errors.New("token 无效")
	// ErrExpiredToken 令牌已过期。
	ErrExpiredToken = errors.New("token 已过期")
	// ErrTokenRevoked 令牌已被吊销。
	ErrTokenRevoked = errors.New("token 已被吊销")
)

// Claims 是本服务的 JWT 载荷。
type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"uname"`
	// TokenType 区分访问令牌与刷新令牌，防止刷新令牌被直接当作访问凭据使用。
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

// Revoker 抽象令牌吊销存储，生产实现基于 Redis。
type Revoker interface {
	// Revoke 将令牌 ID 加入黑名单，ttl 为剩余有效期。
	Revoke(ctx context.Context, tokenID string, ttl time.Duration) error
	// IsRevoked 判断令牌 ID 是否已被吊销。
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

// Manager 负责签发与解析令牌。
type Manager struct {
	secret     []byte
	prevSecret []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	revoker    Revoker
}

// NewManager 构造令牌管理器。
func NewManager(cfg config.JWTConfig, revoker Revoker) (*Manager, error) {
	if len(cfg.Secret) < 32 {
		return nil, errors.New("JWT 密钥长度不足 32 字节")
	}
	return &Manager{
		secret:     []byte(cfg.Secret),
		prevSecret: []byte(cfg.PrevSecret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		revoker:    revoker,
	}, nil
}

// TokenPair 是一次登录签发的令牌对。
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Issue 为指定用户签发访问令牌 + 刷新令牌。
func (m *Manager) Issue(userID, username string) (*TokenPair, error) {
	access, err := m.sign(userID, username, TokenTypeAccess, m.accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := m.sign(userID, username, TokenTypeRefresh, m.refreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(m.accessTTL.Seconds()),
	}, nil
}

func (m *Manager) sign(userID, username, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("签发令牌失败: %w", err)
	}
	return signed, nil
}

// Parse 校验令牌并返回载荷；expectedType 用于强制校验令牌类型。
func (m *Manager) Parse(ctx context.Context, raw, expectedType string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		// 主密钥验签，失败时回退到轮换前的旧密钥（只验不签）
		return m.secret, nil
	}

	token, err := parser.ParseWithClaims(raw, claims, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		// 主密钥失败时尝试旧密钥，支持平滑轮换
		if len(m.prevSecret) >= 32 {
			claimsPrev := &Claims{}
			prevToken, prevErr := parser.ParseWithClaims(raw, claimsPrev, func(t *jwt.Token) (interface{}, error) {
				return m.prevSecret, nil
			})
			if prevErr == nil && prevToken.Valid {
				token, claims = prevToken, claimsPrev
				err = nil
			}
		}
		if err != nil {
			return nil, ErrInvalidToken
		}
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}
	if claims.ID == "" || claims.UserID == "" {
		return nil, ErrInvalidToken
	}

	if m.revoker != nil {
		revoked, revokeErr := m.revoker.IsRevoked(ctx, claims.ID)
		if revokeErr != nil {
			// 吊销状态不可判定时按已吊销处理（fail-closed），避免故障期间放行已注销令牌
			return nil, fmt.Errorf("%w: 无法确认吊销状态", ErrInvalidToken)
		}
		if revoked {
			return nil, ErrTokenRevoked
		}
	}
	return claims, nil
}

// Revoke 主动吊销令牌（用于登出、改密、强制下线）。
func (m *Manager) Revoke(ctx context.Context, raw string) error {
	if m.revoker == nil {
		return errors.New("未配置吊销存储")
	}
	claims := &Claims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if _, err := parser.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	}); err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return ErrInvalidToken
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	return m.revoker.Revoke(ctx, claims.ID, ttl)
}

// PasswordVerifier 密码校验抽象，便于测试与迁移期双算法支持。
type PasswordVerifier interface {
	// Verify 校验明文密码与存储散列是否匹配。
	Verify(plain, stored, salt string) (bool, error)
}

// LegacyPBEVerifier 兼容 Java 版 PasswordUtil（PBEWithMD5AndDES）的历史散列。
//
// 说明：该算法强度不足，仅用于迁移期让存量用户可登录；登录成功后应立即
// 用新的散列算法重写密码字段（rehash-on-login），存量散列清零后可废弃本实现。
type LegacyPBEVerifier struct {
	// Salt 对应 Java 版 PasswordUtil.Salt（固定盐）
	Salt string
}

// ConstantTimeEqual 常量时间比较，避免因比较短路泄露信息。
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
