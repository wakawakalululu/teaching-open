// seed_probe_test.go 种子探针纯函数单测。
package cmd

import (
	"strings"
	"testing"
)

func TestEmptyTableWarnings(t *testing.T) {
	all := map[string]int{}
	for _, t := range adminCriticalTables {
		all[t] = 1
	}
	if w := emptyTableWarnings(all); len(w) != 0 {
		t.Fatalf("全表非空时不应有告警: %v", w)
	}
	all["sys_position"] = 0
	all["sys_data_source"] = 0
	w := emptyTableWarnings(all)
	if len(w) != 2 {
		t.Fatalf("应有 2 条告警，实际 %d: %v", len(w), w)
	}
	joined := strings.Join(w, ",")
	if !strings.Contains(joined, "sys_position") || !strings.Contains(joined, "sys_data_source") {
		t.Fatalf("告警内容缺少空表名: %v", w)
	}
	if len(adminCriticalTables) < 10 {
		t.Fatalf("关键表清单过少: %d", len(adminCriticalTables))
	}
}
