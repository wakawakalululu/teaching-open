// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingScratchAssets 对应 teaching_scratch_assets 表（api/db/teachingopen2.8.sql：
// Scratch 素材库；asset_type 1背景 2声音 3造型 4角色；udpate_time 为表列
// 原始拼写——Java 实体字段名同步保留）。
type TeachingScratchAssets struct {
	Id         string `json:"id"         orm:"id,primary"   description:"ID"`
	AssetType  int    `json:"assetType"  orm:"asset_type"   description:"素材类型 1背景 2声音 3造型 4角色"`
	AssetName  string `json:"assetName"  orm:"asset_name"   description:"素材名称"`
	AssetData  string `json:"assetData"  orm:"asset_data"   description:"素材元数据JSON"`
	Md5Ext     string `json:"md5Ext"     orm:"md5_ext"      description:"md5扩展名"`
	Tags       string `json:"tags"       orm:"tags"         description:"标签"`
	CreateBy   string `json:"createBy"   orm:"create_by"    description:"创建人"`
	CreateTime string `json:"createTime" orm:"create_time"  description:"创建时间"`
	UpdateBy   string `json:"updateBy"   orm:"update_by"    description:"更新人"`
	UdpateTime string `json:"udpateTime" orm:"udpate_time"  description:"更新时间(表列原拼写)"`
	DelFlag    int    `json:"delFlag"    orm:"del_flag"     description:"删除标记"`
}
