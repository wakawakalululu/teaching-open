// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysCategory 对应 sys_category 表。
type SysCategory struct {
	Id         string `json:"id"         orm:"id,primary"     description:"主键"`
	Pid        string `json:"pid"        orm:"pid"            description:"父级节点"`
	Name       string `json:"name"       orm:"name"           description:"类型名称"`
	Code       string `json:"code"       orm:"code"           description:"类型编码"`
	CreateBy   string `json:"createBy"   orm:"create_by"      description:"创建人"`
	CreateTime string `json:"createTime" orm:"create_time"    description:"创建日期"`
	UpdateBy   string `json:"updateBy"   orm:"update_by"      description:"更新人"`
	UpdateTime string `json:"updateTime" orm:"update_time"    description:"更新日期"`
	SysOrgCode string `json:"sysOrgCode" orm:"sys_org_code"   description:"所属部门"`
	HasChild   string `json:"hasChild"   orm:"has_child"      description:"是否有子节点"`
	Value      string `json:"value"      orm:"value"          description:"自定义值"`
	Comment    string `json:"comment"    orm:"comment"        description:"备注"`
}
