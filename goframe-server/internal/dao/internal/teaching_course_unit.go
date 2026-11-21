// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingCourseUnitDao is the data access object for table teaching_course_unit.
type TeachingCourseUnitDao struct {
	table   string                    // table is the underlying table name of the DAO.
	group   string                    // group is the database configuration group name of current DAO.
	columns TeachingCourseUnitColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingCourseUnitColumns defines and stores column names for table teaching_course_unit.
type TeachingCourseUnitColumns struct {
	Id                string // 主键id
	CreateBy          string // 创建人
	CreateTime        string // 创建时间
	UpdateBy          string // 更新人
	UpdateTime        string // 更新时间
	SysOrgCode        string // 所属部门
	DelFlag           string // 删除状态
	UnitName          string // 单元名称
	UnitIntro         string // 单元简介
	UnitCover         string // 单元封面
	CourseId          string // 课程id
	CourseVideo       string // 课程视频(key/url/html)
	CourseVideoSource string // 视频来源 1七牛 2外链URL 3外链HTML
	ShowCourseVideo   string // 学生可见视频
	CourseCase        string // 课程案例
	ShowCourseCase    string // 学生可见案例
	CoursePpt         string // 课件(逗号分隔key)
	ShowCoursePpt     string // 学生可见课件
	CourseWorkType    string // 作业类型
	CourseWork        string // 课程作业文件
	CourseWorkAnswer  string // 作业答案文件
	CoursePlan        string // 教案文件
	ShowCoursePlan    string // 学生可见教案
	MapX              string // 地图X
	MapY              string // 地图Y
	MediaContent      string // 富文本内容
	OrderNum          string // 排序
}

// teachingCourseUnitColumns holds the columns for table teaching_course_unit.
var teachingCourseUnitColumns = TeachingCourseUnitColumns{
	Id:                "id",
	CreateBy:          "create_by",
	CreateTime:        "create_time",
	UpdateBy:          "update_by",
	UpdateTime:        "update_time",
	SysOrgCode:        "sys_org_code",
	DelFlag:           "del_flag",
	UnitName:          "unit_name",
	UnitIntro:         "unit_intro",
	UnitCover:         "unit_cover",
	CourseId:          "course_id",
	CourseVideo:       "course_video",
	CourseVideoSource: "course_video_source",
	ShowCourseVideo:   "show_course_video",
	CourseCase:        "course_case",
	ShowCourseCase:    "show_course_case",
	CoursePpt:         "course_ppt",
	ShowCoursePpt:     "show_course_ppt",
	CourseWorkType:    "course_work_type",
	CourseWork:        "course_work",
	CourseWorkAnswer:  "course_work_answer",
	CoursePlan:        "course_plan",
	ShowCoursePlan:    "show_course_plan",
	MapX:              "map_x",
	MapY:              "map_y",
	MediaContent:      "media_content",
	OrderNum:          "order_num",
}

// NewTeachingCourseUnitDao creates and returns a new DAO object for table data access.
func NewTeachingCourseUnitDao() *TeachingCourseUnitDao {
	return &TeachingCourseUnitDao{
		group:   "default",
		table:   "teaching_course_unit",
		columns: teachingCourseUnitColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingCourseUnitDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingCourseUnitDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingCourseUnitDao) Columns() TeachingCourseUnitColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingCourseUnitDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingCourseUnitDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingCourseUnitDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
