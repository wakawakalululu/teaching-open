// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysFileDao is the data access object for table sys_file.
type SysFileDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns SysFileColumns // columns contains all the column names of Table for convenient usage.
}

// SysFileColumns defines and stores column names for table sys_file.
type SysFileColumns struct {
	Id           string // ID
	CreateTime   string // 创建时间
	UpdateBy     string // 更新人
	CreateBy     string // 创建人
	UpdateTime   string // 更新时间
	SysOrgCode   string // 所属部门
	FileType     string // 文件类型
	FileName     string // 文件名称
	FilePath     string // 文件路径/key
	FileLocation string // 文件位置 1本地 2七牛
	FileTag      string // 文件标签
	DelFlag      string // 删除状态
}

// sysFileColumns holds the columns for table sys_file.
var sysFileColumns = SysFileColumns{
	Id:           "id",
	CreateTime:   "create_time",
	UpdateBy:     "update_by",
	CreateBy:     "create_by",
	UpdateTime:   "update_time",
	SysOrgCode:   "sys_org_code",
	FileType:     "file_type",
	FileName:     "file_name",
	FilePath:     "file_path",
	FileLocation: "file_location",
	FileTag:      "file_tag",
	DelFlag:      "del_flag",
}

// NewSysFileDao creates and returns a new DAO object for table data access.
func NewSysFileDao() *SysFileDao {
	return &SysFileDao{
		group:   "default",
		table:   "sys_file",
		columns: sysFileColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysFileDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysFileDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysFileDao) Columns() SysFileColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysFileDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysFileDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysFileDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
