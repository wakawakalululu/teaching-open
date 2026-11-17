// Package notify 提供飞书 Webhook 告警通知。
//
// 用于安全事件（登录失败锁定、Redis 故障、配置异常等）的即时推送，
// 不阻塞主流程：所有通知异步执行，失败仅记录日志。
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// FeishuNotifier 飞书机器人告警。
type FeishuNotifier struct {
	webhook string
	client  *http.Client
}

// NewFeishu 构造飞书告警器；webhook 为空时禁用。
func NewFeishu(webhook string) *FeishuNotifier {
	return &FeishuNotifier{
		webhook: webhook,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// Enabled 判断是否已配置 webhook。
func (n *FeishuNotifier) Enabled() bool {
	return n != nil && n.webhook != ""
}

// feishuTextPayload 飞书 text 消息体。
type feishuTextPayload struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Send 同步发送一条文本消息。
func (n *FeishuNotifier) Send(ctx context.Context, text string) error {
	if !n.Enabled() {
		return nil
	}
	payload := feishuTextPayload{MsgType: "text"}
	payload.Content.Text = text

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &FeishuError{Status: resp.StatusCode}
	}
	return nil
}

// SendAsync 异步发送，不阻塞主流程；失败仅记录日志。
func (n *FeishuNotifier) SendAsync(ctx context.Context, text string) {
	if !n.Enabled() {
		return
	}
	go func() {
		if err := n.Send(context.Background(), text); err != nil {
			g.Log().Warningf(ctx, "飞书告警发送失败 err=%s", err.Error())
		}
	}()
}

// FeishuError 飞书接口错误。
type FeishuError struct {
	Status int
}

func (e *FeishuError) Error() string {
	return "飞书 webhook 返回状态码 " + string(rune(e.Status))
}
