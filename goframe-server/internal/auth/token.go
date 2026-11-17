// Package auth 实现令牌签发、校验与登录失败限流。
//
// 与 Java 版的关键差异：
//  1. 独立高熵服务端密钥（环境变量注入），而非「用户密码的 MD5」；
//  2. 短效 AccessToken + 长效 RefreshToken，支持主动吊销；
//  3. 限流维度「用户名 + IP」，计数与判定原子化。
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	glog "github.com/gogf/gf/v2/os/glog"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/config"
)

// 令牌类型。
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	// ErrInvalidToken 令牌无效。
	ErrInvalidToken = errors.New("token 无效")
	// ErrExpiredToken 令牌已过期。
	ErrExpiredToken = errors.New("token 已过期")
	// ErrTokenRevoked 令牌已被吊销。
	ErrTokenRevoked = errors.New("token 已被吊销")
)

// Claims 是本服务的 JWT 载荷。
type Claims struct {
	UserID    string `json:"uid"`
	Username  string `json:"uname"`
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
	// revokerFailOpen 吊销存储不可用时是否放行令牌。
	// 生产环境应保持 false（fail-closed）；无 Redis 的开发/演示环境设为 true。
	revokerFailOpen bool
}

// Option 构造选项。
type Option func(*Manager)

// WithRevokerFailOpen 设置吊销存储故障时的放行策略。
func WithRevokerFailOpen(open bool) Option {
	return func(m *Manager) { m.revokerFailOpen = open }
}

// NewManager 构造令牌管理器。
func NewManager(cfg config.JWTConfig, revoker Revoker, opts ...Option) (*Manager, error) {
	if len(cfg.Secret) < 32 {
		return nil, errors.New("JWT 密钥长度不足 32 字节")
	}
	m := &Manager{
		secret:     []byte(cfg.Secret),
		prevSecret: []byte(cfg.PrevSecret),
		issuer:     cfg.Issuer,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		revoker:    revoker,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m, nil
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

	token, err := parser.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
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
			// 吊销存储不可用时的策略由 Manager 配置决定：
			//   - fail-closed（默认，生产推荐）：拒绝令牌，避免故障期放行已注销令牌；
			//   - fail-open（无 Redis 的开发/演示环境）：放行，否则受保护接口全部 401。
			if m.revokerFailOpen {
				glog.Warningf(ctx, "吊销状态查询失败，按未吊销处理（fail-open）: %s", revokeErr.Error())
			} else {
				return nil, fmt.Errorf("%w: 无法确认吊销状态", ErrInvalidToken)
			}
		} else if revoked {
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
