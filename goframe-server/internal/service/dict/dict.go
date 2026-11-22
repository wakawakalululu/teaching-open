// Package dict 承载字典与字典项业务用例（契约蓝本：SysDictController/SysDictItemController）。
//
// 使用 GoFrame gdb ORM（通过 dao.SysDict/dao.SysDictItem），全部占位符绑定；
// 表字典查询的表名/列名为动态标识符，统一经过 identRegexp 严格白名单校验后
// 才允许拼接，从根源上杜绝 Java 版 ${} 插值注入面。
package dict

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误（HTTP 层映射为 Result{success:false}）。
var (
	// ErrDictCodeInvalid 字典 Code 格式不正确（Java 话术保持一致）。
	ErrDictCodeInvalid = errors.New("字典Code格式不正确！")
	// ErrEntityNotFound 未找到对应实体。
	ErrEntityNotFound = errors.New("未找到对应实体")
	// ErrTableDictFilterUnsupported 带过滤条件的表字典暂未迁移。
	ErrTableDictFilterUnsupported = errors.New("暂不支持带过滤条件的表字典")
	// ErrIDsRequired 批量操作缺少 ids 参数。
	ErrIDsRequired = errors.New("参数不识别！")
)

// identRegexp 限定动态 SQL 标识符：仅字母/数字/下划线，杜绝注入与表达式。
// Java 版走黑名单过滤（SqlInjectionUtil），此处升级为白名单，更严。
var identRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ValidIdent 暴露标识符校验，供控制器层提前拦截与单测使用。
func ValidIdent(s string) bool {
	return identRegexp.MatchString(s)
}

// DictModel 字典项模型（对应 Java org.jeecg.common.system.vo.DictModel）。
type DictModel struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// DictTree 树形字典节点（对应 Java SysDictTree）。
type DictTree struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Id          string `json:"id,omitempty"`
	Type        int    `json:"type,omitempty"`
	DictName    string `json:"dictName,omitempty"`
	DictCode    string `json:"dictCode,omitempty"`
	Description string `json:"description,omitempty"`
	DelFlag     int    `json:"delFlag,omitempty"`
	CreateTime  string `json:"createTime,omitempty"`
	UpdateTime  string `json:"updateTime,omitempty"`
}

// PageParam 分页参数。
type PageParam = page.Param

// PageResult 分页返回（对应 Java IPage 的前端消费字段）。
type PageResult = page.Result

// Service 字典业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// QueryAllDictItems 一次性返回全部字典（登录后前端初始化强依赖）。
//
// 语义对齐 SysDictServiceImpl#queryAllDictItems：
//   - 字典主表：del_flag = 0（Java @TableLogic 自动过滤）；
//   - 字典项：status = 1，按 sort_order 升序；
//   - 未启用/空字典也输出空数组键，保证前端取值安全。
func (s *Service) QueryAllDictItems(ctx context.Context) (map[string][]DictModel, error) {
	var (
		dicts []entity.SysDict
		items []entity.SysDictItem
	)
	if err := dao.SysDict.Ctx(ctx).Where(dao.SysDict.Columns().DelFlag, 0).Scan(&dicts); err != nil {
		return nil, fmt.Errorf("查询字典主表失败: %w", err)
	}
	if err := dao.SysDictItem.Ctx(ctx).
		Where(dao.SysDictItem.Columns().Status, 1).
		Order(dao.SysDictItem.Columns().SortOrder).
		Scan(&items); err != nil {
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}

	itemsByDict := make(map[string][]DictModel, len(dicts))
	for _, item := range items {
		itemsByDict[item.DictId] = append(itemsByDict[item.DictId], DictModel{
			Text:  item.ItemText,
			Value: item.ItemValue,
		})
	}
	res := make(map[string][]DictModel, len(dicts))
	for _, d := range dicts {
		res[d.DictCode] = itemsByDict[d.Id]
	}
	return res, nil
}

// GetDictItems 按字典 Code 获取字典项。
//
// dictCode 两种形态（与 Java 契约一致）：
//  1. 普通字典 Code：查 sys_dict_item；
//  2. 表字典 "表名,文本列,值列[,过滤条件]"：动态查目标表。
//     第 4 段过滤条件为 Java 版遗留的原始 SQL 片段，Go 版明确不支持并返回错误。
func (s *Service) GetDictItems(ctx context.Context, dictCode string) ([]DictModel, error) {
	if strings.Contains(dictCode, ",") {
		params := strings.Split(dictCode, ",")
		switch {
		case len(params) < 3:
			return nil, ErrDictCodeInvalid
		case len(params) == 4:
			return nil, ErrTableDictFilterUnsupported
		case len(params) > 4:
			return nil, ErrDictCodeInvalid
		}
		return s.queryTableDictItemsByCode(ctx, params[0], params[1], params[2])
	}
	return s.queryDictItemsByCode(ctx, dictCode)
}

// queryDictItemsByCode 普通字典：按 dict_code 查其下字典项，sort_order 升序。
// Java 版此处不过滤 status，保持一致。
func (s *Service) queryDictItemsByCode(ctx context.Context, code string) ([]DictModel, error) {
	dictCols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).
		Where(dictCols.DictCode, code).
		Where(dictCols.DelFlag, 0).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询字典失败: %w", err)
	}
	if one.IsEmpty() {
		return []DictModel{}, nil
	}
	var items []entity.SysDictItem
	if err := dao.SysDictItem.Ctx(ctx).
		Where(dao.SysDictItem.Columns().DictId, one["id"].String()).
		Order(dao.SysDictItem.Columns().SortOrder).
		Scan(&items); err != nil {
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}
	list := make([]DictModel, 0, len(items))
	for _, item := range items {
		list = append(list, DictModel{Text: item.ItemText, Value: item.ItemValue})
	}
	return list, nil
}

// queryTableDictItemsByCode 表字典：SELECT {text} AS text, {code} AS value FROM {table}。
// 标识符白名单校验后拼接（只可能是 [a-zA-Z0-9_] 组成），值全部走占位符。
func (s *Service) queryTableDictItemsByCode(ctx context.Context, table, text, code string) ([]DictModel, error) {
	for _, ident := range []string{table, text, code} {
		if !ValidIdent(ident) {
			return nil, ErrDictCodeInvalid
		}
	}
	records, err := dao.SysDict.DB().GetAll(ctx,
		fmt.Sprintf("SELECT %s AS `text`, %s AS `value` FROM %s", text, code, table))
	if err != nil {
		return nil, fmt.Errorf("查询表字典失败: %w", err)
	}
	list := make([]DictModel, 0, len(records))
	for _, rec := range records {
		list = append(list, DictModel{
			Text:  rec["text"].String(),
			Value: rec["value"].String(),
		})
	}
	return list, nil
}

// GetDictText 按字典 Code 与键值取字典文本。
func (s *Service) GetDictText(ctx context.Context, dictCode, key string) (string, error) {
	dictCols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).
		Where(dictCols.DictCode, dictCode).
		Where(dictCols.DelFlag, 0).
		One()
	if err != nil {
		return "", fmt.Errorf("查询字典失败: %w", err)
	}
	if one.IsEmpty() {
		return "", nil
	}
	itemCols := dao.SysDictItem.Columns()
	rec, err := dao.SysDictItem.Ctx(ctx).
		Fields(itemCols.ItemText).
		Where(itemCols.DictId, one["id"].String()).
		Where(itemCols.ItemValue, key).
		One()
	if err != nil {
		return "", fmt.Errorf("查询字典项失败: %w", err)
	}
	return rec[itemCols.ItemText].String(), nil
}

// LoadDict 表字典异步搜索：文本列模糊匹配关键字。
func (s *Service) LoadDict(ctx context.Context, dictCode, keyword string) ([]DictModel, error) {
	params := strings.Split(dictCode, ",")
	if len(params) != 3 {
		return nil, ErrDictCodeInvalid
	}
	table, text, code := params[0], params[1], params[2]
	for _, ident := range []string{table, text, code} {
		if !ValidIdent(ident) {
			return nil, ErrDictCodeInvalid
		}
	}
	records, err := dao.SysDict.DB().GetAll(ctx,
		fmt.Sprintf("SELECT %s AS `text`, %s AS `value` FROM %s WHERE %s LIKE ?", text, code, table, text),
		"%"+keyword+"%")
	if err != nil {
		return nil, fmt.Errorf("查询表字典失败: %w", err)
	}
	list := make([]DictModel, 0, len(records))
	for _, rec := range records {
		list = append(list, DictModel{
			Text:  rec["text"].String(),
			Value: rec["value"].String(),
		})
	}
	return list, nil
}

// LoadDictItem 表字典回显：按值列的多个键取文本列，输出顺序与入参键一致。
func (s *Service) LoadDictItem(ctx context.Context, dictCode string, keys []string) ([]string, error) {
	params := strings.Split(dictCode, ",")
	if len(params) != 3 {
		return nil, ErrDictCodeInvalid
	}
	table, text, code := params[0], params[1], params[2]
	for _, ident := range []string{table, text, code} {
		if !ValidIdent(ident) {
			return nil, ErrDictCodeInvalid
		}
	}
	records, err := dao.SysDict.DB().GetAll(ctx,
		fmt.Sprintf("SELECT %s AS `text`, %s AS `value` FROM %s WHERE %s IN (?)", text, code, table, code),
		keys)
	if err != nil {
		return nil, fmt.Errorf("查询表字典失败: %w", err)
	}
	textByKey := make(map[string]string, len(records))
	for _, rec := range records {
		textByKey[rec["value"].String()] = rec["text"].String()
	}
	texts := make([]string, 0, len(keys))
	for _, key := range keys {
		texts = append(texts, textByKey[key])
	}
	return texts, nil
}

// ListPage 字典分页列表（del_flag = 0）。
//
// 过滤语义对齐 QueryGenerator 字符串规则：值含 `*` 时按 `*`→`%` 走 LIKE，否则精确匹配。
func (s *Service) ListPage(ctx context.Context, param PageParam, dictName, dictCode string) (*PageResult, error) {
	cols := dao.SysDict.Columns()
	m := dao.SysDict.Ctx(ctx).Where(cols.DelFlag, 0)
	m = applyStarFilter(m, cols.DictName, dictName)
	m = applyStarFilter(m, cols.DictCode, dictCode)

	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计字典数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysDict{}, 0, param.PageNo, param.PageSize), nil
	}
	var list []entity.SysDict
	if err := m.Page(param.PageNo, param.PageSize).Scan(&list); err != nil {
		return nil, fmt.Errorf("查询字典列表失败: %w", err)
	}
	return page.New(list, total, param.PageNo, param.PageSize), nil
}

// TreeList 树形字典列表（按创建时间倒序）。
func (s *Service) TreeList(ctx context.Context, dictName string) ([]DictTree, error) {
	cols := dao.SysDict.Columns()
	m := dao.SysDict.Ctx(ctx).Where(cols.DelFlag, 0)
	m = applyStarFilter(m, cols.DictName, dictName)
	m = m.Order(cols.CreateTime + " DESC")

	var list []entity.SysDict
	if err := m.Scan(&list); err != nil {
		return nil, fmt.Errorf("查询字典树失败: %w", err)
	}
	tree := make([]DictTree, 0, len(list))
	for _, d := range list {
		tree = append(tree, DictTree{
			Key:         d.Id,
			Title:       d.DictName,
			Id:          d.Id,
			Type:        d.Type,
			DictName:    d.DictName,
			DictCode:    d.DictCode,
			Description: d.Description,
			DelFlag:     d.DelFlag,
			CreateTime:  d.CreateTime,
			UpdateTime:  d.UpdateTime,
		})
	}
	return tree, nil
}

// GetByCode 按字典编码取单条。
func (s *Service) GetByCode(ctx context.Context, dictCode string) (*entity.SysDict, error) {
	cols := dao.SysDict.Columns()
	var one entity.SysDict
	err := dao.SysDict.Ctx(ctx).
		Where(cols.DictCode, dictCode).
		Where(cols.DelFlag, 0).
		Scan(&one)
	if err != nil {
		return nil, fmt.Errorf("查询字典失败: %w", err)
	}
	if one.Id == "" {
		return nil, nil
	}
	return &one, nil
}

// Add 新增字典（del_flag=0，时间戳由服务端写入）。
func (s *Service) Add(ctx context.Context, in *do.SysDict) error {
	if in.DictCode == nil || strings.TrimSpace(gconv.String(in.DictCode)) == "" {
		return ErrDictCodeInvalid
	}
	data := *in
	data.Id = guid.S()
	data.DelFlag = 0
	data.CreateTime = gtime.Now()
	_, err := dao.SysDict.Ctx(ctx).Data(data).Insert()
	if err != nil {
		return fmt.Errorf("保存字典失败: %w", err)
	}
	return nil
}

// Edit 编辑字典：不存在即报错；更新仅作用于请求中显式给出的字段（DO nil 字段自动忽略）。
func (s *Service) Edit(ctx context.Context, id string, in *do.SysDict) error {
	cols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).Where(cols.DelFlag, 0).One()
	if err != nil {
		return fmt.Errorf("查询字典失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	data := *in
	data.Id = id
	data.UpdateTime = gtime.Now()
	if _, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).Data(data).Update(); err != nil {
		return fmt.Errorf("更新字典失败: %w", err)
	}
	return nil
}

// Delete 软删除字典（del_flag=1，语义同 Java @TableLogic removeById）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).Where(cols.DelFlag, 0).One()
	if err != nil {
		return fmt.Errorf("查询字典失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	if _, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).Data(do.SysDict{DelFlag: 1}).Update(); err != nil {
		return fmt.Errorf("删除字典失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量软删除。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return ErrIDsRequired
	}
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteList 回收站列表（del_flag=1）。
func (s *Service) DeleteList(ctx context.Context) ([]entity.SysDict, error) {
	var list []entity.SysDict
	if err := dao.SysDict.Ctx(ctx).
		Where(dao.SysDict.Columns().DelFlag, 1).
		Scan(&list); err != nil {
		return nil, fmt.Errorf("查询回收站失败: %w", err)
	}
	return list, nil
}

// DeletePhysic 物理删除字典并级联清理其全部字典项（事务，语义同 deleteOneById）。
func (s *Service) DeletePhysic(ctx context.Context, id string) error {
	cols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).One()
	if err != nil {
		return fmt.Errorf("查询字典失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	if err := dao.SysDict.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return err
		}
		if _, err := dao.SysDictItem.Ctx(ctx).
			Where(dao.SysDictItem.Columns().DictId, id).
			Delete(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("物理删除字典失败: %w", err)
	}
	return nil
}

// Back 从回收站恢复（del_flag 1 → 0）。
func (s *Service) Back(ctx context.Context, id string) error {
	cols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).One()
	if err != nil {
		return fmt.Errorf("查询字典失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	if _, err := dao.SysDict.Ctx(ctx).Where(cols.Id, id).Data(do.SysDict{DelFlag: 0}).Update(); err != nil {
		return fmt.Errorf("恢复字典失败: %w", err)
	}
	return nil
}

// ItemList 字典项分页列表（按 sort_order 升序，与 Java 契约一致）。
func (s *Service) ItemList(ctx context.Context, param PageParam, dictId, itemText string) (*PageResult, error) {
	cols := dao.SysDictItem.Columns()
	m := dao.SysDictItem.Ctx(ctx)
	m = applyStarFilter(m, cols.DictId, dictId)
	m = applyStarFilter(m, cols.ItemText, itemText)

	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计字典项数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysDictItem{}, 0, param.PageNo, param.PageSize), nil
	}
	var list []entity.SysDictItem
	if err := m.Page(param.PageNo, param.PageSize).Order(cols.SortOrder).Scan(&list); err != nil {
		return nil, fmt.Errorf("查询字典项列表失败: %w", err)
	}
	return page.New(list, total, param.PageNo, param.PageSize), nil
}

// ItemAdd 新增字典项。
func (s *Service) ItemAdd(ctx context.Context, in *do.SysDictItem) error {
	if in.DictId == nil || strings.TrimSpace(gconv.String(in.DictId)) == "" {
		return ErrDictCodeInvalid
	}
	data := *in
	data.Id = guid.S()
	data.CreateTime = gtime.Now()
	if _, err := dao.SysDictItem.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("保存字典项失败: %w", err)
	}
	return nil
}

// ItemEdit 编辑字典项。
func (s *Service) ItemEdit(ctx context.Context, id string, in *do.SysDictItem) error {
	cols := dao.SysDictItem.Columns()
	one, err := dao.SysDictItem.Ctx(ctx).Where(cols.Id, id).One()
	if err != nil {
		return fmt.Errorf("查询字典项失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	data := *in
	data.Id = id
	data.UpdateTime = gtime.Now()
	if _, err := dao.SysDictItem.Ctx(ctx).Where(cols.Id, id).Data(data).Update(); err != nil {
		return fmt.Errorf("更新字典项失败: %w", err)
	}
	return nil
}

// ItemDelete 删除字典项（sys_dict_item 无软删列，物理删除，语义同 Java removeById）。
func (s *Service) ItemDelete(ctx context.Context, id string) error {
	cols := dao.SysDictItem.Columns()
	one, err := dao.SysDictItem.Ctx(ctx).Where(cols.Id, id).One()
	if err != nil {
		return fmt.Errorf("查询字典项失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	if _, err := dao.SysDictItem.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除字典项失败: %w", err)
	}
	return nil
}

// ItemDeleteBatch 批量删除字典项。
func (s *Service) ItemDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return ErrIDsRequired
	}
	for _, id := range ids {
		if err := s.ItemDelete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// applyStarFilter 应用字符串过滤：值含 `*` 时按 `*`→`%` 走 LIKE，否则精确匹配。
// 空值直接跳过。语义对齐 JeecgBoot QueryGenerator 的通配符约定。
func applyStarFilter(m *gdb.Model, col, value string) *gdb.Model {
	value = strings.TrimSpace(value)
	if value == "" {
		return m
	}
	if strings.Contains(value, "*") {
		return m.WhereLike(col, strings.ReplaceAll(value, "*", "%"))
	}
	return m.Where(col, value)
}

// TreeSelectModel 通用树形下拉节点（对应 Java TreeSelectModel）。
type TreeSelectModel struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	IsLeaf   bool   `json:"isLeaf"`
	ParentId string `json:"parentId,omitempty"`
	Value    string `json:"value,omitempty"`
	Code     string `json:"code,omitempty"`
}

// treeTableWhitelist 通用树查询允许的表（Java 版是任意 tableName 拼接，
// 存在注入面；Go 侧收紧为显式白名单，其余表名直接拒绝）。
var treeTableWhitelist = map[string]struct{}{
	"sys_category":   {},
	"sys_dict":       {},
	"sys_depart":     {},
	"sys_permission": {},
	"sys_role":       {},
}

// LoadTreeData 通用表树加载（契约蓝本 SysDictController#loadTreeData）。
//
// 参数与 Java 一致：pid/pidField/tableName/text/code/hasChildField/condition。
// 表名必须命中白名单，pidField/text/code/hasChildField 必须通过标识符校验；
// condition 为 JSON 对象，其键同样校验后以占位符绑定值。
func (s *Service) LoadTreeData(ctx context.Context, pid, pidField, table, text, code, hasChildField, condition string) ([]TreeSelectModel, error) {
	if _, ok := treeTableWhitelist[table]; !ok {
		return nil, ErrDictCodeInvalid
	}
	for _, ident := range []string{pidField, text, code, hasChildField} {
		if ident != "" && !ValidIdent(ident) {
			return nil, ErrDictCodeInvalid
		}
	}
	if pidField == "" {
		pidField = "pid"
	}
	if text == "" || code == "" {
		return nil, ErrDictCodeInvalid
	}
	// isLeaf：有 hasChildField 时按其值取反；无则视作叶子（与 Java 默认一致）。
	leafExpr := "1"
	if hasChildField != "" {
		leafExpr = fmt.Sprintf("(CASE WHEN %s = '1' THEN 0 ELSE 1 END)", hasChildField)
	}
	query := fmt.Sprintf(
		"SELECT %s AS `key`, %s AS `title`, %s AS `parentId`, %s AS `code`, %s AS isLeaf FROM %s WHERE %s = ?",
		code, text, pidField, code, leafExpr, table, pidField)
	args := []interface{}{pid}

	if condition != "" {
		cond := map[string]string{}
		if err := jsonUnmarshalHelper(condition, &cond); err != nil {
			return nil, ErrTableDictFilterUnsupported
		}
		for k, v := range cond {
			if !ValidIdent(k) {
				return nil, ErrDictCodeInvalid
			}
			query += " AND " + k + " = ?"
			args = append(args, v)
		}
	}
	records, err := dao.SysDict.DB().GetAll(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询树数据失败: %w", err)
	}
	var list []TreeSelectModel
	if err := records.Structs(&list); err != nil {
		return nil, fmt.Errorf("解析树数据失败: %w", err)
	}
	return list, nil
}

// jsonUnmarshalHelper JSON 反序列化（避免新增依赖导入位置分散）。
func jsonUnmarshalHelper(data string, v interface{}) error {
	return gjson.Unmarshal([]byte(data), v)
}
