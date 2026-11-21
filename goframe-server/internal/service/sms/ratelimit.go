// Package sms 中的行为限流：进程内滑动窗口计数。
//
// 轻量自研实现（参考 login 域 LoginLimiter 的限流先例，但不依赖 Redis——
// 行为限流属于「尽力而为」的防护层，进程重启即清零，不影响验证码主链路
// 的 fail-closed 语义），供 /sys/sms 与注册接口共用，避免改动 login 包。
package sms

import (
	"sync"
	"time"
)

// RateWindow 按 key 的滑动窗口限流器（同 key 在窗口内最多放行 limit 次）。
type RateWindow struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
	now    func() time.Time // 可注入时钟（单测用）
}

// maxRateKeys 触发全量清扫的 key 数阈值（防长期运行下 map 无界增长）。
const maxRateKeys = 10000

// NewRateWindow 构造限流器；limit 最小为 1，window 最小为 1 分钟。
func NewRateWindow(limit int, window time.Duration) *RateWindow {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateWindow{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
		now:    time.Now,
	}
}

// Allow 记录一次请求并判断是否放行（放行即计数，拒绝不计数）。
func (w *RateWindow) Allow(key string) bool {
	now := w.now()
	w.mu.Lock()
	defer w.mu.Unlock()

	valid := w.hits[key][:0]
	for _, ts := range w.hits[key] {
		if now.Sub(ts) < w.window {
			valid = append(valid, ts)
		}
	}
	if len(valid) == 0 {
		delete(w.hits, key)
	} else {
		w.hits[key] = valid
	}
	if len(w.hits) > maxRateKeys {
		w.sweepLocked(now)
	}
	if len(valid) >= w.limit {
		return false
	}
	w.hits[key] = append(valid, now)
	return true
}

// sweepLocked 全量清理窗口外 key（调用方必须已持锁）。
func (w *RateWindow) sweepLocked(now time.Time) {
	for key, stamps := range w.hits {
		valid := stamps[:0]
		for _, ts := range stamps {
			if now.Sub(ts) < w.window {
				valid = append(valid, ts)
			}
		}
		if len(valid) == 0 {
			delete(w.hits, key)
		} else {
			w.hits[key] = valid
		}
	}
}
