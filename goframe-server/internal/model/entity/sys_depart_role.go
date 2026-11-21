// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysDepartRole 对应 sys_depart_role 表（api/db/teachingopen2.8.sql L604：
// 9 列，部门级角色定义；无 del_flag）。
type SysDepartRole struct {
	Id          string `json:"id"          orm:"id,primary"   description:"ID"`
	DepartId    string `json:"departId"    orm:"depart_id"    description:"部门id"`
	RoleName    string `json:"roleName"    orm:"role_name"    description:"角色名称"`
	RoleCode    string `json:"roleCode"    orm:"role_code"    description:"角色编码"`
	Description string `json:"description" orm:"description"  description:"描述"`
	CreateBy    string `json:"createBy"    orm:"create_by"    description:"创建人"`
	CreateTime  string `json:"createTime"  orm:"create_time"  description:"创建时间"`
	UpdateBy    string `json:"updateBy"    orm:"update_by"    description:"修改人"`
	UpdateTime  string `json:"updateTime"  orm:"update_time"  description:"修改时间"`
}
