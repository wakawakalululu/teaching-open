// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysDictItem 对应 sys_dict_item 表。
type SysDictItem struct {
	Id          string `json:"id"          orm:"id,primary"    description:"主键"`
	DictId      string `json:"dictId"      orm:"dict_id"       description:"字典id"`
	ItemText    string `json:"itemText"    orm:"item_text"     description:"字典项文本"`
	ItemValue   string `json:"itemValue"   orm:"item_value"    description:"字典项值"`
	Description string `json:"description" orm:"description"   description:"描述"`
	SortOrder   int    `json:"sortOrder"   orm:"sort_order"    description:"排序"`
	Status      int    `json:"status"      orm:"status"        description:"状态 1启用 0不启用"`
	CreateBy    string `json:"createBy"    orm:"create_by"     description:"创建人"`
	CreateTime  string `json:"createTime"  orm:"create_time"   description:"创建时间"`
	UpdateBy    string `json:"updateBy"    orm:"update_by"     description:"更新人"`
	UpdateTime  string `json:"updateTime"  orm:"update_time"   description:"更新时间"`
}
