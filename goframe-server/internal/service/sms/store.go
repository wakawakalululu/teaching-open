// Package sms 中的验证码存储实现：Redis（主）与进程内存（无 Redis 时兜底）。
//
// fail-closed 原则：
//   - REDIS_ENABLED=true 时使用 redisCodeStore；运行期任何 Redis 读写错误
//     由调用方（Service.Verify/VerifyReset/Send）按「校验失败/发送失败」处理，
//     不降级放行；
//   - REDIS_ENABLED=false 时使用 memCodeStore（参考 login 域限流降级先例），
//     保证开发/演示环境自服务链路可用。
//
// 键命名：Java 直接以裸手机号作为 Redis key；Go 版统一加 sms:code: 前缀做
// 命名空间隔离（纯进程内实现细节，不影响任何外部契约）。
package sms

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CodeStore 验证码存储抽象。
type CodeStore interface {
	// Save 写入验证码并设置 TTL。
	Save(ctx context.Context, key, code string, ttl time.Duration) error
	// Load 读取验证码；found 表示是否存在未过期记录；
	// err 仅表示存储基础设施故障（由调用方 fail-closed 处理）。
	Load(ctx context.Context, key string) (code string, found bool, err error)
	// Delete 立即删除验证码。
	Delete(ctx context.Context, key string) error
}

// redisCodeStore 基于 Redis 的验证码存储。
type redisCodeStore struct{}

// codeKeyPrefix Redis 键前缀（Java 版用裸手机号做 key，此处加命名空间）。
const codeKeyPrefix = "sms:code:"

// newRedisCodeStore 构造 Redis 存储。
func newRedisCodeStore() *redisCodeStore {
	return &redisCodeStore{}
}

func (s *redisCodeStore) fullKey(key string) string {
	return codeKeyPrefix + key
}

// Save 写入验证码（SET key value EX ttl，GoFrame SetEX 秒级 TTL）。
func (s *redisCodeStore) Save(ctx context.Context, key, code string, ttl time.Duration) error {
	return g.Redis().SetEX(ctx, s.fullKey(key), code, int64(ttl.Seconds()))
}

// Load 读取验证码；键不存在返回 found=false（不视为错误）。
func (s *redisCodeStore) Load(ctx context.Context, key string) (string, bool, error) {
	v, err := g.Redis().Get(ctx, s.fullKey(key))
	if err != nil {
		return "", false, err
	}
	if v == nil || v.IsNil() {
		return "", false, nil
	}
	return v.String(), true, nil
}

// Delete 删除验证码。
func (s *redisCodeStore) Delete(ctx context.Context, key string) error {
	_, err := g.Redis().Del(ctx, s.fullKey(key))
	return err
}

// memEntry 进程内验证码条目。
type memEntry struct {
	code      string
	expiresAt time.Time
}

// memCodeStore 进程内验证码存储（无 Redis 时的兜底实现）。
//
// 过期清理为惰性策略：Load/Save 时顺手清理命中的过期键；Save 时若条目数
// 超过 memStoreMaxKeys 则全量清扫一轮，防止长期运行下内存无界增长。
type memCodeStore struct {
	mu      sync.Mutex
	entries map[string]memEntry
	now     func() time.Time // 可注入时钟（单测用）
}

// memStoreMaxKeys 触发全量清扫的条目数阈值。
const memStoreMaxKeys = 10000

// newMemCodeStore 构造进程内存储。
func newMemCodeStore() *memCodeStore {
	return &memCodeStore{
		entries: make(map[string]memEntry),
		now:     time.Now,
	}
}

// Save 写入验证码；同键覆盖旧值（与 Redis SET 语义一致）。
func (s *memCodeStore) Save(_ context.Context, key, code string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = memEntry{code: code, expiresAt: s.now().Add(ttl)}
	if len(s.entries) > memStoreMaxKeys {
		s.sweepLocked()
	}
	return nil
}

// Load 读取验证码；命中过期键时顺带删除。
func (s *memCodeStore) Load(_ context.Context, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok {
		return "", false, nil
	}
	if s.now().After(entry.expiresAt) {
		delete(s.entries, key)
		return "", false, nil
	}
	return entry.code, true, nil
}

// Delete 删除验证码（键不存在视为成功，与 Redis DEL 语义一致）。
func (s *memCodeStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
	return nil
}

// sweepLocked 全量清理过期条目（调用方必须已持锁）。
func (s *memCodeStore) sweepLocked() {
	now := s.now()
	for key, entry := range s.entries {
		if now.After(entry.expiresAt) {
			delete(s.entries, key)
		}
	}
}
