// registry_test.go Excel 域规格一致性回归测试（P1-2）。
//
// 断言：每个已注册 Excel 域（specs）的 base 都能映射到物理表，且其每一列
// 都真实存在于 api/db/teachingopen2.8.sql 的 CREATE TABLE 中。
// 测试直接解析 DDL 文件，无需数据库连接，可进 CI。
package excelx

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	reCreateTable = regexp.MustCompile("CREATE TABLE `([a-z0-9_]+)` \\(")
	reColumn      = regexp.MustCompile("^\\s+`([a-z0-9_]+)`\\s")
)

// TestSpecsMatchSchema 校验 Excel 列规格与 DDL 列定义一致。
func TestSpecsMatchSchema(t *testing.T) {
	ddl := findDDL(t)
	raw, err := os.ReadFile(ddl)
	if err != nil {
		t.Fatalf("读取 DDL 失败: %v", err)
	}
	// 解析 CREATE TABLE → 列名集合
	tables := map[string]map[string]struct{}{}
	lines := strings.Split(string(raw), "\n")
	var cur string
	for _, line := range lines {
		if m := reCreateTable.FindStringSubmatch(line); m != nil {
			cur = m[1]
			tables[cur] = map[string]struct{}{}
			continue
		}
		if cur == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), ")") {
			cur = ""
			continue
		}
		if m := reColumn.FindStringSubmatch(line); m != nil {
			tables[cur][m[1]] = struct{}{}
		}
	}
	if len(tables) == 0 {
		t.Fatalf("DDL 解析结果为空: %s", ddl)
	}

	var problems []string
	for _, spec := range specs {
		table := TableByBase(spec.Base)
		if table == "" {
			problems = append(problems, spec.Base+" 缺少 base→table 映射")
			continue
		}
		cols, ok := tables[table]
		if !ok {
			problems = append(problems, spec.Base+" 映射的表 "+table+" 不在 DDL 中")
			continue
		}
		if len(spec.Cols) == 0 {
			problems = append(problems, spec.Base+" 列规格为空")
			continue
		}
		for _, col := range spec.Cols {
			if _, ok := cols[col.Field]; !ok {
				problems = append(problems, spec.Base+" 列 "+col.Field+" 不存在于表 "+table)
			}
		}
	}
	// 反向：每个映射的 base 都应有 spec（避免「注册了但无列规格」）
	for base := range SpecTables() {
		if SpecByBase(base) == nil {
			problems = append(problems, base+" 有表映射但缺少 Spec 列规格")
		}
	}
	if len(problems) > 0 {
		t.Fatalf("Excel 规格与 DDL 不一致 %d 处:\n  %s", len(problems), strings.Join(problems, "\n  "))
	}
	t.Logf("Excel 域 %d 个，列定义与 DDL 全部一致", len(specs))
}

// findDDL 向上查找 api/db/teachingopen2.8.sql（相对 goframe-server 运行目录）。
func findDDL(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"../../../../api/db/teachingopen2.8.sql",
		"../../../api/db/teachingopen2.8.sql",
		"../../../api/db/teachingopen2.8.sql",
		"../../../db/teachingopen2.8.sql",
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if st, err := os.Stat(abs); err == nil && !st.IsDir() {
				return abs
			}
		}
	}
	t.Skip("未找到 teachingopen2.8.sql（独立检出场景跳过）")
	return ""
}
