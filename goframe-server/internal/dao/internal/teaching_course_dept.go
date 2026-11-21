// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingCourseDeptDao is the data access object for table teaching_course_dept.
type TeachingCourseDeptDao struct {
	table   string                    // table is the underlying table name of the DAO.
	group   string                    // group is the database configuration group name of current DAO.
	columns TeachingCourseDeptColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingCourseDeptColumns defines and stores column names for table teaching_course_dept.
type TeachingCourseDeptColumns struct {
	Id         string // ID
	CreateBy   string // 创建人
	CreateTime string // 创建时间
	UpdateBy   string // 更新人
	UpdateTime string // 更新时间
	SysOrgCode string // 所属部门
	DeptId     string // 班级部门id
	CourseId   string // 课程id
	OpenTime   string // 开课时间
}

// teachingCourseDeptColumns holds the columns for table teaching_course_dept.
var teachingCourseDeptColumns = TeachingCourseDeptColumns{
	Id:         "id",
	CreateBy:   "create_by",
	CreateTime: "create_time",
	UpdateBy:   "update_by",
	UpdateTime: "update_time",
	SysOrgCode: "sys_org_code",
	DeptId:     "dept_id",
	CourseId:   "course_id",
	OpenTime:   "open_time",
}

// NewTeachingCourseDeptDao creates and returns a new DAO object for table data access.
func NewTeachingCourseDeptDao() *TeachingCourseDeptDao {
	return &TeachingCourseDeptDao{
		group:   "default",
		table:   "teaching_course_dept",
		columns: teachingCourseDeptColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingCourseDeptDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingCourseDeptDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingCourseDeptDao) Columns() TeachingCourseDeptColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingCourseDeptDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingCourseDeptDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingCourseDeptDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
