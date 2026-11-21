// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysSms 对应 sys_sms 表（api/db/teachingopen2.8.sql L821：消息发送记录）。
type SysSms struct {
	Id           string `json:"id"           orm:"id,primary"    description:"ID"`
	EsTitle      string `json:"esTitle"      orm:"es_title"      description:"消息标题"`
	EsType       string `json:"esType"       orm:"es_type"       description:"发送方式 1短信 2邮件 3微信"`
	EsReceiver   string `json:"esReceiver"   orm:"es_receiver"   description:"接收人"`
	EsParam      string `json:"esParam"      orm:"es_param"      description:"消息参数"`
	EsContent    string `json:"esContent"    orm:"es_content"    description:"消息内容"`
	EsSendTime   string `json:"esSendTime"   orm:"es_send_time"  description:"发送时间"`
	EsSendStatus string `json:"esSendStatus" orm:"es_send_status" description:"发送状态 0未发送 1已发送 2发送失败"`
	EsSendNum    int    `json:"esSendNum"    orm:"es_send_num"   description:"发送次数"`
	EsResult     string `json:"esResult"     orm:"es_result"     description:"发送结果"`
	Remark       string `json:"remark"       orm:"remark"        description:"备注"`
	CreateBy     string `json:"createBy"     orm:"create_by"     description:"创建人"`
	CreateTime   string `json:"createTime"   orm:"create_time"   description:"创建时间"`
	UpdateBy     string `json:"updateBy"     orm:"update_by"     description:"更新人"`
}
