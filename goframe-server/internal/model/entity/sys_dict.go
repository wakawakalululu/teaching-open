// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysDict 对应 sys_dict 表（列结构以 api/db/teachingopen2.8.sql 为准）。
type SysDict struct {
	Id          string `json:"id"          orm:"id,primary"    description:"主键"`
	DictName    string `json:"dictName"    orm:"dict_name"     description:"字典名称"`
	DictCode    string `json:"dictCode"    orm:"dict_code"     description:"字典编码"`
	Description string `json:"description" orm:"description"   description:"描述"`
	DelFlag     int    `json:"delFlag"     orm:"del_flag"      description:"删除状态 0正常 1已删除"`
	CreateBy    string `json:"createBy"    orm:"create_by"     description:"创建人"`
	CreateTime  string `json:"createTime"  orm:"create_time"   description:"创建时间"`
	UpdateBy    string `json:"updateBy"    orm:"update_by"     description:"更新人"`
	UpdateTime  string `json:"updateTime"  orm:"update_time"   description:"更新时间"`
	Type        int    `json:"type"        orm:"type"          description:"字典类型 0为string 1为number"`
}
