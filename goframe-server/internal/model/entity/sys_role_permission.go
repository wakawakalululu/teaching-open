// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysRolePermission 对应 sys_role_permission 表（api/db/teachingopen2.8.sql L811：
// 4 列；data_rule_ids 为逗号分隔的数据规则 id 串，varchar(1000)）。
type SysRolePermission struct {
	Id           string `json:"id"          orm:"id,primary"    description:"主键id"`
	RoleId       string `json:"roleId"      orm:"role_id"       description:"角色id"`
	PermissionId string `json:"permissionId" orm:"permission_id" description:"权限id"`
	DataRuleIds  string `json:"dataRuleIds" orm:"data_rule_ids" description:"数据规则id串"`
}
