// Package common 提供业务域控制器共享的响应结构。
//
// 对齐 Java org.jeecg.common.api.vo.Result<T> 的完整字段
// （success/code/message/result/timestamp），前端 axios 拦截器与
// 通知组件按 success + message 工作，无需任何适配。
//
// 与登录控制器（controller/login）的差异：业务域遵循 Java 语义，
// 业务失败返回 HTTP 200 + success:false（Java error500 即如此），
// HTTP 状态码仅用于登录等网关级语义。
package common

import (
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// OK 成功响应（默认话术「操作成功！」）。
func OK(r *ghttp.Request, data interface{}) {
	OKMsg(r, "操作成功！", data)
}

// OKMsg 成功响应（自定义话术 + 数据，对应 Java result.success(msg)）。
func OKMsg(r *ghttp.Request, message string, data interface{}) {
	r.Response.WriteStatus(http.StatusOK)
	r.Response.WriteJson(g.Map{
		"success":   true,
		"message":   message,
		"code":      200,
		"result":    data,
		"timestamp": time.Now().UnixMilli(),
	})
}

// Fail 业务失败响应（HTTP 200 + success:false，对应 Java result.error500）。
func Fail(r *ghttp.Request, message string) {
	r.Response.WriteStatus(http.StatusOK)
	r.Response.WriteJson(g.Map{
		"success":   false,
		"message":   message,
		"code":      500,
		"result":    nil,
		"timestamp": time.Now().UnixMilli(),
	})
}
