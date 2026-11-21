// departpermission_test.go 部门权限域零 DB 依赖单测：TreeModel 五参构造器
// 路径的字段语义与树构建。
package departpermission

import (
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

func perm(id, parentID string, isLeaf int, ruleFlag int) entity.SysPermission {
	return entity.SysPermission{Id: id, ParentId: parentID, Name: "菜单" + id,
		IsLeaf: isLeaf, RuleFlag: ruleFlag, SortNo: 1}
}

func TestNewDeptRoleTreeModel(t *testing.T) {
	// String 五参构造器路径：title=null、label=null、icon=null、slotTitle=name、
	// value=key、ruleFlag 透传、scopedSlots 固定 {"title":"hasDatarule"}
	node := newDeptRoleTreeModel(perm("P1", "", 1, 1))
	if node.Title != nil || node.Label != nil || node.Icon != nil {
		t.Fatalf("title/label/icon 应为 null（对齐 Java 五参构造器）: %+v", node)
	}
	if node.SlotTitle != "菜单P1" || node.Value != "P1" || node.Key != "P1" {
		t.Fatalf("slotTitle/value/key 映射异常: %+v", node)
	}
	if node.RuleFlag != 1 {
		t.Fatalf("ruleFlag 应透传: %+v", node)
	}
	if node.ScopedSlots == nil || node.ScopedSlots["title"] != "hasDatarule" {
		t.Fatalf("scopedSlots 应为 {title: hasDatarule}: %+v", node.ScopedSlots)
	}
	if !node.IsLeaf || node.Children != nil {
		t.Fatalf("叶子应 isLeaf=true 且 children=null: %+v", node)
	}
}

func TestBuildDeptRoleTree(t *testing.T) {
	perms := []entity.SysPermission{
		perm("A", "", 0, 0),
		perm("A1", "A", 1, 0),
		perm("B", "", 1, 0),
		perm("C", "GHOST", 1, 0),
	}
	tree := buildDeptRoleTree(perms)
	if len(tree) != 2 {
		t.Fatalf("根节点数 = %d, 期望 2（孤儿丢弃）", len(tree))
	}
	if tree[0].IsLeaf || len(tree[0].Children) != 1 || tree[0].Children[0].Key != "A1" {
		t.Fatalf("A 子树形态异常: %+v", tree[0])
	}
	if !tree[1].IsLeaf || tree[1].Children != nil {
		t.Fatalf("叶子 B 形态异常: %+v", tree[1])
	}
}

func TestDiffIDs(t *testing.T) {
	got := diffIDs("a,b", "b,c,d")
	if len(got) != 2 || got[0] != "c" || got[1] != "d" {
		t.Fatalf("diffIDs = %v, 期望 [c d]", got)
	}
	if got := diffIDs("a", ""); got != nil {
		t.Fatalf("diff 空应返回 nil, got %v", got)
	}
}
