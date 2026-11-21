// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysAnnouncementSend is the golang structure of table sys_announcement_send for DAO operations like Where/Data.
type SysAnnouncementSend struct {
	g.Meta     `orm:"table:sys_announcement_send, do:true"`
	Id         interface{} // ID
	AnntId     interface{} // 通告id
	UserId     interface{} // 用户id
	ReadFlag   interface{} // 已读 0未读 1已读
	ReadTime   interface{} // 阅读时间
	CreateBy   interface{} // 创建人
	CreateTime interface{} // 创建时间
	UpdateBy   interface{} // 更新人
	UpdateTime interface{} // 更新时间
}
