// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingAdditionalWork is the golang structure of table teaching_additional_work for DAO operations like Where/Data.
type TeachingAdditionalWork struct {
	g.Meta          `orm:"table:teaching_additional_work, do:true"`
	Id              interface{} // ID
	CreateBy        interface{} // 创建人
	CreateTime      interface{} // 创建时间
	UpdateBy        interface{} // 更新人
	UpdateTime      interface{} // 更新时间
	SysOrgCode      interface{} // 所属部门
	CodeType        interface{} // 作品类型
	WorkName        interface{} // 作业名称
	WorkDesc        interface{} // 作业描述
	WorkCover       interface{} // 作业封面
	WorkUrl         interface{} // 作业地址
	WorkDept        interface{} // 授权班级id串
	Status          interface{} // 状态 0未发布 1已发布 2已结束
	WorkDocumentUrl interface{} // 作业文档
}
