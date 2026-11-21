// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysAnnouncementSend 对应 sys_announcement_send 表（api/db/teachingopen2.8.sql：
// 用户通告阅读标记表）。
type SysAnnouncementSend struct {
	Id         string `json:"id"         orm:"id,primary"   description:"ID"`
	AnntId     string `json:"anntId"     orm:"annt_id"      description:"通告id"`
	UserId     string `json:"userId"     orm:"user_id"      description:"用户id"`
	ReadFlag   string `json:"readFlag"   orm:"read_flag"    description:"已读 0未读 1已读"`
	ReadTime   string `json:"readTime"   orm:"read_time"    description:"阅读时间"`
	CreateBy   string `json:"createBy"   orm:"create_by"    description:"创建人"`
	CreateTime string `json:"createTime" orm:"create_time"  description:"创建时间"`
	UpdateBy   string `json:"updateBy"   orm:"update_by"    description:"更新人"`
	UpdateTime string `json:"updateTime" orm:"update_time"  description:"更新时间"`
}
