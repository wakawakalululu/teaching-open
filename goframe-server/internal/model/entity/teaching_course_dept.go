// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingCourseDept 对应 teaching_course_dept 表（api/db/teachingopen2.8.sql L971：
// 班级开课授权表；open_time 为开课时间，学生端 mineCourse 以 now() 比较）。
type TeachingCourseDept struct {
	Id         string `json:"id"          orm:"id,primary"   description:"ID"`
	CreateBy   string `json:"createBy"    orm:"create_by"    description:"创建人"`
	CreateTime string `json:"createTime"  orm:"create_time"  description:"创建时间"`
	UpdateBy   string `json:"updateBy"    orm:"update_by"    description:"更新人"`
	UpdateTime string `json:"updateTime"  orm:"update_time"  description:"更新时间"`
	SysOrgCode string `json:"sysOrgCode"  orm:"sys_org_code" description:"所属部门"`
	DeptId     string `json:"deptId"      orm:"dept_id"      description:"班级部门id"`
	CourseId   string `json:"courseId"    orm:"course_id"    description:"课程id"`
	OpenTime   string `json:"openTime"    orm:"open_time"    description:"开课时间"`
}
