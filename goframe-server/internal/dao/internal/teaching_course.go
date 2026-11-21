// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingCourseDao is the data access object for table teaching_course.
type TeachingCourseDao struct {
	table   string                // table is the underlying table name of the DAO.
	group   string                // group is the database configuration group name of current DAO.
	columns TeachingCourseColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingCourseColumns defines and stores column names for table teaching_course.
type TeachingCourseColumns struct {
	Id             string // 主键id
	CreateBy       string // 创建人
	CreateTime     string // 创建时间
	UpdateBy       string // 更新人
	UpdateTime     string // 更新时间
	SysOrgCode     string // 所属部门
	DelFlag        string // 删除状态
	CourseName     string // 课程名称
	CourseDesc     string // 课程描述(富文本)
	CourseIcon     string // 课程图标
	CourseCover    string // 课程封面
	ShowType       string // 展示类型
	CourseMap      string // 课程地图
	IsShared       string // 是否共享课程
	ShowHome       string // 首页展示
	OrderNum       string // 排序
	DepartIds      string // 授权部门id串(逗号分隔)
	CourseType     string // 课程类型
	CourseCategory string // 课程分类
}

// teachingCourseColumns holds the columns for table teaching_course.
var teachingCourseColumns = TeachingCourseColumns{
	Id:             "id",
	CreateBy:       "create_by",
	CreateTime:     "create_time",
	UpdateBy:       "update_by",
	UpdateTime:     "update_time",
	SysOrgCode:     "sys_org_code",
	DelFlag:        "del_flag",
	CourseName:     "course_name",
	CourseDesc:     "course_desc",
	CourseIcon:     "course_icon",
	CourseCover:    "course_cover",
	ShowType:       "show_type",
	CourseMap:      "course_map",
	IsShared:       "is_shared",
	ShowHome:       "show_home",
	OrderNum:       "order_num",
	DepartIds:      "depart_ids",
	CourseType:     "course_type",
	CourseCategory: "course_category",
}

// NewTeachingCourseDao creates and returns a new DAO object for table data access.
func NewTeachingCourseDao() *TeachingCourseDao {
	return &TeachingCourseDao{
		group:   "default",
		table:   "teaching_course",
		columns: teachingCourseColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingCourseDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingCourseDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingCourseDao) Columns() TeachingCourseColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingCourseDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingCourseDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingCourseDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
