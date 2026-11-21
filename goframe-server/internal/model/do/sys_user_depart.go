// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysUserDepart is the golang structure of table sys_user_depart for DAO operations like Where/Data.
type SysUserDepart struct {
	g.Meta `orm:"table:sys_user_depart, do:true"`
	Id     interface{} // 主键id（表列为大写 ID）
	UserId interface{} // 用户id
	DepId  interface{} // 部门id
}
