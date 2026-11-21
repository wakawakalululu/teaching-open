// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingAdditionalWorkDao is the data access object for table teaching_additional_work.
type TeachingAdditionalWorkDao struct {
	table   string                        // table is the underlying table name of the DAO.
	group   string                        // group is the database configuration group name of current DAO.
	columns TeachingAdditionalWorkColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingAdditionalWorkColumns defines and stores column names for table teaching_additional_work.
type TeachingAdditionalWorkColumns struct {
	Id              string // ID:id
	CreateBy        string // 创建人:create_by
	CreateTime      string // 创建时间:create_time
	UpdateBy        string // 更新人:update_by
	UpdateTime      string // 更新时间:update_time
	SysOrgCode      string // 所属部门:sys_org_code
	CodeType        string // 作品类型:code_type
	WorkName        string // 作业名称:work_name
	WorkDesc        string // 作业描述:work_desc
	WorkCover       string // 作业封面:work_cover
	WorkUrl         string // 作业地址:work_url
	WorkDept        string // 授权班级id串:work_dept
	Status          string // 状态:status
	WorkDocumentUrl string // 作业文档:work_document_url
}

// teachingadditionalworkColumns holds the columns for table teaching_additional_work.
var teachingadditionalworkColumns = TeachingAdditionalWorkColumns{
	Id:              "id",
	CreateBy:        "create_by",
	CreateTime:      "create_time",
	UpdateBy:        "update_by",
	UpdateTime:      "update_time",
	SysOrgCode:      "sys_org_code",
	CodeType:        "code_type",
	WorkName:        "work_name",
	WorkDesc:        "work_desc",
	WorkCover:       "work_cover",
	WorkUrl:         "work_url",
	WorkDept:        "work_dept",
	Status:          "status",
	WorkDocumentUrl: "work_document_url",
}

// NewTeachingAdditionalWorkDao creates and returns a new DAO object for table data access.
func NewTeachingAdditionalWorkDao() *TeachingAdditionalWorkDao {
	return &TeachingAdditionalWorkDao{
		group:   "default",
		table:   "teaching_additional_work",
		columns: teachingadditionalworkColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingAdditionalWorkDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingAdditionalWorkDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingAdditionalWorkDao) Columns() TeachingAdditionalWorkColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingAdditionalWorkDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingAdditionalWorkDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingAdditionalWorkDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
