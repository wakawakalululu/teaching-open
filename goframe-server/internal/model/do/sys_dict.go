// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDict is the golang structure of table sys_dict for DAO operations like Where/Data.
type SysDict struct {
	g.Meta      `orm:"table:sys_dict, do:true"`
	Id          interface{} // 主键
	DictName    interface{} // 字典名称
	DictCode    interface{} // 字典编码
	Description interface{} // 描述
	DelFlag     interface{} // 删除状态 0正常 1已删除
	CreateBy    interface{} // 创建人
	CreateTime  interface{} // 创建时间
	UpdateBy    interface{} // 更新人
	UpdateTime  interface{} // 更新时间
	Type        interface{} // 字典类型 0为string 1为number
}
