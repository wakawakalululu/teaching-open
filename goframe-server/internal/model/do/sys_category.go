// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysCategory is the golang structure of table sys_category for DAO operations like Where/Data.
type SysCategory struct {
	g.Meta     `orm:"table:sys_category, do:true"`
	Id         interface{} // 主键
	Pid        interface{} // 父级节点
	Name       interface{} // 类型名称
	Code       interface{} // 类型编码
	CreateBy   interface{} // 创建人
	CreateTime interface{} // 创建日期
	UpdateBy   interface{} // 更新人
	UpdateTime interface{} // 更新日期
	SysOrgCode interface{} // 所属部门
	HasChild   interface{} // 是否有子节点
	Value      interface{} // 自定义值
	Comment    interface{} // 备注
}
