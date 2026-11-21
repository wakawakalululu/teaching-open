// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysDepartRolePermissionDao is the data access object for table sys_depart_role_permission.
type SysDepartRolePermissionDao struct {
	table   string                         // table is the underlying table name of the DAO.
	group   string                         // group is the database configuration group name of current DAO.
	columns SysDepartRolePermissionColumns // columns contains all the column names of Table for convenient usage.
}

// SysDepartRolePermissionColumns defines and stores column names for table sys_depart_role_permission.
type SysDepartRolePermissionColumns struct {
	Id           string // ID
	DepartId     string // 部门id
	RoleId       string // 部门角色id
	PermissionId string // 菜单id
	DataRuleIds  string // 数据规则id串
}

// sysDepartRolePermissionColumns holds the columns for table sys_depart_role_permission.
var sysDepartRolePermissionColumns = SysDepartRolePermissionColumns{
	Id:           "id",
	DepartId:     "depart_id",
	RoleId:       "role_id",
	PermissionId: "permission_id",
	DataRuleIds:  "data_rule_ids",
}

// NewSysDepartRolePermissionDao creates and returns a new DAO object for table data access.
func NewSysDepartRolePermissionDao() *SysDepartRolePermissionDao {
	return &SysDepartRolePermissionDao{
		group:   "default",
		table:   "sys_depart_role_permission",
		columns: sysDepartRolePermissionColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysDepartRolePermissionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysDepartRolePermissionDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysDepartRolePermissionDao) Columns() SysDepartRolePermissionColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysDepartRolePermissionDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysDepartRolePermissionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysDepartRolePermissionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
