// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysConfigDao is the data access object for table sys_config.
type SysConfigDao struct {
	table   string           // table is the underlying table name of the DAO.
	group   string           // group is the database configuration group name of current DAO.
	columns SysConfigColumns // columns contains all the column names of Table for convenient usage.
}

// SysConfigColumns defines and stores the column names for table sys_config.
type SysConfigColumns struct {
	Id            string // 主键
	ConfigKey     string // 配置属性
	ConfigValue   string // 配置值
	ConfigEnabled string // 是否生效 1生效 0失效
	Comment       string // 备注
}

// sysConfigColumns holds the columns for table sys_config.
var sysConfigColumns = SysConfigColumns{
	Id:            "id",
	ConfigKey:     "config_key",
	ConfigValue:   "config_value",
	ConfigEnabled: "config_enabled",
	Comment:       "comment",
}

// NewSysConfigDao creates and returns a new DAO object for table data access.
func NewSysConfigDao() *SysConfigDao {
	return &SysConfigDao{
		group:   "default",
		table:   "sys_config",
		columns: sysConfigColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysConfigDao) Columns() SysConfigColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysConfigDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
