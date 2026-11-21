// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserDepartDao is the data access object for table sys_user_depart.
type SysUserDepartDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns SysUserDepartColumns // columns contains all the column names of Table for convenient usage.
}

// SysUserDepartColumns defines and stores the column names for table sys_user_depart.
// 列名以 api/db/teachingopen2.8.sql L913 为准：主键列为大写 `ID`（无 del_flag、无时间列）。
type SysUserDepartColumns struct {
	Id     string // 主键id（大写 ID）
	UserId string // 用户id
	DepId  string // 部门id
}

// sysUserDepartColumns holds the columns for table sys_user_depart.
var sysUserDepartColumns = SysUserDepartColumns{
	Id:     "ID",
	UserId: "user_id",
	DepId:  "dep_id",
}

// NewSysUserDepartDao creates and returns a new DAO object for table data access.
func NewSysUserDepartDao() *SysUserDepartDao {
	return &SysUserDepartDao{
		group:   "default",
		table:   "sys_user_depart",
		columns: sysUserDepartColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysUserDepartDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysUserDepartDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysUserDepartDao) Columns() SysUserDepartColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysUserDepartDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysUserDepartDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysUserDepartDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
