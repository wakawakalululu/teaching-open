// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysAnnouncementSendDao is the data access object for table sys_announcement_send.
type SysAnnouncementSendDao struct {
	table   string                     // table is the underlying table name of the DAO.
	group   string                     // group is the database configuration group name of current DAO.
	columns SysAnnouncementSendColumns // columns contains all the column names of Table for convenient usage.
}

// SysAnnouncementSendColumns defines and stores column names for table sys_announcement_send.
type SysAnnouncementSendColumns struct {
	Id         string // ID:id
	AnntId     string // 通告id:annt_id
	UserId     string // 用户id:user_id
	ReadFlag   string // 已读标记:read_flag
	ReadTime   string // 阅读时间:read_time
	CreateBy   string // 创建人:create_by
	CreateTime string // 创建时间:create_time
	UpdateBy   string // 更新人:update_by
	UpdateTime string // 更新时间:update_time
}

// sysannouncementsendColumns holds the columns for table sys_announcement_send.
var sysannouncementsendColumns = SysAnnouncementSendColumns{
	Id:         "id",
	AnntId:     "annt_id",
	UserId:     "user_id",
	ReadFlag:   "read_flag",
	ReadTime:   "read_time",
	CreateBy:   "create_by",
	CreateTime: "create_time",
	UpdateBy:   "update_by",
	UpdateTime: "update_time",
}

// NewSysAnnouncementSendDao creates and returns a new DAO object for table data access.
func NewSysAnnouncementSendDao() *SysAnnouncementSendDao {
	return &SysAnnouncementSendDao{
		group:   "default",
		table:   "sys_announcement_send",
		columns: sysannouncementsendColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysAnnouncementSendDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysAnnouncementSendDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysAnnouncementSendDao) Columns() SysAnnouncementSendColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysAnnouncementSendDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysAnnouncementSendDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysAnnouncementSendDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
