// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysSms is the golang structure of table sys_sms for DAO operations like Where/Data.
type SysSms struct {
	g.Meta       `orm:"table:sys_sms, do:true"`
	Id           interface{} // ID
	EsTitle      interface{} // 消息标题
	EsType       interface{} // 发送方式 1短信 2邮件 3微信
	EsReceiver   interface{} // 接收人
	EsParam      interface{} // 消息参数
	EsContent    interface{} // 消息内容
	EsSendTime   interface{} // 发送时间
	EsSendStatus interface{} // 发送状态 0未发送 1已发送 2发送失败
	EsSendNum    interface{} // 发送次数
	EsResult     interface{} // 发送结果
	Remark       interface{} // 备注
	CreateBy     interface{} // 创建人
	CreateTime   interface{} // 创建时间
	UpdateBy     interface{} // 更新人
}
