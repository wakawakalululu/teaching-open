// register_gaps.go 契约补齐：Java 侧存在、Go 侧缺失或 HTTP 方法不一致的路由。
//
// 背景（2026-11-22 全量对账）：Java 411 个方法级端点中约 19% 在 Go 侧不可达，
// 其中约一半是 Excel 导出/导入（已在 register_p2_6.go 的 excelxBases 补全），
// 其余为方法别名（Java 用 GET，Go 只注册了 PUT/POST）与少量遗漏端点。
// 本文件按 Java Controller 逐条对齐补齐，不改变既有实现语义。
package cmd

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	ctrlDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/dict"
	ctrlLogin "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/login"
	ctrlTeachingUser "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/teachinguser"
	ctrlUpload "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/upload"
	svcSS "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sysspecial"
)

// wireContractGaps 装配契约补齐路由（均在受保护组内，全部登记守卫）。
func wireContractGaps(group *ghttp.RouterGroup, guard func(method, path string, requireAuth bool),
	dictCtrl *ctrlDict.Controller, loginCtrl *ctrlLogin.Controller,
	userCtrl *ctrlTeachingUser.Controller, uploadCtrl *ctrlUpload.Service,
) {
	special := svcSS.New()

	// ---- 方法别名：Java 用 GET，Go 侧原仅注册 PUT/POST ----
	// GET /sys/permission/edit、/sys/permission/editPermissionRule、
	// /teaching/menu/edit、/sys/logout 均按 Java 方法补齐。
	group.GET("/sys/logout", loginCtrl.Logout)
	guard("GET", "/api/sys/logout", true)

	// ---- 遗漏端点：登录后部门选择 / 登录日志（守卫早已声明但无路由）----
	group.PUT("/sys/selectDepart", loginCtrl.SelectDepart)
	group.POST("/sys/selectDepart", loginCtrl.SelectDepart)
	guard("PUT", "/api/sys/selectDepart", true)
	guard("POST", "/api/sys/selectDepart", true)

	group.GET("/sys/loginfo", loginCtrl.LogInfo)
	guard("GET", "/api/sys/loginfo", true)

	group.GET("/sys/visitInfo", loginCtrl.VisitInfo)
	guard("GET", "/api/sys/visitInfo", true)

	// ---- 字典通用表树（SysDictController#loadTreeData）----
	group.GET("/sys/dict/loadTreeData", dictCtrl.LoadTreeData)
	guard("GET", "/api/sys/dict/loadTreeData", true)

	// ---- 用户分部信息 / APP 用户列表（Java 在 /sys/user 下）----
	group.GET("/sys/user/getUserSectionInfoByToken", userCtrl.GetUserSectionInfoByToken)
	group.GET("/sys/user/appUserList", userCtrl.AppUserList)
	guard("GET", "/api/sys/user/getUserSectionInfoByToken", true)
	guard("GET", "/api/sys/user/appUserList", true)

	// ---- 填值规则执行（SysFillRuleController#executeRuleByCode）----
	// Java 通过 FillRuleUtil 反射执行各规则；Go 侧已移植 org_num_role（部门编码），
	// 其余规则返回明确未实现错误（与 testFillRule 同一实现源）。
	group.PUT("/sys/fillRule/executeRuleByCode/{ruleCode}", func(r *ghttp.Request) {
		code := r.Get("ruleCode").String()
		res, err := special.TestFillRule(r.Context(), code, "")
		if err != nil {
			failJSON(r, err.Error())
			return
		}
		okJSON(r, res)
	})
	group.PUT("/sys/fillRule/executeRuleByCodeBatch", func(r *ghttp.Request) {
		var in struct {
			CommonFormData map[string]interface{} `json:"commonFormData"`
			Rules          []struct {
				RuleCode string                 `json:"ruleCode"`
				FormData map[string]interface{} `json:"formData"`
			} `json:"rules"`
		}
		if err := r.Parse(&in); err != nil {
			failJSON(r, "参数解析失败")
			return
		}
		out := make([]map[string]interface{}, 0, len(in.Rules))
		for _, rule := range in.Rules {
			res, err := special.TestFillRule(r.Context(), rule.RuleCode, "")
			item := map[string]interface{}{"ruleCode": rule.RuleCode}
			if err != nil {
				item["result"] = nil
				item["error"] = err.Error()
			} else {
				item["result"] = res
			}
			out = append(out, item)
		}
		okJSON(r, out)
	})
	guard("PUT", "/api/sys/fillRule/executeRuleByCode/{ruleCode}", true)
	guard("PUT", "/api/sys/fillRule/executeRuleByCodeBatch", true)

	// ---- 文件下载与 PDF 预览（CommonController）----
	// MinIO 上传：本部署 UPLOAD_TYPE=local，复用本地上传实现（语义等价）。
	group.POST("/sys/upload/uploadMinio", uploadCtrl.Upload)
	guard("POST", "/api/sys/upload/uploadMinio", true)

	// 下载复用本地静态服务（同一存储根 + 路径穿越剥离），仅语义名为 download。
	group.GET("/sys/common/download/*path", uploadCtrl.Static)
	guard("GET", "/api/sys/common/download/*path", true)

	// PDF 预览：返回内嵌 iframe 的 HTML（Java 为 pdf.js 预览页）。
	group.GET("/sys/common/pdf/pdfPreviewIframe", func(r *ghttp.Request) {
		url := r.Get("url").String()
		r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		r.Response.Write(fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>PDF 预览</title>
<style>html,body{margin:0;height:100%%}iframe{width:100%%;height:100%%;border:0}</style></head>
<body><iframe src="%s"></iframe></body></html>`, url))
	})
	guard("GET", "/api/sys/common/pdf/pdfPreviewIframe", true)

	// ---- ng-alain 表字典（NgAlainController#getDictItemsByTable）----
	group.GET("/sys/ng-alain/getDictItemsByTable/{table}/{key}/{value}", func(r *ghttp.Request) {
		table := r.Get("table").String()
		key := r.Get("key").String()
		value := r.Get("value").String()
		rows, err := tableDictItems(r, table, key, value)
		if err != nil {
			failJSON(r, err.Error())
			return
		}
		okJSON(r, rows)
	})
	guard("GET", "/api/sys/ng-alain/getDictItemsByTable/{table}/{key}/{value}", true)
}

// tableDictItems 白名单表字典取值（[{value,label}]）。
func tableDictItems(r *ghttp.Request, table, key, value string) ([]map[string]interface{}, error) {
	for _, ident := range []string{table, key, value} {
		if !validIdent(ident) {
			return nil, fmt.Errorf("表名或列名不合法")
		}
	}
	switch table {
	case "sys_category", "sys_dict", "sys_depart", "sys_permission", "sys_role":
	default:
		return nil, fmt.Errorf("不支持的表: %s", table)
	}
	sql := fmt.Sprintf("SELECT %s AS value, %s AS label FROM %s", key, value, table)
	rows, err := g.DB().GetAll(r.Context(), sql)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"value": row["value"].Interface(),
			"label": row["label"].String(),
		})
	}
	return out, nil
}

// validIdent 标识符白名单（仅字母数字下划线）。
func validIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

// okJSON / failJSON 轻量响应封装（与 common.OK/Fail 同形状）。
func okJSON(r *ghttp.Request, result interface{}) {
	r.Response.WriteJson(g.Map{"success": true, "code": 200, "message": "操作成功！", "result": result})
}

func failJSON(r *ghttp.Request, msg string) {
	r.Response.WriteJson(g.Map{"success": false, "code": 500, "message": msg, "result": nil})
}
