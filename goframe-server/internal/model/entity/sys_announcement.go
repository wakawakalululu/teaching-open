// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysAnnouncement 对应 sys_announcement 表（api/db/teachingopen2.8.sql：
// 系统通告；titile 为表列原始拼写；user_ids 非本表列（Java @TableField
// exist=false），仅作 add/edit 入参）。
type SysAnnouncement struct {
	Id          string `json:"id"          orm:"id,primary"   description:"ID"`
	Titile      string `json:"titile"      orm:"titile"       description:"标题(表列原拼写)"`
	MsgContent  string `json:"msgContent"  orm:"msg_content"  description:"内容"`
	StartTime   string `json:"startTime"   orm:"start_time"   description:"开始时间"`
	EndTime     string `json:"endTime"     orm:"end_time"     description:"结束时间"`
	Sender      string `json:"sender"      orm:"sender"       description:"发布人"`
	Priority    string `json:"priority"    orm:"priority"     description:"优先级 H/L"`
	MsgCategory string `json:"msgCategory" orm:"msg_category" description:"消息分类 1通知公告 2系统消息"`
	MsgType     string `json:"msgType"     orm:"msg_type"     description:"消息类型 ALL/USER"`
	SendStatus  string `json:"sendStatus"  orm:"send_status"  description:"发布状态 0未发布 1发布中 2已撤销"`
	SendTime    string `json:"sendTime"    orm:"send_time"    description:"发布时间"`
	CancelTime  string `json:"cancelTime"  orm:"cancel_time"  description:"撤销时间"`
	DelFlag     string `json:"delFlag"     orm:"del_flag"     description:"删除状态"`
	BusType     string `json:"busType"     orm:"bus_type"     description:"业务类型"`
	BusId       string `json:"busId"       orm:"bus_id"       description:"业务id"`
	OpenType    string `json:"openType"    orm:"open_type"    description:"打开方式"`
	OpenPage    string `json:"openPage"    orm:"open_page"    description:"组件/路由地址"`
	CreateBy    string `json:"createBy"    orm:"create_by"    description:"创建人"`
	CreateTime  string `json:"createTime"  orm:"create_time"  description:"创建时间"`
	UpdateBy    string `json:"updateBy"    orm:"update_by"    description:"更新人"`
	UpdateTime  string `json:"updateTime"  orm:"update_time"  description:"更新时间"`
	MsgAbstract string `json:"msgAbstract" orm:"msg_abstract" description:"摘要"`
}
