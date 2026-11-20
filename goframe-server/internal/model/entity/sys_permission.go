// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysPermission 对应 sys_permission 表（菜单权限）。
type SysPermission struct {
	Id                 string  `json:"id"                orm:"id,primary"           description:"主键id"`
	ParentId           string  `json:"parentId"          orm:"parent_id"            description:"父id"`
	Name               string  `json:"name"              orm:"name"                 description:"菜单标题"`
	Url                string  `json:"url"               orm:"url"                  description:"路径"`
	Component          string  `json:"component"         orm:"component"            description:"组件"`
	ComponentName      string  `json:"componentName"     orm:"component_name"       description:"组件名字"`
	Redirect           string  `json:"redirect"          orm:"redirect"             description:"一级菜单跳转地址"`
	MenuType           int     `json:"menuType"          orm:"menu_type"            description:"菜单类型 0一级菜单 1子菜单 2按钮权限"`
	Perms              string  `json:"perms"             orm:"perms"                description:"菜单权限编码"`
	PermsType          string  `json:"permsType"         orm:"perms_type"           description:"权限策略 1显示 2禁用"`
	SortNo             float64 `json:"sortNo"            orm:"sort_no"              description:"菜单排序"`
	AlwaysShow         int     `json:"alwaysShow"        orm:"always_show"          description:"聚合子路由 1是 0否"`
	Icon               string  `json:"icon"              orm:"icon"                 description:"菜单图标"`
	IsRoute            int     `json:"isRoute"           orm:"is_route"             description:"是否路由菜单 0不是 1是"`
	IsLeaf             int     `json:"isLeaf"            orm:"is_leaf"              description:"是否叶子节点 1是 0否"`
	KeepAlive          int     `json:"keepAlive"         orm:"keep_alive"           description:"是否缓存该页面 1是 0否"`
	Hidden             int     `json:"hidden"            orm:"hidden"               description:"是否隐藏路由 0否 1是"`
	Description        string  `json:"description"       orm:"description"          description:"描述"`
	CreateBy           string  `json:"createBy"          orm:"create_by"            description:"创建人"`
	CreateTime         string  `json:"createTime"        orm:"create_time"          description:"创建时间"`
	UpdateBy           string  `json:"updateBy"          orm:"update_by"            description:"更新人"`
	UpdateTime         string  `json:"updateTime"        orm:"update_time"          description:"更新时间"`
	DelFlag            int     `json:"delFlag"           orm:"del_flag"             description:"删除状态 0正常 1已删除"`
	RuleFlag           int     `json:"ruleFlag"          orm:"rule_flag"            description:"是否添加数据权限 1是 0否"`
	Status             string  `json:"status"            orm:"status"               description:"按钮权限状态 0无效 1有效"`
	InternalOrExternal int     `json:"internalOrExternal" orm:"internal_or_external" description:"外链打开方式 0内部 1外部"`
}
