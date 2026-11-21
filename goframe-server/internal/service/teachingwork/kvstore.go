// kvstore.go 作业标签/点赞/班级日志计数用的 KV+Set 存储：Redis（主，
// 键格式与 Java 完全一致以兼容灰度双写）与进程内存（REDIS_ENABLED=false
// 时兜底，参考 sms 域 store 先例）。
package teachingwork

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// KVStore 作业域存储抽象。
type KVStore interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, val string)
	SetEx(ctx context.Context, key, val string, ttl time.Duration)
	SGet(ctx context.Context, key string) []string
	SAdd(ctx context.Context, key, member string)
	SHasKey(ctx context.Context, key, member string) bool
	SRemove(ctx context.Context, key, member string)
	Del(ctx context.Context, key string)
}

// NewKVStore 按开关选择实现（redisEnabled=false 走进程内存）。
func NewKVStore(redisEnabled bool) KVStore {
	if redisEnabled {
		return &redisStore{}
	}
	return &memStore{sets: map[string]map[string]bool{}, vals: map[string]memVal{}}
}

// redisStore Redis 实现（键名直接使用 Java 原格式，不做命名空间改写）。
type redisStore struct{}

func (s *redisStore) Get(ctx context.Context, key string) (string, bool) {
	v, err := g.Redis().Do(ctx, "GET", key)
	if err != nil || v == nil {
		return "", false
	}
	return v.String(), true
}

func (s *redisStore) Set(ctx context.Context, key, val string) {
	_, _ = g.Redis().Do(ctx, "SET", key, val)
}

func (s *redisStore) SetEx(ctx context.Context, key, val string, ttl time.Duration) {
	_, _ = g.Redis().Do(ctx, "SET", key, val, "EX", int(ttl.Seconds()))
}

func (s *redisStore) SGet(ctx context.Context, key string) []string {
	v, err := g.Redis().Do(ctx, "SMEMBERS", key)
	if err != nil || v == nil {
		return nil
	}
	out := make([]string, 0)
	for _, item := range v.Strings() {
		out = append(out, item)
	}
	return out
}

func (s *redisStore) SAdd(ctx context.Context, key, member string) {
	_, _ = g.Redis().Do(ctx, "SADD", key, member)
}

func (s *redisStore) SHasKey(ctx context.Context, key, member string) bool {
	v, err := g.Redis().Do(ctx, "SISMEMBER", key, member)
	if err != nil {
		return false
	}
	return v.Int() == 1
}

func (s *redisStore) SRemove(ctx context.Context, key, member string) {
	_, _ = g.Redis().Do(ctx, "SREM", key, member)
}

// memVal 内存值（含过期）。
type memVal struct {
	val    string
	expire time.Time
}

// memStore 进程内存实现（无过期清理线程，读取时惰性判断）。
type memStore struct {
	mu   sync.RWMutex
	vals map[string]memVal
	sets map[string]map[string]bool
}

func (s *memStore) Get(_ context.Context, key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.vals[key]
	if !ok || (!v.expire.IsZero() && time.Now().After(v.expire)) {
		return "", false
	}
	return v.val, true
}

func (s *memStore) Set(_ context.Context, key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vals[key] = memVal{val: val}
}

func (s *memStore) SetEx(_ context.Context, key, val string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vals[key] = memVal{val: val, expire: time.Now().Add(ttl)}
}

func (s *memStore) SGet(_ context.Context, key string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.sets[key]))
	for m := range s.sets[key] {
		out = append(out, m)
	}
	return out
}

func (s *memStore) SAdd(_ context.Context, key, member string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sets[key] == nil {
		s.sets[key] = map[string]bool{}
	}
	s.sets[key][member] = true
}

func (s *memStore) SHasKey(_ context.Context, key, member string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sets[key][member]
}

func (s *memStore) SRemove(_ context.Context, key, member string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sets[key], member)
}

// Del 删除 string 键（setWorkTag 清除分支用）。
func (s *redisStore) Del(ctx context.Context, key string) {
	_, _ = g.Redis().Do(ctx, "DEL", key)
}

// Del 删除 string 键（内存实现）。
func (s *memStore) Del(_ context.Context, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.vals, key)
}
