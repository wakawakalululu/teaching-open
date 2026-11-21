// excelx_test.go Excel 引擎零 DB 依赖单测：三行头渲染/解析回环、单元格名。
package excelx

import (
	"strings"
	"testing"
)

func TestRenderAndParseRoundtrip(t *testing.T) {
	headers := []string{"登录账号", "真实姓名", "手机号"}
	rows := [][]interface{}{
		{"stu1", "学生一", "13800000001"},
		{"stu2", "学生二", "13800000002"},
	}
	data, err := Render("学生导入", "admin", headers, rows)
	if err != nil {
		t.Fatalf("Render 失败: %v", err)
	}
	gotHeaders, gotRows, err := Parse(data, 2, 1)
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}
	if len(gotHeaders) != 3 || gotHeaders[0] != "登录账号" {
		t.Fatalf("表头解析异常: %v", gotHeaders)
	}
	if len(gotRows) != 2 {
		t.Fatalf("数据行数 = %d, 期望 2", len(gotRows))
	}
	if gotRows[0][0] != "stu1" || gotRows[0][2] != "13800000001" {
		t.Fatalf("数据行内容异常: %v", gotRows[0])
	}
}

func TestParseInsufficientRows(t *testing.T) {
	// 仅 2 行（不足 titleRows+headRows+数据）应报错
	minimal, err := Render("t", "", []string{"A"}, [][]interface{}{})
	if err != nil {
		t.Fatalf("Render 失败: %v", err)
	}
	_, _, err = Parse(minimal, 2, 1)
	if err == nil || !strings.Contains(err.Error(), "不足") {
		t.Fatalf("行不足应报错: %v", err)
	}
}

func TestCellName(t *testing.T) {
	if cellName(1, 1) != "A1" || cellName(3, 4) != "C4" {
		t.Fatal("单元格名异常")
	}
}
