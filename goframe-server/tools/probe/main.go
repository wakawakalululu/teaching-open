// Command probe 探测数据库连接与表结构，用于确认环境可用性。
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
		link = os.Getenv("MYSQL_DSN")
	}
	if link == "" {
		fmt.Println("请设置 MYSQL_LINK 环境变量")
		os.Exit(1)
	}

	ctx := context.Background()

	// 创建独立数据库配置
	cfg := gdb.ConfigNode{
		Link: link,
		Type: "mysql",
	}

	db, err := gdb.New(cfg)
	if err != nil {
		fmt.Printf("创建数据库实例失败: %v\n", err)
		os.Exit(1)
	}

	// 测试连接
	if err := db.PingMaster(); err != nil {
		fmt.Printf("连接失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("连接成功")

	// 列出已有表
	tables, err := db.Tables(ctx)
	if err != nil {
		fmt.Printf("列出表失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("当前库共有 %d 张表\n", len(tables))
	for _, t := range tables {
		fmt.Println("  -", t)
	}
}
