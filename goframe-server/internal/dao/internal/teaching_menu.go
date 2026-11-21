// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingMenuDao is the data access object for table teaching_menu.
type TeachingMenuDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns TeachingMenuColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingMenuColumns defines and stores column names for table teaching_menu.
type TeachingMenuColumns struct {
	Id                 string // ID
	ParentId           string // 父菜单id
	Name               string // 菜单标题
	Icon               string // 菜单图标
	Url                string // 菜单路径
	SortNo             string // 排序
	MenuType           string // 菜单类型 0一级 1子菜单
	Leaf               string // 是否叶子节点
	Route              string // 是否生成路由
	Description        string // 描述
	CreateBy           string // 创建人
	DelFlag            string // 删除状态
	Hidden             string // 是否隐藏
	NeedLogin          string // 是否需要登录
	CreateTime         string // 创建时间
	UpdateBy           string // 更新人
	UpdateTime         string // 更新时间
	InternalOrExternal string // 外链打开方式
}

// teachingMenuColumns holds the columns for table teaching_menu.
var teachingMenuColumns = TeachingMenuColumns{
	Id:                 "id",
	ParentId:           "parent_id",
	Name:               "name",
	Icon:               "icon",
	Url:                "url",
	SortNo:             "sort_no",
	MenuType:           "menu_type",
	Leaf:               "is_leaf",
	Route:              "is_route",
	Description:        "description",
	CreateBy:           "create_by",
	DelFlag:            "del_flag",
	Hidden:             "hidden",
	NeedLogin:          "need_login",
	CreateTime:         "create_time",
	UpdateBy:           "update_by",
	UpdateTime:         "update_time",
	InternalOrExternal: "internal_or_external",
}

// NewTeachingMenuDao creates and returns a new DAO object for table data access.
func NewTeachingMenuDao() *TeachingMenuDao {
	return &TeachingMenuDao{
		group:   "default",
		table:   "teaching_menu",
		columns: teachingMenuColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingMenuDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingMenuDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingMenuDao) Columns() TeachingMenuColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingMenuDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingMenuDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingMenuDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
