// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SysFile 对应 sys_file 表（api/db/teachingopen2.8.sql L664：
// 文件记录；file_location 1=本地 2=七牛）。
type SysFile struct {
	Id           string `json:"id"           orm:"id,primary"     description:"ID"`
	CreateTime   string `json:"createTime"   orm:"create_time"    description:"创建时间"`
	UpdateBy     string `json:"updateBy"     orm:"update_by"      description:"更新人"`
	CreateBy     string `json:"createBy"     orm:"create_by"      description:"创建人"`
	UpdateTime   string `json:"updateTime"   orm:"update_time"    description:"更新时间"`
	SysOrgCode   string `json:"sysOrgCode"   orm:"sys_org_code"   description:"所属部门"`
	FileType     string `json:"fileType"     orm:"file_type"      description:"文件类型"`
	FileName     string `json:"fileName"     orm:"file_name"      description:"文件名称"`
	FilePath     string `json:"filePath"     orm:"file_path"      description:"文件路径/key"`
	FileLocation int    `json:"fileLocation" orm:"file_location"  description:"文件位置 1本地 2七牛"`
	FileTag      string `json:"fileTag"      orm:"file_tag"       description:"文件标签"`
	DelFlag      int    `json:"delFlag"      orm:"del_flag"       description:"删除状态"`
}
