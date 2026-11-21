// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysAnnouncementDao is the data access object for table sys_announcement.
type SysAnnouncementDao struct {
	table   string                 // table is the underlying table name of the DAO.
	group   string                 // group is the database configuration group name of current DAO.
	columns SysAnnouncementColumns // columns contains all the column names of Table for convenient usage.
}

// SysAnnouncementColumns defines and stores column names for table sys_announcement.
type SysAnnouncementColumns struct {
	Id          string // ID:id
	Titile      string // 标题:titile
	MsgContent  string // 内容:msg_content
	StartTime   string // 开始时间:start_time
	EndTime     string // 结束时间:end_time
	Sender      string // 发布人:sender
	Priority    string // 优先级:priority
	MsgCategory string // 消息分类:msg_category
	MsgType     string // 消息类型:msg_type
	SendStatus  string // 发布状态:send_status
	SendTime    string // 发布时间:send_time
	CancelTime  string // 撤销时间:cancel_time
	DelFlag     string // 删除状态:del_flag
	BusType     string // 业务类型:bus_type
	BusId       string // 业务id:bus_id
	OpenType    string // 打开方式:open_type
	OpenPage    string // 组件路由地址:open_page
	CreateBy    string // 创建人:create_by
	CreateTime  string // 创建时间:create_time
	UpdateBy    string // 更新人:update_by
	UpdateTime  string // 更新时间:update_time
	MsgAbstract string // 摘要:msg_abstract
}

// sysannouncementColumns holds the columns for table sys_announcement.
var sysannouncementColumns = SysAnnouncementColumns{
	Id:          "id",
	Titile:      "titile",
	MsgContent:  "msg_content",
	StartTime:   "start_time",
	EndTime:     "end_time",
	Sender:      "sender",
	Priority:    "priority",
	MsgCategory: "msg_category",
	MsgType:     "msg_type",
	SendStatus:  "send_status",
	SendTime:    "send_time",
	CancelTime:  "cancel_time",
	DelFlag:     "del_flag",
	BusType:     "bus_type",
	BusId:       "bus_id",
	OpenType:    "open_type",
	OpenPage:    "open_page",
	CreateBy:    "create_by",
	CreateTime:  "create_time",
	UpdateBy:    "update_by",
	UpdateTime:  "update_time",
	MsgAbstract: "msg_abstract",
}

// NewSysAnnouncementDao creates and returns a new DAO object for table data access.
func NewSysAnnouncementDao() *SysAnnouncementDao {
	return &SysAnnouncementDao{
		group:   "default",
		table:   "sys_announcement",
		columns: sysannouncementColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysAnnouncementDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysAnnouncementDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysAnnouncementDao) Columns() SysAnnouncementColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysAnnouncementDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysAnnouncementDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysAnnouncementDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
