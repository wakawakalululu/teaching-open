// menu_test.go 菜单权限域零 DB 依赖单测：intelligentProcessData 预处理、
// 授权 diff 算法、菜单树构建。契约语义出处见 menu.go 各函数注释。
package permission

import (
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

func strp(s string) *string { return &s }
func intp(i int) *int       { return &i }

func TestIntelligentProcessDataComponent(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"去首个前导斜杠后再被 views/ 命中（Java 顺序执行）", "/views/Foo", "Foo"},
		{"views/ 前缀剥离", "views/Foo", "Foo"},
		{"Java 固有怪癖 src/views 先被 views/ 命中", "src/views/Foo", "src/Foo"},
		{"vue 后缀剥离", "views/Foo.vue", "Foo"},
		{"无前缀保持", "system/UserList", "system/UserList"},
	}
	for _, tc := range cases {
		got := intelligentProcessData(MenuInput{Component: strp(tc.in)})
		if *got.Component != tc.want {
			t.Errorf("%s: component %q → %q, 期望 %q", tc.name, tc.in, *got.Component, tc.want)
		}
	}
}

func TestIntelligentProcessDataUrl(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"相对路径补斜杠", "dashboard/analysis", "/dashboard/analysis"},
		{"已有斜杠保持", "/dashboard", "/dashboard"},
		{"http 外链保持", "http://a.com/x", "http://a.com/x"},
		{"https 外链保持", "https://a.com/x", "https://a.com/x"},
		{"vue 后缀剥离", "dashboard.vue", "/dashboard"},
		{"JS 片段占位保持", "{{ JS 代码 }}", "{{ JS 代码 }}"},
		{"带前导空格的占位保持（trim 判定命中，不补斜杠）", "  {{JS}}", "  {{JS}}"},
	}
	for _, tc := range cases {
		got := intelligentProcessData(MenuInput{Url: strp(tc.in)})
		if *got.Url != tc.want {
			t.Errorf("%s: url %q → %q, 期望 %q", tc.name, tc.in, *got.Url, tc.want)
		}
	}
}

func TestIntelligentProcessDataDefaultComponent(t *testing.T) {
	// menuType=0 且组件为空 → 默认 layouts/RouteView
	got := intelligentProcessData(MenuInput{MenuType: intp(0), Component: strp("")})
	if *got.Component != "layouts/RouteView" {
		t.Errorf("一级菜单空组件应取默认值, got %q", *got.Component)
	}
	// 组件非空不覆盖
	got = intelligentProcessData(MenuInput{MenuType: intp(0), Component: strp("views/Foo")})
	if *got.Component != "Foo" {
		t.Errorf("一级菜单非空组件不应被默认值覆盖, got %q", *got.Component)
	}
	// menuType 未提供不触发默认值
	got = intelligentProcessData(MenuInput{Component: strp("")})
	if *got.Component != "" {
		t.Errorf("menuType 缺失不应赋默认组件, got %q", *got.Component)
	}
}

func TestDiffPermIDs(t *testing.T) {
	cases := []struct {
		name string
		main string
		diff string
		want []string
	}{
		{"diff 为空返回 nil", "a,b", "", nil},
		{"main 为空返回 diff 全量", "", "a,b", []string{"a", "b"}},
		{"取 diff 中 main 没有的元素", "a,b", "b,c,d", []string{"c", "d"}},
		{"全等返回空", "a,b", "b,a", nil},
		{"空段跳过", "a", ",b,", []string{"b"}},
	}
	for _, tc := range cases {
		got := diffPermIDs(tc.main, tc.diff)
		if len(got) != len(tc.want) {
			t.Errorf("%s: diffPermIDs(%q,%q) = %v, 期望 %v", tc.name, tc.main, tc.diff, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s: diffPermIDs(%q,%q)[%d] = %q, 期望 %q", tc.name, tc.main, tc.diff, i, got[i], tc.want[i])
			}
		}
	}
}

func perm(id, parentID string, isLeaf int) entity.SysPermission {
	return entity.SysPermission{Id: id, ParentId: parentID, Name: "菜单" + id, IsLeaf: isLeaf, SortNo: 1, MenuType: 1}
}

func TestBuildMenuTree(t *testing.T) {
	// 结构：根A(非叶) ── A1(叶) / A2(非叶) ── A2a(叶)；根B(叶)；孤儿C(丢弃)
	perms := []entity.SysPermission{
		perm("A", "", 0),
		perm("A1", "A", 1),
		perm("A2", "A", 0),
		perm("A2a", "A2", 1),
		perm("B", "", 1),
		perm("C", "GHOST", 1),
	}
	tree := buildMenuTree(perms)
	if len(tree) != 2 {
		t.Fatalf("根节点数 = %d, 期望 2: %+v", len(tree), tree)
	}
	a, b := tree[0], tree[1]
	if a.Key != "A" || a.Id != "A" || a.Title != "菜单A" {
		t.Fatalf("SysPermissionTree 字段映射异常: %+v", a)
	}
	if !b.IsLeaf || b.Children != nil {
		t.Fatalf("叶子 B 应 isLeaf=true 且 children=null: %+v", b)
	}
	if a.IsLeaf || len(a.Children) != 2 || a.Children[1].Children[0].Id != "A2a" {
		t.Fatalf("A 子树形态异常: %+v", a)
	}
}

func TestBuildMenuTreeListModel(t *testing.T) {
	perms := []entity.SysPermission{
		perm("A", "", 0),
		perm("A1", "A", 1),
		perm("B", "", 1),
	}
	tree := buildMenuTreeListModel(perms)
	if len(tree) != 2 {
		t.Fatalf("根节点数 = %d, 期望 2", len(tree))
	}
	if tree[0].SlotTitle != tree[0].Title || tree[0].Value != tree[0].Key || tree[0].Label != tree[0].Title {
		t.Fatalf("TreeModel 字段映射异常: %+v", tree[0])
	}
	if tree[0].IsLeaf || len(tree[0].Children) != 1 {
		t.Fatalf("TreeModel 子节点挂载异常: %+v", tree[0])
	}
}
