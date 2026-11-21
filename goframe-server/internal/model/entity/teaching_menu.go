// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingMenu 对应 teaching_menu 表（api/db/teachingopen2.8.sql：
// 门户菜单；leaf/route/hidden/internalOrExternal 为 tinyint 布尔语义）。
type TeachingMenu struct {
	Id                 string  `json:"id"                 orm:"id,primary"          description:"ID"`
	ParentId           string  `json:"parentId"           orm:"parent_id"           description:"父菜单id"`
	Name               string  `json:"name"               orm:"name"                description:"菜单标题"`
	Icon               string  `json:"icon"               orm:"icon"                description:"菜单图标"`
	Url                string  `json:"url"                orm:"url"                 description:"菜单路径"`
	SortNo             float64 `json:"sortNo"             orm:"sort_no"             description:"排序"`
	MenuType           int     `json:"menuType"           orm:"menu_type"           description:"菜单类型 0一级 1子菜单"`
	Leaf               bool    `json:"leaf"               orm:"is_leaf"             description:"是否叶子节点"`
	Route              bool    `json:"route"              orm:"is_route"            description:"是否生成路由"`
	Description        string  `json:"description"        orm:"description"         description:"描述"`
	CreateBy           string  `json:"createBy"           orm:"create_by"           description:"创建人"`
	DelFlag            int     `json:"delFlag"            orm:"del_flag"            description:"删除状态"`
	Hidden             bool    `json:"hidden"             orm:"hidden"              description:"是否隐藏"`
	NeedLogin          int     `json:"needLogin"          orm:"need_login"          description:"是否需要登录"`
	CreateTime         string  `json:"createTime"         orm:"create_time"         description:"创建时间"`
	UpdateBy           string  `json:"updateBy"           orm:"update_by"           description:"更新人"`
	UpdateTime         string  `json:"updateTime"         orm:"update_time"         description:"更新时间"`
	InternalOrExternal bool    `json:"internalOrExternal" orm:"internal_or_external" description:"外链打开方式"`
}
