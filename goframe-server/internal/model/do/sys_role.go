// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysRole is the golang structure of table sys_role for DAO operations like Where/Data.
type SysRole struct {
	g.Meta      `orm:"table:sys_role, do:true"`
	Id          interface{} // 主键id
	RoleName    interface{} // 角色名称
	RoleCode    interface{} // 角色编码
	Description interface{} // 描述
	CreateBy    interface{} // 创建人
	CreateTime  interface{} // 创建时间
	UpdateBy    interface{} // 更新人
	UpdateTime  interface{} // 更新时间
	RoleLevel   interface{} // 角色等级
}
