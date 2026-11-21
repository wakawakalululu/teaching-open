// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingAdditionalWork 对应 teaching_additional_work 表（api/db/teachingopen2.8.sql L931：
// 附加作业表；work_dept 为逗号分隔班级 id 串）。
type TeachingAdditionalWork struct {
	Id              string `json:"id"             orm:"id,primary"        description:"ID"`
	CreateBy        string `json:"createBy"       orm:"create_by"         description:"创建人"`
	CreateTime      string `json:"createTime"     orm:"create_time"       description:"创建时间"`
	UpdateBy        string `json:"updateBy"       orm:"update_by"         description:"更新人"`
	UpdateTime      string `json:"updateTime"     orm:"update_time"       description:"更新时间"`
	SysOrgCode      string `json:"sysOrgCode"     orm:"sys_org_code"      description:"所属部门"`
	CodeType        int    `json:"codeType"       orm:"code_type"         description:"作品类型"`
	WorkName        string `json:"workName"       orm:"work_name"         description:"作业名称"`
	WorkDesc        string `json:"workDesc"       orm:"work_desc"         description:"作业描述"`
	WorkCover       string `json:"workCover"      orm:"work_cover"        description:"作业封面"`
	WorkUrl         string `json:"workUrl"        orm:"work_url"          description:"作业地址"`
	WorkDept        string `json:"workDept"       orm:"work_dept"         description:"授权班级id串"`
	Status          int    `json:"status"         orm:"status"            description:"状态 0未发布 1已发布 2已结束"`
	WorkDocumentUrl string `json:"workDocumentUrl" orm:"work_document_url" description:"作业文档"`
}
