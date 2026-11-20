// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysCategoryDao is the data access object for table sys_category.
type SysCategoryDao struct {
	table   string             // table is the underlying table name of the DAO.
	group   string             // group is the database configuration group name of current DAO.
	columns SysCategoryColumns // columns contains all the column names of Table for convenient usage.
}

// SysCategoryColumns defines and stores the column names for table sys_category.
type SysCategoryColumns struct {
	Id         string // 主键
	Pid        string // 父级节点
	Name       string // 类型名称
	Code       string // 类型编码
	CreateBy   string // 创建人
	CreateTime string // 创建日期
	UpdateBy   string // 更新人
	UpdateTime string // 更新日期
	SysOrgCode string // 所属部门
	HasChild   string // 是否有子节点
	Value      string // 自定义值
	Comment    string // 备注
}

// sysCategoryColumns holds the columns for table sys_category.
var sysCategoryColumns = SysCategoryColumns{
	Id:         "id",
	Pid:        "pid",
	Name:       "name",
	Code:       "code",
	CreateBy:   "create_by",
	CreateTime: "create_time",
	UpdateBy:   "update_by",
	UpdateTime: "update_time",
	SysOrgCode: "sys_org_code",
	HasChild:   "has_child",
	Value:      "value",
	Comment:    "comment",
}

// NewSysCategoryDao creates and returns a new DAO object for table data access.
func NewSysCategoryDao() *SysCategoryDao {
	return &SysCategoryDao{
		group:   "default",
		table:   "sys_category",
		columns: sysCategoryColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysCategoryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysCategoryDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysCategoryDao) Columns() SysCategoryColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysCategoryDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysCategoryDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysCategoryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
