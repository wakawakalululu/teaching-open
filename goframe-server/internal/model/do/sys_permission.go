// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysPermission is the golang structure of table sys_permission for DAO operations like Where/Data.
type SysPermission struct {
	g.Meta             `orm:"table:sys_permission, do:true"`
	Id                 interface{} // 主键id
	ParentId           interface{} // 父id
	Name               interface{} // 菜单标题
	Url                interface{} // 路径
	Component          interface{} // 组件
	ComponentName      interface{} // 组件名字
	Redirect           interface{} // 一级菜单跳转地址
	MenuType           interface{} // 菜单类型 0一级菜单 1子菜单 2按钮权限
	Perms              interface{} // 菜单权限编码
	PermsType          interface{} // 权限策略 1显示 2禁用
	SortNo             interface{} // 菜单排序
	AlwaysShow         interface{} // 聚合子路由 1是 0否
	Icon               interface{} // 菜单图标
	IsRoute            interface{} // 是否路由菜单 0不是 1是
	IsLeaf             interface{} // 是否叶子节点 1是 0否
	KeepAlive          interface{} // 是否缓存该页面 1是 0否
	Hidden             interface{} // 是否隐藏路由 0否 1是
	Description        interface{} // 描述
	CreateBy           interface{} // 创建人
	CreateTime         interface{} // 创建时间
	UpdateBy           interface{} // 更新人
	UpdateTime         interface{} // 更新时间
	DelFlag            interface{} // 删除状态 0正常 1已删除
	RuleFlag           interface{} // 是否添加数据权限 1是 0否
	Status             interface{} // 按钮权限状态 0无效 1有效
	InternalOrExternal interface{} // 外链打开方式 0内部 1外部
}
