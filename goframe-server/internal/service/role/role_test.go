// role_test.go 角色域零 DB 依赖单测：树构建、编码冲突判定、排序白名单、
// drChecked 规整。契约语义出处见 role.go 各方法注释。
package role

import (
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

func perm(id, parentID string, isLeaf int) entity.SysPermission {
	return entity.SysPermission{Id: id, ParentId: parentID, Name: "菜单" + id, IsLeaf: isLeaf, SortNo: 1}
}

func TestBuildTreeModelTree(t *testing.T) {
	// 结构：根A(非叶) ── A1(叶) / A2(非叶) ── A2a(叶)；根B(叶)；孤儿C(父不存在，丢弃)
	perms := []entity.SysPermission{
		perm("A", "", 0),
		perm("A1", "A", 1),
		perm("A2", "A", 0),
		perm("A2a", "A2", 1),
		perm("B", "", 1),
		perm("C", "GHOST", 1),
	}
	tree := buildTreeModelTree(perms)

	if len(tree) != 2 {
		t.Fatalf("根节点数 = %d, 期望 2（孤儿节点应被丢弃）: %+v", len(tree), tree)
	}
	a, b := tree[0], tree[1]
	if a.Key != "A" || b.Key != "B" {
		t.Fatalf("根节点顺序异常: %s, %s", a.Key, b.Key)
	}
	if !b.IsLeaf || b.Children != nil {
		t.Fatalf("叶子节点 B 应 isLeaf=true 且 children=null（对齐 Java 未初始化字段）: %+v", b)
	}
	if a.IsLeaf || len(a.Children) != 2 {
		t.Fatalf("非叶节点 A 应挂 2 个子节点: %+v", a)
	}
	a1, a2 := a.Children[0], a.Children[1]
	if a1.Key != "A1" || !a1.IsLeaf || a1.Children != nil {
		t.Fatalf("A1 叶子节点形态异常: %+v", a1)
	}
	if a2.IsLeaf || len(a2.Children) != 1 || a2.Children[0].Key != "A2a" {
		t.Fatalf("A2 子树形态异常: %+v", a2)
	}
	// TreeModel 字段映射（TreeModel(SysPermission) 构造器）
	if a.Title != a.SlotTitle || a.Value != a.Key || a.Label != a.Title {
		t.Fatalf("TreeModel 字段映射异常: %+v", a)
	}
	if a.ParentId != "" {
		t.Fatalf("根节点 parentId 应为空串（parent_id 列 NOT NULL DEFAULT ''）: %q", a.ParentId)
	}
}

func TestCodeConflict(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		foundID string
		want    bool
	}{
		{"新增模式编码已存在", "", "role1", true},
		{"新增模式编码不存在", "", "", false},
		{"编辑模式编码属于自己", "role1", "role1", false},
		{"编辑模式编码被他人占用", "role2", "role1", true},
	}
	for _, tc := range cases {
		if got := codeConflict(tc.id, tc.foundID); got != tc.want {
			t.Errorf("%s: codeConflict(%q,%q) = %v, 期望 %v", tc.name, tc.id, tc.foundID, got, tc.want)
		}
	}
}

func TestNormalizeOrder(t *testing.T) {
	cases := []struct {
		name      string
		column    string
		orderType string
		wantCol   string
		wantDesc  bool
		wantOK    bool
	}{
		{"空列名不排序", "", "asc", "", false, false},
		{"空排序类型不排序", "roleName", "", "", false, false},
		{"驼峰转下划线升序", "roleName", "asc", "role_name", false, true},
		{"dictText 后缀剥离", "roleName_dictText", "desc", "role_name", true, true},
		{"url 后缀剥离", "createTime_url", "ASC", "create_time", false, true},
		{"非白名单列拒绝", "password", "asc", "", false, false},
		{"注入列拒绝", "role_name;drop", "asc", "", false, false},
		{"非 ASC 判为降序", "roleCode", "DESC", "role_code", true, true},
	}
	for _, tc := range cases {
		col, desc, ok := normalizeOrder(tc.column, tc.orderType)
		if ok != tc.wantOK || col != tc.wantCol || desc != tc.wantDesc {
			t.Errorf("%s: normalizeOrder(%q,%q) = (%q,%v,%v), 期望 (%q,%v,%v)",
				tc.name, tc.column, tc.orderType, col, desc, ok, tc.wantCol, tc.wantDesc, tc.wantOK)
		}
	}
}

func TestNormalizeDrChecked(t *testing.T) {
	cases := []struct {
		raw      string
		want     string
		wantHave bool
	}{
		{"", "", false},
		{"rule1,rule2", "rule1,rule2", true},
		{"rule1,rule2,", "rule1,rule2", true},
		{"rule1,,", "rule1,", true}, // 只去末尾一个逗号（对齐 Java substring）
	}
	for _, tc := range cases {
		got, have := normalizeDrChecked(tc.raw)
		if got != tc.want || have != tc.wantHave {
			t.Errorf("normalizeDrChecked(%q) = (%q,%v), 期望 (%q,%v)", tc.raw, got, have, tc.want, tc.wantHave)
		}
	}
}

func TestNewTreeModelLeafFlag(t *testing.T) {
	// is_leaf=1 → 叶子（children null）；is_leaf=0 → 非叶（children 空数组非 null）
	leaf := newTreeModel(perm("L", "", 1))
	if !leaf.IsLeaf || leaf.Children != nil {
		t.Fatalf("叶子节点形态异常: %+v", leaf)
	}
	branch := newTreeModel(perm("N", "", 0))
	if branch.IsLeaf || branch.Children == nil || len(branch.Children) != 0 {
		t.Fatalf("非叶节点应初始化空 children 数组（对齐 Java new ArrayList）: %+v", branch)
	}
}
