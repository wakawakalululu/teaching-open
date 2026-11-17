// Command e2e 端到端验证登录与改密链路（含密码迁移）。
//
// 流程：
//  1. 创建/重置一个测试账号，密码用 legacy DES 密文（模拟存量数据）；
//  2. 调用 /api/sys/login 用已知密码登录，验证 rehash-on-login 生效；
//  3. 确认库中密码已升级为 bcrypt；
//  4. 调用 /api/sys/user/updatePassword 改密，再用新密码登录；
//  5. 调用 /api/sys/user/changePassword 管理员重置，验证权限控制；
//  6. 清理测试账号。
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

	cleanup()
	fmt.Println("✓ 清理测试账号")
	fmt.Println("\n=== E2E 全部通过 ===")
}

func login(username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := postJSON(baseURL+"/api/sys/login", "POST", body, "")
	if err != nil {
		return "", err
	}
	if resp["success"] != true {
		return "", fmt.Errorf("登录失败: %v", resp["message"])
	}
	result, _ := resp["result"].(map[string]interface{})
	token, _ := result["token"].(string)
	if token == "" {
		return "", fmt.Errorf("响应无 token")
	}
	return token, nil
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
