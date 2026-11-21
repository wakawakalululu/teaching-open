// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserAgentDao is the data access object for table sys_user_agent.
type SysUserAgentDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns SysUserAgentColumns // columns contains all the column names of Table for convenient usage.
}

// SysUserAgentColumns defines and stores column names for table sys_user_agent.
type SysUserAgentColumns struct {
	Id            string // ID:id
	UserName      string // 用户名:user_name
	AgentUserName string // 代理人用户名:agent_user_name
	StartTime     string // 代理开始时间:start_time
	EndTime       string // 代理结束时间:end_time
	Status        string // 状态:status
	CreateName    string // 创建人姓名:create_name
	CreateBy      string // 创建人:create_by
	CreateTime    string // 创建时间:create_time
	UpdateName    string // 更新人姓名:update_name
	UpdateBy      string // 更新人:update_by
	UpdateTime    string // 更新时间:update_time
}

// sysuseragentColumns holds the columns for table sys_user_agent.
var sysuseragentColumns = SysUserAgentColumns{
	Id:            "id",
	UserName:      "user_name",
	AgentUserName: "agent_user_name",
	StartTime:     "start_time",
	EndTime:       "end_time",
	Status:        "status",
	CreateName:    "create_name",
	CreateBy:      "create_by",
	CreateTime:    "create_time",
	UpdateName:    "update_name",
	UpdateBy:      "update_by",
	UpdateTime:    "update_time",
}

// NewSysUserAgentDao creates and returns a new DAO object for table data access.
func NewSysUserAgentDao() *SysUserAgentDao {
	return &SysUserAgentDao{
		group:   "default",
		table:   "sys_user_agent",
		columns: sysuseragentColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysUserAgentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysUserAgentDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysUserAgentDao) Columns() SysUserAgentColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysUserAgentDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysUserAgentDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysUserAgentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
