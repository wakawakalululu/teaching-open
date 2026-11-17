// Package login 中的 Redis 令牌吊销存储。
package login

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// RedisRevoker 基于 Redis 的令牌吊销存储。
type RedisRevoker struct{}

// NewRedisRevoker 构造吊销存储。
func NewRedisRevoker() *RedisRevoker {
	return &RedisRevoker{}
}

// revokePrefix 吊销标记的键前缀。
const revokePrefix = "auth:revoked:"

// Revoke 写入吊销标记，TTL 与令牌剩余有效期一致。
func (r *RedisRevoker) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	if tokenID == "" {
		return errors.New("令牌ID为空")
	}
	if ttl <= 0 {
		return nil
	}
	return g.Redis().SetEX(ctx, revokePrefix+tokenID, "1", int64(ttl.Seconds()))
}

// IsRevoked 判断令牌是否已被吊销。
func (r *RedisRevoker) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	if tokenID == "" {
		return true, nil
	}
	n, err := g.Redis().Exists(ctx, revokePrefix+tokenID)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
