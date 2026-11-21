// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingDepartDayLog is the golang structure of table teaching_depart_day_log for DAO operations like Where/Data.
type TeachingDepartDayLog struct {
	g.Meta                     `orm:"table:teaching_depart_day_log, do:true"`
	Id                         interface{} // ID
	DepartId                   interface{} // 班级部门id
	DepartName                 interface{} // 班级名称
	UnitOpenCount              interface{} // 单元打开数
	CourseWorkAssignCount      interface{} // 课程作业布置数
	AdditionalWorkAssignCount  interface{} // 附加作业布置数
	CourseWorkCorrectCount     interface{} // 课程作业批改数
	AdditionalWorkCorrectCount interface{} // 附加作业批改数
	CourseWorkSubmitCount      interface{} // 课程作业提交数
	AdditionalWorkSubmitCount  interface{} // 附加作业提交数
	CreateTime                 interface{} // 创建时间
}
