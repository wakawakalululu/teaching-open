// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysDepartRolePermission 对应 sys_depart_role_permission 表（api/db/teachingopen2.8.sql L616：
// 5 列；data_rule_ids 为逗号分隔的数据规则 id 串）。
type SysDepartRolePermission struct {
	Id           string `json:"id"           orm:"id,primary"    description:"ID"`
	DepartId     string `json:"departId"     orm:"depart_id"     description:"部门id"`
	RoleId       string `json:"roleId"       orm:"role_id"       description:"部门角色id"`
	PermissionId string `json:"permissionId" orm:"permission_id" description:"菜单id"`
	DataRuleIds  string `json:"dataRuleIds"  orm:"data_rule_ids" description:"数据规则id串"`
}
