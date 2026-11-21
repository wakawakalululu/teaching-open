// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingNews 对应 teaching_news 表（api/db/teachingopen2.8.sql L1053：
// 资讯表；news_status 0草稿 1已发布）。
type TeachingNews struct {
	Id          string `json:"id"          orm:"id,primary"   description:"ID"`
	NewsTitle   string `json:"newsTitle"   orm:"news_title"   description:"资讯标题"`
	NewsContent string `json:"newsContent" orm:"news_content" description:"资讯内容(富文本)"`
	NewsStatus  int    `json:"newsStatus"  orm:"news_status"  description:"资讯状态 0草稿 1已发布"`
	CreateBy    string `json:"createBy"    orm:"create_by"    description:"创建人"`
	CreateTime  string `json:"createTime"  orm:"create_time"  description:"创建时间"`
	UpdateBy    string `json:"updateBy"    orm:"update_by"    description:"更新人"`
	UpdateTime  string `json:"updateTime"  orm:"update_time"  description:"更新时间"`
}
