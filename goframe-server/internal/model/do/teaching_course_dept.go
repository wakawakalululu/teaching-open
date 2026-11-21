// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingCourseDept is the golang structure of table teaching_course_dept for DAO operations like Where/Data.
type TeachingCourseDept struct {
	g.Meta     `orm:"table:teaching_course_dept, do:true"`
	Id         interface{} // ID
	CreateBy   interface{} // 创建人
	CreateTime interface{} // 创建时间
	UpdateBy   interface{} // 更新人
	UpdateTime interface{} // 更新时间
	SysOrgCode interface{} // 所属部门
	DeptId     interface{} // 班级部门id
	CourseId   interface{} // 课程id
	OpenTime   interface{} // 开课时间
}
