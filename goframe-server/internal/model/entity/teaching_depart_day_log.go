// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingDepartDayLog 对应 teaching_depart_day_log 表（api/db/teachingopen2.8.sql：
// 班级每日教学计数；create_time 由 addLog 写 datetime 但查询用 yyyyMMdd
// 字符串比较——Java 格式错位现状）。
type TeachingDepartDayLog struct {
	Id                         string `json:"id"                         orm:"id,primary"                    description:"ID"`
	DepartId                   string `json:"departId"                   orm:"depart_id"                     description:"班级部门id"`
	DepartName                 string `json:"departName"                 orm:"depart_name"                   description:"班级名称"`
	UnitOpenCount              int    `json:"unitOpenCount"              orm:"unit_open_count"               description:"单元打开数"`
	CourseWorkAssignCount      int    `json:"courseWorkAssignCount"      orm:"course_work_assign_count"      description:"课程作业布置数"`
	AdditionalWorkAssignCount  int    `json:"additionalWorkAssignCount"  orm:"additional_work_assign_count"  description:"附加作业布置数"`
	CourseWorkCorrectCount     int    `json:"courseWorkCorrectCount"     orm:"course_work_correct_count"     description:"课程作业批改数"`
	AdditionalWorkCorrectCount int    `json:"additionalWorkCorrectCount" orm:"additional_work_correct_count" description:"附加作业批改数"`
	CourseWorkSubmitCount      int    `json:"courseWorkSubmitCount"      orm:"course_work_submit_count"      description:"课程作业提交数"`
	AdditionalWorkSubmitCount  int    `json:"additionalWorkSubmitCount"  orm:"additional_work_submit_count"  description:"附加作业提交数"`
	CreateTime                 string `json:"createTime"                 orm:"create_time"                   description:"创建时间"`
}
