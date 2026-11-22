// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OssFileDao is the data access object for table oss_file.
type OssFileDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns OssFileColumns // columns contains all the column names of Table for convenient usage.
}

// OssFileColumns defines and stores column names for table oss_file.
type OssFileColumns struct {
	Id         string // 主键id:id
	FileName   string // 文件名称:file_name
	Url        string // 文件地址:url
	CreateBy   string // 创建人:create_by
	CreateTime string // 创建时间:create_time
	UpdateBy   string // 更新人:update_by
	UpdateTime string // 更新时间:update_time
}

// ossfileColumns holds the columns for table oss_file.
var ossfileColumns = OssFileColumns{
	Id:         "id",
	FileName:   "file_name",
	Url:        "url",
	CreateBy:   "create_by",
	CreateTime: "create_time",
	UpdateBy:   "update_by",
	UpdateTime: "update_time",
}

// NewOssFileDao creates and returns a new DAO object for table data access.
func NewOssFileDao() *OssFileDao {
	return &OssFileDao{
		group:   "default",
		table:   "oss_file",
		columns: ossfileColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *OssFileDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *OssFileDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *OssFileDao) Columns() OssFileColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *OssFileDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *OssFileDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *OssFileDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
