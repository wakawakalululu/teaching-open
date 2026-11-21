// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingCourse is the golang structure of table teaching_course for DAO operations like Where/Data.
type TeachingCourse struct {
	g.Meta         `orm:"table:teaching_course, do:true"`
	Id             interface{} // 主键id
	CreateBy       interface{} // 创建人
	CreateTime     interface{} // 创建时间
	UpdateBy       interface{} // 更新人
	UpdateTime     interface{} // 更新时间
	SysOrgCode     interface{} // 所属部门
	DelFlag        interface{} // 删除状态
	CourseName     interface{} // 课程名称
	CourseDesc     interface{} // 课程描述(富文本)
	CourseIcon     interface{} // 课程图标
	CourseCover    interface{} // 课程封面
	ShowType       interface{} // 展示类型
	CourseMap      interface{} // 课程地图
	IsShared       interface{} // 是否共享课程
	ShowHome       interface{} // 首页展示
	OrderNum       interface{} // 排序
	DepartIds      interface{} // 授权部门id串(逗号分隔)
	CourseType     interface{} // 课程类型
	CourseCategory interface{} // 课程分类
}
