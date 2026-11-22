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
	// ---- 2026-11-22 补全：管理面板缺失的 Excel 域（前端「导出/导入」按钮 404）----
	{Base: "/sys/dict", Title: "数据字典列表", Cols: []ExcelCol{
		{Field: "dict_name", Header: "字典名称"}, {Field: "dict_code", Header: "字典编码"},
		{Field: "description", Header: "描述"}, {Field: "type", Header: "字典类型"},
		{Field: "create_by", Header: "创建人"}, {Field: "create_time", Header: "创建时间"},
		{Field: "update_by", Header: "更新人"}, {Field: "update_time", Header: "更新时间"},
	}},
	{Base: "/sys/dictItem", Title: "字典项列表", Cols: []ExcelCol{
		{Field: "item_text", Header: "字典项文本"}, {Field: "item_value", Header: "字典项值"},
		{Field: "description", Header: "描述"}, {Field: "sort_order", Header: "排序"},
		{Field: "status", Header: "状态", Dic: "dict_item_status"},
	}},
	{Base: "/sys/category", Title: "分类字典列表", Cols: []ExcelCol{
		{Field: "name", Header: "分类名称"}, {Field: "code", Header: "分类编码"},
		{Field: "pid", Header: "父级节点"}, {Field: "has_child", Header: "是否有子节点"},
		{Field: "comment", Header: "描述"}, {Field: "create_by", Header: "创建人"},
		{Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/role", Title: "角色列表", Cols: []ExcelCol{
		{Field: "role_name", Header: "角色名称"}, {Field: "role_code", Header: "角色编码"},
		{Field: "description", Header: "描述"}, {Field: "role_level", Header: "角色等级"},
		{Field: "create_by", Header: "创建人"}, {Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/sysDepart", Title: "部门列表", Cols: []ExcelCol{
		{Field: "depart_name", Header: "机构名称"}, {Field: "depart_name_en", Header: "机构英文名称"},
		{Field: "depart_name_abbr", Header: "机构简称"}, {Field: "org_code", Header: "机构编码"},
		{Field: "parent_id", Header: "父机构ID"}, {Field: "depart_order", Header: "排序"},
		{Field: "mobile", Header: "手机号"}, {Field: "fax", Header: "传真"},
		{Field: "address", Header: "地址"}, {Field: "memo", Header: "备注"},
		{Field: "status", Header: "状态", Dic: "depart_status"},
	}},
	{Base: "/sys/sysDepartPermission", Title: "部门权限列表", Cols: []ExcelCol{
		{Field: "depart_id", Header: "部门id"}, {Field: "permission_id", Header: "权限id"},
		{Field: "data_rule_ids", Header: "数据规则id"},
	}},
	{Base: "/sys/sysDepartRole", Title: "部门角色列表", Cols: []ExcelCol{
		{Field: "depart_id", Header: "部门id"}, {Field: "role_name", Header: "角色名称"},
		{Field: "role_code", Header: "角色编码"}, {Field: "description", Header: "描述"},
	}},
	{Base: "/system/sysFile", Title: "文件列表", Cols: []ExcelCol{
		{Field: "file_name", Header: "文件名称"}, {Field: "file_path", Header: "文件地址"},
		{Field: "file_type", Header: "文件类型"}, {Field: "file_location", Header: "存储位置"},
		{Field: "file_tag", Header: "文件标识"}, {Field: "create_by", Header: "创建人"},
		{Field: "create_time", Header: "创建时间"},
	}},
	{Base: "/sys/sysUserAgent", Title: "用户代理人列表", Cols: []ExcelCol{
		{Field: "user_name", Header: "用户名"}, {Field: "agent_user_name", Header: "代理人用户名"},
		{Field: "start_time", Header: "开始时间"}, {Field: "end_time", Header: "结束时间"},
		{Field: "status", Header: "状态", Dic: "valid_status"},
	}},
	{Base: "/message/sysMessage", Title: "消息列表", Cols: []ExcelCol{
		{Field: "es_title", Header: "消息标题"}, {Field: "es_type", Header: "消息类型", Dic: "msgType"},
		{Field: "es_receiver", Header: "接收人"}, {Field: "es_content", Header: "消息内容"},
		{Field: "es_send_time", Header: "发送时间"}, {Field: "es_send_status", Header: "发送状态", Dic: "send_status"},
		{Field: "es_send_num", Header: "发送次数"}, {Field: "es_result", Header: "发送结果"},
	}},
	{Base: "/message/sysMessageTemplate", Title: "消息模板列表", Cols: []ExcelCol{
		{Field: "template_name", Header: "模板名称"}, {Field: "template_code", Header: "模板编码"},
		{Field: "template_type", Header: "模板类型", Dic: "msgType"},
		{Field: "template_content", Header: "模板内容"},
	}},
	{Base: "/teaching/teachingDepartDayLog", Title: "班级日志列表", Cols: []ExcelCol{
		{Field: "depart_name", Header: "班级名称"}, {Field: "unit_open_count", Header: "课程单元开放数"},
		{Field: "course_work_assign_count", Header: "课程作业布置数"},
		{Field: "additional_work_assign_count", Header: "附加作业布置数"},
		{Field: "course_work_correct_count", Header: "课程作业批改数"},
		{Field: "course_work_submit_count", Header: "课程作业提交数"},
		{Field: "create_time", Header: "统计日期"},
	}},
	{Base: "/teaching/teachingCourseDept", Title: "课程授权列表", Cols: []ExcelCol{
		{Field: "dept_id", Header: "授权部门"},
		{Field: "course_id", Header: "课程"},
		{Field: "open_time", Header: "开放时间"},
		{Field: "create_by", Header: "创建人"},
		{Field: "create_time", Header: "创建日期"},
	}},
	{Base: "/teaching/teachingCourseUnit", Title: "课程单元列表", Cols: []ExcelCol{
		{Field: "unit_name", Header: "单元名称"},
		{Field: "unit_intro", Header: "单元简介"},
		{Field: "unit_cover", Header: "单元封面"},
		{Field: "course_id", Header: "所属课程"},
		{Field: "course_video", Header: "课程视频"},
		{Field: "course_case", Header: "课程案例"},
		{Field: "course_ppt", Header: "课程PPT"},
		{Field: "course_work", Header: "课程作业"},
		{Field: "course_plan", Header: "课程教案"},
		{Field: "map_x", Header: "地图X坐标"},
		{Field: "map_y", Header: "地图Y坐标"},
		{Field: "order_num", Header: "排序"},
		{Field: "create_by", Header: "创建人"},
		{Field: "create_time", Header: "创建日期"},
	}},
}

// baseTables 域基路径 → 物理表名（Excel 导出/导入的唯一映射源，
// 供 controller 与一致性测试共用）。
var baseTables = map[string]string{
	"/sys/fillRule":                   "sys_fill_rule",
	"/sys/checkRule":                  "sys_check_rule",
	"/sys/dataSource":                 "sys_data_source",
	"/sys/position":                   "sys_position",
	"/sys/log":                        "sys_log",
	"/sys/dataLog":                    "sys_data_log",
	"/sys/quartzJob":                  "sys_quartz_job",
	"/sys/annountCement":              "sys_announcement",
	"/sys/user":                       "sys_user",
	"/teaching/teachingNews":          "teaching_news",
	"/teaching/teachingCourse":        "teaching_course",
	"/teaching/teachingCourseUnit":    "teaching_course_unit",
	"/teaching/teachingCourseDept":    "teaching_course_dept",
	"/teaching/teachingWork":          "teaching_work",
	"/teaching/teachingScratchAssets": "teaching_scratch_assets",
	"/sys/dict":                       "sys_dict",
	"/sys/dictItem":                   "sys_dict_item",
	"/sys/category":                   "sys_category",
	"/sys/role":                       "sys_role",
	"/sys/sysDepart":                  "sys_depart",
	"/sys/sysDepartPermission":        "sys_depart_permission",
	"/sys/sysDepartRole":              "sys_depart_role",
	"/system/sysFile":                 "sys_file",
	"/sys/sysUserAgent":               "sys_user_agent",
	"/message/sysMessage":             "sys_sms",
	"/message/sysMessageTemplate":     "sys_sms_template",
	"/teaching/teachingDepartDayLog":  "teaching_depart_day_log",
}

// TableByBase 返回域对应的物理表名（不存在返回空串）。
func TableByBase(base string) string {
	return baseTables[base]
}

// SpecTables 返回全部已注册域的 base→table 映射（只读副本，测试用）。
func SpecTables() map[string]string {
	out := make(map[string]string, len(baseTables))
	for k, v := range baseTables {
		out[k] = v
	}
	return out
}
