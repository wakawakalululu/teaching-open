// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysAnnouncement is the golang structure of table sys_announcement for DAO operations like Where/Data.
type SysAnnouncement struct {
	g.Meta      `orm:"table:sys_announcement, do:true"`
	Id          interface{} // ID
	Titile      interface{} // 标题(表列原拼写)
	MsgContent  interface{} // 内容
	StartTime   interface{} // 开始时间
	EndTime     interface{} // 结束时间
	Sender      interface{} // 发布人
	Priority    interface{} // 优先级 H/L
	MsgCategory interface{} // 消息分类 1通知公告 2系统消息
	MsgType     interface{} // 消息类型 ALL/USER
	SendStatus  interface{} // 发布状态 0未发布 1发布中 2已撤销
	SendTime    interface{} // 发布时间
	CancelTime  interface{} // 撤销时间
	DelFlag     interface{} // 删除状态
	BusType     interface{} // 业务类型
	BusId       interface{} // 业务id
	OpenType    interface{} // 打开方式
	OpenPage    interface{} // 组件/路由地址
	CreateBy    interface{} // 创建人
	CreateTime  interface{} // 创建时间
	UpdateBy    interface{} // 更新人
	UpdateTime  interface{} // 更新时间
	MsgAbstract interface{} // 摘要
}
