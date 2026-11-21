// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysPermissionDataRule 对应 sys_permission_data_rule 表（api/db/teachingopen2.8.sql L754：
// 11 列，菜单级数据权限规则定义；status 为 varchar "1"有效/"0"无效）。
type SysPermissionDataRule struct {
	Id             string `json:"id"             orm:"id,primary"       description:"ID"`
	PermissionId   string `json:"permissionId"   orm:"permission_id"    description:"菜单ID"`
	RuleName       string `json:"ruleName"       orm:"rule_name"        description:"规则名称"`
	RuleColumn     string `json:"ruleColumn"     orm:"rule_column"      description:"字段"`
	RuleConditions string `json:"ruleConditions" orm:"rule_conditions"  description:"条件"`
	RuleValue      string `json:"ruleValue"      orm:"rule_value"       description:"规则值"`
	Status         string `json:"status"         orm:"status"           description:"权限有效状态1有0否"`
	CreateTime     string `json:"createTime"     orm:"create_time"      description:"创建时间"`
	CreateBy       string `json:"createBy"       orm:"create_by"        description:"创建人"`
	UpdateTime     string `json:"updateTime"     orm:"update_time"      description:"修改时间"`
	UpdateBy       string `json:"updateBy"       orm:"update_by"        description:"修改人"`
}
