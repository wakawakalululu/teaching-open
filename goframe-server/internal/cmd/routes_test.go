// routes_test.go 路由一致性回归测试（P1-1）。
//
// 背景：历史上出现过「守卫已声明、路由未挂」的漂移（/sys/selectDepart、
// /sys/loginfo、/sys/visitInfo 曾 404 但守卫存在）。本测试以源码为输入，
// 断言每条 registerGuard/guard 声明的 <METHOD, PATH> 都能在路由注册中找到
// 对应项；/api/sys 与 /api 两个路由组按缩进跟踪前缀，动态注册的域
// （excelxBases / p2_6 Base）按已知后缀展开后再比对。
package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var (
	reRegistered = regexp.MustCompile(`group\.(GET|POST|PUT|DELETE)\("([^"]+)"`)
	reGuard      = regexp.MustCompile(`(?:registerGuard|guard)\(\s*"(GET|POST|PUT|DELETE)"\s*,\s*"/api([^"]*)"`)
	reBaseLit    = regexp.MustCompile(`Base:\s*"([^"]+)"`)
	reStrLit     = regexp.MustCompile(`"(/[a-zA-Z0-9_/]+)"`)
	reGroupOpen  = regexp.MustCompile(`s\.Group\("(/api[^"]*)"`)
	reGroupClose = regexp.MustCompile(`^\t\}\)$`)
)

// TestRouteGuardsAreRegistered 断言守卫声明与路由注册一一对应。
func TestRouteGuardsAreRegistered(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("列举源文件失败: %v", err)
	}
	registered := map[string]struct{}{}
	guards := map[string]string{}

	var excelBases, p26Bases []string

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", f, err)
		}
		lines := strings.Split(string(raw), "\n")

		// 仅 cmd.go 使用 s.Group 分组；其余文件在既有 group 内相对注册。
		prefix := ""
		for _, line := range lines {
			if m := reGroupOpen.FindStringSubmatch(line); m != nil && f == "cmd.go" {
				prefix = strings.TrimSuffix(m[1], "/")
				continue
			}
			if f == "cmd.go" && reGroupClose.MatchString(line) {
				prefix = ""
				continue
			}
			for _, m := range reRegistered.FindAllStringSubmatch(line, -1) {
				p := m[2]
				// /api/sys 组内的 "/sys/login" 等写法已是绝对语义，仅纯相对路径拼接前缀
				if prefix != "" && !strings.HasPrefix(p, "/sys/") && prefix != "/api" {
					p = prefix + p
				}
				registered[m[1]+" "+strings.TrimPrefix(p, "/api")] = struct{}{}
			}
		}

		for _, m := range reGuard.FindAllStringSubmatch(string(raw), -1) {
			if strings.TrimSpace(m[2]) == "" {
				continue
			}
			guards[m[1]+" "+m[2]] = f
		}

		if f == "register_p2_6.go" {
			if idx := strings.Index(string(raw), "var excelxBases = []string{"); idx >= 0 {
				if end := strings.Index(string(raw)[idx:], "\n}"); end > 0 {
					for _, m := range reStrLit.FindAllStringSubmatch(string(raw)[idx:idx+end], -1) {
						excelBases = append(excelBases, m[1])
					}
				}
			}
			for _, m := range reBaseLit.FindAllStringSubmatch(string(raw), -1) {
				p26Bases = append(p26Bases, m[1])
			}
		}
	}

	// 动态展开：Excel 能力
	for _, b := range excelBases {
		registered["GET "+b+"/exportXls"] = struct{}{}
		registered["POST "+b+"/exportXls"] = struct{}{}
		registered["POST "+b+"/importExcel"] = struct{}{}
	}
	// 动态展开：p2_6 泛型域 CRUD
	for _, b := range p26Bases {
		for _, c := range []struct{ method, suffix string }{
			{"GET", "/list"}, {"POST", "/add"}, {"PUT", "/edit"},
			{"DELETE", "/delete"}, {"DELETE", "/deleteBatch"}, {"GET", "/queryById"},
		} {
			registered[c.method+" "+b+c.suffix] = struct{}{}
		}
	}

	// 通配符归一：/** 与 *path 视为同一端点
	normWild := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "*path", "*"), "/**", "/*")
	}
	registeredNorm := map[string]struct{}{}
	for k := range registered {
		registeredNorm[normWild(k)] = struct{}{}
	}

	var missing []string
	for key, file := range guards {
		if _, ok := registeredNorm[normWild(key)]; ok {
			continue
		}
		missing = append(missing, key+"  ("+file+")")
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("存在「守卫已声明但路由未注册」的端点 %d 条:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
	t.Logf("守卫 %d 条 / 注册 %d 条，全部对应", len(guards), len(registered))
}

// collectRegisteredRoutes 解析源码并展开动态域，返回 "METHOD /path"（无 /api 前缀）。
func collectRegisteredRoutes(t *testing.T) map[string]struct{} {
	t.Helper()
	registered := map[string]struct{}{}
	files, _ := filepath.Glob("*.go")
	var excelBases, p26Bases []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		prefix := ""
		for _, line := range strings.Split(string(raw), "\n") {
			if m := reGroupOpen.FindStringSubmatch(line); m != nil && f == "cmd.go" {
				prefix = strings.TrimSuffix(m[1], "/")
				continue
			}
			if f == "cmd.go" && reGroupClose.MatchString(line) {
				prefix = ""
				continue
			}
			for _, m := range reRegistered.FindAllStringSubmatch(line, -1) {
				p := m[2]
				if prefix != "" && !strings.HasPrefix(p, "/sys/") && prefix != "/api" {
					p = prefix + p
				}
				registered[m[1]+" "+strings.TrimPrefix(p, "/api")] = struct{}{}
			}
		}
		if f == "register_p2_6.go" {
			if idx := strings.Index(string(raw), "var excelxBases = []string{"); idx >= 0 {
				if end := strings.Index(string(raw)[idx:], "\n}"); end > 0 {
					for _, m := range reStrLit.FindAllStringSubmatch(string(raw)[idx:idx+end], -1) {
						excelBases = append(excelBases, m[1])
					}
				}
			}
			for _, m := range reBaseLit.FindAllStringSubmatch(string(raw), -1) {
				p26Bases = append(p26Bases, m[1])
			}
		}
	}
	for _, b := range excelBases {
		registered["GET "+b+"/exportXls"] = struct{}{}
		registered["POST "+b+"/exportXls"] = struct{}{}
		registered["POST "+b+"/importExcel"] = struct{}{}
	}
	for _, b := range p26Bases {
		for _, c := range []struct{ method, suffix string }{
			{"GET", "/list"}, {"POST", "/add"}, {"PUT", "/edit"},
			{"DELETE", "/delete"}, {"DELETE", "/deleteBatch"}, {"GET", "/queryById"},
		} {
			registered[c.method+" "+b+c.suffix] = struct{}{}
		}
	}
	return registered
}
