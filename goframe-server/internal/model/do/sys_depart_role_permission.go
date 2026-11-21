// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDepartRolePermission is the golang structure of table sys_depart_role_permission for DAO operations like Where/Data.
type SysDepartRolePermission struct {
	g.Meta       `orm:"table:sys_depart_role_permission, do:true"`
	Id           interface{} // ID
	DepartId     interface{} // 部门id
	RoleId       interface{} // 部门角色id
	PermissionId interface{} // 菜单id
	DataRuleIds  interface{} // 数据规则id串
}
