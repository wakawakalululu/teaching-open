// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingCourse 对应 teaching_course 表（api/db/teachingopen2.8.sql L948：
// 课程主表；is_shared/show_home 为 tinyint 布尔语义；del_flag 列存在但业务
// 不读写（删除即物理删），无 @TableLogic）。
type TeachingCourse struct {
	Id             string `json:"id"            orm:"id,primary"      description:"主键id"`
	CreateBy       string `json:"createBy"      orm:"create_by"       description:"创建人"`
	CreateTime     string `json:"createTime"    orm:"create_time"     description:"创建时间"`
	UpdateBy       string `json:"updateBy"      orm:"update_by"       description:"更新人"`
	UpdateTime     string `json:"updateTime"    orm:"update_time"     description:"更新时间"`
	SysOrgCode     string `json:"sysOrgCode"    orm:"sys_org_code"    description:"所属部门"`
	DelFlag        int    `json:"delFlag"       orm:"del_flag"        description:"删除状态"`
	CourseName     string `json:"courseName"    orm:"course_name"     description:"课程名称"`
	CourseDesc     string `json:"courseDesc"    orm:"course_desc"     description:"课程描述(富文本)"`
	CourseIcon     string `json:"courseIcon"    orm:"course_icon"     description:"课程图标"`
	CourseCover    string `json:"courseCover"   orm:"course_cover"    description:"课程封面"`
	ShowType       int    `json:"showType"      orm:"show_type"       description:"展示类型"`
	CourseMap      string `json:"courseMap"     orm:"course_map"      description:"课程地图"`
	IsShared       bool   `json:"isShared"      orm:"is_shared"       description:"是否共享课程"`
	ShowHome       bool   `json:"showHome"      orm:"show_home"       description:"首页展示"`
	OrderNum       int    `json:"orderNum"      orm:"order_num"       description:"排序"`
	DepartIds      string `json:"departIds"     orm:"depart_ids"      description:"授权部门id串(逗号分隔)"`
	CourseType     string `json:"courseType"    orm:"course_type"     description:"课程类型"`
	CourseCategory string `json:"courseCategory" orm:"course_category" description:"课程分类"`
}
