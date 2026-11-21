// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingWorkComment 对应 teaching_work_comment 表（api/db/teachingopen2.8.sql L1114：
// 作品评论子表）。
type TeachingWorkComment struct {
	Id         string `json:"id"         orm:"id,primary"   description:"ID"`
	CreateBy   string `json:"createBy"   orm:"create_by"    description:"创建人"`
	CreateTime string `json:"createTime" orm:"create_time"  description:"创建时间"`
	UpdateBy   string `json:"updateBy"   orm:"update_by"    description:"更新人"`
	UpdateTime string `json:"updateTime" orm:"update_time"  description:"更新时间"`
	SysOrgCode string `json:"sysOrgCode" orm:"sys_org_code" description:"所属部门"`
	WorkId     string `json:"workId"     orm:"work_id"      description:"作业id"`
	Comment    string `json:"comment"    orm:"comment"      description:"评论内容"`
	UserId     string `json:"userId"     orm:"user_id"      description:"评论人id"`
}
