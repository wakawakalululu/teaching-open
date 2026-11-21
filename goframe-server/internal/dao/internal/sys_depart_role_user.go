// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysDepartRoleUserDao is the data access object for table sys_depart_role_user.
type SysDepartRoleUserDao struct {
	table   string                   // table is the underlying table name of the DAO.
	group   string                   // group is the database configuration group name of current DAO.
	columns SysDepartRoleUserColumns // columns contains all the column names of Table for convenient usage.
}

// SysDepartRoleUserColumns defines and stores column names for table sys_depart_role_user.
type SysDepartRoleUserColumns struct {
	Id      string // ID
	UserId  string // 用户id
	DroleId string // 部门角色id
}

// sysDepartRoleUserColumns holds the columns for table sys_depart_role_user.
var sysDepartRoleUserColumns = SysDepartRoleUserColumns{
	Id:      "id",
	UserId:  "user_id",
	DroleId: "drole_id",
}

// NewSysDepartRoleUserDao creates and returns a new DAO object for table data access.
func NewSysDepartRoleUserDao() *SysDepartRoleUserDao {
	return &SysDepartRoleUserDao{
		group:   "default",
		table:   "sys_depart_role_user",
		columns: sysDepartRoleUserColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysDepartRoleUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysDepartRoleUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysDepartRoleUserDao) Columns() SysDepartRoleUserColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysDepartRoleUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysDepartRoleUserDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysDepartRoleUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
