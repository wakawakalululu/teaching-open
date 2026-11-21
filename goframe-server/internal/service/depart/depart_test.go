// depart_test.go 部门域零 DB 依赖单测：树构建（isLeaf 由子节点存在性计算）、
// YouBian 编码副本、输入拷贝。
package depart

import (
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

func dep(id, parentID, orgCode string) entity.SysDepart {
	return entity.SysDepart{Id: id, ParentId: parentID, DepartName: "部门" + id,
		OrgCode: orgCode, OrgType: "1", DelFlag: "0", DepartOrder: 1}
}

func TestBuildDepartTree(t *testing.T) {
	// 结构：根A(orgCode A01,非叶) ── A1(A011,叶) / A2(A012,非叶) ── A2a；根B；孤儿C
	rows := []entity.SysDepart{
		dep("A", "", "A01"),
		dep("A1", "A", "A011"),
		dep("A2", "A", "A012"),
		dep("A2a", "A2", "A0121"),
		dep("B", "", "B01"),
		dep("C", "GHOST", "C01"),
	}
	tree := buildDepartTree(rows)
	if len(tree) != 2 {
		t.Fatalf("根节点数 = %d, 期望 2（孤儿丢弃）: %+v", len(tree), tree)
	}
	a, b := tree[0], tree[1]
	if a.Key != "A" || a.Value != "A" || a.Title != "部门A" {
		t.Fatalf("SysDepartTreeModel 字段映射异常: %+v", a)
	}
	// isLeaf 由子节点存在性计算（sys_depart 无 is_leaf 列）
	if !b.IsLeaf || b.Children != nil {
		t.Fatalf("叶子 B 应 isLeaf=true 且 children=null: %+v", b)
	}
	if a.IsLeaf || len(a.Children) != 2 {
		t.Fatalf("非叶 A 应挂 2 个子节点: %+v", a)
	}
	if a.Children[0].Id != "A1" || !a.Children[0].IsLeaf || a.Children[0].Children != nil {
		t.Fatalf("A1 叶子形态异常: %+v", a.Children[0])
	}
	if a.Children[1].IsLeaf || a.Children[1].Children[0].Id != "A2a" {
		t.Fatalf("A2 子树形态异常: %+v", a.Children[1])
	}
	// ParentId 指针语义：根为空串
	if a.ParentId == nil || *a.ParentId != "" {
		t.Fatalf("根节点 parentId 应为空串指针: %+v", a.ParentId)
	}
}

func TestBuildDepartTreeMyDeptRootOverride(t *testing.T) {
	// 我的部门树：命中前缀的部门 parentId 被置空成根（服务层已改写后传入）
	rows := []entity.SysDepart{
		dep("A", "", "A01"),
		dep("A1", "A", "A011"),
	}
	tree := buildDepartTree(rows)
	if len(tree) != 1 || tree[0].Id != "A" || len(tree[0].Children) != 1 {
		t.Fatalf("前缀根树形态异常: %+v", tree)
	}
}

func TestYouBianDepartCopy(t *testing.T) {
	// 与 category 域同一移植源的冒烟校验
	cases := []struct{ in, want string }{
		{"", "A01"},
		{"A01", "A02"},
		{"A99", "B01"},
		{"Z99", "Z99A01"},
		{"A01Z99", "A01Z99A01"},
	}
	for _, tc := range cases {
		if got := nextYouBianCode(tc.in); got != tc.want {
			t.Errorf("nextYouBianCode(%q) = %q, 期望 %q", tc.in, got, tc.want)
		}
	}
	if got := subYouBianCode("A01", ""); got != "A01A01" {
		t.Errorf("subYouBianCode 首子级 = %q, 期望 A01A01", got)
	}
	if got := subYouBianCode("A01", "A01B01"); got != "A01B02" {
		t.Errorf("subYouBianCode 同级递增 = %q, 期望 A01B02", got)
	}
}

func TestApplyDepartInput(t *testing.T) {
	name := "研发部"
	order := 3
	data := do.SysDepart{}
	applyDepartInput(&data, DepartInput{DepartName: &name, DepartOrder: &order})
	if data.DepartName != name || data.DepartOrder != order {
		t.Fatalf("字段拷贝异常: %+v", data)
	}
	// 未提供字段保持 nil（MP NOT_NULL 语义）
	if data.Mobile != nil || data.OrgCode != nil {
		t.Fatalf("未提供字段应为 nil: %+v", data)
	}
}
