// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysDepartRoleDao is the data access object for table sys_depart_role.
type SysDepartRoleDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns SysDepartRoleColumns // columns contains all the column names of Table for convenient usage.
}

// SysDepartRoleColumns defines and stores column names for table sys_depart_role.
type SysDepartRoleColumns struct {
	Id          string // ID
	DepartId    string // 部门id
	RoleName    string // 角色名称
	RoleCode    string // 角色编码
	Description string // 描述
	CreateBy    string // 创建人
	CreateTime  string // 创建时间
	UpdateBy    string // 修改人
	UpdateTime  string // 修改时间
}

// sysDepartRoleColumns holds the columns for table sys_depart_role.
var sysDepartRoleColumns = SysDepartRoleColumns{
	Id:          "id",
	DepartId:    "depart_id",
	RoleName:    "role_name",
	RoleCode:    "role_code",
	Description: "description",
	CreateBy:    "create_by",
	CreateTime:  "create_time",
	UpdateBy:    "update_by",
	UpdateTime:  "update_time",
}

// NewSysDepartRoleDao creates and returns a new DAO object for table data access.
func NewSysDepartRoleDao() *SysDepartRoleDao {
	return &SysDepartRoleDao{
		group:   "default",
		table:   "sys_depart_role",
		columns: sysDepartRoleColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysDepartRoleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysDepartRoleDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysDepartRoleDao) Columns() SysDepartRoleColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysDepartRoleDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysDepartRoleDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysDepartRoleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
