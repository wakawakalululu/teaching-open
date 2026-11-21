// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUserRole 对应 sys_user_role 表（api/db/teachingopen2.8.sql L922：
// 仅 id/user_id/role_id 三列，varchar(32)，无 del_flag/create_time 等字段）。
type SysUserRole struct {
	Id     string `json:"id"     orm:"id,primary" description:"主键id"`
	UserId string `json:"userId" orm:"user_id"     description:"用户id"`
	RoleId string `json:"roleId" orm:"role_id"     description:"角色id"`
}
