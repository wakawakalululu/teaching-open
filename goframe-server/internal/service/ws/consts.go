// consts.go WebSocket 消息常量与类型（对齐 WebsocketConst）。
package ws

import (
	"github.com/gogf/gf/v2/frame/g"
)

// gwebsocketTextMessage gorilla/websocket TextMessage 值（1）。
const TextMessage = 1

// TopicMsg 全员通知消息（{cmd:"topic", msgId, txt}）。
func TopicMsg(msgID, txt string) []byte {
	b, _ := g.NewVar(g.Map{"cmd": "topic", "msgId": msgID, "txt": txt}).MarshalJSON()
	return b
}

// UserMsg 单人通知消息（{cmd:"user", msgId, txt}）。
func UserMsg(msgID, txt string) []byte {
	b, _ := g.NewVar(g.Map{"cmd": "user", "msgId": msgID, "txt": txt}).MarshalJSON()
	return b
}
