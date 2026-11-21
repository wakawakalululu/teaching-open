// registry.go 各域 Excel 列规格注册表
// （Field=表列、Header=Java @Excel 中文名/DDL 注释；导入按 Header 反查列）。
package excelx

// SpecByBase 按路由基路径取域规格；未注册域返回 nil。
func SpecByBase(base string) *Spec {
	for i := range specs {
		if specs[i].Base == base {
			return &specs[i]
		}
	}
	return nil
}

// specs 域注册表（顺序即导出表头顺序；列取自各实体 @Excel 注解/DDL 注释）。
var specs = []Spec{
	{Base: "/sys/fillRule", Title: "填充规则列表", Cols: []ExcelCol{
		{"rule_name", "规则名称"}, {"rule_code", "规则编码"},
		{"rule_class", "规则类"}, {"rule_params", "规则参数"},
		{"create_by", "创建人"}, {"create_time", "创建时间"}}},
	{Base: "/sys/checkRule", Title: "编码校验规则列表", Cols: []ExcelCol{
		{"rule_name", "规则名称"}, {"rule_code", "规则编码"},
		{"rule_json", "规则JSON"}, {"rule_description", "规则描述"}}},
	{Base: "/sys/dataSource", Title: "数据源列表", Cols: []ExcelCol{
		{"code", "编码"}, {"name", "名称"}, {"db_type", "数据库类型"},
		{"db_driver", "驱动类"}, {"db_url", "连接地址"}, {"db_name", "数据库名"},
		{"db_username", "用户名"}, {"remark", "备注"}}},
	{Base: "/sys/position", Title: "职务列表", Cols: []ExcelCol{
		{"code", "职务编码"}, {"name", "职务名称"}, {"post_rank", "职级"},
		{"company_id", "公司id"}}},
	{Base: "/sys/log", Title: "系统日志列表", Cols: []ExcelCol{
		{"log_content", "日志内容"}, {"log_type", "日志类型"}, {"operate_type", "操作类型"},
		{"userid", "操作人id"}, {"username", "操作人"}, {"ip", "IP"},
		{"method", "方法"}, {"request_url", "请求路径"}, {"request_param", "请求参数"},
		{"request_type", "请求类型"}, {"cost_time", "耗时"}, {"create_time", "创建时间"}}},
	{Base: "/sys/dataLog", Title: "数据变更历史列表", Cols: []ExcelCol{
		{"data_table", "表名"}, {"data_id", "数据id"}, {"data_content", "数据内容"},
		{"data_version", "版本号"}, {"create_time", "创建时间"}}},
	{Base: "/sys/quartzJob", Title: "定时任务列表", Cols: []ExcelCol{
		{"job_class_name", "任务类名"}, {"cron_expression", "cron表达式"},
		{"parameter", "参数"}, {"description", "描述"}, {"status", "状态"}}},
	{Base: "/sys/annountCement", Title: "系统通告列表", Cols: []ExcelCol{
		{"titile", "标题"}, {"msg_content", "内容"}, {"msg_category", "消息分类"},
		{"msg_type", "消息类型"}, {"sender", "发布人"}, {"priority", "优先级"},
		{"send_status", "发布状态"}, {"start_time", "开始时间"}, {"end_time", "结束时间"},
		{"msg_abstract", "摘要"}}},
	{Base: "/teaching/teachingNews", Title: "资讯列表", Cols: []ExcelCol{
		{"news_title", "资讯标题"}, {"news_status", "资讯状态"}, {"create_by", "创建人"}}},
	{Base: "/teaching/teachingCourse", Title: "课程列表", Cols: []ExcelCol{
		{"course_name", "课程名称"}, {"course_desc", "课程描述"},
		{"course_type", "课程类型"}, {"course_category", "课程分类"},
		{"is_shared", "是否共享"}, {"show_home", "首页展示"}, {"order_num", "排序"},
		{"depart_ids", "授权班级"}}},
	{Base: "/teaching/teachingCourseUnit", Title: "课程单元列表", Cols: []ExcelCol{
		{"unit_name", "单元名称"}, {"unit_intro", "单元简介"}, {"course_id", "课程id"},
		{"course_work_type", "作业类型"}, {"order_num", "排序"}, {"map_x", "地图X"}, {"map_y", "地图Y"}}},
	{Base: "/teaching/teachingCourseDept", Title: "班级课程列表", Cols: []ExcelCol{
		{"dept_id", "班级id"}, {"course_id", "课程id"}, {"open_time", "开课时间"}}},
	{Base: "/teaching/teachingWork", Title: "作业列表列表", Cols: []ExcelCol{
		{"work_name", "作业名称"}, {"work_type", "作业类型"}, {"work_status", "作业状态"},
		{"user_id", "学生id"}, {"depart_id", "班级id"}, {"view_num", "浏览数"},
		{"star_num", "点赞数"}, {"create_by", "创建人"}, {"create_time", "创建时间"}}},
	{Base: "/teaching/teachingScratchAssets", Title: "Scratch素材库列表", Cols: []ExcelCol{
		{"asset_type", "素材类型"}, {"asset_name", "素材名称"}, {"md5_ext", "md5扩展名"},
		{"tags", "标签"}}},
	{Base: "/sys/user", Title: "用户列表", Cols: []ExcelCol{
		{"username", "登录账号"}, {"realname", "真实姓名"}, {"sex", "性别"},
		{"birthday", "生日"}, {"email", "邮箱"}, {"phone", "手机号"},
		{"org_code", "部门编码"}, {"work_no", "工号"}, {"status", "状态"}}},
}
