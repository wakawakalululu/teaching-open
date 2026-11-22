// seed_probe.go 启动期种子完备度探针（P2）。
//
// 背景：全量替换 Mock 数据时曾清空系统表，导致管理页「接口正常但表格为空」，
// 用户感知为「后端能力缺失」。本探针在启动时统计管理页依赖表的行数，
// 空表打 WARN 并在日志中给出补救命令，把「感觉功能坏了」变成可观测告警。
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// adminCriticalTables 管理页依赖的关键表（空表即提示）。
var adminCriticalTables = []string{
	"sys_user", "sys_role", "sys_permission", "sys_depart", "sys_dict",
	"sys_position", "sys_data_source", "sys_sms_template", "sys_depart_role",
	"teaching_course", "teaching_course_unit", "sys_announcement",
}

// emptyTableWarnings 纯函数：输入表→行数，返回空表告警文本（便于单测）。
func emptyTableWarnings(counts map[string]int) []string {
	var warns []string
	for _, t := range adminCriticalTables {
		if counts[t] == 0 {
			warns = append(warns, t)
		}
	}
	return warns
}

// probeAdminTables 启动探针：统计关键表行数，空表 WARN。
// 表名来自固定白名单，不做动态拼接。
func probeAdminTables(ctx context.Context) {
	counts := make(map[string]int, len(adminCriticalTables))
	for _, t := range adminCriticalTables {
		rec, err := g.DB().GetOne(ctx, "SELECT COUNT(*) AS c FROM "+t)
		if err != nil {
			g.Log().Warningf(ctx, "种子探针：统计表 %s 失败 err=%v", t, err)
			continue
		}
		counts[t] = rec["c"].Int()
	}
	warns := emptyTableWarnings(counts)
	if len(warns) == 0 {
		g.Log().Infof(ctx, "种子探针：%d 张关键表均有数据", len(adminCriticalTables))
		return
	}
	g.Log().Warningf(ctx, "种子探针：%d 张管理页关键表为空 [%s]，相关页面会显示「暂无数据」。"+
		"请执行 tools/chseed -layer base（及 -layer demo）或 api/db/chunhui_seed_{base,demo}.sql 回填。",
		len(warns), strings.Join(warns, ", "))
	_ = fmt.Sprintf
}
