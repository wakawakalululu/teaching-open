// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingNews is the golang structure of table teaching_news for DAO operations like Where/Data.
type TeachingNews struct {
	g.Meta      `orm:"table:teaching_news, do:true"`
	Id          interface{} // ID
	NewsTitle   interface{} // 资讯标题
	NewsContent interface{} // 资讯内容(富文本)
	NewsStatus  interface{} // 资讯状态 0草稿 1已发布
	CreateBy    interface{} // 创建人
	CreateTime  interface{} // 创建时间
	UpdateBy    interface{} // 更新人
	UpdateTime  interface{} // 更新时间
}
