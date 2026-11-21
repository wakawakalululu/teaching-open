// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingWork is the golang structure of table teaching_work for DAO operations like Where/Data.
type TeachingWork struct {
	g.Meta       `orm:"table:teaching_work, do:true"`
	Id           interface{} // 主键id
	CreateBy     interface{} // 创建人
	CreateTime   interface{} // 创建时间
	UpdateBy     interface{} // 更新人
	UpdateTime   interface{} // 更新时间
	SysOrgCode   interface{} // 所属部门
	UserId       interface{} // 用户id
	DepartId     interface{} // 班级部门id
	CourseId     interface{} // 课程单元id
	WorkName     interface{} // 作业名称
	WorkType     interface{} // 作业类型
	WorkFile     interface{} // 作业文件(sys_file id)
	WorkCover    interface{} // 作业封面(sys_file id)
	WorkStatus   interface{} // 作业状态 0保存 1待批改 2已批改 3首页展示 4精选
	StarNum      interface{} // 点赞数
	CollectNum   interface{} // 收藏数
	DelFlag      interface{} // 删除标记
	ViewNum      interface{} // 浏览数
	AdditionalId interface{} // 附加作业id
	WorkScene    interface{} // 来源场景 course/additional/create
	HasCloudData interface{} // 是否有云数据
}
