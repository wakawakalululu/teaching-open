// Command e2e 端到端验证登录与改密链路（含密码迁移）。
//
// 流程：
//  1. 创建/重置一个测试账号，密码用 legacy DES 密文（模拟存量数据）；
//  2. 调用 /api/sys/login 用已知密码登录，验证 rehash-on-login 生效；
//  3. 确认库中密码已升级为 bcrypt；
//  4. 调用 /api/sys/user/updatePassword 改密，再用新密码登录；
//  5. 调用 /api/sys/user/changePassword 管理员重置，验证权限控制；
//  6. 登录响应校验 sysAllDictItems（Java 契约字段）；
//  7. 绑定角色后校验 getUserPermissionByToken 返回非空菜单树；
//  8. 清理测试账号。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
)

const (
	testUsername = "e2e_test_user"
	testPassword = "LegacyP@ss123"
	baseURL      = "http://127.0.0.1:8080"
)

func main() {
	link := os.Getenv("MYSQL_LINK")
	if link == "" {
		fmt.Println("请设置 MYSQL_LINK 环境变量")
		os.Exit(1)
	}

	ctx := context.Background()
	_ = ctx
	db, err := gdb.New(gdb.ConfigNode{Link: link, Type: "mysql"})
	if err != nil {
		fatal("创建数据库实例失败", err)
	}

	// 1. 准备测试账号（密码用 legacy DES 密文，模拟存量数据）
	const salt = "87654321"
	cipherText, err := passwd.NewLegacyPBE().Encrypt(testUsername, testPassword, salt)
	if err != nil {
		fatal("生成 legacy 密文失败", err)
	}
	cleanup := func() {
		_, _ = db.Model("sys_user").Where("username", testUsername).Delete()
	}
	cleanup()
	_, err = db.Model("sys_user").Data(g.Map{
		"id":       "e2e00000000000000000000000000001",
		"username": testUsername,
		"realname": "E2E测试",
		"password": cipherText,
		"salt":     salt,
		"status":   1,
		"del_flag": 0,
	}).Insert()
	if err != nil {
		fatal("创建测试账号失败", err)
	}
	fmt.Printf("✓ 步骤1 创建测试账号（legacy DES 密文, salt=%s）\n", salt)

	// 2. 用已知密码登录，触发 rehash-on-login
	token, err := login(testUsername, testPassword)
	if err != nil {
		fatal("步骤2 legacy 密码登录失败（rehash 未生效）", err)
	}
	fmt.Println("✓ 步骤2 用 legacy 密码登录成功（rehash-on-login 已触发）")

	// 3. 确认密码已升级为 bcrypt
	pwd, _ := db.Model("sys_user").Where("username", testUsername).Value("password")
	if passwd.Detect(pwd.String()) != passwd.SchemeBcrypt {
		fatal("步骤3 密码未升级为 bcrypt", fmt.Errorf("当前方案: %v, 值前 10 位: %.10s", passwd.Detect(pwd.String()), pwd.String()))
	}
	fmt.Println("✓ 步骤3 库中密码已升级为 bcrypt（legacy 迁移成功）")

	// 4. 自助改密（校验旧密码）
	const newPwd = "NewP@ss456"
	if err := changePassword(token, testUsername, testPassword, newPwd, newPwd); err != nil {
		fatal("步骤4 自助改密失败", err)
	}
	fmt.Println("✓ 步骤4 自助改密成功")

	// 5. 用新密码登录
	if _, err := login(testUsername, newPwd); err != nil {
		fatal("步骤5 新密码登录失败", err)
	}
	fmt.Println("✓ 步骤5 用新密码登录成功")

	// 6. 旧密码必须失败
	if _, err := login(testUsername, testPassword); err == nil {
		fatal("步骤6 旧密码仍可登录（改密未真正生效）", nil)
	}
	fmt.Println("✓ 步骤6 旧密码已失效")

	// 7. 越权改他人密码必须被拒（非 admin）
	otherToken, _ := login(testUsername, newPwd)
	if err := changePassword(otherToken, "admin", newPwd, "Hack@123", "Hack@123"); err == nil {
		fatal("步骤7 非管理员重置他人密码未被拒绝", nil)
	}
	fmt.Println("✓ 步骤7 非管理员越权重置他人密码被拒绝")

	// 8. 登录响应携带 sysAllDictItems（Java 登录契约，前端缓存全量字典）
	full, err := loginResult(testUsername, newPwd)
	if err != nil {
		fatal("步骤8 登录失败", err)
	}
	dictItems, _ := full["sysAllDictItems"].(map[string]interface{})
	if len(dictItems) == 0 {
		fatal("步骤8 登录响应缺少 sysAllDictItems", fmt.Errorf("got %T", full["sysAllDictItems"]))
	}
	fmt.Printf("✓ 步骤8 登录响应携带 sysAllDictItems（%d 个字典）\n", len(dictItems))

	// 9. 菜单权限接口：绑定角色后返回非空路由树与按钮权限
	bindTestUserRole(db)
	permResp, err := postJSON(baseURL+"/api/sys/permission/getUserPermissionByToken", "GET", nil, otherToken)
	if err != nil {
		fatal("步骤9 菜单权限接口请求失败", err)
	}
	if permResp["success"] != true {
		fatal("步骤9 菜单权限接口返回失败", fmt.Errorf("%v", permResp["message"]))
	}
	permResult, _ := permResp["result"].(map[string]interface{})
	menu, _ := permResult["menu"].([]interface{})
	auth, _ := permResult["auth"].([]interface{})
	allAuth, _ := permResult["allAuth"].([]interface{})
	if len(menu) == 0 {
		fatal("步骤9 菜单树为空（角色绑定未生效）", nil)
	}
	fmt.Printf("✓ 步骤9 getUserPermissionByToken：menu=%d auth=%d allAuth=%d\n",
		len(menu), len(auth), len(allAuth))

	// 10. 字典/分类域抽查（真实种子数据）
	dictResp, err := postJSON(baseURL+"/api/sys/dict/queryAllDictItems", "GET", nil, otherToken)
	if err != nil {
		fatal("步骤10 全量字典请求失败", err)
	}
	dictMap, _ := dictResp["result"].(map[string]interface{})
	if len(dictMap) == 0 {
		fatal("步骤10 全量字典为空", nil)
	}
	fmt.Printf("✓ 步骤10a queryAllDictItems：%d 个字典\n", len(dictMap))

	sexResp, err := postJSON(baseURL+"/api/sys/dict/getDictItems/sex", "GET", nil, otherToken)
	if err != nil {
		fatal("步骤10 sex 字典请求失败", err)
	}
	sexList, _ := sexResp["result"].([]interface{})
	if len(sexList) == 0 {
		fatal("步骤10 sex 字典项为空", nil)
	}
	fmt.Printf("✓ 步骤10b getDictItems/sex：%d 项\n", len(sexList))

	tableDictResp, err := postJSON(baseURL+"/api/sys/dict/getDictItems/sys_user,realname,id", "GET", nil, otherToken)
	if err != nil {
		fatal("步骤10 表字典请求失败", err)
	}
	tableDictList, _ := tableDictResp["result"].([]interface{})
	if len(tableDictList) == 0 {
		fatal("步骤10 表字典（sys_user,realname,id）为空", nil)
	}
	fmt.Printf("✓ 步骤10c 表字典 sys_user,realname,id：%d 项\n", len(tableDictList))

	catResp, err := postJSON(baseURL+"/api/sys/category/loadAllData?code=A03", "GET", nil, otherToken)
	if err != nil {
		fatal("步骤10 分类数据请求失败", err)
	}
	catList, _ := catResp["result"].([]interface{})
	if len(catList) == 0 {
		fatal("步骤10 分类数据（code=A03）为空", nil)
	}
	fmt.Printf("✓ 步骤10d category/loadAllData?code=A03：%d 项\n", len(catList))

	cleanup()
	cleanupUserRole(db)
	fmt.Println("✓ 清理测试账号")
	fmt.Println("\n=== E2E 全部通过 ===")
}

// bindTestUserRole 把测试账号绑定到库内第一个角色（使菜单权限非空）。
func bindTestUserRole(db gdb.DB) {
	_, _ = db.Model("sys_user_role").Where("id", "e2e0000000000000000000000000r001").Delete()
	role, err := db.Model("sys_role").Limit(1).One()
	if err != nil || role.IsEmpty() {
		fatal("绑定测试角色失败（sys_role 无数据）", err)
	}
	_, err = db.Model("sys_user_role").Data(g.Map{
		"id":      "e2e0000000000000000000000000r001",
		"user_id": "e2e00000000000000000000000000001",
		"role_id": role["id"].String(),
	}).Insert()
	if err != nil {
		fatal("绑定测试角色失败", err)
	}
}

// cleanupUserRole 清理测试账号的角色绑定。
func cleanupUserRole(db gdb.DB) {
	_, _ = db.Model("sys_user_role").Where("user_id", "e2e00000000000000000000000000001").Delete()
}

func login(username, password string) (string, error) {
	result, err := loginResult(username, password)
	if err != nil {
		return "", err
	}
	token, _ := result["token"].(string)
	if token == "" {
		return "", fmt.Errorf("响应无 token")
	}
	return token, nil
}

// loginResult 登录并返回完整 result（含 sysAllDictItems 等 Java 契约字段）。
func loginResult(username, password string) (map[string]interface{}, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := postJSON(baseURL+"/api/sys/login", "POST", body, "")
	if err != nil {
		return nil, err
	}
	if resp["success"] != true {
		return nil, fmt.Errorf("登录失败: %v", resp["message"])
	}
	result, _ := resp["result"].(map[string]interface{})
	if result == nil {
		return nil, fmt.Errorf("响应无 result")
	}
	return result, nil
}

func changePassword(token, username, oldPwd, newPwd, confirmPwd string) error {
	body, _ := json.Marshal(map[string]string{
		"username":        username,
		"oldpassword":     oldPwd,
		"password":        newPwd,
		"confirmpassword": confirmPwd,
	})
	resp, err := postJSON(baseURL+"/api/sys/user/updatePassword", "PUT", body, token)
	if err != nil {
		return err
	}
	if resp["success"] != true {
		return fmt.Errorf("改密失败: %v", resp["message"])
	}
	return nil
}

func postJSON(url, method string, body []byte, token string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Access-Token", token)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("响应非 JSON (status=%d): %.200s", res.StatusCode, string(raw))
	}
	return out, nil
}

func fatal(msg string, err error) {
	if err != nil {
		fmt.Printf("✗ %s: %v\n", msg, err)
	} else {
		fmt.Printf("✗ %s\n", msg)
	}
	os.Exit(1)
}
