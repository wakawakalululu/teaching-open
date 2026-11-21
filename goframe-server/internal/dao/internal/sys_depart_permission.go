// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysDepartPermissionDao is the data access object for table sys_depart_permission.
type SysDepartPermissionDao struct {
	table   string                     // table is the underlying table name of the DAO.
	group   string                     // group is the database configuration group name of current DAO.
	columns SysDepartPermissionColumns // columns contains all the column names of Table for convenient usage.
}

// SysDepartPermissionColumns defines and stores column names for table sys_depart_permission.
type SysDepartPermissionColumns struct {
	Id           string // ID
	DepartId     string // 部门id
	PermissionId string // 权限id
	DataRuleIds  string // 数据规则id串
}

// sysDepartPermissionColumns holds the columns for table sys_depart_permission.
var sysDepartPermissionColumns = SysDepartPermissionColumns{
	Id:           "id",
	DepartId:     "depart_id",
	PermissionId: "permission_id",
	DataRuleIds:  "data_rule_ids",
}

// NewSysDepartPermissionDao creates and returns a new DAO object for table data access.
func NewSysDepartPermissionDao() *SysDepartPermissionDao {
	return &SysDepartPermissionDao{
		group:   "default",
		table:   "sys_depart_permission",
		columns: sysDepartPermissionColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysDepartPermissionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysDepartPermissionDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysDepartPermissionDao) Columns() SysDepartPermissionColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysDepartPermissionDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysDepartPermissionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysDepartPermissionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
