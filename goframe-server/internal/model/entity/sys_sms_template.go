// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysSmsTemplate 对应 sys_sms_template 表（api/db/teachingopen2.8.sql L843：
// 消息模板表；Java 实体 @TableName("sys_sms_template")）。
type SysSmsTemplate struct {
	Id               string `json:"id"               orm:"id,primary"        description:"ID"`
	TemplateName     string `json:"templateName"     orm:"template_name"     description:"模板名称"`
	TemplateCode     string `json:"templateCode"     orm:"template_code"     description:"模板编码"`
	TemplateType     string `json:"templateType"     orm:"template_type"     description:"模板类型"`
	TemplateContent  string `json:"templateContent"  orm:"template_content"  description:"模板内容"`
	TemplateTestJson string `json:"templateTestJson" orm:"template_test_json" description:"模板测试JSON"`
	CreateBy         string `json:"createBy"         orm:"create_by"         description:"创建人"`
	CreateTime       string `json:"createTime"       orm:"create_time"       description:"创建时间"`
	UpdateBy         string `json:"updateBy"         orm:"update_by"         description:"更新人"`
	UpdateTime       string `json:"updateTime"       orm:"update_time"       description:"更新时间"`
}
