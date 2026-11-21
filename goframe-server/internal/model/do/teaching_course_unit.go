// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingCourseUnit is the golang structure of table teaching_course_unit for DAO operations like Where/Data.
type TeachingCourseUnit struct {
	g.Meta            `orm:"table:teaching_course_unit, do:true"`
	Id                interface{} // 主键id
	CreateBy          interface{} // 创建人
	CreateTime        interface{} // 创建时间
	UpdateBy          interface{} // 更新人
	UpdateTime        interface{} // 更新时间
	SysOrgCode        interface{} // 所属部门
	DelFlag           interface{} // 删除状态
	UnitName          interface{} // 单元名称
	UnitIntro         interface{} // 单元简介
	UnitCover         interface{} // 单元封面
	CourseId          interface{} // 课程id
	CourseVideo       interface{} // 课程视频(key/url/html)
	CourseVideoSource interface{} // 视频来源 1七牛 2外链URL 3外链HTML
	ShowCourseVideo   interface{} // 学生可见视频
	CourseCase        interface{} // 课程案例
	ShowCourseCase    interface{} // 学生可见案例
	CoursePpt         interface{} // 课件(逗号分隔key)
	ShowCoursePpt     interface{} // 学生可见课件
	CourseWorkType    interface{} // 作业类型
	CourseWork        interface{} // 课程作业文件
	CourseWorkAnswer  interface{} // 作业答案文件
	CoursePlan        interface{} // 教案文件
	ShowCoursePlan    interface{} // 学生可见教案
	MapX              interface{} // 地图X
	MapY              interface{} // 地图Y
	MediaContent      interface{} // 富文本内容
	OrderNum          interface{} // 排序
}
