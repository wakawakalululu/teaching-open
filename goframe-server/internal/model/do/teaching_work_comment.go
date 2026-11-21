// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingWorkComment is the golang structure of table teaching_work_comment for DAO operations like Where/Data.
type TeachingWorkComment struct {
	g.Meta     `orm:"table:teaching_work_comment, do:true"`
	Id         interface{} // ID
	CreateBy   interface{} // 创建人
	CreateTime interface{} // 创建时间
	UpdateBy   interface{} // 更新人
	UpdateTime interface{} // 更新时间
	SysOrgCode interface{} // 所属部门
	WorkId     interface{} // 作业id
	Comment    interface{} // 评论内容
	UserId     interface{} // 评论人id
}
