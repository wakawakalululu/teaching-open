// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDepartPermission is the golang structure of table sys_depart_permission for DAO operations like Where/Data.
type SysDepartPermission struct {
	g.Meta       `orm:"table:sys_depart_permission, do:true"`
	Id           interface{} // ID
	DepartId     interface{} // 部门id
	PermissionId interface{} // 权限id
	DataRuleIds  interface{} // 数据规则id串
}
