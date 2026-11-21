// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysCheckRule 对应 sys_check_rule 表（api/db/teachingopen2.8.sql：
// 编码校验规则表）。
type SysCheckRule struct {
	Id              string `json:"id"          orm:"id,primary"   description:"ID"`
	RuleName        string `json:"ruleName"    orm:"rule_name"    description:"规则名称"`
	RuleCode        string `json:"ruleCode"    orm:"rule_code"    description:"规则编码"`
	RuleJson        string `json:"ruleJson"    orm:"rule_json"    description:"规则JSON"`
	RuleDescription string `json:"ruleDescription" orm:"rule_description" description:"规则描述"`
	UpdateBy        string `json:"updateBy"    orm:"update_by"    description:"更新人"`
	UpdateTime      string `json:"updateTime"  orm:"update_time"  description:"更新时间"`
	CreateBy        string `json:"createBy"    orm:"create_by"    description:"创建人"`
	CreateTime      string `json:"createTime"  orm:"create_time"  description:"创建时间"`
}
