// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysUser is the golang structure of table sys_user for DAO operations like Where/Data.
type SysUser struct {
	g.Meta       `orm:"table:sys_user, do:true"`
	Id           interface{} // 主键id
	Username     interface{} // 登录账号
	Realname     interface{} // 真实姓名
	Password     interface{} // 密码散列/密文
	Salt         interface{} // 盐
	Avatar       interface{} // 头像
	Birthday     interface{} // 生日
	Sex          interface{} // 性别 0未知 1男 2女
	Email        interface{} // 邮箱
	Phone        interface{} // 手机号
	OrgCode      interface{} // 机构编码
	Status       interface{} // 状态 1正常 2冻结
	DelFlag      interface{} // 删除状态 0正常 1已删除
	ThirdId      interface{} // 第三方登录唯一标识
	ThirdType    interface{} // 第三方类型
	ActivitiSync interface{} // 同步工作流 1同步 0不同步
	WorkNo       interface{} // 工号
	Post         interface{} // 职务
	School       interface{} // 学校
	Telephone    interface{} // 座机号
	CreateBy     interface{} // 创建人
	CreateTime   interface{} // 创建时间
	UpdateBy     interface{} // 更新人
	UpdateTime   interface{} // 更新时间
	UserIdentity interface{} // 身份 1普通成员 2上级
	DepartIds    interface{} // 负责部门
}
