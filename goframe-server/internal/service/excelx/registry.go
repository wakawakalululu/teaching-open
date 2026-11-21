// registry.go 各域 Excel 列规格注册表
// （Field=表列名、Header=Java @Excel(name=...) 中文表头、Dic=可选字典编码；
// 列顺序与 @Excel 注解出现顺序逐语义对齐；导入按 Header 反查列）。
package excelx

// ExcelCol 导出/导入列定义。
type ExcelCol struct {
	Field  string // 表列名
	Header string // 中文表头（@Excel name）
	Dic    string // dicCode（非空时导出取字典文本）
}

// Spec 域的 Excel 列规格。
type Spec struct {
	Base  string
	Title string
	Cols  []ExcelCol
}

// SpecByBase 按路由基路径取域规格。
func SpecByBase(base string) *Spec {
	for i := range specs {
		if specs[i].Base == base {
			return &specs[i]
		}
	}
	return nil
}

// specs 域注册表（列顺序与 @Excel 注解出现顺序逐语义对齐）。
var specs = []Spec{
	{Base: "/sys/fillRule", Title: "填充规则列表", Cols: []ExcelCol{
		{Field: "rule_name", Header: "规则名称"}, {Field: "rule_code", Header: "规则Code"},
		{Field: "rule_class", Header: "规则实现类"}, {Field: "rule_params", Header: "规则参数"},
		{Field: "update_by", Header: "修改人"}, {Field: "update_time", Header: "修改时间"},
		{Field: "create_by", Header: "创建人"}, {Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/checkRule", Title: "编码校验规则列表", Cols: []ExcelCol{
		{Field: "rule_name", Header: "规则名称"}, {Field: "rule_code", Header: "规则Code"},
		{Field: "rule_json", Header: "规则JSON"}, {Field: "rule_description", Header: "规则描述"},
		{Field: "update_by", Header: "更新人"}, {Field: "update_time", Header: "更新时间"},
		{Field: "create_by", Header: "创建人"}, {Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/dataSource", Title: "数据源列表", Cols: []ExcelCol{
		{Field: "code", Header: "数据源编码"}, {Field: "name", Header: "数据源名称"},
		{Field: "remark", Header: "备注"}, {Field: "db_type", Header: "数据库类型"},
		{Field: "db_driver", Header: "驱动类"}, {Field: "db_url", Header: "数据源地址"},
		{Field: "db_name", Header: "数据库名称"}, {Field: "db_username", Header: "用户名"},
		{Field: "db_password", Header: "密码"}, {Field: "sys_org_code", Header: "所属部门"},
	}},
	{Base: "/sys/position", Title: "职务列表", Cols: []ExcelCol{
		{Field: "code", Header: "职务编码"}, {Field: "name", Header: "职务名称"},
		{Field: "post_rank", Header: "职级", Dic: "position_rank"},
		{Field: "company_id", Header: "公司id"},
		{Field: "sys_org_code", Header: "组织机构编码"},
	}},
	{Base: "/sys/annountCement", Title: "系统通告列表", Cols: []ExcelCol{
		{Field: "titile", Header: "标题"}, {Field: "msg_content", Header: "内容"},
		{Field: "start_time", Header: "开始时间"}, {Field: "end_time", Header: "结束时间"},
		{Field: "sender", Header: "发布人"}, {Field: "priority", Header: "优先级", Dic: "priority"},
		{Field: "msg_category", Header: "消息类型", Dic: "msg_category"},
		{Field: "msg_type", Header: "通告对象类型", Dic: "msg_type"},
		{Field: "send_status", Header: "发布状态", Dic: "send_status"},
		{Field: "send_time", Header: "发布时间"}, {Field: "cancel_time", Header: "撤销时间"},
	}},
	{Base: "/sys/user", Title: "用户列表", Cols: []ExcelCol{
		{Field: "username", Header: "登录账号"}, {Field: "realname", Header: "真实姓名"},
		{Field: "birthday", Header: "生日"}, {Field: "sex", Header: "性别", Dic: "sex"},
		{Field: "email", Header: "电子邮件"}, {Field: "phone", Header: "电话"},
		{Field: "org_code", Header: "所在部门"}, {Field: "status", Header: "状态", Dic: "user_status"},
		{Field: "del_flag", Header: "删除状态", Dic: "del_flag"},
		{Field: "work_no", Header: "工号"}, {Field: "post", Header: "职务"},
		{Field: "school", Header: "学校"}, {Field: "telephone", Header: "座机号"},
		{Field: "user_identity", Header: "（1普通成员 2上级）"},
		{Field: "depart_ids", Header: "负责部门"},
	}},
	{Base: "/sys/log", Title: "系统日志列表", Cols: []ExcelCol{
		{Field: "log_content", Header: "日志内容"}, {Field: "log_type", Header: "日志类型"},
		{Field: "operate_type", Header: "操作类型"}, {Field: "userid", Header: "操作人id"},
		{Field: "username", Header: "操作人"}, {Field: "ip", Header: "IP"},
		{Field: "method", Header: "方法"}, {Field: "request_url", Header: "请求路径"},
		{Field: "request_param", Header: "请求参数"}, {Field: "request_type", Header: "请求类型"},
		{Field: "cost_time", Header: "耗时"}, {Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/dataLog", Title: "数据变更历史列表", Cols: []ExcelCol{
		{Field: "data_table", Header: "表名"}, {Field: "data_id", Header: "数据id"},
		{Field: "data_content", Header: "数据内容"}, {Field: "data_version", Header: "版本号"},
		{Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/quartzJob", Title: "定时任务列表", Cols: []ExcelCol{
		{Field: "job_class_name", Header: "任务类名"}, {Field: "cron_expression", Header: "cron表达式"},
		{Field: "parameter", Header: "参数"}, {Field: "description", Header: "描述"},
		{Field: "status", Header: "状态"},
	}},
	{Base: "/teaching/teachingNews", Title: "资讯列表", Cols: []ExcelCol{
		{Field: "news_title", Header: "标题"}, {Field: "news_content", Header: "内容"},
		{Field: "news_status", Header: "状态", Dic: "news_status"},
	}},
	{Base: "/teaching/teachingCourse", Title: "课程列表", Cols: []ExcelCol{
		{Field: "create_by", Header: "创建人"}, {Field: "create_time", Header: "创建日期"},
		{Field: "update_by", Header: "更新人"}, {Field: "update_time", Header: "更新日期"},
		{Field: "sys_org_code", Header: "所属部门"}, {Field: "is_shared", Header: "是否共享"},
		{Field: "show_home", Header: "首页展示"}, {Field: "course_name", Header: "课程名"},
		{Field: "course_desc", Header: "课程介绍"}, {Field: "course_icon", Header: "课程图标"},
		{Field: "course_cover", Header: "课程封面"}, {Field: "course_map", Header: "课程地图"},
		{Field: "depart_ids", Header: "授权部门"},
		{Field: "course_type", Header: "课程类型", Dic: "course_type"},
		{Field: "course_category", Header: "课程分类", Dic: "course_category"},
	}},
	{Base: "/teaching/teachingWork", Title: "作业列表列表", Cols: []ExcelCol{
		{Field: "user_id", Header: "用户"}, {Field: "depart_id", Header: "班级"},
		{Field: "course_id", Header: "课程"}, {Field: "additional_id", Header: "附加作业"},
		{Field: "work_name", Header: "作业名"}, {Field: "work_type", Header: "作业类型", Dic: "work_type"},
		{Field: "work_status", Header: "作业状态", Dic: "work_status"},
		{Field: "work_file", Header: "作文文件"}, {Field: "work_scene", Header: "提交场景"},
		{Field: "create_time", Header: "创建日期"}, {Field: "update_by", Header: "更新人"},
		{Field: "update_time", Header: "更新日期"},
	}},
	{Base: "/teaching/teachingScratchAssets", Title: "Scratch素材库列表", Cols: []ExcelCol{
		{Field: "asset_type", Header: "素材类型"}, {Field: "asset_name", Header: "素材名称"},
		{Field: "md5_ext", Header: "md5扩展名"}, {Field: "tags", Header: "标签"},
	}},
}
