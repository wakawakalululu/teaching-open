// Package teachingnews 承载资讯域业务用例
// （契约蓝本：TeachingNewsController 的 8 个核心端点；exportXls/importExcel
// 有意未迁移；service 无自定义逻辑，纯 MP 语义）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - newsList（anon）：固定 news_status>=1 + QueryGenerator 过滤 + 每条
//     newsContent 置 null（列表不回大字段）；DictAspect 注入 newsStatus_dictText；
//   - newsDetail（anon）：全量含 content；newsStatus<=0（含默认零值）→
//     「未找到对应数据」；单实体不注入；
//   - list（jwt）：无状态过滤 + dict 注入；add/edit/delete/deleteBatch/
//     queryById 纯 MP CRUD。
package teachingnews

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// ErrNotFound 资讯不存在（Java newsDetail/queryById error 原话术）。
var ErrNotFound = errors.New("未找到对应数据")

// NewsInput add/edit 请求体（TeachingNews 可编辑字段）。
type NewsInput struct {
	Id          string
	NewsTitle   *string
	NewsContent *string
	NewsStatus  *int
}

// newsRecord IPage 白名单输出行（实体全字段 + newsStatus_dictText 注入键）。
type newsRecord struct {
	entity.TeachingNews
	NewsStatusDictText string `json:"newsStatus_dictText"`
}

// Service 资讯业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// NewsList 已发布资讯分页（anon）。
//
// 契约蓝本 TeachingNewsController#getNewsList：固定 news_status>=1；
// 每条 newsContent 置 null（列表不回大字段）；dict 注入 newsStatus_dictText。
func (s *Service) NewsList(ctx context.Context, param page.Param, newsTitle string) (*page.Result, error) {
	cols := dao.TeachingNews.Columns()
	m := dao.TeachingNews.Ctx(ctx).
		Where(cols.NewsStatus+" >= ?", 1)
	if newsTitle != "" {
		m = applyQueryRule(m, cols.NewsTitle, newsTitle)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计资讯数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]newsRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingNews
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询资讯列表失败: %w", err)
	}
	statusTexts, err := s.dictTextMap(ctx, "news_status")
	if err != nil {
		return nil, err
	}
	records := make([]newsRecord, 0, len(rows))
	for _, r := range rows {
		rec := newsRecord{TeachingNews: r, NewsStatusDictText: statusTexts[intKey(r.NewsStatus)]}
		// 列表不回大字段（对齐 Java setNewsContent(null)）
		rec.NewsContent = ""
		records = append(records, rec)
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// NewsDetail 资讯详情（anon，全量含 content，不注入 dict）。
// 契约蓝本 TeachingNewsController#newsDetail：newsStatus<=0 →「未找到对应数据」。
func (s *Service) NewsDetail(ctx context.Context, id string) (*entity.TeachingNews, error) {
	var row entity.TeachingNews
	if err := dao.TeachingNews.Ctx(ctx).
		Where(dao.TeachingNews.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询资讯失败: %w", err)
	}
	if row.Id == "" || row.NewsStatus <= 0 {
		return nil, ErrNotFound
	}
	return &row, nil
}

// List 资讯管理分页（无状态过滤，dict 注入）。
func (s *Service) List(ctx context.Context, param page.Param, newsTitle string) (*page.Result, error) {
	cols := dao.TeachingNews.Columns()
	m := dao.TeachingNews.Ctx(ctx)
	if newsTitle != "" {
		m = applyQueryRule(m, cols.NewsTitle, newsTitle)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计资讯数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]newsRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingNews
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询资讯列表失败: %w", err)
	}
	statusTexts, err := s.dictTextMap(ctx, "news_status")
	if err != nil {
		return nil, err
	}
	records := make([]newsRecord, 0, len(rows))
	for _, r := range rows {
		records = append(records, newsRecord{
			TeachingNews:       r,
			NewsStatusDictText: statusTexts[intKey(r.NewsStatus)],
		})
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// Add 新增资讯（纯 save，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in NewsInput) error {
	data := do.TeachingNews{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyInput(&data, in)
	if _, err := dao.TeachingNews.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑资讯（纯 updateById）。
func (s *Service) Edit(ctx context.Context, in NewsInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.TeachingNews{}
	applyInput(&data, in)
	cols := dao.TeachingNews.Columns()
	if _, err := dao.TeachingNews.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除资讯。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.TeachingNews.Columns()
	if _, err := dao.TeachingNews.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除资讯失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除资讯。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingNews.Columns()
	if _, err := dao.TeachingNews.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除资讯失败: %w", err)
	}
	return nil
}

// QueryByID 资讯详情（管理端）。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.TeachingNews, error) {
	var row entity.TeachingNews
	if err := dao.TeachingNews.Ctx(ctx).
		Where(dao.TeachingNews.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询资讯失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// applyInput 请求体字段拷贝进 DO。
func applyInput(data *do.TeachingNews, in NewsInput) {
	if in.NewsTitle != nil {
		data.NewsTitle = *in.NewsTitle
	}
	if in.NewsContent != nil {
		data.NewsContent = *in.NewsContent
	}
	if in.NewsStatus != nil {
		data.NewsStatus = *in.NewsStatus
	}
}

// intKey int → 字典键字符串。
func intKey(v int) string {
	return fmt.Sprintf("%d", v)
}

// dictTextMap 读取普通字典的 键值→文本 映射（与既有域同款实现）。
func (s *Service) dictTextMap(ctx context.Context, code string) (map[string]string, error) {
	texts := make(map[string]string)
	dCols := dao.SysDict.Columns()
	var d entity.SysDict
	if err := dao.SysDict.Ctx(ctx).Where(dCols.DictCode, code).Scan(&d); err != nil {
		return nil, fmt.Errorf("查询字典失败: %w", err)
	}
	if d.Id == "" {
		return texts, nil
	}
	itemCols := dao.SysDictItem.Columns()
	var items []entity.SysDictItem
	if err := dao.SysDictItem.Ctx(ctx).Where(itemCols.DictId, d.Id).Scan(&items); err != nil {
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}
	for _, item := range items {
		texts[item.ItemValue] = item.ItemText
	}
	return texts, nil
}

// applyQueryRule 按 QueryGenerator 值规则应用过滤条件（与既有域同款）。
func applyQueryRule(m *gdb.Model, column, raw string) *gdb.Model {
	kind, val, vals := parseQueryRule(raw)
	switch kind {
	case "ne":
		return m.WhereNot(column, val)
	case "like":
		return m.WhereLike(column, val)
	case "in":
		return m.WhereIn(column, vals)
	default:
		return m.Where(column, val)
	}
}

// parseQueryRule QueryGenerator 值规则纯判定（与既有域同款）。
func parseQueryRule(raw string) (kind, val string, vals []string) {
	switch {
	case strings.HasPrefix(raw, "!="):
		return "ne", strings.TrimPrefix(raw, "!="), nil
	case strings.Contains(raw, "*"):
		return "like", strings.ReplaceAll(raw, "*", "%"), nil
	case strings.Contains(raw, ","):
		return "in", "", strings.Split(raw, ",")
	default:
		return "eq", raw, nil
	}
}
