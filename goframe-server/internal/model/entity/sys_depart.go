// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysDepart 对应 sys_depart 表（列结构以 api/db/teachingopen2.8.sql L570 为准；
// JSON 字段名对齐 Java org.jeecg.modules.system.entity.SysDepart 的 Jackson 序列化）。
type SysDepart struct {
	Id             string `json:"id"             orm:"id,primary"        description:"ID"`
	ParentId       string `json:"parentId"       orm:"parent_id"         description:"父机构ID"`
	DepartName     string `json:"departName"     orm:"depart_name"       description:"机构/部门名称"`
	DepartNameEn   string `json:"departNameEn"   orm:"depart_name_en"    description:"英文名"`
	DepartNameAbbr string `json:"departNameAbbr" orm:"depart_name_abbr"  description:"缩写"`
	DepartOrder    int    `json:"departOrder"    orm:"depart_order"      description:"排序"`
	Description    string `json:"description"    orm:"description"       description:"描述"`
	OrgCategory    string `json:"orgCategory"    orm:"org_category"      description:"机构类别 1组织机构，2岗位"`
	OrgType        string `json:"orgType"        orm:"org_type"          description:"机构类型 1一级部门 2子部门"`
	OrgCode        string `json:"orgCode"        orm:"org_code"          description:"机构编码"`
	Mobile         string `json:"mobile"         orm:"mobile"            description:"手机号"`
	Fax            string `json:"fax"            orm:"fax"               description:"传真"`
	Address        string `json:"address"        orm:"address"           description:"地址"`
	Memo           string `json:"memo"           orm:"memo"              description:"备注"`
	Status         string `json:"status"         orm:"status"            description:"状态（1启用，0不启用）"`
	DelFlag        string `json:"delFlag"        orm:"del_flag"          description:"删除状态（0，正常，1已删除）"`
	CreateBy       string `json:"createBy"       orm:"create_by"         description:"创建人"`
	CreateTime     string `json:"createTime"     orm:"create_time"       description:"创建日期"`
	UpdateBy       string `json:"updateBy"       orm:"update_by"         description:"更新人"`
	UpdateTime     string `json:"updateTime"     orm:"update_time"       description:"更新日期"`
}
