// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPermissionDao is the data access object for table sys_permission.
type SysPermissionDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns SysPermissionColumns // columns contains all the column names of Table for convenient usage.
}

// SysPermissionColumns defines and stores the column names for table sys_permission.
type SysPermissionColumns struct {
	Id                 string // 主键id
	ParentId           string // 父id
	Name               string // 菜单标题
	Url                string // 路径
	Component          string // 组件
	ComponentName      string // 组件名字
	Redirect           string // 一级菜单跳转地址
	MenuType           string // 菜单类型 0一级菜单 1子菜单 2按钮权限
	Perms              string // 菜单权限编码
	PermsType          string // 权限策略 1显示 2禁用
	SortNo             string // 菜单排序
	AlwaysShow         string // 聚合子路由 1是 0否
	Icon               string // 菜单图标
	IsRoute            string // 是否路由菜单 0不是 1是
	IsLeaf             string // 是否叶子节点 1是 0否
	KeepAlive          string // 是否缓存该页面 1是 0否
	Hidden             string // 是否隐藏路由 0否 1是
	Description        string // 描述
	CreateBy           string // 创建人
	CreateTime         string // 创建时间
	UpdateBy           string // 更新人
	UpdateTime         string // 更新时间
	DelFlag            string // 删除状态 0正常 1已删除
	RuleFlag           string // 是否添加数据权限 1是 0否
	Status             string // 按钮权限状态 0无效 1有效
	InternalOrExternal string // 外链打开方式 0内部 1外部
}

// sysPermissionColumns holds the columns for table sys_permission.
var sysPermissionColumns = SysPermissionColumns{
	Id:                 "id",
	ParentId:           "parent_id",
	Name:               "name",
	Url:                "url",
	Component:          "component",
	ComponentName:      "component_name",
	Redirect:           "redirect",
	MenuType:           "menu_type",
	Perms:              "perms",
	PermsType:          "perms_type",
	SortNo:             "sort_no",
	AlwaysShow:         "always_show",
	Icon:               "icon",
	IsRoute:            "is_route",
	IsLeaf:             "is_leaf",
	KeepAlive:          "keep_alive",
	Hidden:             "hidden",
	Description:        "description",
	CreateBy:           "create_by",
	CreateTime:         "create_time",
	UpdateBy:           "update_by",
	UpdateTime:         "update_time",
	DelFlag:            "del_flag",
	RuleFlag:           "rule_flag",
	Status:             "status",
	InternalOrExternal: "internal_or_external",
}

// NewSysPermissionDao creates and returns a new DAO object for table data access.
func NewSysPermissionDao() *SysPermissionDao {
	return &SysPermissionDao{
		group:   "default",
		table:   "sys_permission",
		columns: sysPermissionColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysPermissionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysPermissionDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysPermissionDao) Columns() SysPermissionColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysPermissionDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysPermissionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysPermissionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
