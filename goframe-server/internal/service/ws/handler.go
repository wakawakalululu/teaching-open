// handler.go WebSocket HTTP 处理器（ghttp 原生升级）。
package ws

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Handler 返回 WebSocket 升级与消息循环的 HTTP 处理器
// （anon——ShiroConfig.java:129 /websocket/** anon）。
// Java 消息 JSON 形状：{cmd, msgId, txt}（WebsocketConst）。
func (h *Hub) Handler(r *ghttp.Request) {
	userID := r.Get("userId").String()
	conn, err := r.WebSocket()
	if err != nil {
		g.Log().Errorf(r.Context(), "WebSocket 升级失败 err=%s", err.Error())
		return
	}
	defer func() { _ = conn.Close() }()
	deregister := h.Register(userID, func(msg []byte) error {
		return conn.WriteMessage(TextMessage, msg)
	})
	defer deregister()
	for {
		if _, _, rerr := conn.NextReader(); rerr != nil {
			// 客户端断开/异常 → 注销连接
			return
		}
	}
}
