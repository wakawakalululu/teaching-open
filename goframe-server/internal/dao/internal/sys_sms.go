// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysSmsDao is the data access object for table sys_sms.
type SysSmsDao struct {
	table   string        // table is the underlying table name of the DAO.
	group   string        // group is the database configuration group name of current DAO.
	columns SysSmsColumns // columns contains all the column names of Table for convenient usage.
}

// SysSmsColumns defines and stores column names for table sys_sms.
type SysSmsColumns struct {
	Id           string // ID:id
	EsTitle      string // 消息标题:es_title
	EsType       string // 发送方式:es_type
	EsReceiver   string // 接收人:es_receiver
	EsParam      string // 消息参数:es_param
	EsContent    string // 消息内容:es_content
	EsSendTime   string // 发送时间:es_send_time
	EsSendStatus string // 发送状态:es_send_status
	EsSendNum    string // 发送次数:es_send_num
	EsResult     string // 发送结果:es_result
	Remark       string // 备注:remark
	CreateBy     string // 创建人:create_by
	CreateTime   string // 创建时间:create_time
	UpdateBy     string // 更新人:update_by
}

// syssmsColumns holds the columns for table sys_sms.
var syssmsColumns = SysSmsColumns{
	Id:           "id",
	EsTitle:      "es_title",
	EsType:       "es_type",
	EsReceiver:   "es_receiver",
	EsParam:      "es_param",
	EsContent:    "es_content",
	EsSendTime:   "es_send_time",
	EsSendStatus: "es_send_status",
	EsSendNum:    "es_send_num",
	EsResult:     "es_result",
	Remark:       "remark",
	CreateBy:     "create_by",
	CreateTime:   "create_time",
	UpdateBy:     "update_by",
}

// NewSysSmsDao creates and returns a new DAO object for table data access.
func NewSysSmsDao() *SysSmsDao {
	return &SysSmsDao{
		group:   "default",
		table:   "sys_sms",
		columns: syssmsColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysSmsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysSmsDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysSmsDao) Columns() SysSmsColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysSmsDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysSmsDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysSmsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
