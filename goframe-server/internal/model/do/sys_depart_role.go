// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDepartRole is the golang structure of table sys_depart_role for DAO operations like Where/Data.
type SysDepartRole struct {
	g.Meta      `orm:"table:sys_depart_role, do:true"`
	Id          interface{} // ID
	DepartId    interface{} // 部门id
	RoleName    interface{} // 角色名称
	RoleCode    interface{} // 角色编码
	Description interface{} // 描述
	CreateBy    interface{} // 创建人
	CreateTime  interface{} // 创建时间
	UpdateBy    interface{} // 修改人
	UpdateTime  interface{} // 修改时间
}
