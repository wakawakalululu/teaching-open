// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysDepartRoleUser 对应 sys_depart_role_user 表（api/db/teachingopen2.8.sql L627：
// 3 列；注意角色列名为 drole_id，Java 实体 getter 为 getDroleId）。
type SysDepartRoleUser struct {
	Id      string `json:"id"     orm:"id,primary" description:"ID"`
	UserId  string `json:"userId" orm:"user_id"    description:"用户id"`
	DroleId string `json:"droleId" orm:"drole_id"  description:"部门角色id"`
}
