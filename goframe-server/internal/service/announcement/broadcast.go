// broadcast.go 通告广播依赖注入（ws 中心由 cmd 装配传入）。
package announcement

import twWS "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/ws"

// SetBroadcaster 注入 WebSocket 广播依赖（cmd.go 装配时调用一次）。
func SetBroadcaster(hub *twWS.Hub) {
	broadcaster = hub
}

// broadcaster 全局广播器（nil 时静默跳过）。
var broadcaster *twWS.Hub

// broadcastTopic 全员广播 topic 消息。
func broadcastTopic(msgID, txt string) {
	if broadcaster != nil {
		broadcaster.SendAllMessage(twWS.TopicMsg(msgID, txt))
	}
}
