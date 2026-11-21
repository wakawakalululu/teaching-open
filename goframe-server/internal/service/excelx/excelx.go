// Package excelx 统一 Excel 导出/导入能力
// （替代 Java AutoPoi：ExportParams(title, secondary, sheet) 三行头结构，
// ImportParams titleRows=2 + headRows=1——跳过两行标题读第三行表头）。
//
// Java exportXls 输出：ModelAndView 二进制流（Content-Disposition attachment）；
// Go 版输出 xlsx 字节流，同款三行头 + 数据行。
// Java importExcel：解析后逐行 save，返回「文件导入成功！数据行数:N」。
//
// 列映射来自 DomainSpec（Field=表列、Header=Java @Excel 中文名，
// 取自各实体注解/DDL 注释），中英双向映射供导出与导入共用。
package excelx

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/xuri/excelize/v2"
)

// Render 渲染三行头 xlsx（行1 标题、行2 导出人、行3 表头、行4+ 数据）。
func Render(title, exporter string, headers []string, rows [][]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.MergeCell(sheet, "A1", cellName(len(headers), 1))
	_ = f.SetCellValue(sheet, "A1", title)
	if exporter != "" {
		_ = f.MergeCell(sheet, "A2", cellName(len(headers), 2))
		_ = f.SetCellValue(sheet, "A2", "导出人:"+exporter)
	}
	for i, h := range headers {
		_ = f.SetCellValue(sheet, cellName(i+1, 3), h)
	}
	for r, row := range rows {
		for c, v := range row {
			_ = f.SetCellValue(sheet, cellName(c+1, r+4), v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("渲染 Excel 失败: %w", err)
	}
	return buf.Bytes(), nil
}

// Parse 解析导入文件（titleRows 行标题 + headRows 行表头 → 数据行），
// 返回表头与数据（均为字符串）。
func Parse(data []byte, titleRows, headRows int) ([]string, [][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("解析 Excel 失败: %w", err)
	}
	sheet := f.GetSheetName(0)
	headerRow := titleRows + headRows
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, nil, fmt.Errorf("读取 Excel 失败: %w", err)
	}
	if len(rows) < headerRow+1 {
		return nil, nil, fmt.Errorf("Excel 内容行不足")
	}
	headers := rows[headerRow-1]
	dataRows := make([][]string, 0)
	for _, row := range rows[headerRow:] {
		nonEmpty := false
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				nonEmpty = true
				break
			}
		}
		if nonEmpty {
			dataRows = append(dataRows, row)
		}
	}
	return headers, dataRows, nil
}

// cellName 列号(1 起)+行号 → 单元格名。
func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

// QueryRows 按域规格查询全量数据（无分页，对齐 Java exportXls 的 list 语义；
// del_flag=0 过滤仅对含该列的表生效由 spec 声明）。
func QueryRows(ctx context.Context, table string, cols []ExcelCol, filters map[string]interface{}, delFlagCol string) ([][]interface{}, error) {
	set := make(map[string]struct{}, len(cols))
	for _, c := range cols {
		set[c.Field] = struct{}{}
	}
	fields := make([]string, 0, len(cols))
	for _, c := range cols {
		fields = append(fields, c.Field)
	}
	m := g.DB().Model(table)
	if delFlagCol != "" {
		m = m.Where(delFlagCol, "0")
	}
	for c, v := range filters {
		if _, ok := set[c]; !ok || v == nil {
			continue
		}
		if str, isStr := v.(string); isStr && str == "" {
			continue
		}
		m = m.Where(c, v)
	}
	rows, err := m.Fields(strings.Join(fields, ",")).All()
	if err != nil {
		return nil, fmt.Errorf("查询导出数据失败: %w", err)
	}
	out := make([][]interface{}, 0, len(rows))
	for _, r := range rows {
		row := make([]interface{}, 0, len(cols))
		for _, c := range cols {
			row = append(row, r[c.Field].Interface())
		}
		out = append(out, row)
	}
	return out, nil
}

// InsertRow 按表头→列映射插入一行（白名单列、id 服务端生成）。
func InsertRow(ctx context.Context, table string, cols []ExcelCol, headers, values []string) error {
	headerToField := make(map[string]string, len(cols))
	for _, c := range cols {
		headerToField[strings.TrimSpace(c.Header)] = c.Field
	}
	data := map[string]interface{}{"id": guid.S()}
	for i, v := range values {
		if i >= len(headers) {
			break
		}
		field, ok := headerToField[strings.TrimSpace(headers[i])]
		if !ok || v == "" {
			continue
		}
		data[field] = v
	}
	if _, err := g.DB().Model(table).Data(data).Insert(); err != nil {
		return fmt.Errorf("插入行失败: %w", err)
	}
	return nil
}
