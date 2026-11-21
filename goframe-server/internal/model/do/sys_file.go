// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// SysFile is the golang structure of table sys_file for DAO operations like Where/Data.
type SysFile struct {
	g.Meta       `orm:"table:sys_file, do:true"`
	Id           interface{} // ID
	CreateTime   interface{} // 创建时间
	UpdateBy     interface{} // 更新人
	CreateBy     interface{} // 创建人
	UpdateTime   interface{} // 更新时间
	SysOrgCode   interface{} // 所属部门
	FileType     interface{} // 文件类型
	FileName     interface{} // 文件名称
	FilePath     interface{} // 文件路径/key
	FileLocation interface{} // 文件位置 1本地 2七牛
	FileTag      interface{} // 文件标签
	DelFlag      interface{} // 删除状态
}
