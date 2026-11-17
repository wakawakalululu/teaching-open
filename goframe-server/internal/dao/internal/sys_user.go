// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysUserDao is the data access object for table sys_user.
type SysUserDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns SysUserColumns // columns contains all the column names of Table for convenient usage.
}

// SysUserColumns defines and stores column names for table sys_user.
type SysUserColumns struct {
	Id         string // 主键
	Username   string // 登录账号
	Realname   string // 真实姓名
	Password   string // 密码散列/密文
	Salt       string // 盐
	Avatar     string // 头像
	Email      string // 邮箱
	Phone      string // 手机号
	Status     string // 状态 1启用 0禁用
	DelFlag    string // 删除标记 0正常
	OrgCode    string // 组织编码
	CreateTime string // 创建时间
}

// sysUserColumns holds the columns for table sys_user.
var sysUserColumns = SysUserColumns{
	Id:         "id",
	Username:   "username",
	Realname:   "realname",
	Password:   "password",
	Salt:       "salt",
	Avatar:     "avatar",
	Email:      "email",
	Phone:      "phone",
	Status:     "status",
	DelFlag:    "del_flag",
	OrgCode:    "org_code",
	CreateTime: "create_time",
}

// NewSysUserDao creates and returns a new DAO object for table data access.
func NewSysUserDao() *SysUserDao {
	return &SysUserDao{
		group:   "default",
		table:   "sys_user",
		columns: sysUserColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysUserDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysUserDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysUserDao) Columns() SysUserColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysUserDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysUserDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
