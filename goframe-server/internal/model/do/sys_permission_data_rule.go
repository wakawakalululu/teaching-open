// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysPermissionDataRule is the golang structure of table sys_permission_data_rule for DAO operations like Where/Data.
type SysPermissionDataRule struct {
	g.Meta         `orm:"table:sys_permission_data_rule, do:true"`
	Id             interface{} // ID
	PermissionId   interface{} // 菜单ID
	RuleName       interface{} // 规则名称
	RuleColumn     interface{} // 字段
	RuleConditions interface{} // 条件
	RuleValue      interface{} // 规则值
	Status         interface{} // 权限有效状态1有0否
	CreateTime     interface{} // 创建时间
	CreateBy       interface{} // 创建人
	UpdateTime     interface{} // 修改时间
	UpdateBy       interface{} // 修改人
}
