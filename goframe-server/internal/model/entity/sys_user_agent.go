// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUserAgent 对应 sys_user_agent 表（api/db/teachingopen2.8.sql：
// 用户代理表——用户 A 代理用户 B 处理事务）。
type SysUserAgent struct {
	Id            string `json:"id"             orm:"id,primary"      description:"ID"`
	UserName      string `json:"userName"       orm:"user_name"       description:"用户名"`
	AgentUserName string `json:"agentUserName"  orm:"agent_user_name" description:"代理人用户名"`
	StartTime     string `json:"startTime"      orm:"start_time"      description:"代理开始时间"`
	EndTime       string `json:"endTime"        orm:"end_time"        description:"代理结束时间"`
	Status        string `json:"status"         orm:"status"          description:"状态"`
	CreateName    string `json:"createName"     orm:"create_name"     description:"创建人姓名"`
	CreateBy      string `json:"createBy"       orm:"create_by"       description:"创建人"`
	CreateTime    string `json:"createTime"     orm:"create_time"     description:"创建时间"`
	UpdateName    string `json:"updateName"     orm:"update_name"     description:"更新人姓名"`
	UpdateBy      string `json:"updateBy"       orm:"update_by"       description:"更新人"`
	UpdateTime    string `json:"updateTime"     orm:"update_time"     description:"更新时间"`
}
