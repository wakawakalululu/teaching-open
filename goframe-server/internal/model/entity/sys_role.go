// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysRole 对应 sys_role 表（api/db/teachingopen2.8.sql L798：
// 9 列，role_code 唯一键 uniq_sys_role_role_code；无 del_flag，删除即物理删）。
type SysRole struct {
	Id          string `json:"id"          orm:"id,primary"    description:"主键id"`
	RoleName    string `json:"roleName"    orm:"role_name"     description:"角色名称"`
	RoleCode    string `json:"roleCode"    orm:"role_code"     description:"角色编码"`
	Description string `json:"description" orm:"description"   description:"描述"`
	CreateBy    string `json:"createBy"    orm:"create_by"     description:"创建人"`
	CreateTime  string `json:"createTime"  orm:"create_time"   description:"创建时间"`
	UpdateBy    string `json:"updateBy"    orm:"update_by"     description:"更新人"`
	UpdateTime  string `json:"updateTime"  orm:"update_time"   description:"更新时间"`
	RoleLevel   int    `json:"roleLevel"   orm:"role_level"    description:"角色等级"`
}
