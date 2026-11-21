// Package ws 实时推送 WebSocket 中心（Hub）。
//
// 契约蓝本：LoginController doReleaseData 的 webSocket.sendAllMessage /
// sendMoreMessage 广播 + 前端 HeaderNotice.vue 连接 /websocket/{userId}。
// 消息 JSON 形状对齐 WebsocketConst：{cmd: "topic"|"user", msgId, txt}。
//
// GoFrame ghttp 原生支持 WebSocket 升级（r.WebSocket()）。
package ws

import (
	"context"
	"sync"
)

// Hub 连接注册表（userId → 连接集合，支持多端登录多连接）。
type Hub struct {
	mu    sync.RWMutex
	conns map[string]map[string]SendFunc // userId -> connID -> sender
}

// SendFunc 向单个连接写入 JSON 消息的函数抽象。
type SendFunc func(msg []byte) error

// sendFuncKey 连接唯一标识。
type sendFuncKey struct{ id string }

// NewHub 构造 Hub。
func New() *Hub {
	return &Hub{conns: map[string]map[string]SendFunc{}}
}

// Register 注册连接（登录连接建立时调用）；返回注销函数。
func (h *Hub) Register(userID string, send SendFunc) (deregister func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	connID := guidLike()
	if h.conns[userID] == nil {
		h.conns[userID] = map[string]SendFunc{}
	}
	h.conns[userID][connID] = send
	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.conns[userID], connID)
		if len(h.conns[userID]) == 0 {
			delete(h.conns, userID)
		}
	}
}

// SendAllMessage 全员广播（对齐 webSocket.sendAllMessage）。
func (h *Hub) SendAllMessage(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conns := range h.conns {
		for _, send := range conns {
			_ = send(msg)
		}
	}
}

// SendMessage 指定用户推送（对齐 webSocket.sendMessage(userId, msg)）。
func (h *Hub) SendMessage(ctx context.Context, userID string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, send := range h.conns[userID] {
		_ = send(msg)
	}
}

// guidLike 生成连接 id（仓库 guid.S 风格，包内独立实现避免导入环）。
func guidLike() string {
	return fmtSprintf("%d", timeNowNano())
}
