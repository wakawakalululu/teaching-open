// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SysSmsTemplateDao is the data access object for table sys_sms_template.
type SysSmsTemplateDao struct {
	table   string                // table is the underlying table name of the DAO.
	group   string                // group is the database configuration group name of current DAO.
	columns SysSmsTemplateColumns // columns contains all the column names of Table for convenient usage.
}

// SysSmsTemplateColumns defines and stores column names for table sys_sms_template.
type SysSmsTemplateColumns struct {
	Id               string // ID:id
	TemplateName     string // 模板名称:template_name
	TemplateCode     string // 模板编码:template_code
	TemplateType     string // 模板类型:template_type
	TemplateContent  string // 模板内容:template_content
	TemplateTestJson string // 模板测试JSON:template_test_json
	CreateBy         string // 创建人:create_by
	CreateTime       string // 创建时间:create_time
	UpdateBy         string // 更新人:update_by
	UpdateTime       string // 更新时间:update_time
}

// syssmstemplateColumns holds the columns for table sys_sms_template.
var syssmstemplateColumns = SysSmsTemplateColumns{
	Id:               "id",
	TemplateName:     "template_name",
	TemplateCode:     "template_code",
	TemplateType:     "template_type",
	TemplateContent:  "template_content",
	TemplateTestJson: "template_test_json",
	CreateBy:         "create_by",
	CreateTime:       "create_time",
	UpdateBy:         "update_by",
	UpdateTime:       "update_time",
}

// NewSysSmsTemplateDao creates and returns a new DAO object for table data access.
func NewSysSmsTemplateDao() *SysSmsTemplateDao {
	return &SysSmsTemplateDao{
		group:   "default",
		table:   "sys_sms_template",
		columns: syssmstemplateColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SysSmsTemplateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SysSmsTemplateDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SysSmsTemplateDao) Columns() SysSmsTemplateColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SysSmsTemplateDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SysSmsTemplateDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SysSmsTemplateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
