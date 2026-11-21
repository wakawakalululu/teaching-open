// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysSmsTemplate is the golang structure of table sys_sms_template for DAO operations like Where/Data.
type SysSmsTemplate struct {
	g.Meta           `orm:"table:sys_sms_template, do:true"`
	Id               interface{} // ID
	TemplateName     interface{} // 模板名称
	TemplateCode     interface{} // 模板编码
	TemplateType     interface{} // 模板类型
	TemplateContent  interface{} // 模板内容
	TemplateTestJson interface{} // 模板测试JSON
	CreateBy         interface{} // 创建人
	CreateTime       interface{} // 创建时间
	UpdateBy         interface{} // 更新人
	UpdateTime       interface{} // 更新时间
}
