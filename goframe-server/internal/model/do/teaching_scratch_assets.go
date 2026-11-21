// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import "github.com/gogf/gf/v2/frame/g"

// TeachingScratchAssets is the golang structure of table teaching_scratch_assets for DAO operations like Where/Data.
type TeachingScratchAssets struct {
	g.Meta     `orm:"table:teaching_scratch_assets, do:true"`
	Id         interface{} // ID
	AssetType  interface{} // 素材类型 1背景 2声音 3造型 4角色
	AssetName  interface{} // 素材名称
	AssetData  interface{} // 素材元数据JSON
	Md5Ext     interface{} // md5扩展名
	Tags       interface{} // 标签
	CreateBy   interface{} // 创建人
	CreateTime interface{} // 创建时间
	UpdateBy   interface{} // 更新人
	UdpateTime interface{} // 更新时间(表列原拼写)
	DelFlag    interface{} // 删除标记
}
