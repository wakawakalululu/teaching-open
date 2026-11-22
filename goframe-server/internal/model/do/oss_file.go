// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// OssFile is the golang structure of table oss_file for DAO operations like Where/Data.
type OssFile struct {
	g.Meta     `orm:"table:oss_file, do:true"`
	Id         interface{} // 主键id
	FileName   interface{} // 文件名称
	Url        interface{} // 文件地址
	CreateBy   interface{} // 创建人
	CreateTime interface{} // 创建时间
	UpdateBy   interface{} // 更新人
	UpdateTime interface{} // 更新时间
}
