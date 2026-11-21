// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysPermissionDataRuleDao is the data access object for table sys_permission_data_rule.
type SysPermissionDataRuleDao struct {
	table   string                       // table is the underlying table name of the DAO.
	group   string                       // group is the database configuration group name of current DAO.
	columns SysPermissionDataRuleColumns // columns contains all the column names of Table for convenient usage.
}

// SysPermissionDataRuleColumns defines and stores column names for table sys_permission_data_rule.
type SysPermissionDataRuleColumns struct {
	Id             string // ID
	PermissionId   string // 菜单ID
	RuleName       string // 规则名称
	RuleColumn     string // 字段
	RuleConditions string // 条件
	RuleValue      string // 规则值
	Status         string // 权限有效状态1有0否
	CreateTime     string // 创建时间
	CreateBy       string // 创建人
	UpdateTime     string // 修改时间
	UpdateBy       string // 修改人
}

// sysPermissionDataRuleColumns holds the columns for table sys_permission_data_rule.
var sysPermissionDataRuleColumns = SysPermissionDataRuleColumns{
	Id:             "id",
	PermissionId:   "permission_id",
	RuleName:       "rule_name",
	RuleColumn:     "rule_column",
	RuleConditions: "rule_conditions",
	RuleValue:      "rule_value",
	Status:         "status",
	CreateTime:     "create_time",
	CreateBy:       "create_by",
	UpdateTime:     "update_time",
	UpdateBy:       "update_by",
}

// NewSysPermissionDataRuleDao creates and returns a new DAO object for table data access.
func NewSysPermissionDataRuleDao() *SysPermissionDataRuleDao {
	return &SysPermissionDataRuleDao{
		group:   "default",
		table:   "sys_permission_data_rule",
		columns: sysPermissionDataRuleColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysPermissionDataRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysPermissionDataRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysPermissionDataRuleDao) Columns() SysPermissionDataRuleColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysPermissionDataRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysPermissionDataRuleDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysPermissionDataRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
