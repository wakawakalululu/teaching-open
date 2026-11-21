// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeachingWorkDao is the data access object for table teaching_work.
type TeachingWorkDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns TeachingWorkColumns // columns contains all the column names of Table for convenient usage.
}

// TeachingWorkColumns defines and stores column names for table teaching_work.
type TeachingWorkColumns struct {
	Id             string // 主键id
	id             string // id
	CreateBy       string // 创建人
	create_by      string // create_by
	CreateTime     string // 创建时间
	create_time    string // create_time
	UpdateBy       string // 更新人
	update_by      string // update_by
	UpdateTime     string // 更新时间
	update_time    string // update_time
	SysOrgCode     string // 所属部门
	sys_org_code   string // sys_org_code
	UserId         string // 用户id
	user_id        string // user_id
	DepartId       string // 班级部门id
	depart_id      string // depart_id
	CourseId       string // 课程单元id
	course_id      string // course_id
	WorkName       string // 作业名称
	work_name      string // work_name
	WorkType       string // 作业类型
	work_type      string // work_type
	WorkFile       string // 作业文件
	work_file      string // work_file
	WorkCover      string // 作业封面
	work_cover     string // work_cover
	WorkStatus     string // 作业状态
	work_status    string // work_status
	StarNum        string // 点赞数
	star_num       string // star_num
	CollectNum     string // 收藏数
	collect_num    string // collect_num
	DelFlag        string // 删除标记
	del_flag       string // del_flag
	ViewNum        string // 浏览数
	view_num       string // view_num
	AdditionalId   string // 附加作业id
	additional_id  string // additional_id
	WorkScene      string // 来源场景
	work_scene     string // work_scene
	HasCloudData   string // 云数据
	has_cloud_data string // has_cloud_data
}

// teachingworkColumns holds the columns for table teaching_work.
var teachingworkColumns = TeachingWorkColumns{
	Id:             "",
	id:             "id",
	CreateBy:       "",
	create_by:      "create_by",
	CreateTime:     "",
	create_time:    "create_time",
	UpdateBy:       "",
	update_by:      "update_by",
	UpdateTime:     "",
	update_time:    "update_time",
	SysOrgCode:     "",
	sys_org_code:   "sys_org_code",
	UserId:         "",
	user_id:        "user_id",
	DepartId:       "",
	depart_id:      "depart_id",
	CourseId:       "",
	course_id:      "course_id",
	WorkName:       "",
	work_name:      "work_name",
	WorkType:       "",
	work_type:      "work_type",
	WorkFile:       "",
	work_file:      "work_file",
	WorkCover:      "",
	work_cover:     "work_cover",
	WorkStatus:     "",
	work_status:    "work_status",
	StarNum:        "",
	star_num:       "star_num",
	CollectNum:     "",
	collect_num:    "collect_num",
	DelFlag:        "",
	del_flag:       "del_flag",
	ViewNum:        "",
	view_num:       "view_num",
	AdditionalId:   "",
	additional_id:  "additional_id",
	WorkScene:      "",
	work_scene:     "work_scene",
	HasCloudData:   "",
	has_cloud_data: "has_cloud_data",
}

// NewTeachingWorkDao creates and returns a new DAO object for table data access.
func NewTeachingWorkDao() *TeachingWorkDao {
	return &TeachingWorkDao{
		group:   "default",
		table:   "teaching_work",
		columns: teachingworkColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *TeachingWorkDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *TeachingWorkDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *TeachingWorkDao) Columns() TeachingWorkColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *TeachingWorkDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *TeachingWorkDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *TeachingWorkDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
