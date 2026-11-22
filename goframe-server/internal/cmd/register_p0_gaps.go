// register_p0_gaps.go P0 收尾：字典表分页、403、transitRESTful、OSS 域、工作台真实数据。
//
// 依据 docs/BACKEND-GAP-ANALYSIS.md §4 的 P0/P3 清单：
//   - /sys/dict/queryTableData、/sys/common/403、/sys/common/transitRESTful；
//   - /oss/file/* 4 端点（本地存储后端）；
//   - 工作台真实数据（前端 /api/api/* 双前缀契约，见 MockController）。
package cmd

import (
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	ctrlDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/dict"
	ctrlOSS "github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/oss"
)

// wireP0Gaps 装配 P0 收尾路由与工作台真实数据端点。
func wireP0Gaps(group *ghttp.RouterGroup, guard func(method, path string, requireAuth bool),
	dictCtrl *ctrlDict.Controller, ossCtrl *ctrlOSS.Controller,
) {
	// ---- 字典表分页 ----
	group.GET("/sys/dict/queryTableData", dictCtrl.QueryTableData)
	guard("GET", "/api/sys/dict/queryTableData", true)

	// ---- 无权限占位（Result.error 语义，供 403 页面轮询）----
	group.GET("/sys/common/403", func(r *ghttp.Request) {
		failJSON(r, "没有权限，请联系管理员授权")
	})
	guard("GET", "/api/sys/common/403", true)

	// ---- transitRESTful：Java 为任意 URL 中转（SSRF 面）----
	// 依项目安全规范（占位符绑定 + 默认拒绝）显式不实现代理能力，
	// 返回明确错误而非静默 404，便于前端提示与运维定位。
	group.GET("/sys/common/transitRESTful", transitRejected)
	group.POST("/sys/common/transitRESTful", transitRejected)
	guard("GET", "/api/sys/common/transitRESTful", true)
	guard("POST", "/api/sys/common/transitRESTful", true)

	// ---- OSS 文件域（4 端点，本地存储后端）----
	group.GET("/oss/file/list", ossCtrl.List)
	group.POST("/oss/file/upload", ossCtrl.Upload)
	group.DELETE("/oss/file/delete", ossCtrl.Delete)
	group.GET("/oss/file/queryById", ossCtrl.QueryByID)
	guard("GET", "/api/oss/file/list", true)
	guard("POST", "/api/oss/file/upload", true)
	guard("DELETE", "/api/oss/file/delete", true)
	guard("GET", "/api/oss/file/queryById", true)

	// ---- 工作台真实数据（前端 MockController 的 /api/* 契约，双前缀 /api/api/*）----
	group.GET("/api/workplace/teams", workplaceTeams)
	group.GET("/api/workplace/activity", workplaceActivity)
	group.GET("/api/workplace/radar", workplaceRadar)
	group.GET("/api/role", workplaceRoles)
	group.GET("/api/user", workplaceUsers)
	group.GET("/api/list/search/projects", workplaceProjects)
	group.GET("/api/task/process", workplaceTaskProcess)
	group.GET("/api/service", workplaceServices)
	for _, p := range []string{
		"/api/api/workplace/teams", "/api/api/workplace/activity", "/api/api/workplace/radar",
		"/api/api/role", "/api/api/user", "/api/api/list/search/projects",
		"/api/api/task/process", "/api/api/service",
	} {
		guard("GET", p, true)
	}
}

// transitRejected 显式中转拒绝（SSRF 防护，功能层面不迁移）。
func transitRejected(r *ghttp.Request) {
	failJSON(r, "中转请求能力未启用（安全策略：禁止服务端任意 URL 代理）")
}

// workplaceTeams 团队/成员概览：按班级聚合学员数。
func workplaceTeams(r *ghttp.Request) {
	// 形状对齐 Workplace.vue：团队卡片读 item.name / item.avatar
	const q = `SELECT d.id, d.depart_name AS name, COUNT(ud.user_id) AS memberCount
FROM sys_depart d LEFT JOIN sys_user_depart ud ON ud.dep_id = d.id
WHERE d.org_category = '3' AND d.del_flag = 0
GROUP BY d.id, d.depart_name ORDER BY d.depart_order ASC`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id": row["id"].String(), "name": row["name"].String(),
			"avatar": nil, "memberCount": row["memberCount"].Int(),
		})
	}
	okJSON(r, out)
}

// workplaceActivity 最近动态：作品提交 + 资讯发布合成时间线。
func workplaceActivity(r *ghttp.Request) {
	// 形状对齐 Workplace.vue：{user:{avatar,nickname}, project:{name,action,event}, time}
	const q = `SELECT * FROM (
  SELECT '提交作品' AS action, w.work_name AS event, u.realname AS nickname, w.create_time AS createTime
  FROM teaching_work w LEFT JOIN sys_user u ON u.id = w.user_id WHERE w.work_status >= 1
  UNION ALL
  SELECT '发布资讯' AS action, n.news_title AS event, COALESCE(n.create_by, 'system') AS nickname, n.create_time AS createTime
  FROM teaching_news n WHERE n.news_status >= 1
) t ORDER BY t.createTime DESC LIMIT 8`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"user":    map[string]interface{}{"avatar": nil, "nickname": row["nickname"].String()},
			"project": map[string]interface{}{"name": "春晖青少年编程中心", "action": row["action"].String(), "event": row["event"].String()},
			"time":    row["createTime"].String(),
		})
	}
	okJSON(r, out)
}

// workplaceRadar 能力雷达：按课程维度统计作业提交/点赞/评论密度。
func workplaceRadar(r *ghttp.Request) {
	// 形状对齐 radar 组件：DataSet fold([个人,团队,部门]) 前每行需含 name + 三个数值
	const q = `SELECT c.course_name AS name,
  COUNT(DISTINCT w.id) AS works,
  COALESCE(SUM(w.star_num), 0) AS stars,
  (SELECT COUNT(*) FROM teaching_work_comment wc JOIN teaching_work w2 ON w2.id = wc.work_id WHERE w2.course_id = c.id) AS comments,
  (SELECT COUNT(*) FROM teaching_course_unit u WHERE u.course_id = c.id AND u.del_flag = 0) AS units
FROM teaching_course c LEFT JOIN teaching_work w ON w.course_id = c.id AND w.del_flag = 0
WHERE c.del_flag = 0 GROUP BY c.id, c.course_name ORDER BY c.order_num ASC`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"name": row["name"].String(),
			"个人":   row["works"].Int() * 10,
			"团队":   row["stars"].Int() * 4,
			"部门":   row["comments"].Int()*8 + row["units"].Int()*5,
		})
	}
	okJSON(r, out)
}

// workplaceRoles 角色分布（工作台 /api/role）。
func workplaceRoles(r *ghttp.Request) {
	const q = `SELECT r.id, r.role_name AS name, r.role_code AS code, COUNT(ur.user_id) AS userCount
FROM sys_role r LEFT JOIN sys_user_role ur ON ur.role_id = r.id GROUP BY r.id, r.role_name, r.role_code`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id": row["id"].String(), "name": row["name"].String(),
			"code": row["code"].String(), "userCount": row["userCount"].Int(),
		})
	}
	okJSON(r, out)
}

// workplaceUsers 用户概览（工作台 /api/user）。
func workplaceUsers(r *ghttp.Request) {
	const q = `SELECT id, username, realname, post, user_identity, status FROM sys_user
WHERE del_flag = 0 ORDER BY create_time ASC LIMIT 20`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id": row["id"].String(), "username": row["username"].String(),
			"realname": row["realname"].String(), "post": row["post"].String(),
			"userIdentity": row["user_identity"].Int(), "status": row["status"].Int(),
		})
	}
	okJSON(r, out)
}

// workplaceProjects 课程项目卡片（工作台 /api/list/search/projects）。
func workplaceProjects(r *ghttp.Request) {
	// 形状对齐 Workplace.vue：res.result.data[]，字段 title/description/cover
	const q = `SELECT c.id, c.course_name AS name, c.course_desc AS description, c.course_cover AS cover,
  c.order_num AS orderNum,
  (SELECT COUNT(*) FROM teaching_course_unit u WHERE u.course_id = c.id AND u.del_flag = 0) AS unitCount,
  (SELECT COUNT(*) FROM teaching_work w WHERE w.course_id = c.id AND w.del_flag = 0) AS workCount
FROM teaching_course c WHERE c.del_flag = 0 ORDER BY c.order_num ASC`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		desc := stripTags(row["description"].String())
		if len([]rune(desc)) > 60 {
			desc = string([]rune(desc)[:60]) + "…"
		}
		items = append(items, map[string]interface{}{
			"id": row["id"].String(), "title": row["name"].String(),
			"description": desc, "cover": row["cover"].String(),
			"unitCount": row["unitCount"].Int(), "workCount": row["workCount"].Int(),
		})
	}
	// Jeecg 前端读 res.result.data
	r.Response.WriteJson(g.Map{"success": true, "code": 200, "message": "操作成功！",
		"result": g.Map{"data": items, "total": len(items)}})
}

// workplaceServices 平台服务健康（工作台 getServiceList 仅打日志，保持契约可调）
func workplaceServices(r *ghttp.Request) {
	okJSON(r, []map[string]interface{}{
		{"id": "api", "name": "教学平台 API", "status": "up"},
		{"id": "db", "name": "MariaDB 数据库", "status": "up"},
		{"id": "storage", "name": "本地文件存储", "status": "up"},
	})
}

// workplaceTaskProcess 流程进度（工作台 /api/task/process）：课程完成度近似。
func workplaceTaskProcess(r *ghttp.Request) {
	const q = `SELECT c.course_name AS name,
  (SELECT COUNT(*) FROM teaching_course_unit u WHERE u.course_id = c.id AND u.del_flag = 0) AS totalUnits,
  (SELECT COUNT(DISTINCT w.course_id) FROM teaching_work w WHERE w.course_id = c.id AND w.work_status >= 3) AS doneUnits
FROM teaching_course c WHERE c.del_flag = 0 ORDER BY c.order_num ASC`
	rows, err := g.DB().GetAll(r.Context(), q)
	if err != nil {
		failJSON(r, "查询失败")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		total := row["totalUnits"].Int()
		done := row["doneUnits"].Int()
		percent := 0
		if total > 0 {
			percent = done * 100 / total
			if percent > 100 {
				percent = 100
			}
		}
		out = append(out, map[string]interface{}{
			"name": row["name"].String(), "total": total, "done": done, "percent": percent,
		})
	}
	okJSON(r, out)
}

// stripTags 去除富文本标签（工作台卡片摘要用）。
func stripTags(s string) string {
	out := strings.Builder{}
	inTag := false
	for _, ch := range s {
		switch {
		case ch == '<':
			inTag = true
		case ch == '>':
			inTag = false
		case !inTag:
			out.WriteRune(ch)
		}
	}
	return strings.TrimSpace(out.String())
}
