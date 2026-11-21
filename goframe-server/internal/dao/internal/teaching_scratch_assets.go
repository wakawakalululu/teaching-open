// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingScratchAssetsDao is the data access object for table teaching_scratch_assets.
type TeachingScratchAssetsDao struct {
	table   string                       // table is the underlying table name of the DAO.
	group   string                       // group is the database configuration group name of current DAO.
	columns TeachingScratchAssetsColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingScratchAssetsColumns defines and stores column names for table teaching_scratch_assets.
type TeachingScratchAssetsColumns struct {
	Id         string // ID:id
	AssetType  string // 素材类型:asset_type
	AssetName  string // 素材名称:asset_name
	AssetData  string // 素材元数据JSON:asset_data
	Md5Ext     string // md5扩展名:md5_ext
	Tags       string // 标签:tags
	CreateBy   string // 创建人:create_by
	CreateTime string // 创建时间:create_time
	UpdateBy   string // 更新人:update_by
	UdpateTime string // 更新时间:udpate_time
	DelFlag    string // 删除标记:del_flag
}

// teachingscratchassetsColumns holds the columns for table teaching_scratch_assets.
var teachingscratchassetsColumns = TeachingScratchAssetsColumns{
	Id:         "id",
	AssetType:  "asset_type",
	AssetName:  "asset_name",
	AssetData:  "asset_data",
	Md5Ext:     "md5_ext",
	Tags:       "tags",
	CreateBy:   "create_by",
	CreateTime: "create_time",
	UpdateBy:   "update_by",
	UdpateTime: "udpate_time",
	DelFlag:    "del_flag",
}

// NewTeachingScratchAssetsDao creates and returns a new DAO object for table data access.
func NewTeachingScratchAssetsDao() *TeachingScratchAssetsDao {
	return &TeachingScratchAssetsDao{
		group:   "default",
		table:   "teaching_scratch_assets",
		columns: teachingscratchassetsColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingScratchAssetsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingScratchAssetsDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingScratchAssetsDao) Columns() TeachingScratchAssetsColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingScratchAssetsDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingScratchAssetsDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingScratchAssetsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
