package permission

import (
	"strings"
	"testing"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// 菜单树组装是前端动态路由的数据源，形状必须与 Java getPermissionJsonArray 一致。
func TestBuildMenuArray(t *testing.T) {
	list := []entity.SysPermission{
		{Id: "1", ParentId: "", Name: "首页", Url: "/dashboard/analysis", Component: "dashboard/Analysis", MenuType: 0, IsRoute: 1, IsLeaf: 1, KeepAlive: 1},
		{Id: "2", ParentId: "1", Name: "分析页", Url: "/dashboard/analysis/detail", Component: "dashboard/Detail", MenuType: 1, IsRoute: 1, IsLeaf: 1},
		{Id: "3", ParentId: "", Name: "系统管理", Url: "/isystem", Component: "layouts/RouteView", MenuType: 0, IsRoute: 1, IsLeaf: 0},
		{Id: "4", ParentId: "3", Name: "用户列表", Url: "/isystem/user", Component: "system/UserList", MenuType: 1, IsRoute: 1, IsLeaf: 1},
		{Id: "5", ParentId: "4", Name: "添加用户", Url: "", Perms: "system:user:add", MenuType: 2, Status: "1"}, // 按钮
		{Id: "6", ParentId: "999", Name: "孤儿菜单", Url: "/orphan", MenuType: 0, IsRoute: 1, IsLeaf: 1},        // 父节点缺失
	}

	root := buildMenuArray(list, "")
	if len(root) != 2 {
		t.Fatalf("顶层节点数 = %d, want 2（孤儿节点应被丢弃）, got %+v", len(root), root)
	}
	if root[0]["path"] != "/dashboard/analysis" {
		t.Errorf("root[0].path = %v", root[0]["path"])
	}
	// 子菜单归入 children
	children, ok := root[1]["children"].([]gMap)
	if !ok || len(children) != 1 {
		t.Fatalf("系统管理 children 数量异常: %+v", root[1])
	}
	if children[0]["path"] != "/isystem/user" {
		t.Errorf("children[0].path = %v", children[0]["path"])
	}
	// 按钮不进入路由树
	if _, has := children[0]["children"]; has {
		t.Errorf("按钮不应进入路由树")
	}
	// meta 形状
	meta := root[0]["meta"].(gMap)
	if meta["title"] != "首页" || meta["keepAlive"] != true || meta["internalOrExternal"] != false {
		t.Errorf("meta 形状异常: %+v", meta)
	}
	// route 标志为字符串
	if root[0]["route"] != "1" {
		t.Errorf("route = %v, want \"1\"", root[0]["route"])
	}
}

// gMap 便于测试内类型断言（g.Map 的别名）。
type gMap = map[string]interface{}

func TestMenuNodeExternalURL(t *testing.T) {
	p := &entity.SysPermission{
		Id: "9", Name: "官网", Url: "https://example.com/docs", MenuType: 0, IsRoute: 1, IsLeaf: 1,
		InternalOrExternal: 1,
	}
	node := menuNodeJSON(p)
	// 外链 path 为 MD5 摘要
	if path, _ := node["path"].(string); len(path) != 32 || strings.ContainsAny(path, "./") {
		t.Errorf("外链 path 应为 32 位 MD5, got %v", node["path"])
	}
	meta := node["meta"].(gMap)
	if meta["url"] != p.Url || meta["internalOrExternal"] != true {
		t.Errorf("外链 meta 异常: %+v", meta)
	}
}

func TestUrlToRouteName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"/isystem/role", "isystem-role"},
		{"isystem/role", "isystem-role"},
		{"/online/cform/:code", "online-cform-@code"},
		{"", ""},
	}
	for _, c := range cases {
		if got := urlToRouteName(c.in); got != c.want {
			t.Errorf("urlToRouteName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHasIndexPage(t *testing.T) {
	if hasIndexPage([]entity.SysPermission{{Component: "system/UserList"}}) {
		t.Error("无 dashboard/Index 时应返回 false")
	}
	if !hasIndexPage([]entity.SysPermission{{Component: indexComponent}}) {
		t.Error("含 dashboard/Index 时应返回 true")
	}
}

func TestBuildAuthArray(t *testing.T) {
	list := []entity.SysPermission{
		{MenuType: 2, Perms: "system:user:add", PermsType: "1", Name: "添加", Status: "1"},
		{MenuType: 2, Perms: "system:user:del", PermsType: "1", Name: "删除", Status: "0"}, // 无效不输出
		{MenuType: 1, Perms: "x", Status: "1"},                                           // 菜单不参与
	}
	auth := buildAuthArray(list)
	if len(auth) != 1 || auth[0]["action"] != "system:user:add" {
		t.Errorf("auth 数量/内容异常: %+v", auth)
	}
}
