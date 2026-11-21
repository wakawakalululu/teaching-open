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
	Id           string // 主键
	Username     string // 登录账号
	Realname     string // 真实姓名
	Password     string // 密码散列/密文
	Salt         string // 盐
	Avatar       string // 头像
	Birthday     string // 生日
	Sex          string // 性别
	Email        string // 邮箱
	Phone        string // 手机号
	OrgCode      string // 组织编码
	Status       string // 状态 1启用 0禁用
	DelFlag      string // 删除标记 0正常
	ThirdId      string // 第三方登录id
	ThirdType    string // 第三方登录类型
	ActivitiSync string // 同步工作流标识
	WorkNo       string // 工号
	Post         string // 职务
	School       string // 学校
	Telephone    string // 座机号
	CreateBy     string // 创建人
	CreateTime   string // 创建时间
	UpdateBy     string // 更新人
	UpdateTime   string // 更新时间
	UserIdentity string // 身份 1普通 2上级
	DepartIds    string // 负责部门id串
}

// sysUserColumns holds the columns for table sys_user.
var sysUserColumns = SysUserColumns{
	Id:           "id",
	Username:     "username",
	Realname:     "realname",
	Password:     "password",
	Salt:         "salt",
	Avatar:       "avatar",
	Birthday:     "birthday",
	Sex:          "sex",
	Email:        "email",
	Phone:        "phone",
	OrgCode:      "org_code",
	Status:       "status",
	DelFlag:      "del_flag",
	ThirdId:      "third_id",
	ThirdType:    "third_type",
	ActivitiSync: "activiti_sync",
	WorkNo:       "work_no",
	Post:         "post",
	School:       "school",
	Telephone:    "telephone",
	CreateBy:     "create_by",
	CreateTime:   "create_time",
	UpdateBy:     "update_by",
	UpdateTime:   "update_time",
	UserIdentity: "user_identity",
	DepartIds:    "depart_ids",
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
