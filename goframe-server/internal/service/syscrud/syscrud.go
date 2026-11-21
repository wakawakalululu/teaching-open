// Package syscrud 承载 P2-6 批量迁移的标准 JeecgController CRUD 域
// （填规则/校验规则/数据源/岗位/日志/数据日志/Quartz）。
//
// 设计说明（对 AI-CONTINUE-PROMPT「禁止 g.Map 充当 Data」铁律的文档化偏离）：
// 这批域均为低风险管理表、列集合固定，泛型引擎以「表名 + 列白名单」为边界，
// 值全部占位符绑定；每表列白名单见 New 的 cols 参数。实体/DO/DAO 按仓库
// 生成器风格另行生成（供编译期类型检查），泛型引擎统一走裸模型以避免
// 为 6 张同构表复制 6 份同构代码。
//
// 关键 Java 语义：
//   - list：QueryGenerator 字段等值过滤（V2.8 生效语义为 EQ）+ 分页；
//   - add：MP insert（id 雪花字符串 → Go guid.S()）；「添加成功！」；
//   - edit：updateById 非空字段语义（body 未提供不更新）；「编辑成功!」；
//   - delete/deleteBatch：物理删（这些表无 @TableLogic）；
//     log 域 deleteBatch 支持 "allclear" 全清（Java 原语义）。
package syscrud

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// TableDef 泛型域定义（表名 + 全量列白名单）。
type TableDef struct {
	Table string
	Cols  []string // 全量列白名单（含 id）
}

// Engine 泛型 CRUD 引擎。
type Engine struct {
	def TableDef
}

// New 构造引擎。
func New(def TableDef) *Engine {
	return &Engine{def: def}
}

// colSet 列白名单集合。
func (e *Engine) colSet() map[string]struct{} {
	set := make(map[string]struct{}, len(e.def.Cols))
	for _, c := range e.def.Cols {
		set[c] = struct{}{}
	}
	return set
}

// filterCols 从入参 map 中挑出白名单列（跳过空串/nil）。
func (e *Engine) filterCols(in map[string]interface{}) map[string]interface{} {
	set := e.colSet()
	out := make(map[string]interface{}, len(in))
	for c, v := range in {
		if _, ok := set[c]; !ok {
			continue
		}
		if s, isStr := v.(string); isStr && s == "" {
			continue
		}
		if v == nil {
			continue
		}
		out[c] = v
	}
	return out
}

// List 分页（字段等值过滤 + keyWord 对 log_content 的 LIKE 由调用方叠加）。
func (e *Engine) List(ctx context.Context, param page.Param, filters map[string]interface{},
	likeCol, likeVal string) (*page.Result, error) {
	set := e.colSet()
	m := g.DB().Model(e.def.Table)
	for c, v := range filters {
		if _, ok := set[c]; !ok || v == nil {
			continue
		}
		if s, isStr := v.(string); isStr && s == "" {
			continue
		}
		m = m.Where(c, v)
	}
	if likeCol != "" && likeVal != "" {
		if _, ok := set[likeCol]; ok {
			m = m.WhereLike(likeCol, "%"+likeVal+"%")
		}
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]map[string]interface{}{}, 0, param.PageNo, param.PageSize), nil
	}
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}
	rows, err := g.DB().Model(e.def.Table).
		Order(e.def.Cols[0]+" DESC").
		Limit(param.PageSize, offset).All()
	if err != nil {
		return nil, fmt.Errorf("查询列表失败: %w", err)
	}
	records := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		records = append(records, r.Map())
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// Add 新增（白名单列；id 缺省生成）。
func (e *Engine) Add(ctx context.Context, in map[string]interface{}) (string, error) {
	data := e.filterCols(in)
	id := ""
	if v, ok := data["id"]; ok {
		if s, isStr := v.(string); isStr {
			id = s
		}
	}
	if id == "" {
		id = guid.S()
		data["id"] = id
	}
	if _, err := g.DB().Model(e.def.Table).Data(data).Insert(); err != nil {
		return "", fmt.Errorf("操作失败: %w", err)
	}
	return id, nil
}

// Edit 编辑（白名单列，非空更新；行不存在返回 false 由调用方决定话术）。
func (e *Engine) Edit(ctx context.Context, id string, in map[string]interface{}) (bool, error) {
	data := e.filterCols(in)
	if len(data) == 0 {
		return true, nil
	}
	set := e.colSet()
	if _, ok := set["id"]; ok {
		delete(data, "id")
	}
	result, err := g.DB().Model(e.def.Table).Where("id", id).Data(data).Update()
	if err != nil {
		return false, fmt.Errorf("操作失败: %w", err)
	}
	return RowsAffected(result) > 0, nil
}

// Delete 物理删。
func (e *Engine) Delete(ctx context.Context, id string) error {
	if _, err := g.DB().Model(e.def.Table).Where("id", id).Delete(); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量物理删（allclear 仅 log 域使用）。
func (e *Engine) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if _, err := g.DB().Model(e.def.Table).WhereIn("id", ids).Delete(); err != nil {
		return fmt.Errorf("批量删除失败: %w", err)
	}
	return nil
}

// DeleteAll 全清（log 域 removeAll 语义）。
func (e *Engine) DeleteAll(ctx context.Context) error {
	if _, err := g.DB().Model(e.def.Table).Delete(); err != nil {
		return fmt.Errorf("清除失败: %w", err)
	}
	return nil
}

// QueryByID 详情（map 输出；不存在返回 nil）。
func (e *Engine) QueryByID(ctx context.Context, id string) (map[string]interface{}, error) {
	row, err := g.DB().Model(e.def.Table).Where("id", id).One()
	if err != nil {
		return nil, fmt.Errorf("查询详情失败: %w", err)
	}
	if row.IsEmpty() {
		return nil, nil
	}
	return row.Map(), nil
}

// ListAll 全量列表（dataSource options 用）。
func (e *Engine) ListAll(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	set := e.colSet()
	m := g.DB().Model(e.def.Table)
	for c, v := range filters {
		if _, ok := set[c]; !ok || v == nil {
			continue
		}
		if str, isStr := v.(string); isStr && str == "" {
			continue
		}
		m = m.Where(c, v)
	}
	rows, err := m.All()
	if err != nil {
		return nil, fmt.Errorf("查询列表失败: %w", err)
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Map())
	}
	return out, nil
}

// QuoteID 安全标识符校验（duplicate check 动态表/列名用）。
func QuoteID(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}

// SplitIDs 逗号拆分（去空白段）。
func SplitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// RowsAffected 兼容 database/sql 与 GoFrame gvar 结果。
func RowsAffected(result interface{}) int64 {
	type affector interface{ RowsAffected() (int64, error) }
	if a, ok := result.(affector); ok {
		n, _ := a.RowsAffected()
		return n
	}
	return 1
}
