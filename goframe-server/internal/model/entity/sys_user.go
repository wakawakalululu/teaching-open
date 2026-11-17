// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUser 对应 sys_user 表（仅列出服务端实际使用的列）。
type SysUser struct {
	Id         string `json:"id"          orm:"id,primary"       description:"主键"`
	Username   string `json:"username"    orm:"username"         description:"登录账号"`
	Realname   string `json:"realname"    orm:"realname"         description:"真实姓名"`
	Password   string `json:"-"           orm:"password"         description:"密码散列/密文"`
	Salt       string `json:"-"           orm:"salt"             description:"盐"`
	Avatar     string `json:"avatar"      orm:"avatar"           description:"头像"`
	Email      string `json:"email"       orm:"email"            description:"邮箱"`
	Phone      string `json:"phone"       orm:"phone"            description:"手机号"`
	Status     int    `json:"status"      orm:"status"           description:"状态 1启用 0禁用"`
	DelFlag    int    `json:"delFlag"     orm:"del_flag"         description:"删除标记 0正常"`
	OrgCode    string `json:"orgCode"     orm:"org_code"         description:"组织编码"`
	CreateTime string `json:"createTime"  orm:"create_time"      description:"创建时间"`
}
