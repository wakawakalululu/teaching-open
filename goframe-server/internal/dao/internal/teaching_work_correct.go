// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingWorkCorrectDao is the data access object for table teaching_work_correct.
type TeachingWorkCorrectDao struct {
	table   string                     // table is the underlying table name of the DAO.
	group   string                     // group is the database configuration group name of current DAO.
	columns TeachingWorkCorrectColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingWorkCorrectColumns defines and stores column names for table teaching_work_correct.
type TeachingWorkCorrectColumns struct {
	Id           string // ID
	id           string // id
	WorkId       string // 作业id
	work_id      string // work_id
	Score        string // 评分
	score        string // score
	Comment      string // 评语
	comment      string // comment
	CreateBy     string // 创建人
	create_by    string // create_by
	CreateTime   string // 创建时间
	create_time  string // create_time
	UpdateBy     string // 更新人
	update_by    string // update_by
	UpdateTime   string // 更新时间
	update_time  string // update_time
	SysOrgCode   string // 所属部门
	sys_org_code string // sys_org_code
}

// teachingworkcorrectColumns holds the columns for table teaching_work_correct.
var teachingworkcorrectColumns = TeachingWorkCorrectColumns{
	Id:           "",
	id:           "id",
	WorkId:       "",
	work_id:      "work_id",
	Score:        "",
	score:        "score",
	Comment:      "",
	comment:      "comment",
	CreateBy:     "",
	create_by:    "create_by",
	CreateTime:   "",
	create_time:  "create_time",
	UpdateBy:     "",
	update_by:    "update_by",
	UpdateTime:   "",
	update_time:  "update_time",
	SysOrgCode:   "",
	sys_org_code: "sys_org_code",
}

// NewTeachingWorkCorrectDao creates and returns a new DAO object for table data access.
func NewTeachingWorkCorrectDao() *TeachingWorkCorrectDao {
	return &TeachingWorkCorrectDao{
		group:   "default",
		table:   "teaching_work_correct",
		columns: teachingworkcorrectColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingWorkCorrectDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingWorkCorrectDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingWorkCorrectDao) Columns() TeachingWorkCorrectColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingWorkCorrectDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingWorkCorrectDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingWorkCorrectDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
