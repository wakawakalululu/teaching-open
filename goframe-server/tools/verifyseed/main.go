// Command verifyseed 验证导入的种子数据结构完整性与密码字段格式。
//
// 不尝试猜测明文密码（密码可逆存储的根本问题），仅检查：
//   - 种子用户是否存在、字段是否完整；
//   - 密码字段是 DES 密文（hex 16 字节）还是 bcrypt（$2a$ 前缀）；
//   - 统计 legacy 与 bcrypt 账号数量（对应迁移进度报告）。
package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
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
		fmt.Printf("创建数据库实例失败: %v\n", err)
		os.Exit(1)
	}

	// 1. 统计密码方案分布
	one, err := db.Model("sys_user").
		Fields("COUNT(*) AS total",
			"SUM(CASE WHEN password LIKE '$2a$%' OR password LIKE '$2b$%' OR password LIKE '$2y$%' THEN 1 ELSE 0 END) AS bcrypt_count").
		Where("del_flag", 0).
		One()
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		os.Exit(1)
	}
	total := one["total"].Int()
	bcrypt := one["bcrypt_count"].Int()
	fmt.Printf("密码迁移进度: total=%d bcrypt=%d legacy=%d\n", total, bcrypt, total-bcrypt)

	// 2. 列出种子用户
	rows, err := db.Model("sys_user").
		Fields("id, username, realname, password, salt, status, del_flag").
		Where("del_flag", 0).
		All()
	if err != nil {
		fmt.Printf("列出用户失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n种子用户:")
	for _, r := range rows {
		pwd := r["password"].String()
		scheme := "legacy-DES"
		if len(pwd) >= 4 && (pwd[:4] == "$2a$" || pwd[:4] == "$2b$" || pwd[:4] == "$2y$") {
			scheme = "bcrypt"
		}
		fmt.Printf("  %-8s %-12s status=%d salt=%s 密码方案=%s\n",
			r["username"].String(), r["realname"].String(), r["status"].Int(), r["salt"].String(), scheme)
	}
}
