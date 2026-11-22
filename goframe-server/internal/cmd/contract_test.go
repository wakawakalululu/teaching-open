// contract_test.go Java→Go 契约覆盖回归测试（P1-3）。
//
// 解析源仓库 api/**/Controller.java 的注解（跳过整行注释），与 Go 路由集合对账，
// 断言覆盖率不低于阈值，并把「有意不迁移」的端点显式列入白名单（附理由）。
// 该测试把 docs/BACKEND-GAP-ANALYSIS.md 的一次性对账脚本固化为持续门禁。
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var (
	reJavaClass  = regexp.MustCompile(`^@RequestMapping\("([^"]*)"\)`)
	reJavaMethod = regexp.MustCompile(`^@(Get|Post|Put|Delete|Request)Mapping\(\s*(?:value\s*=\s*)?"([^"]*)"(?:[^)]*method\s*=\s*RequestMethod\.(\w+))?`)
)

// excludedPrefixes 有意不迁移的 Java 端点（前缀匹配 + 理由）。
var excludedPrefixes = map[string]string{
	"/api/":               "MockController：前端本地 mock 配套接口（工作台示例组件）",
	"webSocketApi/":       "TestController：测试用 WS 推送样例",
	"/login":              "TestWebController：鉴权演示",
	"/article":            "TestWebController：鉴权演示",
	"/require_auth":       "TestWebController：鉴权演示",
	"/require_role":       "TestWebController：鉴权演示",
	"/require_permission": "TestWebController：鉴权演示",
	"/actuator/":          "Spring Actuator 专属（Go 无对应运行时）",
	"/thirdLogin/":        "第三方 OAuth 未配置凭据",
	"/cas/client/":        "CAS SSO 未部署",
}

// minJavaCoverage 最低覆盖率（低于即失败）。
const minJavaCoverage = 0.85

// TestJavaContractCoverage 断言 Go 路由对 Java 方法级端点的覆盖率。
func TestJavaContractCoverage(t *testing.T) {
	apiRoot := findAPIRoot(t)
	javaEps, err := parseJavaEndpoints(apiRoot)
	if err != nil {
		t.Fatalf("解析 Java 控制器失败: %v", err)
	}
	if len(javaEps) == 0 {
		t.Fatalf("未解析到 Java 端点")
	}

	registered := collectRegisteredRoutes(t)
	normSet := map[string]struct{}{}
	for k := range registered {
		normSet[normParams(k)] = struct{}{}
	}

	total, covered := 0, 0
	var missing []string
	for _, ep := range javaEps {
		key := ep.verb + " " + strings.TrimPrefix(ep.path, "/api")
		if isExcluded(ep.path) {
			continue
		}
		total++
		if _, ok := normSet[normParams(key)]; ok {
			covered++
			continue
		}
		missing = append(missing, ep.cls+"  "+ep.verb+" "+ep.path)
	}
	if total == 0 {
		t.Fatalf("排除后无可比对的 Java 端点")
	}
	cov := float64(covered) / float64(total)
	t.Logf("Java 端点 %d（排除 demo/test 后 %d），已覆盖 %d，覆盖率 %.1f%%",
		len(javaEps), total, covered, cov*100)

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Logf("未覆盖端点 %d 条:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}
	if cov < minJavaCoverage {
		t.Fatalf("契约覆盖率 %.1f%% 低于阈值 %.0f%%，缺失 %d 条:\n  %s",
			cov*100, minJavaCoverage*100, len(missing), strings.Join(missing, "\n  "))
	}
}

// javaEndpoint Java 方法级端点。
type javaEndpoint struct {
	cls  string
	verb string
	path string
}

// parseJavaEndpoints 解析全部 Controller.java 的 base + 方法注解。
func parseJavaEndpoints(apiRoot string) ([]javaEndpoint, error) {
	var out []javaEndpoint
	err := filepath.Walk(apiRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, "Controller.java") || strings.Contains(path, "/test/") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		cls := strings.TrimSuffix(filepath.Base(path), ".java")
		base := ""
		for _, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue // Java 侧注释掉的方法不计入契约
			}
			if m := reJavaClass.FindStringSubmatch(trimmed); m != nil {
				if base == "" {
					base = m[1]
				}
				continue
			}
			m := reJavaMethod.FindStringSubmatch(trimmed)
			if m == nil {
				continue
			}
			verb := strings.ToUpper(m[1])
			if verb == "REQUEST" {
				verb = strings.ToUpper(m[3])
				if verb == "" {
					verb = "GET"
				}
			}
			p := m[2]
			if !strings.HasPrefix(p, "/") {
				continue
			}
			full := strings.ReplaceAll(base+p, "//", "/")
			out = append(out, javaEndpoint{cls: cls, verb: verb, path: full})
		}
		return nil
	})
	return out, err
}

// normParams 归一化路径参数（{id}、:id、** 统一为 :P）。
func normParams(p string) string {
	reParam := regexp.MustCompile(`\{[^}]*\}`)
	reColon := regexp.MustCompile(`:[a-zA-Z0-9_]+`)
	reWild := regexp.MustCompile(`\*\*|\*[a-zA-Z0-9_]*|\*`)
	s := reParam.ReplaceAllString(p, ":P")
	s = reColon.ReplaceAllString(s, ":P")
	s = reWild.ReplaceAllString(s, ":P")
	return strings.TrimSuffix(s, "/")
}

// isExcluded 判断是否属于有意不迁移的端点。
func isExcluded(path string) bool {
	for prefix := range excludedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// findAPIRoot 定位源仓库 Java 代码根目录。
func findAPIRoot(t *testing.T) string {
	t.Helper()
	candidates := []string{"../../../api", "../../api", "../../../../api"}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if st, err := os.Stat(abs); err == nil && st.IsDir() {
				return abs
			}
		}
	}
	t.Skip("未找到 Java 源码目录 api/（独立检出场景跳过）")
	return ""
}

var _ = fmt.Sprintf
