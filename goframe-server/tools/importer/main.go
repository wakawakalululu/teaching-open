// Command importer 将 teaching-open 的 SQL 文件导入到目标 MySQL。
//
// 处理策略：
//   - 跳过 LOCK/UNLOCK TABLES、SET @saved、/*!...*/ 条件注释（这些是 mysqldump 的会话控制，
//     在单连接下无需、且可能因多语句混用出错）；
//   - 按分号拆分，逐条执行 CREATE TABLE / DROP TABLE / INSERT 语句；
//   - 出错即停止并打印上下文，便于定位。
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
)

func main() {
	link := os.Getenv("MYSQL_LINK")
	if link == "" {
		link = os.Getenv("MYSQL_DSN")
	}
	sqlPath := os.Getenv("SQL_FILE")
	if sqlPath == "" {
		sqlPath = "../api/db/teachingopen2.6.sql"
	}
	if link == "" {
		fmt.Println("请设置 MYSQL_LINK 环境变量")
		os.Exit(1)
	}

	// 读取 SQL 文件
	data, err := os.ReadFile(sqlPath)
	if err != nil {
		fmt.Printf("读取 SQL 文件失败: %v\n", err)
		os.Exit(1)
	}
	text := string(data)

	ctx := context.Background()
	cfg := gdb.ConfigNode{Link: link, Type: "mysql"}
	db, err := gdb.New(cfg)
	if err != nil {
		fmt.Printf("创建数据库实例失败: %v\n", err)
		os.Exit(1)
	}
	if err := db.PingMaster(); err != nil {
		fmt.Printf("连接失败: %v\n", err)
		os.Exit(1)
	}
	// 关闭外键检查，避免 Quartz 表之间的外键依赖顺序问题
	if _, err := db.Exec(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		fmt.Printf("关闭外键检查失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Exec(ctx, "SET FOREIGN_KEY_CHECKS=1")
	fmt.Println("连接成功，开始导入...")

	// 按行处理，跳过会话控制指令，其余按分号拆分执行
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024) // 支持超长行

	var (
		stmt      strings.Builder
		execCount int
		skipLines int
	)

	flush := func() error {
		s := strings.TrimSpace(stmt.String())
		stmt.Reset()
		if s == "" {
			return nil
		}
		if _, err := db.Exec(ctx, s); err != nil {
			return fmt.Errorf("执行失败: %v\n语句: %.200s...", err, s)
		}
		execCount++
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 跳过纯注释与空行
		if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
			// 但 /*!...*/ 是 mysqldump 的条件执行，跳过
			skipLines++
			continue
		}
		// 跳过会话控制语句
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "LOCK TABLES") ||
			strings.HasPrefix(upper, "UNLOCK TABLES") ||
			strings.HasPrefix(upper, "SET @") ||
			strings.HasPrefix(upper, "SET SQL_") ||
			strings.HasPrefix(upper, "SET NAMES") ||
			strings.HasPrefix(upper, "SET CHARACTER_SET") ||
			strings.HasPrefix(upper, "SET FOREIGN_KEY_CHECKS") ||
			strings.HasPrefix(upper, "SET UNIQUE_CHECKS") {
			skipLines++
			continue
		}

		stmt.WriteString(line)
		stmt.WriteString("\n")

		// 遇到分号结尾即执行（简单判断：行尾含分号）
		if strings.HasSuffix(trimmed, ";") {
			if err := flush(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	// 兜底：剩余语句
	if err := flush(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("导入完成：执行 %d 条语句，跳过 %d 行会话控制\n", execCount, skipLines)
}
