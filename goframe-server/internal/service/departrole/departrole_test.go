// departrole_test.go 部门角色域零 DB 依赖单测：diff 算法与 QueryGenerator
// 值规则判定。
package departrole

import "testing"

func TestDiffIDs(t *testing.T) {
	cases := []struct {
		name string
		main string
		diff string
		want []string
	}{
		{"diff 空", "a,b", "", nil},
		{"main 空", "", "a,b", []string{"a", "b"}},
		{"常规差集", "a,b", "b,c", []string{"c"}},
		{"全等", "b,a", "a,b", nil},
		{"空段过滤", "a", ",c,", []string{"c"}},
	}
	for _, tc := range cases {
		got := diffIDs(tc.main, tc.diff)
		if len(got) != len(tc.want) {
			t.Errorf("%s: diffIDs(%q,%q) = %v, 期望 %v", tc.name, tc.main, tc.diff, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s: diffIDs(%q,%q)[%d] = %q, 期望 %q", tc.name, tc.main, tc.diff, i, got[i], tc.want[i])
			}
		}
	}
}

func TestParseQueryRule(t *testing.T) {
	cases := []struct {
		raw      string
		wantKind string
		wantVal  string
	}{
		{"admin", "eq", "admin"},
		{"*管理*", "like", "%管理%"},
		{"a,b", "in", ""},
		{"!=x", "ne", "x"},
	}
	for _, tc := range cases {
		kind, val, _ := parseQueryRule(tc.raw)
		if kind != tc.wantKind || (tc.wantVal != "" && val != tc.wantVal) {
			t.Errorf("parseQueryRule(%q) = (%q,%q), 期望 (%q,%q)", tc.raw, kind, val, tc.wantKind, tc.wantVal)
		}
	}
}
