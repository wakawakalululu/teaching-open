// Package departpermission 承载部门权限域业务用例
// （契约蓝本：SysDepartPermissionController 的 11 个核心端点；exportXls/
// importExcel 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - datarule GET 的规则来源为「该菜单全部规则」（sys_permission_data_rule
//     by permission_id），drChecked 来自 sys_depart_permission.data_rule_ids；
//   - saveDeptRolePermission 的 diff 新增不写 depart_id（Java 两参构造器落 NULL，
//     queryDeptRolePermission 只消费 permission_id，行为不受影响，逐语义保留）；
//   - queryTreeListForDeptRole 的 inSql 字符串拼接（注入面）改为「先查授权
//     permission_id 集合再参数化 IN」，结果等价；
//   - TreeModel 使用 String 五参构造器路径：title=null、slotTitle=name、
//     icon=null、ruleFlag=菜单 rule_flag、scopedSlots={"title":"hasDatarule"}。
package departpermission

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

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrNoPermConfig 菜单无数据规则（Java datarule GET error500 原话术）。
	ErrNoPermConfig = errors.New("未找到权限配置信息")
	// ErrSaveDeptPermFirst 部门菜单授权关系缺失（Java datarule POST error500 原话术）。
	ErrSaveDeptPermFirst = errors.New("请先保存部门菜单权限!")
	// ErrSaveFailed datarule POST 异常兜底（Java catch 原话术）。
	ErrSaveFailed = errors.New("保存失败")
)

// DepartPermissionInput add/edit 请求体（SysDepartPermission 可编辑字段）。
type DepartPermissionInput struct {
	Id           string
	DepartId     *string
	PermissionId *string
	DataRuleIds  *string
}

// departTreeModel 对齐 Java TreeModel（String 五参构造器路径，供
// queryTreeListForDeptRole 使用）：
// title=null、slotTitle=name、icon=null、value=key、label=null；
// scopedSlots 固定 {"title":"hasDatarule"}；叶子 children=null。
type departTreeModel struct {
	Key         string            `json:"key"`
	Title       *string           `json:"title"`
	SlotTitle   string            `json:"slotTitle"`
	Value       string            `json:"value"`
	Label       *string           `json:"label"`
	ParentId    string            `json:"parentId"`
	Icon        *string           `json:"icon"`
	RuleFlag    int               `json:"ruleFlag"`
	IsLeaf      bool              `json:"isLeaf"`
	ScopedSlots map[string]string `json:"scopedSlots"`
	Children    []departTreeModel `json:"children"`
}

// DataruleResult GET datarule 的返回。
type DataruleResult struct {
	Rules      []entity.SysPermissionDataRule
	DrChecked  string
	HasChecked bool
}

// Service 部门权限业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// List 部门授权分页列表。
//
// 契约蓝本 SysDepartPermissionController#list（GET /sys/sysDepartPermission/list）：
// QueryGenerator 实体过滤消费面（departId/permissionId 精确 + 其余字符串规则）。
func (s *Service) List(ctx context.Context, param page.Param, departID, permissionID string) (*page.Result, error) {
	cols := dao.SysDepartPermission.Columns()
	m := dao.SysDepartPermission.Ctx(ctx)
	if departID != "" {
		m = m.Where(cols.DepartId, departID)
	}
	if permissionID != "" {
		m = m.Where(cols.PermissionId, permissionID)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计部门授权数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysDepartPermission{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysDepartPermission
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门授权失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增部门授权（纯 save，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in DepartPermissionInput) error {
	data := do.SysDepartPermission{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyInput(&data, in)
	if _, err := dao.SysDepartPermission.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑部门授权（纯 updateById，无存在性校验，对齐 Java）。
func (s *Service) Edit(ctx context.Context, in DepartPermissionInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.SysDepartPermission{}
	applyInput(&data, in)
	cols := dao.SysDepartPermission.Columns()
	if _, err := dao.SysDepartPermission.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除部门授权（无级联，对齐 Java removeById）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.SysDepartPermission.Columns()
	if _, err := dao.SysDepartPermission.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除部门授权失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除部门授权（无级联，对齐 Java removeByIds）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.SysDepartPermission.Columns()
	if _, err := dao.SysDepartPermission.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除部门授权失败: %w", err)
	}
	return nil
}

// QueryById 部门授权详情（Result.ok(entity)：不存在时 result=null 仍 success）。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.SysDepartPermission, error) {
	var row entity.SysDepartPermission
	if err := dao.SysDepartPermission.Ctx(ctx).Where(dao.SysDepartPermission.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询部门授权失败: %w", err)
	}
	if row.Id == "" {
		return nil, nil
	}
	return &row, nil
}

// DataruleList 部门的菜单数据规则与已勾选项。
//
// 契约蓝本 SysDepartPermissionController#datarule GET：
//   - datarule = 该菜单全部规则（sys_permission_data_rule by permission_id，
//     create_time 倒序）；为空 →「未找到权限配置信息」；
//   - drChecked 来自 sys_depart_permission (depart_id, permission_id) 的
//     data_rule_ids，非空时去末尾一个逗号。
func (s *Service) DataruleList(ctx context.Context, permissionID, departID string) (*DataruleResult, error) {
	drCols := dao.SysPermissionDataRule.Columns()
	var rules []entity.SysPermissionDataRule
	if err := dao.SysPermissionDataRule.Ctx(ctx).
		Where(drCols.PermissionId, permissionID).
		OrderDesc(drCols.CreateTime).
		Scan(&rules); err != nil {
		return nil, fmt.Errorf("查询数据规则失败: %w", err)
	}
	if len(rules) == 0 {
		return nil, ErrNoPermConfig
	}
	res := &DataruleResult{Rules: rules}
	dpCols := dao.SysDepartPermission.Columns()
	var dp entity.SysDepartPermission
	if err := dao.SysDepartPermission.Ctx(ctx).
		Where(dpCols.DepartId, departID).
		Where(dpCols.PermissionId, permissionID).
		Scan(&dp); err != nil {
		return nil, fmt.Errorf("查询部门授权失败: %w", err)
	}
	if dp.Id != "" && dp.DataRuleIds != "" {
		res.DrChecked = strings.TrimSuffix(dp.DataRuleIds, ",")
		res.HasChecked = true
	}
	return res, nil
}

// DataruleSave 保存部门的菜单数据规则勾选。
//
// 契约蓝本 SysDepartPermissionController#datarule POST（body {permissionId,
// departId, dataRuleIds}）：sys_depart_permission 必须已有 (depart_id,
// permission_id) 行（update-only）→「请先保存部门菜单权限!」；多行命中 →
// 「保存失败」（Java getOne throwEx）。
func (s *Service) DataruleSave(ctx context.Context, permissionID, departID string, dataRuleIDs *string) error {
	dpCols := dao.SysDepartPermission.Columns()
	var rows []entity.SysDepartPermission
	if err := dao.SysDepartPermission.Ctx(ctx).
		Where(dpCols.DepartId, departID).
		Where(dpCols.PermissionId, permissionID).
		Scan(&rows); err != nil {
		return fmt.Errorf("查询部门授权失败: %w", err)
	}
	if len(rows) == 0 {
		return ErrSaveDeptPermFirst
	}
	if len(rows) > 1 {
		return ErrSaveFailed
	}
	if dataRuleIDs == nil {
		return nil
	}
	if _, err := dao.SysDepartPermission.Ctx(ctx).
		Where(dpCols.Id, rows[0].Id).
		Data(do.SysDepartPermission{DataRuleIds: *dataRuleIDs}).Update(); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}
	return nil
}

// QueryDeptRolePermission 部门角色已授权的菜单 ID 列表。
//
// 契约蓝本 SysDepartPermissionController#queryDeptRolePermission（GET ?roleId=）：
// sys_depart_role_permission WHERE role_id → permission_id 列表。
func (s *Service) QueryDeptRolePermission(ctx context.Context, roleID string) ([]string, error) {
	cols := dao.SysDepartRolePermission.Columns()
	var rows []entity.SysDepartRolePermission
	if err := dao.SysDepartRolePermission.Ctx(ctx).
		Fields(cols.PermissionId).
		Where(cols.RoleId, roleID).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门角色授权失败: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.PermissionId)
	}
	return out, nil
}

// SaveDeptRolePermission 保存部门角色的菜单授权（diff 增删）。
//
// 契约蓝本 SysDepartPermissionController#saveDeptRolePermission（body
// {roleId, permissionIds, lastpermissionIds} 小写 p）+ SysDepartRolePermissionServiceImpl：
//   - 新增 = 在新不在旧 → 插入 (role_id, permission_id)，depart_id 不写
//     （Java 两参构造器落 NULL，逐语义保留）；
//   - 删除 = 在旧不在新 → 按 (role_id, permission_id) 删除；
//   - Java 无事务 → Go 包单事务（文档化改进）。
func (s *Service) SaveDeptRolePermission(ctx context.Context, roleID, permIDs, lastPermIDs string) error {
	addList := diffIDs(lastPermIDs, permIDs)
	delList := diffIDs(permIDs, lastPermIDs)
	if len(addList) == 0 && len(delList) == 0 {
		return nil
	}
	return dao.SysDepartRolePermission.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.SysDepartRolePermission.Columns()
		for _, pid := range addList {
			if pid == "" {
				continue
			}
			if _, err := dao.SysDepartRolePermission.Ctx(ctx).Data(do.SysDepartRolePermission{
				Id:           guid.S(),
				RoleId:       roleID,
				PermissionId: pid,
				// depart_id 有意不写：对齐 Java 两参构造器的 NULL 落库
			}).Insert(); err != nil {
				return fmt.Errorf("保存部门角色授权失败: %w", err)
			}
		}
		for _, pid := range delList {
			if pid == "" {
				continue
			}
			if _, err := dao.SysDepartRolePermission.Ctx(ctx).
				Where(cols.RoleId, roleID).
				Where(cols.PermissionId, pid).
				Delete(); err != nil {
				return fmt.Errorf("删除部门角色授权失败: %w", err)
			}
		}
		return nil
	})
}

// QueryTreeListForDeptRole 部门角色授权弹窗的菜单树。
//
// 契约蓝本 SysDepartPermissionController#queryTreeListForDeptRole（GET ?departId=）：
//   - 菜单范围 = sys_depart_permission WHERE depart_id 的 permission_id 集合
//     （Java inSql 字符串拼接 → Go 先查集合再参数化 IN，结果等价）；
//   - sys_permission del_flag=0 + sort_no ASC；TreeModel 为 String 五参构造器
//     路径（title=null、slotTitle=name、ruleFlag、scopedSlots 固定值）；
//   - result {treeList, ids}。
func (s *Service) QueryTreeListForDeptRole(ctx context.Context, departID string) ([]departTreeModel, []string, error) {
	dpCols := dao.SysDepartPermission.Columns()
	var permIDs []string
	if err := dao.SysDepartPermission.Ctx(ctx).
		Fields(dpCols.PermissionId).
		Where(dpCols.DepartId, departID).
		Scan(&permIDs); err != nil {
		return nil, nil, fmt.Errorf("查询部门授权失败: %w", err)
	}
	if len(permIDs) == 0 {
		return []departTreeModel{}, []string{}, nil
	}
	cols := dao.SysPermission.Columns()
	var rows []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Where(cols.DelFlag, 0).
		WhereIn(cols.Id, permIDs).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, nil, fmt.Errorf("查询菜单权限失败: %w", err)
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	tree := buildDeptRoleTree(rows)
	return tree, ids, nil
}

// buildDeptRoleTree 构建部门角色授权树（对齐 SysDepartPermissionController#
// getTreeModelList 递归：根 = parentId 空；非叶才有 children 容器并递归）。
func buildDeptRoleTree(perms []entity.SysPermission) []departTreeModel {
	var out []departTreeModel
	appendDeptRoleTree(&out, perms, nil)
	return out
}

// appendDeptRoleTree 树递归挂载（parent 为 nil 时收集根节点，孤儿丢弃）。
func appendDeptRoleTree(dest *[]departTreeModel, perms []entity.SysPermission, parent *departTreeModel) {
	for i := range perms {
		p := perms[i]
		node := newDeptRoleTreeModel(p)
		switch {
		case parent == nil && p.ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !last.IsLeaf {
				appendDeptRoleTree(&last.Children, perms, last)
			}
		case parent != nil && p.ParentId == parent.Key:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !last.IsLeaf {
				appendDeptRoleTree(&last.Children, perms, last)
			}
		}
	}
}

// newDeptRoleTreeModel 对齐 TreeModel(String key, String parentId, String
// slotTitle, Integer ruleFlag, boolean isLeaf) 构造器 + scopedSlots 默认值。
func newDeptRoleTreeModel(p entity.SysPermission) departTreeModel {
	node := departTreeModel{
		Key:         p.Id,
		SlotTitle:   p.Name,
		Value:       p.Id,
		ParentId:    p.ParentId,
		RuleFlag:    p.RuleFlag,
		IsLeaf:      p.IsLeaf == 1,
		ScopedSlots: map[string]string{"title": "hasDatarule"},
	}
	if !node.IsLeaf {
		node.Children = []departTreeModel{}
	}
	return node
}

// diffIDs 纯函数（与 departrole 域同语义）：取 diff 中 main 没有的元素。
func diffIDs(main, diff string) []string {
	if diff == "" {
		return nil
	}
	diffArr := strings.Split(diff, ",")
	if main == "" {
		return diffArr
	}
	set := make(map[string]struct{}, len(main))
	for _, s := range strings.Split(main, ",") {
		set[s] = struct{}{}
	}
	var res []string
	for _, s := range diffArr {
		if s != "" {
			if _, ok := set[s]; !ok {
				res = append(res, s)
			}
		}
	}
	return res
}

// applyInput 请求体字段拷贝进 DO。
func applyInput(data *do.SysDepartPermission, in DepartPermissionInput) {
	if in.DepartId != nil {
		data.DepartId = *in.DepartId
	}
	if in.PermissionId != nil {
		data.PermissionId = *in.PermissionId
	}
	if in.DataRuleIds != nil {
		data.DataRuleIds = *in.DataRuleIds
	}
}
