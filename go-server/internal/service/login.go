// Package service 承载业务用例。
package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/wakawakalululu/teaching-open/go-server/internal/auth"
	"github.com/wakawakalululu/teaching-open/go-server/internal/passwd"
	"github.com/wakawakalululu/teaching-open/go-server/internal/store"
)

// 登录错误（对外统一话术，不区分账号不存在与密码错误）。
var (
	// ErrInvalidCredential 账号或密码错误。
	ErrInvalidCredential = errors.New("用户名或密码错误")
	// ErrAccountLocked 账号被锁定。
	ErrAccountLocked = errors.New("密码错误次数过多，账号已临时锁定，请稍后再试")
	// ErrAccountDisabled 账号被禁用/冻结。
	ErrAccountDisabled = errors.New("账号已被锁定，请联系管理员")
)

// LoginResult 登录成功后的返回。
type LoginResult struct {
	Tokens *auth.TokenPair
	User   *store.User
	Roles  []string
}

// LoginService 处理登录用例。
type LoginService struct {
	users   *store.UserRepo
	tokens  *auth.Manager
	limiter *auth.LoginLimiter
	legacy  *passwd.LegacyPBE
	logger  *slog.Logger
}

// NewLoginService 构造服务。
func NewLoginService(users *store.UserRepo, tokens *auth.Manager, limiter *auth.LoginLimiter, logger *slog.Logger) *LoginService {
	return &LoginService{
		users:   users,
		tokens:  tokens,
		limiter: limiter,
		legacy:  passwd.NewLegacyPBE(),
		logger:  logger,
	}
}

// Login 执行账号密码登录。
//
// 安全要点：
//  1. 限流前置：先判断是否已锁定，再做任何数据库查询，避免被用于探测账号存在性；
//  2. 统一错误话术：账号不存在与密码错误返回同一文案，防止账号枚举；
//  3. 时间恒定：账号不存在时也执行一次 bcrypt 校验，抹平响应时间差异；
//  4. rehash-on-login：存量 PBE 密文校验通过后立即升级为 bcrypt。
func (s *LoginService) Login(ctx context.Context, username, password, clientIP string) (*LoginResult, error) {
	limitKey := s.limiter.Key(hashSubject(username), clientIP)

	blocked, count, err := s.limiter.Blocked(ctx, limitKey)
	if err != nil {
		// Redis 不可用时 fail-open，仅告警：避免缓存故障导致全站无法登录
		s.logger.Warn("登录限流检查失败，按放行处理", slog.String("error", err.Error()))
	}
	if blocked {
		s.logger.Warn("账号因失败次数过多被锁定",
			slog.String("username", maskText(username)),
			slog.String("client_ip", clientIP),
			slog.Int("fail_count", count),
		)
		return nil, ErrAccountLocked
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
		// 账号不存在：仍执行一次等价的散列运算，抹平时间差
		_, _ = passwd.Hash(password)
		s.recordFailure(ctx, limitKey, username, clientIP)
		return nil, ErrInvalidCredential
	}

	if user.Status != 1 {
		return nil, ErrAccountDisabled
	}

	if err := s.verifyPassword(ctx, user, password); err != nil {
		s.recordFailure(ctx, limitKey, username, clientIP)
		return nil, ErrInvalidCredential
	}

	if err := s.limiter.Reset(ctx, limitKey); err != nil {
		s.logger.Warn("清除登录失败计数失败", slog.String("error", err.Error()))
	}

	roles, err := s.users.ListRoles(ctx, user.ID)
	if err != nil {
		// 角色查询失败不阻断登录，按最小权限处理
		s.logger.Warn("查询用户角色失败，按无角色处理",
			slog.String("user_id", user.ID), slog.String("error", err.Error()))
		roles = nil
	}

	tokens, err := s.tokens.Issue(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("签发令牌失败: %w", err)
	}

	return &LoginResult{Tokens: tokens, User: user, Roles: roles}, nil
}

// verifyPassword 按存储方案校验密码，并在存量格式命中时触发升级。
func (s *LoginService) verifyPassword(ctx context.Context, user *store.User, plain string) error {
	switch passwd.Detect(user.Password) {
	case passwd.SchemeBcrypt:
		return passwd.VerifyBcrypt(plain, user.Password)

	case passwd.SchemeLegacyPBE:
		// 存量：PBEWithMD5AndDES，密文解密后应等于用户名
		if err := s.legacy.Verify(user.Username, plain, user.Password, user.Salt); err != nil {
			return err
		}
		// 校验通过，立即升级为 bcrypt
		newHash, hashErr := passwd.Hash(plain)
		if hashErr != nil {
			s.logger.Warn("密码散列升级失败，本次登录仍放行",
				slog.String("user_id", user.ID), slog.String("error", hashErr.Error()))
			return nil
		}
		if updErr := s.users.UpdatePassword(ctx, user.ID, newHash, ""); updErr != nil {
			s.logger.Warn("回写升级后密码失败",
				slog.String("user_id", user.ID), slog.String("error", updErr.Error()))
			return nil
		}
		s.logger.Info("密码散列已升级为 bcrypt", slog.String("user_id", user.ID))
		return nil

	default:
		return passwd.ErrMismatch
	}
}

func (s *LoginService) recordFailure(ctx context.Context, key, username, clientIP string) {
	n, err := s.limiter.Fail(ctx, key)
	if err != nil {
		s.logger.Warn("记录登录失败计数失败", slog.String("error", err.Error()))
		return
	}
	s.logger.Info("登录失败",
		slog.String("username", maskText(username)),
		slog.String("client_ip", clientIP),
		slog.Int("fail_count", n),
	)
}

// Refresh 用刷新令牌换取新的令牌对，并吊销旧刷新令牌（一次性使用）。
func (s *LoginService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	claims, err := s.tokens.Parse(ctx, refreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidCredential
	}
	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidCredential
	}
	if user.Status != 1 {
		return nil, ErrAccountDisabled
	}
	if revokeErr := s.tokens.Revoke(ctx, refreshToken); revokeErr != nil {
		s.logger.Warn("吊销旧刷新令牌失败", slog.String("error", revokeErr.Error()))
	}
	roles, _ := s.users.ListRoles(ctx, user.ID)
	tokens, err := s.tokens.Issue(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("签发令牌失败: %w", err)
	}
	return &LoginResult{Tokens: tokens, User: user, Roles: roles}, nil
}

// Logout 吊销当前访问令牌。
func (s *LoginService) Logout(ctx context.Context, accessToken string) error {
	return s.tokens.Revoke(ctx, accessToken)
}

// hashSubject 对用户名做不可逆摘要，避免明文写入 Redis 键名。
func hashSubject(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
}

// maskText 对日志中的账号做脱敏（保留首尾各 1 字符）。
func maskText(s string) string {
	r := []rune(s)
	if len(r) <= 2 {
		return "**"
	}
	return string(r[:1]) + "***" + string(r[len(r)-1:])
}

// RedisRevoker 基于 Redis 的令牌吊销存储。
type RedisRevoker struct {
	rdb *redis.Client
}

// NewRedisRevoker 构造吊销存储。
func NewRedisRevoker(rdb *redis.Client) *RedisRevoker {
	return &RedisRevoker{rdb: rdb}
}

// revokePrefix 吊销标记的键前缀。
const revokePrefix = "auth:revoked:"

// Revoke 写入吊销标记，TTL 与令牌剩余有效期一致。
func (r *RedisRevoker) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	if tokenID == "" {
		return errors.New("令牌ID为空")
	}
	if ttl <= 0 {
		// 已过期的令牌无需吊销
		return nil
	}
	return r.rdb.Set(ctx, revokePrefix+tokenID, "1", ttl).Err()
}

// IsRevoked 判断令牌是否已被吊销。
func (r *RedisRevoker) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return true, nil
	}
	n, err := r.rdb.Exists(ctx, revokePrefix+tokenID).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
