// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysDepart is the golang structure of table sys_depart for DAO operations like Where/Data.
type SysDepart struct {
	g.Meta         `orm:"table:sys_depart, do:true"`
	Id             interface{} // ID
	ParentId       interface{} // 父机构ID
	DepartName     interface{} // 机构/部门名称
	DepartNameEn   interface{} // 英文名
	DepartNameAbbr interface{} // 缩写
	DepartOrder    interface{} // 排序
	Description    interface{} // 描述
	OrgCategory    interface{} // 机构类别 1组织机构，2岗位
	OrgType        interface{} // 机构类型 1一级部门 2子部门
	OrgCode        interface{} // 机构编码
	Mobile         interface{} // 手机号
	Fax            interface{} // 传真
	Address        interface{} // 地址
	Memo           interface{} // 备注
	Status         interface{} // 状态（1启用，0不启用）
	DelFlag        interface{} // 删除状态（0，正常，1已删除）
	CreateBy       interface{} // 创建人
	CreateTime     interface{} // 创建日期
	UpdateBy       interface{} // 更新人
	UpdateTime     interface{} // 更新日期
}
