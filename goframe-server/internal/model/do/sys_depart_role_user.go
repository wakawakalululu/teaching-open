// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDepartRoleUser is the golang structure of table sys_depart_role_user for DAO operations like Where/Data.
type SysDepartRoleUser struct {
	g.Meta  `orm:"table:sys_depart_role_user, do:true"`
	Id      interface{} // ID
	UserId  interface{} // 用户id
	DroleId interface{} // 部门角色id
}
