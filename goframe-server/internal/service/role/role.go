// Package role 承载系统角色管理域业务用例
// （契约蓝本：SysRoleController，GET/POST/PUT/DELETE /sys/role/**）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - sys_role 无 del_flag 列，删除为物理级联（sys_user_role → sys_role_permission
//     → sys_role），与 P0-2/P0-3 已有的 sys_user_role 读写端点构成表级耦合；
//   - add/edit/delete/deleteBatch 在 Java 带 @RequiresRoles({"admin"})，违规经
//     JeecgBootExceptionHandler 返回 code=510「没有权限，请联系管理员授权」；
//   - Java 控制器不写 create_by/update_by（仓库无 MetaObjectHandler 自动填充），
//     Go 版保持一致，仅服务端写 create_time/update_time；
//   - 列表查询对齐 QueryGenerator 的实际消费面：字符串字段 LIKE、roleLevel 精确、
//     createTime _begin/_end 区间、column+order 排序（列白名单，无默认排序）。
//
// 全部数据库操作走 dao/do 占位符绑定；写路径 id 生成沿用仓库 guid.S() 先例。
package role

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrNoAuth 非管理员执行管理操作（Java @RequiresRoles 违规 → Result.noauth，code 510）。
	ErrNoAuth = errors.New("没有权限，请联系管理员授权")
	// ErrNotFound 实体不存在（Java queryById/edit 的 error500 原话术）。
	ErrNotFound = errors.New("未找到对应实体")
	// ErrNoRoleInfo 角色列表为空（Java queryall/queryMySubRole error500 原话术）。
	ErrNoRoleInfo = errors.New("未找到角色信息")
	// ErrRoleNotSelected 批量删除未传 ids（Java deleteBatch error500 原话术）。
	ErrRoleNotSelected = errors.New("未选中角色！")
	// ErrNoPermConfig 菜单下无数据规则（Java datarule GET error500 原话术）。
	ErrNoPermConfig = errors.New("未找到权限配置信息")
	// ErrSaveRolePermFirst 角色菜单授权关系缺失（Java datarule POST error500 原话术）。
	ErrSaveRolePermFirst = errors.New("请先保存角色菜单权限!")
	// ErrSaveFailed datarule POST 异常兜底（Java catch 原话术）。
	ErrSaveFailed = errors.New("保存失败")
)

// identRegexp 动态标识符白名单（排序列防注入，与 usercore 域同一正则）。
var identRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// orderColumns 排序列白名单 = sys_role 全量列（无敏感列，del_flag 本就不存在）。
var orderColumns = map[string]struct{}{
	"id":          {},
	"role_name":   {},
	"role_code":   {},
	"description": {},
	"create_by":   {},
	"create_time": {},
	"update_by":   {},
	"update_time": {},
	"role_level":  {},
}

// SaveInput add/edit 请求体。
//
// 指针字段对齐 MyBatis-Plus 默认 NOT_NULL 字段策略：JSON 未提供的字段不参与
// INSERT/UPDATE（DO 相应字段保持 nil 即自动跳过），避免把「未提供」误写成空串。
type SaveInput struct {
	Id          string
	RoleName    *string
	RoleCode    *string
	Description *string
	RoleLevel   *int
}

// ListFilter list 查询过滤条件（对齐 QueryGenerator 实际消费面）。
type ListFilter struct {
	RoleName        string // LIKE %v%
	RoleCode        string // LIKE %v%
	Description     string // LIKE %v%
	RoleLevel       *int   // 精确匹配
	CreateTimeBegin string // create_time >=（Java _begin 后缀规则 → GE）
	CreateTimeEnd   string // create_time <=（Java _end 后缀规则 → LE）
	OrderColumn     string // column 参数（驼峰，服务端转下划线并过白名单）
	OrderType       string // order 参数（asc/desc，大小写不敏感）
}

// treeModel 对齐 Java TreeModel 的前端消费字段
// （UserRoleModal/DepartAuthModal 的 ant Tree：key/title/children/isLeaf）。
// Java 侧 ruleFlag/scopedSlots 在该构造路径下恒为 null，此处有意省略（前端未消费）。
type treeModel struct {
	Key       string      `json:"key"`
	Title     string      `json:"title"`
	SlotTitle string      `json:"slotTitle"`
	Value     string      `json:"value"`
	Label     string      `json:"label"`
	ParentId  string      `json:"parentId"`
	Icon      string      `json:"icon"`
	IsLeaf    bool        `json:"isLeaf"`
	Children  []treeModel `json:"children"` // 叶子为 null（对齐 Java 未初始化字段）
}

// DataruleResult GET /sys/role/datarule/{permissionId}/{roleId} 的返回。
type DataruleResult struct {
	Rules      []entity.SysPermissionDataRule
	DrChecked  string
	HasChecked bool
}

// Service 角色管理业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// List 角色分页列表。
//
// 契约蓝本 SysRoleController#list（GET /sys/role/list）：
//   - QueryGenerator 动态条件 → Go 收敛为前端实际消费面（RoleList.vue：
//     roleName 模糊 + createTime_begin/end 区间；RoleUserList.vue 仅分页）；
//   - Java 无默认排序（无 ORDER BY），Go 保持一致，仅当 column/order 传入且
//     过白名单时排序；
//   - result：IPage {records,total,size,current,pages}，records 为 SysRole 全字段。
func (s *Service) List(ctx context.Context, param page.Param, f ListFilter) (*page.Result, error) {
	cols := dao.SysRole.Columns()
	m := dao.SysRole.Ctx(ctx)
	if f.RoleName != "" {
		m = m.WhereLike(cols.RoleName, "%"+f.RoleName+"%")
	}
	if f.RoleCode != "" {
		m = m.WhereLike(cols.RoleCode, "%"+f.RoleCode+"%")
	}
	if f.Description != "" {
		m = m.WhereLike(cols.Description, "%"+f.Description+"%")
	}
	if f.RoleLevel != nil {
		m = m.Where(cols.RoleLevel, *f.RoleLevel)
	}
	if f.CreateTimeBegin != "" {
		m = m.WhereGTE(cols.CreateTime, f.CreateTimeBegin)
	}
	if f.CreateTimeEnd != "" {
		m = m.WhereLTE(cols.CreateTime, f.CreateTimeEnd)
	}
	m = applyOrder(m, f.OrderColumn, f.OrderType)
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计角色数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysRole{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysRole
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增角色。
//
// 契约蓝本 SysRoleController#add（POST /sys/role/add，@RequiresRoles admin）：
//   - create_time 服务端写入；create_by 不写（Java 控制器未设置且无自动填充）；
//   - roleCode 唯一性由 DB 唯一键 uniq_sys_role_role_code 兜底，冲突返回
//     通用「操作失败」（Java catch → error500("操作失败")，不区分话术）；
//   - result 为 null（Java 仅 success("添加成功！")）。
func (s *Service) Add(ctx context.Context, operatorID string, in SaveInput) error {
	if err := s.requireAdmin(ctx, operatorID); err != nil {
		return err
	}
	data := do.SysRole{
		Id:         guid.S(),
		CreateTime: gtime.Now(),
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
	if in.RoleLevel != nil {
		data.RoleLevel = *in.RoleLevel
	}
	if _, err := dao.SysRole.Ctx(ctx).Data(data).Insert(); err != nil {
		// 唯一键冲突等数据库错误统一脱敏为 Java 同款话术
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑角色。
//
// 契约蓝本 SysRoleController#edit（PUT /sys/role/edit，@RequiresRoles admin）：
//   - 先 getById，不存在 → 「未找到对应实体」（HTTP 200 + success:false）；
//   - MyBatis-Plus updateById 仅更新非 null 字段 → DO 未赋值字段自动跳过；
//   - update_time 服务端写入；updateById 返回 false 时 Java 不改写响应
//     （仍 success:true），Go 版 Update 零行受影响同样不视为错误。
func (s *Service) Edit(ctx context.Context, operatorID string, in SaveInput) error {
	if err := s.requireAdmin(ctx, operatorID); err != nil {
		return err
	}
	cols := dao.SysRole.Columns()
	var old entity.SysRole
	if err := dao.SysRole.Ctx(ctx).Where(cols.Id, in.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询角色失败: %w", err)
	}
	if old.Id == "" {
		return ErrNotFound
	}
	data := do.SysRole{UpdateTime: gtime.Now()}
	if in.RoleName != nil {
		data.RoleName = *in.RoleName
	}
	if in.RoleCode != nil {
		data.RoleCode = *in.RoleCode
	}
	if in.Description != nil {
		data.Description = *in.Description
	}
	if in.RoleLevel != nil {
		data.RoleLevel = *in.RoleLevel
	}
	if _, err := dao.SysRole.Ctx(ctx).Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除角色（物理级联）。
//
// 契约蓝本 SysRoleController#delete（DELETE /sys/role/delete?id=，@RequiresRoles admin）
// + SysRoleServiceImpl#deleteRole（@Transactional）：
//   - 固定顺序三级联：sys_user_role(role_id) → sys_role_permission(role_id)
//     → sys_role(id)，全部物理 DELETE（sys_role 无 del_flag）；
//   - 角色不存在同样返回成功（Java removeById 布尔返回值被忽略）；
//   - 级联清空 sys_user_role 会连带影响 P0-3 的角色分配端点数据（表级耦合，
//     语义与 Java 一致，无需额外处理）。
func (s *Service) Delete(ctx context.Context, operatorID, id string) error {
	if err := s.requireAdmin(ctx, operatorID); err != nil {
		return err
	}
	return s.cascadeDelete(ctx, []string{id})
}

// DeleteBatch 批量删除角色（物理级联）。
//
// 契约蓝本 SysRoleController#deleteBatch（DELETE /sys/role/deleteBatch?ids=逗号分隔，
// @RequiresRoles admin）：空 ids → 「未选中角色！」；否则同单删级联，
// Java 批删走 IN 数组 SQL，Go 用 WhereIn 等价实现，单事务包裹。
func (s *Service) DeleteBatch(ctx context.Context, operatorID string, ids []string) error {
	if err := s.requireAdmin(ctx, operatorID); err != nil {
		return err
	}
	if len(ids) == 0 {
		return ErrRoleNotSelected
	}
	return s.cascadeDelete(ctx, ids)
}

// cascadeDelete 三表级联物理删除（单删/批删共用，单事务，顺序对齐 Java ServiceImpl）。
func (s *Service) cascadeDelete(ctx context.Context, roleIDs []string) error {
	return dao.SysRole.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		urCols := dao.SysUserRole.Columns()
		if _, err := dao.SysUserRole.Ctx(ctx).
			WhereIn(urCols.RoleId, roleIDs).Delete(); err != nil {
			return fmt.Errorf("级联删除用户角色失败: %w", err)
		}
		rpCols := dao.SysRolePermission.Columns()
		if _, err := dao.SysRolePermission.Ctx(ctx).
			WhereIn(rpCols.RoleId, roleIDs).Delete(); err != nil {
			return fmt.Errorf("级联删除角色权限失败: %w", err)
		}
		cols := dao.SysRole.Columns()
		if _, err := dao.SysRole.Ctx(ctx).
			WhereIn(cols.Id, roleIDs).Delete(); err != nil {
			return fmt.Errorf("删除角色失败: %w", err)
		}
		return nil
	})
}

// QueryById 按 ID 查询角色详情。
//
// 契约蓝本 SysRoleController#queryById（GET /sys/role/queryById?id=）：
// 不存在 → 「未找到对应实体」（HTTP 200 + success:false）。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.SysRole, error) {
	var row entity.SysRole
	if err := dao.SysRole.Ctx(ctx).Where(dao.SysRole.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// QueryAll 全量角色（下拉选择用）。
//
// 契约蓝本 SysRoleController#queryall（GET /sys/role/queryall）：
// 全表 list()，无过滤无排序；空列表 → 「未找到角色信息」。
func (s *Service) QueryAll(ctx context.Context) ([]entity.SysRole, error) {
	var rows []entity.SysRole
	if err := dao.SysRole.Ctx(ctx).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrNoRoleInfo
	}
	return rows, nil
}

// QueryMySubRole 查询当前用户等级之下的角色（角色授权下拉用）。
//
// 契约蓝本 SysRoleController#queryMySubRole（GET /sys/role/queryMySubRole）：
//   - 操作人具备 admin 或 dev 角色（Shiro hasRole）→ 返回全部角色；
//   - 否则 role_level < 我的角色等级（getUserRoleLevel = max(role_level)，
//     无角色时为 -1；严格小于，等级相等看不到任何角色）；
//   - 空列表 → 「未找到角色信息」。
func (s *Service) QueryMySubRole(ctx context.Context, operatorID string) ([]entity.SysRole, error) {
	privileged, err := s.hasAnyRoleCode(ctx, operatorID, "admin", "dev")
	if err != nil {
		return nil, err
	}
	cols := dao.SysRole.Columns()
	m := dao.SysRole.Ctx(ctx)
	if !privileged {
		level, err := s.userRoleLevel(ctx, operatorID)
		if err != nil {
			return nil, err
		}
		m = m.WhereLT(cols.RoleLevel, level)
	}
	var rows []entity.SysRole
	if err := m.Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrNoRoleInfo
	}
	return rows, nil
}

// CheckRoleCode 角色编码唯一性校验。
//
// 契约蓝本 SysRoleController#checkRoleCode（GET /sys/role/checkRoleCode）：
//   - 返回 (conflict, error)：conflict=true 表示编码已被占用；
//   - Java 语义：result 字段恒为 true（布尔语义在 success+message 上），新增模式
//     下编码已存在即冲突；编辑模式下查得记录 id 与传入 id 不一致才冲突；
//   - roleCode 为空：Java eq role_code=NULL 永不命中 → 无冲突，此处一致；
//   - 调用方（RoleModal 的表单校验器）只消费 success/message，不读 result。
func (s *Service) CheckRoleCode(ctx context.Context, id, roleCode string) (bool, error) {
	if roleCode == "" {
		return false, nil
	}
	cols := dao.SysRole.Columns()
	var byCode entity.SysRole
	if err := dao.SysRole.Ctx(ctx).Where(cols.RoleCode, roleCode).Scan(&byCode); err != nil {
		return false, fmt.Errorf("查询角色失败: %w", err)
	}
	if byCode.Id == "" {
		return false, nil
	}
	return codeConflict(id, byCode.Id), nil
}

// codeConflict 编码冲突纯判定（便于单测）：
// 编码未被占用（foundID 为空）→ 无冲突；新增模式（id 为空）只要编码存在即冲突；
// 编辑模式仅当编码属于其他角色才冲突。
func codeConflict(id, foundID string) bool {
	if foundID == "" {
		return false
	}
	if id == "" {
		return true
	}
	return id != foundID
}

// QueryPermissionTreeList 菜单权限树（角色授权弹窗数据源）。
//
// 契约蓝本 SysRoleController#queryTreeList（GET /sys/role/queryTreeList）：
//   - sys_permission WHERE del_flag=0 ORDER BY sort_no ASC，内存拼树（parentId
//     空为根，非叶节点递归挂子）；
//   - result：{treeList:[TreeModel], ids:[全部权限 id]}（ids 供 ant Tree 默认勾选）。
func (s *Service) QueryPermissionTreeList(ctx context.Context) ([]treeModel, []string, error) {
	cols := dao.SysPermission.Columns()
	var perms []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&perms); err != nil {
		return nil, nil, fmt.Errorf("查询菜单权限失败: %w", err)
	}
	ids := make([]string, 0, len(perms))
	for _, p := range perms {
		ids = append(ids, p.Id)
	}
	return buildTreeModelTree(perms), ids, nil
}

// buildTreeModelTree 由扁平菜单列表构建 TreeModel 树（对齐 Java getTreeModelList 递归）。
func buildTreeModelTree(perms []entity.SysPermission) []treeModel {
	var out []treeModel
	appendTreeModel(&out, perms, nil)
	return out
}

// appendTreeModel 把 parent 的子节点追加进 dest（parent 为 nil 时收集根节点）。
func appendTreeModel(dest *[]treeModel, perms []entity.SysPermission, parent *treeModel) {
	for i := range perms {
		p := perms[i]
		node := newTreeModel(p)
		switch {
		case parent == nil && p.ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !last.IsLeaf {
				appendTreeModel(&last.Children, perms, last)
			}
		case parent != nil && p.ParentId == parent.Key:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !last.IsLeaf {
				appendTreeModel(&last.Children, perms, last)
			}
		}
	}
}

// newTreeModel 由 SysPermission 构造树节点（对齐 Java TreeModel(SysPermission) 构造器）。
func newTreeModel(p entity.SysPermission) treeModel {
	node := treeModel{
		Key:       p.Id,
		Title:     p.Name,
		SlotTitle: p.Name,
		Value:     p.Id,
		Label:     p.Name,
		ParentId:  p.ParentId,
		Icon:      p.Icon,
		IsLeaf:    p.IsLeaf == 1,
	}
	if !node.IsLeaf {
		node.Children = []treeModel{}
	}
	return node
}

// DataruleList 查询角色的菜单数据规则与已勾选项。
//
// 契约蓝本 SysRoleController#loadDatarule（GET /sys/role/datarule/{permissionId}/{roleId}）：
//   - datarule = sys_permission_data_rule WHERE permission_id，按 create_time 倒序
//     （SysPermissionDataRuleImpl#getPermRuleListByPermId）；列表为空 →
//     「未找到权限配置信息」（即使角色已配置，前端依赖该短路行为）；
//   - drChecked 仅当 sys_role_permission 存在 (permission_id, role_id) 且
//     data_rule_ids 非空时出现，并去掉末尾一个逗号（Java 原样 substring）。
func (s *Service) DataruleList(ctx context.Context, permissionID, roleID string) (*DataruleResult, error) {
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
	rpCols := dao.SysRolePermission.Columns()
	var rp entity.SysRolePermission
	if err := dao.SysRolePermission.Ctx(ctx).
		Where(rpCols.PermissionId, permissionID).
		Where(rpCols.RoleId, roleID).
		WhereNotNull(rpCols.DataRuleIds).
		Scan(&rp); err != nil {
		return nil, fmt.Errorf("查询角色菜单配置失败: %w", err)
	}
	if rp.Id != "" && rp.DataRuleIds != "" {
		res.DrChecked, res.HasChecked = normalizeDrChecked(rp.DataRuleIds)
	}
	return res, nil
}

// normalizeDrChecked drChecked 纯判定（便于单测）：非空时去掉末尾一个逗号
// （对齐 Java oConvertUtils.isNotEmpty + endsWith 判断 + substring 原语义）。
func normalizeDrChecked(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	return strings.TrimSuffix(raw, ","), true
}

// DataruleSave 保存角色的菜单数据规则勾选。
//
// 契约蓝本 SysRoleController#saveDatarule（POST /sys/role/datarule）：
//   - sys_role_permission 必须已有 (permission_id, role_id) 行，否则
//     「请先保存角色菜单权限!」（update-only，不会 insert）；
//   - dataRuleIds 原串整段覆盖写入（尾逗号由前端负责，服务端不 trim）；
//     请求未携带该字段（nil）时 Java setDataRuleIds(null) 被 MP 字段策略跳过
//     → 值不变，Go 版以 no-op 等价实现；
//   - (permission_id, role_id) 命中多行时 Java getOne 抛异常 → 「保存失败」，
//     Go 版显式判定，行为一致。
func (s *Service) DataruleSave(ctx context.Context, permissionID, roleID string, dataRuleIDs *string) error {
	rpCols := dao.SysRolePermission.Columns()
	var rows []entity.SysRolePermission
	if err := dao.SysRolePermission.Ctx(ctx).
		Where(rpCols.PermissionId, permissionID).
		Where(rpCols.RoleId, roleID).
		Scan(&rows); err != nil {
		return fmt.Errorf("查询角色菜单配置失败: %w", err)
	}
	if len(rows) == 0 {
		return ErrSaveRolePermFirst
	}
	if len(rows) > 1 {
		// Java getOne 默认 throwEx=true：多行同样进 catch 返回「保存失败」
		return ErrSaveFailed
	}
	if dataRuleIDs == nil {
		return nil
	}
	if _, err := dao.SysRolePermission.Ctx(ctx).
		Where(rpCols.Id, rows[0].Id).
		Data(do.SysRolePermission{DataRuleIds: *dataRuleIDs}).Update(); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}
	return nil
}

// requireAdmin 校验操作人具备 admin 角色（对齐 Java @RequiresRoles({"admin"})）。
// 操作人上下文缺失视为无权限（fail-closed）。
func (s *Service) requireAdmin(ctx context.Context, operatorID string) error {
	ok, err := s.hasAnyRoleCode(ctx, operatorID, "admin")
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoAuth
	}
	return nil
}

// hasAnyRoleCode 判断用户是否拥有任一指定角色编码。
func (s *Service) hasAnyRoleCode(ctx context.Context, userID string, roleCodes ...string) (bool, error) {
	if userID == "" || len(roleCodes) == 0 {
		return false, nil
	}
	rec, err := g.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Where("ur.user_id", userID).
		WhereIn("r.role_code", roleCodes).
		Limit(1).
		One()
	if err != nil {
		return false, fmt.Errorf("查询用户角色失败: %w", err)
	}
	return !rec.IsEmpty(), nil
}

// userRoleLevel 查询用户角色等级（max(role_level)，无角色为 -1）。
//
// 与 userrole 域的私有实现同语义（SysUserRoleMapper#getUserRoleLevel），
// 各域自持副本是本仓库既定模式（跨表聚合走裸模型 + 占位符）。
func (s *Service) userRoleLevel(ctx context.Context, userID string) (int, error) {
	rec, err := g.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Fields("MAX(r.role_level) AS role_level").
		Where("ur.user_id", userID).
		One()
	if err != nil {
		return -1, fmt.Errorf("查询用户角色等级失败: %w", err)
	}
	if rec.IsEmpty() || rec["role_level"].IsNil() {
		return -1, nil
	}
	return rec["role_level"].Int(), nil
}

// applyOrder 排序参数白名单应用（对齐 usercore 域同款实现与 QueryGenerator 行为）：
// 剔除 _dictText/_url 后缀 → 驼峰转下划线 → 正则 + 白名单双重校验；
// 未通过任何校验或 orderType 为空时保持无排序（对齐 Java 无默认 ORDER BY）。
func applyOrder(m *gdb.Model, column, orderType string) *gdb.Model {
	col, desc, ok := normalizeOrder(column, orderType)
	if !ok {
		return m
	}
	if desc {
		return m.OrderDesc(col)
	}
	return m.OrderAsc(col)
}

// normalizeOrder 排序参数纯判定（便于单测）：返回 (下划线列名, 是否降序, 是否生效)。
func normalizeOrder(column, orderType string) (string, bool, bool) {
	column = strings.TrimSpace(column)
	if column == "" || strings.TrimSpace(orderType) == "" {
		return "", false, false
	}
	column = strings.TrimSuffix(column, "_dictText")
	column = strings.TrimSuffix(column, "_url")
	col := camelToUnderline(column)
	if !identRegexp.MatchString(col) {
		return "", false, false
	}
	if _, ok := orderColumns[col]; !ok {
		return "", false, false
	}
	return col, !strings.Contains(strings.ToUpper(orderType), "ASC"), true
}

// camelToUnderline 驼峰转下划线（同 Java oConvertUtils#isCamelCase 转换语义）。
func camelToUnderline(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
