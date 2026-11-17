// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysUser is the golang structure of table sys_user for DAO operations like Where/Data.
type SysUser struct {
	g.Meta     `orm:"table:sys_user, do:true"`
	Id         interface{} // 主键
	Username   interface{} // 登录账号
	Realname   interface{} // 真实姓名
	Password   interface{} // 密码散列/密文
	Salt       interface{} // 盐
	Avatar     interface{} // 头像
	Email      interface{} // 邮箱
	Phone      interface{} // 手机号
	Status     interface{} // 状态 1启用 0禁用
	DelFlag    interface{} // 删除标记 0正常
	OrgCode    interface{} // 组织编码
	CreateTime interface{} // 创建时间
}
