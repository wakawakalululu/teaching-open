// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUserDepart 对应 sys_user_depart 表（列结构以 api/db/teachingopen2.8.sql L913 为准；
// 主键列为大写 ID，无 del_flag/时间列，Java 版为物理删除）。
type SysUserDepart struct {
	Id     string `json:"id"     orm:"ID,primary" description:"主键id"`
	UserId string `json:"userId" orm:"user_id"     description:"用户id"`
	DepId  string `json:"depId"  orm:"dep_id"      description:"部门id"`
}
