// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDictItem is the golang structure of table sys_dict_item for DAO operations like Where/Data.
type SysDictItem struct {
	g.Meta      `orm:"table:sys_dict_item, do:true"`
	Id          interface{} // 主键
	DictId      interface{} // 字典id
	ItemText    interface{} // 字典项文本
	ItemValue   interface{} // 字典项值
	Description interface{} // 描述
	SortOrder   interface{} // 排序
	Status      interface{} // 状态 1启用 0不启用
	CreateBy    interface{} // 创建人
	CreateTime  interface{} // 创建时间
	UpdateBy    interface{} // 更新人
	UpdateTime  interface{} // 更新时间
}
