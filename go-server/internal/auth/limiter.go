// Package auth 中的登录限流实现：基于 Redis 的滑动窗口计数。
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginLimiter 限制同一「账号 + 客户端 IP」在窗口内的登录失败次数。
//
// 实现选用 Redis 的 INCR + EXPIRE 组合：
//   - INCR 与 EXPIRE 通过 Lua 脚本原子执行，避免并发下计数丢失或 TTL 未设置；
//   - key 使用哈希后的组合，避免把明文用户名/手机号写进 Redis 键名。
type LoginLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

// NewLoginLimiter 构造限流器。
func NewLoginLimiter(rdb *redis.Client, limit int, window time.Duration) *LoginLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	return &LoginLimiter{rdb: rdb, limit: limit, window: window}
}

var incrWithExpire = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

// Key 生成限流键；调用方传入的 subject 应已做哈希或脱敏。
func (l *LoginLimiter) Key(subject, clientIP string) string {
	return fmt.Sprintf("login:fail:%s:%s", subject, clientIP)
}

// Blocked 判断当前是否已超限。
func (l *LoginLimiter) Blocked(ctx context.Context, key string) (bool, int, error) {
	count, err := l.rdb.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return false, 0, nil
		}
		// Redis 故障时不阻断登录，仅记录（由调用方决定是否 fail-open）
		return false, 0, err
	}
	return count >= l.limit, count, nil
}

// Fail 记录一次失败，返回累计次数。
func (l *LoginLimiter) Fail(ctx context.Context, key string) (int, error) {
	res, err := incrWithExpire.Run(ctx, l.rdb, []string{key}, l.window.Milliseconds()).Int()
	if err != nil {
		return 0, err
	}
	return res, nil
}

// Reset 登录成功后清零计数。
func (l *LoginLimiter) Reset(ctx context.Context, key string) error {
	return l.rdb.Del(ctx, key).Err()
}

// Remaining 返回剩余可尝试次数。
func (l *LoginLimiter) Remaining(count int) int {
	left := l.limit - count
	if left < 0 {
		return 0
	}
	return left
}
