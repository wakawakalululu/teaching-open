// register_p2_6.go P2-6 各域路由装配（泛型 CRUD 绑定 + 特殊端点）。
package cmd

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/syscrud"
	svcSC "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/syscrud"
	svcSS "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sysspecial"
)

// crudDomainDef 泛型域装配定义。
type crudDomainDef struct {
	Base    string // 路由基路径（不含 /api 前缀）
	Table   string
	Cols    []string
	LikeCol string // list 的 LIKE 过滤列（空则无）
}

// p2_6_domains P2-6 泛型域清单。
var p2_6_domains = []crudDomainDef{
	{Base: "/sys/fillRule", Table: "sys_fill_rule", Cols: []string{
		"id", "rule_name", "rule_code", "rule_class", "rule_params",
		"update_by", "update_time", "create_by", "create_time"}},
	{Base: "/sys/checkRule", Table: "sys_check_rule", Cols: []string{
		"id", "rule_name", "rule_code", "rule_json", "rule_description",
		"update_by", "update_time", "create_by", "create_time"}},
	{Base: "/sys/dataSource", Table: "sys_data_source", Cols: []string{
		"id", "code", "name", "remark", "db_type", "db_driver", "db_url",
		"db_name", "db_username", "db_password", "create_by", "create_time"}},
	{Base: "/sys/position", Table: "sys_position", Cols: []string{
		"id", "code", "name", "post_rank", "company_id", "create_by",
		"create_time", "update_by", "update_time", "sys_org_code"}},
	{Base: "/sys/log", Table: "sys_log", Cols: []string{
		"id", "log_type", "log_content", "operate_type", "userid", "username",
		"ip", "method", "request_url", "request_param", "request_type",
		"cost_time", "create_by", "create_time", "update_by", "update_time"},
		LikeCol: "log_content"},
	{Base: "/sys/dataLog", Table: "sys_data_log", Cols: []string{
		"id", "data_table", "data_id", "data_content", "data_version",
		"create_by", "create_time", "update_by", "update_time"}},
	{Base: "/sys/quartzJob", Table: "sys_quartz_job", Cols: []string{
		"id", "create_by", "create_time", "del_flag", "update_by", "update_time",
		"job_class_name", "cron_expression", "parameter", "description", "status"}},
}

// excelxBases 已接入 Excel 能力的域基路径。
var excelxBases = []string{
	"/sys/fillRule", "/sys/checkRule", "/sys/dataSource", "/sys/position",
	"/sys/log", "/sys/dataLog", "/sys/quartzJob", "/sys/annountCement",
	"/sys/user", "/teaching/teachingNews", "/teaching/teachingCourse",
	"/teaching/teachingCourseUnit", "/teaching/teachingCourseDept",
	"/teaching/teachingWork", "/teaching/teachingScratchAssets",
	// 2026-11-22 补全：管理面板前端「导出/导入」按钮对应域
	"/sys/dict", "/sys/dictItem", "/sys/category", "/sys/role",
	"/sys/sysDepart", "/sys/sysDepartPermission", "/sys/sysDepartRole",
	"/system/sysFile", "/sys/sysUserAgent",
	"/message/sysMessage", "/message/sysMessageTemplate",
	"/teaching/teachingDepartDayLog",
}

// wireP2_6Domains 装配 P2-6 泛型域路由、守卫与特殊端点。
func wireP2_6Domains(group *ghttp.RouterGroup, guard func(method, path string, requireAuth bool),
	special *svcSS.Service) {
	var dataSourceEngine *syscrud.Domain
	for _, def := range p2_6_domains {
		d := syscrud.NewDomain(svcSC.TableDef{Table: def.Table, Cols: def.Cols})
		if def.Table == "sys_data_source" {
			dataSourceEngine = d
		}
		likeCol := def.LikeCol
		group.GET(def.Base+"/list", func(r *ghttp.Request) { d.List(r, likeCol) })
		group.POST(def.Base+"/add", d.Add)
		group.PUT(def.Base+"/edit", d.Edit)
		group.DELETE(def.Base+"/delete", d.Delete)
		group.DELETE(def.Base+"/deleteBatch", d.DeleteBatch)
		group.GET(def.Base+"/queryById", d.QueryByID)
		guard("GET", "/api"+def.Base+"/list", true)
		guard("POST", "/api"+def.Base+"/add", true)
		guard("PUT", "/api"+def.Base+"/edit", true)
		guard("DELETE", "/api"+def.Base+"/delete", true)
		guard("DELETE", "/api"+def.Base+"/deleteBatch", true)
		guard("GET", "/api"+def.Base+"/queryById", true)
	}
	// 特殊端点：填充规则试算
	group.GET("/sys/fillRule/testFillRule", func(r *ghttp.Request) {
		res, err := special.TestFillRule(r.Context(), r.Get("ruleCode").String(),
			r.Get("testValue").String())
		if err != nil {
			if errorsIs(err, svcSS.ErrNotImplemented) {
				common.Fail(r, err.Error())
				return
			}
			common.Fail(r, "操作失败")
			return
		}
		common.OK(r, res)
	})
	guard("GET", "/api/sys/fillRule/testFillRule", true)
	// 特殊端点：编码校验规则试算
	group.GET("/sys/checkRule/checkByCode", func(r *ghttp.Request) {
		res, err := special.CheckByCode(r.Context(), r.Get("ruleCode").String(),
			r.Get("value").String())
		if err != nil {
			if errorsIs(err, svcSS.ErrNotFound) {
				common.Fail(r, err.Error())
				return
			}
			common.Fail(r, "操作失败")
			return
		}
		if res.Passed {
			common.OK(r, nil)
			return
		}
		// Java：error(message) + result=明细
		common.FailMsg(r, res.Message, res.Detail)
	})
	guard("GET", "/api/sys/checkRule/checkByCode", true)
	// 特殊端点：数据源 options（裸数组）
	group.GET("/sys/dataSource/options", func(r *ghttp.Request) {
		res, err := dataSourceEngine.ListAllOptionsEngine(r.Context(), r.Get("name").String())
		if err != nil {
			common.Fail(r, "操作失败")
			return
		}
		common.OK(r, res)
	})
	guard("GET", "/api/sys/dataSource/options", true)
	// 特殊端点：数据日志对比与版本
	group.GET("/sys/dataLog/queryCompareList", func(r *ghttp.Request) {
		res, err := special.QueryCompareList(r.Context(),
			r.Get("dataId1").String(), r.Get("dataId2").String())
		if err != nil {
			common.Fail(r, "操作失败")
			return
		}
		common.OK(r, res)
	})
	guard("GET", "/api/sys/dataLog/queryCompareList", true)
	group.GET("/sys/dataLog/queryDataVerList", func(r *ghttp.Request) {
		res, err := special.QueryDataVerList(r.Context(),
			r.Get("dataTable").String(), r.Get("dataId").String())
		if err != nil {
			common.Fail(r, "操作失败")
			return
		}
		if len(res) == 0 {
			common.Fail(r, "未找到版本信息")
			return
		}
		common.OK(r, res)
	})
	guard("GET", "/api/sys/dataLog/queryDataVerList", true)
	// 特殊端点：重复校验
	group.GET("/sys/duplicate/check", func(r *ghttp.Request) {
		usable, err := special.DuplicateCheck(r.Context(),
			r.Get("tableName").String(), r.Get("fieldName").String(),
			r.Get("fieldVal").String(), r.Get("dataId").String())
		if err != nil {
			common.Fail(r, "操作失败")
			return
		}
		if usable {
			common.OKMsg(r, "该值可用！", nil)
			return
		}
		common.Fail(r, "该值不可用，系统中已存在！")
	})
	guard("GET", "/api/sys/duplicate/check", true)
	// 特殊端点：Quartz 任务暂停/恢复（状态落库——Go 无调度器后端，文档化）
	group.GET("/sys/quartzJob/pause", func(r *ghttp.Request) {
		if err := special.QuartzSetStatus(r.Context(),
			svcSS.TrimJobClassName(r.Get("jobClassName").String()), -1); err != nil {
			common.Fail(r, "操作失败")
			return
		}
		common.OKMsg(r, "停止任务成功", nil)
	})
	guard("GET", "/api/sys/quartzJob/pause", true)
	group.GET("/sys/quartzJob/resume", func(r *ghttp.Request) {
		if err := special.QuartzSetStatus(r.Context(),
			svcSS.TrimJobClassName(r.Get("jobClassName").String()), 0); err != nil {
			common.Fail(r, "操作失败")
			return
		}
		common.OKMsg(r, "恢复任务成功", nil)
	})
	guard("GET", "/api/sys/quartzJob/resume", true)
}

// dOptions 数据源 options 委托。
func dOptions(r *ghttp.Request) {
	if p2_6_optionsFn != nil {
		p2_6_optionsFn(r)
		return
	}
	common.OK(r, []interface{}{})
}

// p2_6_optionsFn 由 mainFunc 注入（避免循环依赖直接持有 dataSource 引擎）。
var p2_6_optionsFn func(r *ghttp.Request)

// errorsIs errors.Is 别名（避免 import 冲突）。
func errorsIs(err, target error) bool {
	return err == target || (err != nil && target != nil &&
		stringsContains(err.Error(), target.Error()))
}

// stringsContains 子串判断。
func stringsContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

// indexOf 子串查找。
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
