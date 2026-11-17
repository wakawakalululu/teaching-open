// Package auth 中的登录限流实现：基于 Redis 的滑动窗口计数。
//
// 使用 GoFrame 的 gredis 组件，INCR + PEXPIRE 通过 Lua 脚本原子执行，
// 避免并发下计数丢失或 TTL 未设置。
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// LoginLimiter 限制同一「账号 + 客户端 IP」在窗口内的登录失败次数。
type LoginLimiter struct {
	limit  int
	window time.Duration
}

// NewLoginLimiter 构造限流器。
func NewLoginLimiter(limit int, window time.Duration) *LoginLimiter {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	return &LoginLimiter{limit: limit, window: window}
}

// incrWithExpire 原子递增并设置过期（Lua 脚本，避免并发竞态）。
const incrWithExpire = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`

// Key 生成限流键；调用方传入的 subject 应已做哈希或脱敏。
func (l *LoginLimiter) Key(subject, clientIP string) string {
	return fmt.Sprintf("login:fail:%s:%s", subject, clientIP)
}

// Blocked 判断当前是否已超限。
func (l *LoginLimiter) Blocked(ctx context.Context, key string) (bool, int, error) {
	if ctx == nil {
		ctx = gctx.GetInitCtx()
	}
	count, err := g.Redis().Get(ctx, key)
	if err != nil {
		// Redis 故障时返回未超限（fail-open），由调用方决定是否告警
		return false, 0, err
	}
	n := count.Int()
	return n >= l.limit, n, nil
}

// Fail 记录一次失败，返回累计次数。
func (l *LoginLimiter) Fail(ctx context.Context, key string) (int, error) {
	if ctx == nil {
		ctx = gctx.GetInitCtx()
	}
	res, err := g.Redis().Eval(ctx, incrWithExpire, 1, []string{key}, []any{l.window.Milliseconds()})
	if err != nil {
		return 0, err
	}
	return res.Int(), nil
}

// Reset 登录成功后清零计数。
func (l *LoginLimiter) Reset(ctx context.Context, key string) error {
	if ctx == nil {
		ctx = gctx.GetInitCtx()
	}
	_, err := g.Redis().Del(ctx, key)
	return err
}

// Remaining 返回剩余可尝试次数。
func (l *LoginLimiter) Remaining(count int) int {
	left := l.limit - count
	if left < 0 {
		return 0
	}
	return left
}
