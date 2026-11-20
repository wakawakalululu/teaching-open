// Package category 承载分类字典业务用例（契约蓝本：SysCategoryController）。
//
// 关键语义与 Java 版逐项对齐：
//   - code 由 YouBian 流水规则在服务端生成（前端传入值被覆盖）；
//   - pid 缺省为 "0"（根节点）；新增/编辑子节点时父节点 has_child 置 "1"；
//   - sys_category 无软删列，delete 为物理删除（同 Java removeById）。
//
// 安全改进：Java 版 loadOne/loadTreeData 把请求参数直接拼进 QueryWrapper
// （${key} 插值存在注入面），Go 版一律按列名白名单校验后才允许使用。
package category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// RootPidValue 根节点 pid（Java ISysCategoryService.ROOT_PID_VALUE）。
const RootPidValue = "0"

// 业务错误。
var (
	// ErrEntityNotFound 未找到对应实体。
	ErrEntityNotFound = errors.New("未找到对应实体")
	// ErrInvalidColumn 查询列不合法（白名单外）。
	ErrInvalidColumn = errors.New("查询字段不合法")
	// ErrLoadTreeParam 加载分类字典树参数有误。
	ErrLoadTreeParam = errors.New("加载分类字典树参数有误")
	// ErrCategoryCodeEmpty 类型编码为空。
	ErrCategoryCodeEmpty = errors.New("错误,类型编码为空!")
)

// TreeSelectModel 树形下拉节点（对应 Java TreeSelectModel）。
type TreeSelectModel struct {
	Key      string            `json:"key"`
	Title    string            `json:"title"`
	IsLeaf   bool              `json:"isLeaf"`
	ParentId string            `json:"parentId,omitempty"`
	Value    string            `json:"value,omitempty"`
	Code     string            `json:"code,omitempty"`
	Children []TreeSelectModel `json:"children,omitempty"`
}

// DictModel 分类数据对（value=id，text=name，同 Java new DictModel(id, name)）。
type DictModel struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

// categoryColumns sys_category 合法列集合（白名单，用于 loadOne/condition 校验）。
var categoryColumns = buildColumnSet()

func buildColumnSet() map[string]struct{} {
	cols := dao.SysCategory.Columns()
	set := make(map[string]struct{}, 16)
	v := reflect.ValueOf(cols)
	for i := 0; i < v.Type().NumField(); i++ {
		set[v.Field(i).String()] = struct{}{}
	}
	return set
}

// Service 分类字典业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// Add 新增分类：code 由服务端按流水规则生成，父节点 has_child 维护在事务内完成。
func (s *Service) Add(ctx context.Context, in *do.SysCategory) error {
	pid := RootPidValue
	if in.Pid != nil && gconv.String(in.Pid) != "" {
		pid = gconv.String(in.Pid)
	}

	if err := dao.SysCategory.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 非根节点：父节点 has_child 置 1
		if pid != RootPidValue {
			if err := markParentHasChild(ctx, pid); err != nil {
				return err
			}
		}
		code, err := s.nextCategoryCode(ctx, pid)
		if err != nil {
			return err
		}
		data := *in
		data.Id = guid.S()
		data.Pid = pid
		data.Code = code
		data.CreateTime = gtime.Now()
		_, err = dao.SysCategory.Ctx(ctx).Data(data).Insert()
		return err
	}); err != nil {
		return fmt.Errorf("保存分类失败: %w", err)
	}
	return nil
}

// Edit 编辑分类：存在性校验 + 父节点 has_child 维护；code 不由编辑重生成（同 Java）。
func (s *Service) Edit(ctx context.Context, id string, in *do.SysCategory) error {
	cols := dao.SysCategory.Columns()
	one, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, id).One()
	if err != nil {
		return fmt.Errorf("查询分类失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}

	pid := RootPidValue
	if in.Pid != nil && gconv.String(in.Pid) != "" {
		pid = gconv.String(in.Pid)
	}
	if pid != RootPidValue {
		if err := markParentHasChild(ctx, pid); err != nil {
			return err
		}
	}

	data := *in
	data.Id = id
	data.Pid = pid
	data.UpdateTime = gtime.Now()
	if _, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, id).Data(data).Update(); err != nil {
		return fmt.Errorf("更新分类失败: %w", err)
	}
	return nil
}

// markParentHasChild 父节点 has_child 置 "1"（已标记则跳过）。
func markParentHasChild(ctx context.Context, pid string) error {
	cols := dao.SysCategory.Columns()
	one, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, pid).One()
	if err != nil {
		return fmt.Errorf("查询父节点失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	if one[cols.HasChild].String() == "1" {
		return nil
	}
	if _, err := dao.SysCategory.Ctx(ctx).
		Where(cols.Id, pid).
		Data(do.SysCategory{HasChild: "1"}).
		Update(); err != nil {
		return fmt.Errorf("更新父节点标记失败: %w", err)
	}
	return nil
}

// nextCategoryCode 生成下一个分类编码（移植 CategoryCodeRule）。
// 三种情况：同级无数据（根/子）、同级有数据（取最大 code 递增）。
func (s *Service) nextCategoryCode(ctx context.Context, pid string) (string, error) {
	cols := dao.SysCategory.Columns()
	one, err := dao.SysCategory.Ctx(ctx).
		Where(cols.Pid, pid).
		WhereNotNull(cols.Code).
		Order(cols.Code + " DESC").
		One()
	if err != nil {
		return "", fmt.Errorf("查询同级编码失败: %w", err)
	}
	if one.IsEmpty() {
		if pid == RootPidValue {
			return nextYouBianCode(""), nil
		}
		parent, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, pid).One()
		if err != nil {
			return "", fmt.Errorf("查询父节点失败: %w", err)
		}
		if parent.IsEmpty() {
			return "", ErrEntityNotFound
		}
		return subYouBianCode(parent[cols.Code].String(), ""), nil
	}
	return nextYouBianCode(one[cols.Code].String()), nil
}

// RootList 根列表分页查询：仅按 pid 过滤（Java rootList 覆盖了 QueryGenerator）。
func (s *Service) RootList(ctx context.Context, pid string, pageNo, pageSize int) (gdb.Result, int, error) {
	cols := dao.SysCategory.Columns()
	if pid == "" {
		pid = RootPidValue
	}
	m := dao.SysCategory.Ctx(ctx).Where(cols.Pid, pid)
	total, err := m.Count()
	if err != nil {
		return nil, 0, fmt.Errorf("统计分类数量失败: %w", err)
	}
	if total == 0 {
		return gdb.Result{}, 0, nil
	}
	list, err := m.Page(pageNo, pageSize).All()
	if err != nil {
		return nil, 0, fmt.Errorf("查询分类列表失败: %w", err)
	}
	return list, total, nil
}

// ChildList 子列表查询：pid 非空按 pid，否则按 code（同 Java 的互斥逻辑）。
func (s *Service) ChildList(ctx context.Context, pid, code string) ([]entity.SysCategory, error) {
	cols := dao.SysCategory.Columns()
	var (
		list []entity.SysCategory
		err  error
	)
	switch {
	case pid != "":
		err = dao.SysCategory.Ctx(ctx).Where(cols.Pid, pid).Scan(&list)
	case code != "":
		err = dao.SysCategory.Ctx(ctx).Where(cols.Code, code).Scan(&list)
	default:
		err = dao.SysCategory.Ctx(ctx).Scan(&list)
	}
	if err != nil {
		return nil, fmt.Errorf("查询分类列表失败: %w", err)
	}
	return list, nil
}

// QueryById 按 id 查询。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.SysCategory, error) {
	one, err := dao.SysCategory.Ctx(ctx).Where(dao.SysCategory.Columns().Id, id).One()
	if err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	if one.IsEmpty() {
		return nil, ErrEntityNotFound
	}
	var entityData *entity.SysCategory
	if err := one.Struct(&entityData); err != nil {
		return nil, fmt.Errorf("解析分类数据失败: %w", err)
	}
	return entityData, nil
}

// Delete 物理删除（sys_category 无软删列）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.SysCategory.Columns()
	one, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, id).One()
	if err != nil {
		return fmt.Errorf("查询分类失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	if _, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除分类失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量物理删除。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// LoadOne 单条回显：field 列名白名单校验；0/多条均为查询失败（同 Java 语义）。
func (s *Service) LoadOne(ctx context.Context, field, val string) (*entity.SysCategory, error) {
	if !validCategoryColumn(field) {
		return nil, ErrInvalidColumn
	}
	list, err := dao.SysCategory.Ctx(ctx).Where(field, val).All()
	if err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	switch len(list) {
	case 0:
		return nil, errors.New("查询无果")
	case 1:
		var entityData *entity.SysCategory
		if err := list[0].Struct(&entityData); err != nil {
			return nil, fmt.Errorf("解析分类数据失败: %w", err)
		}
		return entityData, nil
	default:
		return nil, fmt.Errorf("查询数据异常,[%s]存在多个值:%s", field, val)
	}
}

// queryListByPid 按 pid 查树节点（可选条件键值对，键必须在白名单内）。
func (s *Service) queryListByPid(ctx context.Context, pid string, condition map[string]string) ([]TreeSelectModel, error) {
	if pid == "" {
		pid = RootPidValue
	}
	query := fmt.Sprintf(
		"SELECT code, name AS `title`, id AS `key`, (CASE WHEN has_child = '1' THEN 0 ELSE 1 END) AS isLeaf, pid AS parentId, value FROM sys_category WHERE pid = ?")
	args := []interface{}{pid}
	for k, v := range condition {
		if !validCategoryColumn(k) {
			return nil, ErrInvalidColumn
		}
		query += " AND " + k + " = ?"
		args = append(args, v)
	}
	records, err := dao.SysCategory.DB().GetAll(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询分类树失败: %w", err)
	}
	var list []TreeSelectModel
	if err := records.Structs(&list); err != nil {
		return nil, fmt.Errorf("解析分类树失败: %w", err)
	}
	return list, nil
}

// LoadTreeChildren 加载指定节点的子数据。
func (s *Service) LoadTreeChildren(ctx context.Context, pid string) ([]TreeSelectModel, error) {
	return s.queryListByPid(ctx, pid, nil)
}

// LoadTreeRoot 加载一级节点；async=false 时递归装载全部子孙节点。
func (s *Service) LoadTreeRoot(ctx context.Context, async bool, pcode string) ([]TreeSelectModel, error) {
	pid, err := s.resolvePidByPcode(ctx, pcode)
	if err != nil {
		return nil, err
	}
	list, err := s.queryListByPid(ctx, pid, nil)
	if err != nil {
		return nil, err
	}
	if !async {
		s.loadAllChildren(ctx, list)
	}
	return list, nil
}

// loadAllChildren 递归求子节点（同步加载用）。
func (s *Service) loadAllChildren(ctx context.Context, list []TreeSelectModel) {
	for i := range list {
		children, err := s.queryListByPid(ctx, list[i].Key, nil)
		if err != nil || len(children) == 0 {
			continue
		}
		list[i].Children = children
		s.loadAllChildren(ctx, children)
	}
}

// LoadTreeData 分类树控件加载：pid 优先；pid 为空时按 pcode 解析；condition 键白名单校验。
func (s *Service) LoadTreeData(ctx context.Context, pid, pcode, condition string) ([]TreeSelectModel, error) {
	if pid == "" {
		if pcode == "" {
			return nil, fmt.Errorf("%w.[null]!", ErrLoadTreeParam)
		}
		var err error
		if pid, err = s.resolvePidByPcode(ctx, pcode); err != nil {
			return nil, err
		}
		if pid == "" {
			return nil, fmt.Errorf("%w.[code]!", ErrLoadTreeParam)
		}
	}
	query := map[string]string{}
	if condition != "" {
		if err := json.Unmarshal([]byte(condition), &query); err != nil {
			return nil, ErrInvalidColumn
		}
	}
	return s.queryListByPid(ctx, pid, query)
}

// resolvePidByPcode 按编码解析 pid：pcode 为根编码时直接返回根。
// 编码不存在或重复时返回与 Java 一致的错误话术。
func (s *Service) resolvePidByPcode(ctx context.Context, pcode string) (string, error) {
	if pcode == RootPidValue {
		return RootPidValue, nil
	}
	if pcode == "" {
		return RootPidValue, nil
	}
	cols := dao.SysCategory.Columns()
	list, err := dao.SysCategory.Ctx(ctx).Fields(cols.Id).Where(cols.Code, pcode).All()
	if err != nil {
		return "", fmt.Errorf("查询编码失败: %w", err)
	}
	switch len(list) {
	case 0:
		return "", fmt.Errorf("该编码【%s】不存在，请核实!", pcode)
	case 1:
		return list[0]["id"].String(), nil
	default:
		return "", fmt.Errorf("该编码【%s】存在多个，请核实!", pcode)
	}
}

// LoadDictItem 数据回显：id 或 value 命中任一即可，且必须属于 pcode 解析出的父节点。
func (s *Service) LoadDictItem(ctx context.Context, ids, pcode string) ([]entity.SysCategory, error) {
	if ids == "" {
		return nil, errors.New("ids 不能为空")
	}
	idList := splitIDs(ids)
	pid, err := s.resolvePidByPcode(ctx, pcode)
	if err != nil {
		return nil, err
	}
	if pid == "" {
		return nil, fmt.Errorf("%w.[code]!", ErrLoadTreeParam)
	}
	cols := dao.SysCategory.Columns()
	var list []entity.SysCategory
	err = dao.SysCategory.Ctx(ctx).
		Where(fmt.Sprintf("(`%s` IN (?) OR `%s` IN (?)) AND `%s` = ?", cols.Id, cols.Value, cols.Pid), idList, idList, pid).
		Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	return list, nil
}

// LoadAllData 列表页数据替换：按 code 前缀过滤，输出 value=id、text=name。
func (s *Service) LoadAllData(ctx context.Context, code string) ([]DictModel, error) {
	cols := dao.SysCategory.Columns()
	m := dao.SysCategory.Ctx(ctx)
	if code != "" && code != RootPidValue {
		m = m.WhereLike(cols.Code, code+"%")
	}
	list, err := m.All()
	if err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	if len(list) == 0 {
		return nil, errors.New("无数据,参数有误.[code]")
	}
	result := make([]DictModel, 0, len(list))
	for _, rec := range list {
		result = append(result, DictModel{
			Value: rec[cols.Id].String(),
			Text:  rec[cols.Name].String(),
		})
	}
	return result, nil
}

// CheckCode 校验子分类编码必须以父分类编码为前缀。
func (s *Service) CheckCode(ctx context.Context, pid, code string) error {
	if code == "" {
		return ErrCategoryCodeEmpty
	}
	if pid == "" {
		return nil
	}
	cols := dao.SysCategory.Columns()
	one, err := dao.SysCategory.Ctx(ctx).Where(cols.Id, pid).One()
	if err != nil {
		return fmt.Errorf("查询父节点失败: %w", err)
	}
	if one.IsEmpty() {
		return ErrEntityNotFound
	}
	parentCode := one[cols.Code].String()
	if len(code) >= len(parentCode) && code[:len(parentCode)] == parentCode {
		return nil
	}
	return fmt.Errorf("编码不符合规范,须以\"%s\"开头!", parentCode)
}

// validCategoryColumn 列名白名单校验。
func validCategoryColumn(col string) bool {
	if col == "" {
		return false
	}
	_, ok := categoryColumns[col]
	return ok
}

// splitIDs 逗号拆分并去除空白项。
func splitIDs(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
