// Package departrole 承载部门角色域业务用例
// （契约蓝本：SysDepartRoleController 的 11 个核心端点；exportXls/importExcel
// 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - delete/deleteBatch 无级联（Java removeById/removeByIds 原样），遗留的
//     sys_depart_role_permission/sys_depart_role_user 孤儿行为逐语义保留；
//   - add/edit 不写审计列（Java 控制器不 setCreateBy/UpdateTime）；
//   - getDeptRoleList 的 SQL 为「用户所在部门 ∪ org_code 前缀子部门」的角色集合；
//   - deptRoleUserAdd 为 diff 增删，Java 无事务 → Go 包单事务（文档化改进）；
//   - datarule GET 的规则来源为「部门已授权规则」（sys_depart_permission.
//     data_rule_ids 拆分 IN），drChecked 来自 sys_depart_role_permission。
package departrole

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrNoPermConfig 部门未授权菜单或菜单无规则（Java datarule GET error500 原话术）。
	ErrNoPermConfig = errors.New("未找到权限配置信息")
	// ErrSaveRolePermFirst 部门角色菜单授权关系缺失（Java datarule POST error500 原话术）。
	ErrSaveRolePermFirst = errors.New("请先保存角色菜单权限!")
	// ErrSaveFailed datarule POST 异常兜底（Java catch 原话术）。
	ErrSaveFailed = errors.New("保存失败")
	// ErrOpFailed departId 无效等运行时异常的脱敏话术（Java NPE → 全局处理器）。
	ErrOpFailed = errors.New("操作失败")
)

// DepartRoleInput add/edit 请求体（SysDepartRole 可编辑字段）。
type DepartRoleInput struct {
	Id          string
	DepartId    *string
	RoleName    *string
	RoleCode    *string
	Description *string
}

// DataruleResult GET datarule 的返回。
type DataruleResult struct {
	Rules      []entity.SysPermissionDataRule
	DrChecked  string
	HasChecked bool
}

// Service 部门角色业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// List 部门角色分页列表。
//
// 契约蓝本 SysDepartRoleController#list（GET /sys/sysDepartRole/list）：
//   - 固定 eq depart_id=deptId（deptId 缺省时 Java eq(null) 拼出 depart_id=NULL
//     永不命中 → 空页，Go 等价短路）；实体字段过滤走 QueryGenerator 消费面；
//   - result：IPage {records,total,size,current,pages}，records 为 SysDepartRole 全字段。
func (s *Service) List(ctx context.Context, param page.Param, deptID, roleName, roleCode, description string) (*page.Result, error) {
	if deptID == "" {
		// Java：eq("depart_id", null) 永不命中 → 空页
		return page.New([]entity.SysDepartRole{}, 0, param.PageNo, param.PageSize), nil
	}
	cols := dao.SysDepartRole.Columns()
	m := dao.SysDepartRole.Ctx(ctx).Where(cols.DepartId, deptID)
	if roleName != "" {
		m = applyQueryRule(m, cols.RoleName, roleName)
	}
	if roleCode != "" {
		m = applyQueryRule(m, cols.RoleCode, roleCode)
	}
	if description != "" {
		m = applyQueryRule(m, cols.Description, description)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计部门角色数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysDepartRole{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysDepartRole
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门角色失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增部门角色。
//
// 契约蓝本 SysDepartRoleController#add：纯 save（不写审计列，body 未提供即
// 落列默认值）；id 为空时服务端生成（Java ID_WORKER_STR 雪花字符串）。
func (s *Service) Add(ctx context.Context, in DepartRoleInput) error {
	data := do.SysDepartRole{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyDepartRoleInput(&data, in)
	if _, err := dao.SysDepartRole.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑部门角色。
//
// 契约蓝本 SysDepartRoleController#edit：纯 updateById（无存在性校验，
// 0 行受影响同样返回「编辑成功!」），不写审计列。
func (s *Service) Edit(ctx context.Context, in DepartRoleInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.SysDepartRole{}
	applyDepartRoleInput(&data, in)
	cols := dao.SysDepartRole.Columns()
	if _, err := dao.SysDepartRole.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除部门角色（无级联，对齐 Java removeById）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.SysDepartRole.Columns()
	if _, err := dao.SysDepartRole.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除部门角色失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除部门角色（无级联，对齐 Java removeByIds）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.SysDepartRole.Columns()
	if _, err := dao.SysDepartRole.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除部门角色失败: %w", err)
	}
	return nil
}

// QueryById 部门角色详情。
//
// 契约蓝本 SysRoleController... 确切说 SysDepartRoleController#queryById：
// Result.ok(entity)（实体不存在时 result=null、仍 success），Go 等价返回
// (nil, nil)。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.SysDepartRole, error) {
	var row entity.SysDepartRole
	if err := dao.SysDepartRole.Ctx(ctx).Where(dao.SysDepartRole.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询部门角色失败: %w", err)
	}
	if row.Id == "" {
		return nil, nil
	}
	return &row, nil
}

// GetDeptRoleList 查询用户在指定部门（含前缀子部门）下可用的部门角色。
//
// 契约蓝本 SysDepartRoleController#getDeptRoleList（GET ?departId=&userId=）+
// SysDepartRoleMapper.xml：
//
//	SELECT * FROM sys_depart_role WHERE depart_id IN (
//	  SELECT id FROM sys_depart
//	  WHERE id IN (SELECT dep_id FROM sys_user_depart WHERE user_id=?)
//	    AND org_code LIKE CONCAT(?,'%'))
//
// departId 无效时 Java getById(...).getOrgCode() NPE → 全局「操作失败」，
// Go 显式返回 ErrOpFailed。
func (s *Service) GetDeptRoleList(ctx context.Context, departID, userID string) ([]entity.SysDepartRole, error) {
	dCols := dao.SysDepart.Columns()
	var depart entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).Where(dCols.Id, departID).Scan(&depart); err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	if depart.Id == "" {
		return nil, ErrOpFailed
	}
	udCols := dao.SysUserDepart.Columns()
	var userDepIDs []string
	if err := dao.SysUserDepart.Ctx(ctx).
		Fields(udCols.DepId).
		Where(udCols.UserId, userID).
		Scan(&userDepIDs); err != nil {
		return nil, fmt.Errorf("查询用户部门失败: %w", err)
	}
	if len(userDepIDs) == 0 {
		return []entity.SysDepartRole{}, nil
	}
	// 用户所属部门 ∩ org_code 前缀匹配的部门（与 Java IN+LIKE 组合等价）
	ids, err := g.DB().Model("sys_depart").
		Fields("id").
		WhereIn("id", userDepIDs).
		WhereLike("org_code", depart.OrgCode+"%").
		Array()
	if err != nil {
		return nil, fmt.Errorf("查询部门范围失败: %w", err)
	}
	if len(ids) == 0 {
		return []entity.SysDepartRole{}, nil
	}
	cols := dao.SysDepartRole.Columns()
	var rows []entity.SysDepartRole
	if err := dao.SysDepartRole.Ctx(ctx).
		WhereIn(cols.DepartId, ids).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门角色失败: %w", err)
	}
	return rows, nil
}

// DeptRoleUserAdd 保存用户与部门角色的绑定（diff 增删）。
//
// 契约蓝本 SysDepartRoleController#deptRoleUserAdd（POST body
// {newRoleId, oldRoleId, userId}）+ SysDepartRoleUserServiceImpl#deptRoleUserAdd：
//   - 新增 = 在新不在旧 → 插入 (user_id, drole_id)；删除 = 在旧不在新 →
//     按 (user_id, drole_id) 删除；
//   - Java 无事务（半途失败留脏数据）→ Go 包单事务（文档化改进）。
func (s *Service) DeptRoleUserAdd(ctx context.Context, userID, newRoleIDs, oldRoleIDs string) error {
	addList := diffIDs(oldRoleIDs, newRoleIDs)
	delList := diffIDs(newRoleIDs, oldRoleIDs)
	if len(addList) == 0 && len(delList) == 0 {
		return nil
	}
	return dao.SysDepartRoleUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.SysDepartRoleUser.Columns()
		for _, roleID := range addList {
			if roleID == "" {
				continue
			}
			if _, err := dao.SysDepartRoleUser.Ctx(ctx).Data(do.SysDepartRoleUser{
				Id:      guid.S(),
				UserId:  userID,
				DroleId: roleID,
			}).Insert(); err != nil {
				return fmt.Errorf("保存部门角色用户失败: %w", err)
			}
		}
		for _, roleID := range delList {
			if roleID == "" {
				continue
			}
			if _, err := dao.SysDepartRoleUser.Ctx(ctx).
				Where(cols.UserId, userID).
				Where(cols.DroleId, roleID).
				Delete(); err != nil {
				return fmt.Errorf("删除部门角色用户失败: %w", err)
			}
		}
		return nil
	})
}

// GetDeptRoleByUserID 查询用户的部门角色绑定列表。
//
// 契约蓝本 SysDepartRoleController#getDeptRoleByUserId（GET ?userId=）：
// sys_depart_role_user WHERE user_id 全查（无排序）。
func (s *Service) GetDeptRoleByUserID(ctx context.Context, userID string) ([]entity.SysDepartRoleUser, error) {
	cols := dao.SysDepartRoleUser.Columns()
	var rows []entity.SysDepartRoleUser
	if err := dao.SysDepartRoleUser.Ctx(ctx).
		Where(cols.UserId, userID).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门角色用户失败: %w", err)
	}
	return rows, nil
}

// DataruleList 部门角色的菜单数据规则与已勾选项。
//
// 契约蓝本 SysDepartRoleController#datarule GET + SysDepartPermissionServiceImpl#
// getPermRuleListByDeptIdAndPermId：
//   - 规则来源 = 部门已授权的 data_rule_ids（sys_depart_permission 行存在才有值，
//     行不存在 → null →「未找到权限配置信息」；空串 → IN(”) 空列表 → 同样报错）；
//   - 规则按 create_time 倒序；drChecked 来自 sys_depart_role_permission
//     (permission_id, role_id) 的 data_rule_ids，非空时去末尾一个逗号。
func (s *Service) DataruleList(ctx context.Context, permissionID, departID, roleID string) (*DataruleResult, error) {
	dpCols := dao.SysDepartPermission.Columns()
	var dp entity.SysDepartPermission
	if err := dao.SysDepartPermission.Ctx(ctx).
		Where(dpCols.DepartId, departID).
		Where(dpCols.PermissionId, permissionID).
		Scan(&dp); err != nil {
		return nil, fmt.Errorf("查询部门授权失败: %w", err)
	}
	if dp.Id == "" {
		return nil, ErrNoPermConfig
	}
	if dp.DataRuleIds == "" {
		// Java split("") → [""] → IN('') 空列表 → 控制器判空报错
		return nil, ErrNoPermConfig
	}
	drCols := dao.SysPermissionDataRule.Columns()
	var rules []entity.SysPermissionDataRule
	if err := dao.SysPermissionDataRule.Ctx(ctx).
		WhereIn(drCols.Id, strings.Split(dp.DataRuleIds, ",")).
		OrderDesc(drCols.CreateTime).
		Scan(&rules); err != nil {
		return nil, fmt.Errorf("查询数据规则失败: %w", err)
	}
	if len(rules) == 0 {
		return nil, ErrNoPermConfig
	}
	res := &DataruleResult{Rules: rules}
	rpCols := dao.SysDepartRolePermission.Columns()
	var rp entity.SysDepartRolePermission
	if err := dao.SysDepartRolePermission.Ctx(ctx).
		Where(rpCols.PermissionId, permissionID).
		Where(rpCols.RoleId, roleID).
		Scan(&rp); err != nil {
		return nil, fmt.Errorf("查询部门角色授权失败: %w", err)
	}
	if rp.Id != "" && rp.DataRuleIds != "" {
		res.DrChecked = strings.TrimSuffix(rp.DataRuleIds, ",")
		res.HasChecked = true
	}
	return res, nil
}

// DataruleSave 保存部门角色的菜单数据规则勾选。
//
// 契约蓝本 SysDepartRoleController#datarule POST（body {permissionId, roleId,
// dataRuleIds}）：sys_depart_role_permission 必须已有 (permission_id, role_id)
// 行（update-only）；dataRuleIds 整段覆盖（nil → 值不变）；多行命中 →
// 「保存失败」（Java getOne throwEx）。
func (s *Service) DataruleSave(ctx context.Context, permissionID, roleID string, dataRuleIDs *string) error {
	rpCols := dao.SysDepartRolePermission.Columns()
	var rows []entity.SysDepartRolePermission
	if err := dao.SysDepartRolePermission.Ctx(ctx).
		Where(rpCols.PermissionId, permissionID).
		Where(rpCols.RoleId, roleID).
		Scan(&rows); err != nil {
		return fmt.Errorf("查询部门角色授权失败: %w", err)
	}
	if len(rows) == 0 {
		return ErrSaveRolePermFirst
	}
	if len(rows) > 1 {
		return ErrSaveFailed
	}
	if dataRuleIDs == nil {
		return nil
	}
	if _, err := dao.SysDepartRolePermission.Ctx(ctx).
		Where(rpCols.Id, rows[0].Id).
		Data(do.SysDepartRolePermission{DataRuleIds: *dataRuleIDs}).Update(); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}
	return nil
}

// diffIDs 纯函数（与 permission 域 diffPermIDs 同语义）：取 diff 中 main 没有的元素。
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

// applyDepartRoleInput 请求体字段拷贝进 DO。
func applyDepartRoleInput(data *do.SysDepartRole, in DepartRoleInput) {
	if in.DepartId != nil {
		data.DepartId = *in.DepartId
	}
	if in.RoleName != nil {
		data.RoleName = *in.RoleName
	}
	if in.RoleCode != nil {
		data.RoleCode = *in.RoleCode
	}
	if in.Description != nil {
		data.Description = *in.Description
	}
}

// applyQueryRule 按 QueryGenerator 值规则应用过滤条件（与 role 域同款实现）。
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

// parseQueryRule QueryGenerator 值规则纯判定（与 role 域同款）：
// * → LIKE（* 换 %）、, → IN、!= 前缀 → NE、其余 EQ。
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
