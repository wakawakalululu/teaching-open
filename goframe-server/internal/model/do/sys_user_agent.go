// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysUserAgent is the golang structure of table sys_user_agent for DAO operations like Where/Data.
type SysUserAgent struct {
	g.Meta        `orm:"table:sys_user_agent, do:true"`
	Id            interface{} // ID
	UserName      interface{} // 用户名
	AgentUserName interface{} // 代理人用户名
	StartTime     interface{} // 代理开始时间
	EndTime       interface{} // 代理结束时间
	Status        interface{} // 状态
	CreateName    interface{} // 创建人姓名
	CreateBy      interface{} // 创建人
	CreateTime    interface{} // 创建时间
	UpdateName    interface{} // 更新人姓名
	UpdateBy      interface{} // 更新人
	UpdateTime    interface{} // 更新时间
}
