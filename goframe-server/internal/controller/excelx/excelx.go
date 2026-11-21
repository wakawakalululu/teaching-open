// Package excelx 提供全域 exportXls/importExcel 的 HTTP 处理器
// （逐域绑定 base → DomainSpec；export 输出 xlsx 附件流，import 解析入表）。
package excelx

import (
	"fmt"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcEX "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/excelx"
)

// Controller Excel 导出/导入控制器。
type Controller struct {
	bcryptHash func(string) string
}

// New 构造控制器。
func New(bcryptHash func(string) string) *Controller {
	return &Controller{bcryptHash: bcryptHash}
}

// operatorName 当前登录用户名（导出人显示）。
func operatorName(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUsername).String()
}

// ExportXls 导出 xlsx 附件流。
// 契约：GET/POST {base}/exportXls → 三行头 + 数据（Java ModelAndView 等价）。
func (c *Controller) ExportXls(base string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		spec := svcEX.SpecByBase(base)
		if spec == nil {
			common.Fail(r, "该域未配置导出")
			return
		}
		filters := map[string]interface{}{
			"titile": r.Get("titile").String(), "news_title": r.Get("newsTitle").String(),
			"course_name": r.Get("courseName").String(), "unit_name": r.Get("unitName").String(),
			"work_name": r.Get("workName").String(), "asset_type": r.Get("assetType").Int(),
			"username": r.Get("username").String(), "msg_category": r.Get("msgCategory").String(),
			"log_type": r.Get("logType").Int(),
		}
		rows, err := svcEX.QueryRows(r.Context(), specTable(base), spec.Cols, filters, delFlagCol(base))
		if err != nil {
			common.Fail(r, "导出失败")
			return
		}
		data, err := svcEX.Render(spec.Title, operatorName(r), headersOf(spec), rows)
		if err != nil {
			common.Fail(r, "导出失败")
			return
		}
		r.Response.Header().Set("Content-Type",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		r.Response.Header().Set("Content-Disposition",
			fmt.Sprintf("attachment;filename=%s.xlsx", spec.Title))
		r.Response.Write(data)
	}
}

// ImportExcel 导入 xlsx。
// 契约：POST {base}/importExcel（multipart file）→「文件导入成功！数据行数:N」。
func (c *Controller) ImportExcel(base string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		spec := svcEX.SpecByBase(base)
		if spec == nil {
			common.Fail(r, "该域未配置导入")
			return
		}
		file := r.GetUploadFile("file")
		if file == nil {
			common.Fail(r, "文件导入失败！")
			return
		}
		data, err := file.Open()
		if err != nil {
			common.Fail(r, "文件导入失败！")
			return
		}
		defer func() { _ = data.Close() }()
		buf := make([]byte, 0, 1024*1024)
		tmp := make([]byte, 32*1024)
		for {
			n, rerr := data.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if rerr != nil {
				break
			}
		}
		headers, rows, perr := svcEX.Parse(buf, 2, 1)
		if perr != nil {
			common.Fail(r, "文件导入失败！")
			return
		}
		inserted := 0
		for _, values := range rows {
			if ierr := svcEX.InsertRow(r.Context(), specTable(base), spec.Cols, headers, values); ierr == nil {
				inserted++
			}
		}
		common.OKMsg(r, fmt.Sprintf("文件导入成功！数据行数:%d", inserted), nil)
	}
}

// ImportStudent 学生批量导入。
// 契约：POST /sys/user/importStudent（multipart file + departIds）。
func (c *Controller) ImportStudent(r *ghttp.Request) {
	file := r.GetUploadFile("file")
	if file == nil {
		common.Fail(r, "文件导入失败！")
		return
	}
	src, err := file.Open()
	if err != nil {
		common.Fail(r, "文件导入失败！")
		return
	}
	defer func() { _ = src.Close() }()
	buf := make([]byte, 0, 1024*1024)
	tmp := make([]byte, 32*1024)
	for {
		n, rerr := src.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if rerr != nil {
			break
		}
	}
	res, err := svcEX.ImportStudent(r.Context(), buf,
		r.Get("departIds").String(), c.bcryptHash)
	if err != nil {
		common.Fail(r, err.Error())
		return
	}
	msg := fmt.Sprintf("成功:%d条,失败:%d条", res.SuccessLines, res.ErrorLines)
	common.OKMsg(r, msg, res.Errors)
}

// specTable/specDelFlag 表名与 del_flag 列映射（与 Spec 注册表 Base 对应）。
func specTable(base string) string {
	m := map[string]string{
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
	}
	return m[base]
}

// delFlagCol 含 del_flag 列的表返回列名（导出过滤已删行），否则空。
func delFlagCol(base string) string {
	switch base {
	case "/sys/annountCement", "/sys/user", "/sys/quartzJob":
		return "del_flag"
	case "/teaching/teachingNews":
		return ""
	}
	return ""
}

// headersOf 提取表头数组。
func headersOf(spec *svcEX.Spec) []string {
	out := make([]string, 0, len(spec.Cols))
	for _, c := range spec.Cols {
		out = append(out, c.Header)
	}
	return out
}
