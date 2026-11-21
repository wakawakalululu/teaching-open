// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// TeachingCourseUnit 对应 teaching_course_unit 表（api/db/teachingopen2.8.sql L983：
// 课程单元；del_flag 列存在但业务不读写；show_xxx 为 tinyint 布尔语义）。
type TeachingCourseUnit struct {
	Id                string `json:"id"               orm:"id,primary"         description:"主键id"`
	CreateBy          string `json:"createBy"         orm:"create_by"          description:"创建人"`
	CreateTime        string `json:"createTime"       orm:"create_time"        description:"创建时间"`
	UpdateBy          string `json:"updateBy"         orm:"update_by"          description:"更新人"`
	UpdateTime        string `json:"updateTime"       orm:"update_time"        description:"更新时间"`
	SysOrgCode        string `json:"sysOrgCode"       orm:"sys_org_code"       description:"所属部门"`
	DelFlag           int    `json:"delFlag"          orm:"del_flag"           description:"删除状态"`
	UnitName          string `json:"unitName"         orm:"unit_name"          description:"单元名称"`
	UnitIntro         string `json:"unitIntro"        orm:"unit_intro"         description:"单元简介"`
	UnitCover         string `json:"unitCover"        orm:"unit_cover"         description:"单元封面"`
	CourseId          string `json:"courseId"         orm:"course_id"          description:"课程id"`
	CourseVideo       string `json:"courseVideo"      orm:"course_video"       description:"课程视频(key/url/html)"`
	CourseVideoSource int    `json:"courseVideoSource" orm:"course_video_source" description:"视频来源 1七牛 2外链URL 3外链HTML"`
	ShowCourseVideo   bool   `json:"showCourseVideo"  orm:"show_course_video"  description:"学生可见视频"`
	CourseCase        string `json:"courseCase"       orm:"course_case"        description:"课程案例"`
	ShowCourseCase    bool   `json:"showCourseCase"   orm:"show_course_case"   description:"学生可见案例"`
	CoursePpt         string `json:"coursePpt"        orm:"course_ppt"         description:"课件(逗号分隔key)"`
	ShowCoursePpt     bool   `json:"showCoursePpt"    orm:"show_course_ppt"    description:"学生可见课件"`
	CourseWorkType    int    `json:"courseWorkType"   orm:"course_work_type"   description:"作业类型"`
	CourseWork        string `json:"courseWork"       orm:"course_work"        description:"课程作业文件"`
	CourseWorkAnswer  string `json:"courseWorkAnswer" orm:"course_work_answer" description:"作业答案文件"`
	CoursePlan        string `json:"coursePlan"       orm:"course_plan"        description:"教案文件"`
	ShowCoursePlan    bool   `json:"showCoursePlan"   orm:"show_course_plan"   description:"学生可见教案"`
	MapX              int    `json:"mapX"             orm:"map_x"              description:"地图X"`
	MapY              int    `json:"mapY"             orm:"map_y"              description:"地图Y"`
	MediaContent      string `json:"mediaContent"     orm:"media_content"      description:"富文本内容"`
	OrderNum          int    `json:"orderNum"         orm:"order_num"          description:"排序"`
}
