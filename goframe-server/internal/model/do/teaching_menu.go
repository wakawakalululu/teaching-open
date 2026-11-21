// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingMenu is the golang structure of table teaching_menu for DAO operations like Where/Data.
type TeachingMenu struct {
	g.Meta             `orm:"table:teaching_menu, do:true"`
	Id                 interface{} // ID
	ParentId           interface{} // 父菜单id
	Name               interface{} // 菜单标题
	Icon               interface{} // 菜单图标
	Url                interface{} // 菜单路径
	SortNo             interface{} // 排序
	MenuType           interface{} // 菜单类型 0一级 1子菜单
	Leaf               interface{} // 是否叶子节点
	Route              interface{} // 是否生成路由
	Description        interface{} // 描述
	CreateBy           interface{} // 创建人
	DelFlag            interface{} // 删除状态
	Hidden             interface{} // 是否隐藏
	NeedLogin          interface{} // 是否需要登录
	CreateTime         interface{} // 创建时间
	UpdateBy           interface{} // 更新人
	UpdateTime         interface{} // 更新时间
	InternalOrExternal interface{} // 外链打开方式
}
