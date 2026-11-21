// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingWork 对应 teaching_work 表（api/db/teachingopen2.8.sql L1087：
// 学生作业主表；work_file/work_cover 存 sys_file.id；del_flag 列存在但业务
// 不读写（物理删除），无 @TableLogic）。
type TeachingWork struct {
	Id           string `json:"id"          orm:"id,primary"     description:"主键id"`
	CreateBy     string `json:"createBy"    orm:"create_by"      description:"创建人"`
	CreateTime   string `json:"createTime"  orm:"create_time"    description:"创建时间"`
	UpdateBy     string `json:"updateBy"    orm:"update_by"      description:"更新人"`
	UpdateTime   string `json:"updateTime"  orm:"update_time"    description:"更新时间"`
	SysOrgCode   string `json:"sysOrgCode"  orm:"sys_org_code"   description:"所属部门"`
	UserId       string `json:"userId"      orm:"user_id"        description:"用户id"`
	DepartId     string `json:"departId"    orm:"depart_id"      description:"班级部门id"`
	CourseId     string `json:"courseId"    orm:"course_id"      description:"课程单元id"`
	WorkName     string `json:"workName"    orm:"work_name"      description:"作业名称"`
	WorkType     string `json:"workType"    orm:"work_type"      description:"作业类型"`
	WorkFile     string `json:"workFile"    orm:"work_file"      description:"作业文件(sys_file id)"`
	WorkCover    string `json:"workCover"   orm:"work_cover"     description:"作业封面(sys_file id)"`
	WorkStatus   int    `json:"workStatus"  orm:"work_status"    description:"作业状态 0保存 1待批改 2已批改 3首页展示 4精选"`
	StarNum      int    `json:"starNum"     orm:"star_num"       description:"点赞数"`
	CollectNum   int    `json:"collectNum"  orm:"collect_num"    description:"收藏数"`
	DelFlag      int    `json:"delFlag"     orm:"del_flag"       description:"删除标记"`
	ViewNum      int    `json:"viewNum"     orm:"view_num"       description:"浏览数"`
	AdditionalId string `json:"additionalId" orm:"additional_id" description:"附加作业id"`
	WorkScene    string `json:"workScene"   orm:"work_scene"     description:"来源场景 course/additional/create"`
	HasCloudData bool   `json:"hasCloudData" orm:"has_cloud_data" description:"是否有云数据"`
}
