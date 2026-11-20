// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysUser 对应 sys_user 表（列结构以 api/db/teachingopen2.8.sql 为准；
// 密码/盐不输出 JSON，仅服务端使用）。
type SysUser struct {
	Id           string `json:"id"           orm:"id,primary"         description:"主键id"`
	Username     string `json:"username"     orm:"username"           description:"登录账号"`
	Realname     string `json:"realname"     orm:"realname"           description:"真实姓名"`
	Password     string `json:"-"            orm:"password"           description:"密码散列/密文"`
	Salt         string `json:"-"            orm:"salt"               description:"盐"`
	Avatar       string `json:"avatar"       orm:"avatar"             description:"头像"`
	Birthday     string `json:"birthday"     orm:"birthday"           description:"生日"`
	Sex          int    `json:"sex"          orm:"sex"                description:"性别 0未知 1男 2女"`
	Email        string `json:"email"        orm:"email"              description:"邮箱"`
	Phone        string `json:"phone"        orm:"phone"              description:"手机号"`
	OrgCode      string `json:"orgCode"      orm:"org_code"           description:"机构编码"`
	Status       int    `json:"status"       orm:"status"             description:"状态 1正常 2冻结"`
	DelFlag      int    `json:"delFlag"      orm:"del_flag"           description:"删除状态 0正常 1已删除"`
	ThirdId      string `json:"thirdId"      orm:"third_id"           description:"第三方登录唯一标识"`
	ThirdType    string `json:"thirdType"    orm:"third_type"         description:"第三方类型"`
	ActivitiSync int    `json:"activitiSync" orm:"activiti_sync"      description:"同步工作流 1同步 0不同步"`
	WorkNo       string `json:"workNo"       orm:"work_no"            description:"工号"`
	Post         string `json:"post"         orm:"post"               description:"职务"`
	School       string `json:"school"       orm:"school"             description:"学校"`
	Telephone    string `json:"telephone"    orm:"telephone"          description:"座机号"`
	CreateBy     string `json:"createBy"     orm:"create_by"          description:"创建人"`
	CreateTime   string `json:"createTime"   orm:"create_time"        description:"创建时间"`
	UpdateBy     string `json:"updateBy"     orm:"update_by"          description:"更新人"`
	UpdateTime   string `json:"updateTime"   orm:"update_time"        description:"更新时间"`
	UserIdentity int    `json:"userIdentity" orm:"user_identity"      description:"身份 1普通成员 2上级"`
	DepartIds    string `json:"departIds"    orm:"depart_ids"         description:"负责部门"`
}
