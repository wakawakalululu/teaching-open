// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// OssFile 对应 oss_file 表（api/db/teachingopen2.8.sql：OSS 对象文件记录）。
type OssFile struct {
	Id         string `json:"id"         orm:"id,primary" description:"主键id"`
	FileName   string `json:"fileName"   orm:"file_name"  description:"文件名称"`
	Url        string `json:"url"        orm:"url"        description:"文件地址"`
	CreateBy   string `json:"createBy"   orm:"create_by"  description:"创建人"`
	CreateTime string `json:"createTime" orm:"create_time" description:"创建时间"`
	UpdateBy   string `json:"updateBy"   orm:"update_by"  description:"更新人"`
	UpdateTime string `json:"updateTime" orm:"update_time" description:"更新时间"`
}
