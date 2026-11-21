// menu.go 承载菜单权限 CRUD / 角色授权 / 数据规则 / 部门授权业务用例
// （契约蓝本：SysPermissionController 的 19 个端点；P0-2 已实现的
// getUserPermissionByToken 不在本文件范围）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - sys_permission 的删除为物理级联（entity 无 @TableLogic，del_flag 仅作查询过滤）：
//     递归删子菜单 + 每个节点的数据规则/角色授权/部门授权/部门角色授权；
//   - is_leaf 维护：add 置父非叶、edit 按父子变化双向修正（setMenuLeaf 注解 SQL）；
//   - 角色授权 saveRolePermission 为 diff 算法（增删交集差集），Java 无事务，
//     Go 包单事务（文档化改进，HTTP 契约一致）；
//   - Java V2.8 的 queryListAsync SQL 引用不存在的 leaf_flag 列，运行必然报错
//     被 catch 吞掉（前端无调用方）。Go 用 is_leaf 修正实现并文档化偏差。
package permission

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrMenuNotFound 菜单不存在（Java deletePermission/editPermission
	// JeecgBootException 原话术）。
	ErrMenuNotFound = errors.New("未找到菜单信息")
	// ErrNoPermInfo 异步子菜单为空（Java queryAsync error500 原话术，
	// 「角色信息」为 Java 源码原文，虽语义怪异仍保持一致）。
	ErrNoPermInfo = errors.New("未找到角色信息")
)

// MenuInput add/edit 请求体（SysPermission 全量可编辑字段）。
//
// 指针字段对齐 MyBatis-Plus 默认 NOT_NULL 字段策略：JSON 未提供的字段不参与
// INSERT/UPDATE（DO 相应字段保持 nil 即自动跳过）。
type MenuInput struct {
	Id                 string
	ParentId           *string
	Name               *string
	Url                *string
	Component          *string
	ComponentName      *string
	Redirect           *string
	MenuType           *int
	Perms              *string
	PermsType          *string
	SortNo             *float64
	AlwaysShow         *int
	Icon               *string
	IsRoute            *int
	KeepAlive          *int
	Hidden             *int
	Description        *string
	InternalOrExternal *int
	Status             *string
}

// DataRuleInput 数据规则 add/edit 请求体（SysPermissionDataRule 可编辑字段）。
type DataRuleInput struct {
	Id             string
	PermissionId   *string
	RuleName       *string
	RuleColumn     *string
	RuleConditions *string
	RuleValue      *string
	Status         *string
}

// permTreeModel 对齐 Java SysPermissionTree 的前端消费字段（菜单管理页表格/树）。
// isLeaf/route/keepAlive/alwaysShow/hidden/internalOrExternal 为 Java boolean
// getter 序列化名；叶子节点 children 为 null（对齐 Java 未初始化字段）。
type permTreeModel struct {
	Id                 string          `json:"id"`
	Key                string          `json:"key"`
	Title              string          `json:"title"`
	ParentId           string          `json:"parentId"`
	Name               string          `json:"name"`
	Perms              string          `json:"perms"`
	PermsType          string          `json:"permsType"`
	Icon               string          `json:"icon"`
	Component          string          `json:"component"`
	Url                string          `json:"url"`
	Redirect           string          `json:"redirect"`
	SortNo             float64         `json:"sortNo"`
	MenuType           int             `json:"menuType"`
	IsLeaf             bool            `json:"isLeaf"`
	Route              bool            `json:"route"`
	KeepAlive          bool            `json:"keepAlive"`
	AlwaysShow         bool            `json:"alwaysShow"`
	Hidden             bool            `json:"hidden"`
	InternalOrExternal bool            `json:"internalOrExternal"`
	Status             string          `json:"status"`
	Description        string          `json:"description"`
	DelFlag            int             `json:"delFlag"`
	CreateBy           string          `json:"createBy"`
	CreateTime         string          `json:"createTime"`
	UpdateBy           string          `json:"updateBy"`
	UpdateTime         string          `json:"updateTime"`
	Children           []permTreeModel `json:"children"`
}

// asyncTreeModel 对齐 Java TreeModel（queryListAsync 路径）：
// key=id、title=name、icon、parentId、isLeaf=is_leaf。
type asyncTreeModel struct {
	Key      string           `json:"key"`
	Title    string           `json:"title"`
	Icon     string           `json:"icon"`
	ParentId string           `json:"parentId"`
	IsLeaf   bool             `json:"isLeaf"`
	Children []asyncTreeModel `json:"children"`
}

// menuTreeList 菜单权限树（角色/部门授权弹窗数据源），对齐 Java TreeModel
// 构造器输出（key/title/slotTitle/value/label/parentId/icon/isLeaf/children）。
type menuTreeList struct {
	Key       string         `json:"key"`
	Title     string         `json:"title"`
	SlotTitle string         `json:"slotTitle"`
	Value     string         `json:"value"`
	Label     string         `json:"label"`
	ParentId  string         `json:"parentId"`
	Icon      string         `json:"icon"`
	IsLeaf    bool           `json:"isLeaf"`
	Children  []menuTreeList `json:"children"`
}

// MenuTreeList 菜单管理页全量树。
//
// 契约蓝本 SysPermissionController#list（GET /sys/permission/list）：
// sys_permission WHERE del_flag=0 ORDER BY sort_no ASC，内存拼树
// （parentId 空为根，非叶节点递归挂子）。
func (s *Service) MenuTreeList(ctx context.Context) ([]permTreeModel, error) {
	rows, err := s.listPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(rows), nil
}

// SystemMenuList 一级菜单列表（菜单管理页懒加载根层）。
//
// 契约蓝本 SysPermissionController#getSystemMenuList（GET /sys/permission/getSystemMenuList）：
// menuType=0 + del_flag=0，sort_no ASC，扁平列表（节点 children 按叶子形态初始化）。
func (s *Service) SystemMenuList(ctx context.Context) ([]permTreeModel, error) {
	cols := dao.SysPermission.Columns()
	var rows []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Where(cols.MenuType, 0).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询一级菜单失败: %w", err)
	}
	out := make([]permTreeModel, 0, len(rows))
	for _, p := range rows {
		out = append(out, newPermTreeModel(p))
	}
	return out, nil
}

// SystemSubmenu 查询子菜单（懒加载展开）。
//
// 契约蓝本 SysPermissionController#getSystemSubmenu（GET /sys/permission/getSystemSubmenu?parentId=）：
// parent_id=? + del_flag=0，sort_no ASC，扁平列表。
func (s *Service) SystemSubmenu(ctx context.Context, parentID string) ([]permTreeModel, error) {
	cols := dao.SysPermission.Columns()
	var rows []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Where(cols.ParentId, parentID).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询子菜单失败: %w", err)
	}
	out := make([]permTreeModel, 0, len(rows))
	for _, p := range rows {
		out = append(out, newPermTreeModel(p))
	}
	return out, nil
}

// SystemSubmenuBatch 批量查询子菜单（展开多行时复用）。
//
// 契约蓝本 SysPermissionController#getSystemSubmenuBatch（GET /sys/permission/getSystemSubmenuBatch?parentIds=逗号分隔）：
// parent_id IN + del_flag=0 + sort_no ASC，result 为 {parentId: [子菜单]} 映射。
func (s *Service) SystemSubmenuBatch(ctx context.Context, parentIDs []string) (map[string][]permTreeModel, error) {
	if len(parentIDs) == 0 {
		return map[string][]permTreeModel{}, nil
	}
	cols := dao.SysPermission.Columns()
	var rows []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		WhereIn(cols.ParentId, parentIDs).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("批量查询子菜单失败: %w", err)
	}
	out := make(map[string][]permTreeModel)
	for _, p := range rows {
		out[p.ParentId] = append(out[p.ParentId], newPermTreeModel(p))
	}
	return out, nil
}

// AddMenu 添加菜单。
//
// 契约蓝本 SysRoleController#add + SysPermissionServiceImpl#addPermission：
//   - intelligentProcessData 预处理（组件/URL 规整 + 一级菜单默认组件）；
//   - menuType=0 → 父 ID 不写入（Java setParentId(null) 被 MP 字段策略跳过）；
//   - 有父 ID → 父节点置非叶（is_leaf=0）；create_time 服务端写入；
//     del_flag=0、is_leaf=1 固定；id 为空时服务端生成（Java ID_WORKER_STR）。
func (s *Service) AddMenu(ctx context.Context, in MenuInput) error {
	in = intelligentProcessData(in)
	data := do.SysPermission{
		Id:         in.Id,
		CreateTime: gtime.Now(),
		DelFlag:    0,
		IsLeaf:     1,
	}
	if in.Id == "" {
		data.Id = guid.S()
	}
	// 一级菜单：Java setParentId(null) → MP 跳过该列（落 DB 默认值），此处不写即等价
	if in.MenuType == nil || *in.MenuType != 0 {
		if in.ParentId != nil {
			data.ParentId = *in.ParentId
		}
	}
	if in.ParentId != nil && *in.ParentId != "" && (in.MenuType == nil || *in.MenuType != 0) {
		if err := s.setMenuLeaf(ctx, *in.ParentId, 0); err != nil {
			return err
		}
	}
	applyMenuInput(&data, in)
	if _, err := dao.SysPermission.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// EditMenu 编辑菜单。
//
// 契约蓝本 SysRoleController#edit + SysPermissionServiceImpl#editPermission：
//   - 先查存在性（不存在 → 「未找到菜单信息」）；update_time 服务端写入；
//   - menuType=0 → parent_id 更新为空串（Java setParentId("")，空串参与 UPDATE）；
//   - 该节点无子节点 → is_leaf=1；
//   - 父节点变化时双向维护 is_leaf：新父置非叶；老父无其他子节点时置叶
//     （对齐 Java 条件 (pid 非空且变化) 或 (pid 空且老父非空)）。
func (s *Service) EditMenu(ctx context.Context, in MenuInput) error {
	if in.Id == "" {
		return ErrMenuNotFound
	}
	in = intelligentProcessData(in)
	cols := dao.SysPermission.Columns()
	var old entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).Where(cols.Id, in.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询菜单失败: %w", err)
	}
	if old.Id == "" {
		return ErrMenuNotFound
	}
	data := do.SysPermission{UpdateTime: gtime.Now()}
	if in.MenuType != nil && *in.MenuType == 0 {
		// Java setParentId("")：空串参与 UPDATE（与 add 的 null 语义不同）
		empty := ""
		data.ParentId = empty
	} else if in.ParentId != nil {
		data.ParentId = *in.ParentId
	}
	// 该节点无子节点 → 叶子（Java Step2，在 updateById 前统计）
	cnt, err := dao.SysPermission.Ctx(ctx).Where(cols.ParentId, in.Id).Count()
	if err != nil {
		return fmt.Errorf("统计子菜单失败: %w", err)
	}
	if cnt == 0 {
		data.IsLeaf = 1
	}
	applyMenuInput(&data, in)
	if _, err := dao.SysPermission.Ctx(ctx).Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	// 父节点变化检测（对齐 Java 条件）
	newPid := ""
	if in.ParentId != nil {
		newPid = *in.ParentId
	}
	if in.MenuType != nil && *in.MenuType == 0 {
		newPid = ""
	}
	parentChanged := (newPid != "" && newPid != old.ParentId) ||
		(newPid == "" && old.ParentId != "")
	if parentChanged {
		if newPid != "" {
			// a. 新父置非叶（newPid 为空串时 Java 的 UPDATE 不命中任何行，跳过等价）
			if err := s.setMenuLeaf(ctx, newPid, 0); err != nil {
				return err
			}
		}
		// b. 老父无其他子节点 → 置叶
		cc, err := dao.SysPermission.Ctx(ctx).Where(cols.ParentId, old.ParentId).Count()
		if err != nil {
			return fmt.Errorf("统计老父节点子菜单失败: %w", err)
		}
		if cc == 0 && old.ParentId != "" {
			if err := s.setMenuLeaf(ctx, old.ParentId, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteMenu 删除菜单（物理级联递归）。
//
// 契约蓝本 SysPermissionController#delete + SysPermissionServiceImpl#deletePermission
// （@Transactional）：
//   - 单事务内：置父叶子标记（如为唯一子）→ 物理删本节点 → 递归删子树并
//     逐节点清理四张关联表（数据规则/角色授权/部门授权/部门角色授权）→
//     清理本节点四张关联表；
//   - 不存在 → 「未找到菜单信息」（HTTP 200 + success:false）。
func (s *Service) DeleteMenu(ctx context.Context, id string) error {
	return dao.SysPermission.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.SysPermission.Columns()
		var p entity.SysPermission
		if err := dao.SysPermission.Ctx(ctx).Where(cols.Id, id).Scan(&p); err != nil {
			return fmt.Errorf("查询菜单失败: %w", err)
		}
		if p.Id == "" {
			return ErrMenuNotFound
		}
		// 若父节点仅剩本子节点，删除后父节点成为叶子
		if p.ParentId != "" {
			cnt, err := dao.SysPermission.Ctx(ctx).Where(cols.ParentId, p.ParentId).Count()
			if err != nil {
				return fmt.Errorf("统计兄弟节点失败: %w", err)
			}
			if cnt == 1 {
				if err := s.setMenuLeaf(ctx, p.ParentId, 1); err != nil {
					return err
				}
			}
		}
		if _, err := dao.SysPermission.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return fmt.Errorf("删除菜单失败: %w", err)
		}
		if err := s.removeMenuChildren(ctx, id); err != nil {
			return err
		}
		return s.deletePermissionRelations(ctx, id)
	})
}

// DeleteMenuBatch 批量删除菜单。
//
// 契约蓝本 SysRoleController#deleteBatch：逐个走单删级联，首个失败即中止。
// Java 异常分支返回 message=「删除成功!」的缺陷话术未照搬（文档化偏差），
// Go 返回通用「操作失败」。
func (s *Service) DeleteMenuBatch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := s.DeleteMenu(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// removeMenuChildren 递归删除子树（对齐 Java removeChildrenBy）：
// 先批量物理删子节点，再逐个清理子节点关联表，仍有下级则递归。
func (s *Service) removeMenuChildren(ctx context.Context, parentID string) error {
	cols := dao.SysPermission.Columns()
	var children []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Where(cols.ParentId, parentID).
		Scan(&children); err != nil {
		return fmt.Errorf("查询子菜单失败: %w", err)
	}
	if len(children) == 0 {
		return nil
	}
	if _, err := dao.SysPermission.Ctx(ctx).
		Where(cols.ParentId, parentID).Delete(); err != nil {
		return fmt.Errorf("删除子菜单失败: %w", err)
	}
	for _, child := range children {
		if err := s.deletePermissionRelations(ctx, child.Id); err != nil {
			return err
		}
		cnt, err := dao.SysPermission.Ctx(ctx).Where(cols.ParentId, child.Id).Count()
		if err != nil {
			return fmt.Errorf("统计下级菜单失败: %w", err)
		}
		if cnt > 0 {
			if err := s.removeMenuChildren(ctx, child.Id); err != nil {
				return err
			}
		}
	}
	return nil
}

// deletePermissionRelations 清理菜单节点的四张关联表（对齐 Java 顺序）。
func (s *Service) deletePermissionRelations(ctx context.Context, permissionID string) error {
	// 数据规则
	drCols := dao.SysPermissionDataRule.Columns()
	if _, err := dao.SysPermissionDataRule.Ctx(ctx).
		Where(drCols.PermissionId, permissionID).Delete(); err != nil {
		return fmt.Errorf("级联删除数据规则失败: %w", err)
	}
	// 角色授权
	rpCols := dao.SysRolePermission.Columns()
	if _, err := dao.SysRolePermission.Ctx(ctx).
		Where(rpCols.PermissionId, permissionID).Delete(); err != nil {
		return fmt.Errorf("级联删除角色授权失败: %w", err)
	}
	// 部门授权
	dpCols := dao.SysDepartPermission.Columns()
	if _, err := dao.SysDepartPermission.Ctx(ctx).
		Where(dpCols.PermissionId, permissionID).Delete(); err != nil {
		return fmt.Errorf("级联删除部门授权失败: %w", err)
	}
	// 部门角色授权（sys_depart_role_permission 与 sys_role_permission 同列名，
	// 经由裸模型访问，列名以 teachingopen2.8.sql L616 为准）
	if _, err := g.DB().Model("sys_depart_role_permission").
		Where("permission_id", permissionID).Delete(); err != nil {
		return fmt.Errorf("级联删除部门角色授权失败: %w", err)
	}
	return nil
}

// setMenuLeaf 维护 is_leaf 列（对齐 SysPermissionMapper#setMenuLeaf 注解 SQL：
// UPDATE sys_permission SET is_leaf = ? WHERE id = ?）。
func (s *Service) setMenuLeaf(ctx context.Context, id string, leaf int) error {
	if id == "" {
		return nil
	}
	if _, err := dao.SysPermission.Ctx(ctx).
		Where(dao.SysPermission.Columns().Id, id).
		Data(do.SysPermission{IsLeaf: leaf}).Update(); err != nil {
		return fmt.Errorf("维护菜单叶子标记失败: %w", err)
	}
	return nil
}

// listPermissions del_flag=0 + sort_no ASC 的全量菜单查询。
func (s *Service) listPermissions(ctx context.Context) ([]entity.SysPermission, error) {
	cols := dao.SysPermission.Columns()
	var rows []entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询菜单权限失败: %w", err)
	}
	return rows, nil
}

// applyMenuInput 把请求体的可编辑字段拷贝进 DO（ParentId 由 add/edit 按
// menuType 语义单独处理，不在此列；Id/CreateTime/UpdateTime 亦由调用方管理）。
func applyMenuInput(data *do.SysPermission, in MenuInput) {
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Url != nil {
		data.Url = *in.Url
	}
	if in.Component != nil {
		data.Component = *in.Component
	}
	if in.ComponentName != nil {
		data.ComponentName = *in.ComponentName
	}
	if in.Redirect != nil {
		data.Redirect = *in.Redirect
	}
	if in.MenuType != nil {
		data.MenuType = *in.MenuType
	}
	if in.Perms != nil {
		data.Perms = *in.Perms
	}
	if in.PermsType != nil {
		data.PermsType = *in.PermsType
	}
	if in.SortNo != nil {
		data.SortNo = *in.SortNo
	}
	if in.AlwaysShow != nil {
		data.AlwaysShow = *in.AlwaysShow
	}
	if in.Icon != nil {
		data.Icon = *in.Icon
	}
	if in.IsRoute != nil {
		data.IsRoute = *in.IsRoute
	}
	if in.KeepAlive != nil {
		data.KeepAlive = *in.KeepAlive
	}
	if in.Hidden != nil {
		data.Hidden = *in.Hidden
	}
	if in.Description != nil {
		data.Description = *in.Description
	}
	if in.InternalOrExternal != nil {
		data.InternalOrExternal = *in.InternalOrExternal
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
}

// intelligentProcessData 菜单数据预处理（对齐 PermissionDataUtil#intelligentProcessData）：
//   - component：去首个 "/" 字符 → replaceFirst("views/") → replaceFirst("src/views/")
//     → endsWith(".vue") 时全局去 ".vue"。Java 按此顺序逐条执行，含其固有怪癖：
//     "src/views/Foo" 会先被 "views/" 命中变成 "src/Foo"，逐语义保留不「修正」；
//   - url：endsWith(".vue") 时全局去 ".vue"；非 http 前缀、非 "/" 开头、
//     trim 后非 "{{" 开头 → 补 "/"（Java trim 判定用裁剪值、拼接用原值，保留）；
//   - menuType=0 且组件为空 → 默认 "layouts/RouteView"。
func intelligentProcessData(in MenuInput) MenuInput {
	if in.Component != nil && *in.Component != "" {
		c := *in.Component
		if strings.HasPrefix(c, "/") {
			c = c[1:]
		}
		c = replaceFirst(c, "views/", "")
		c = replaceFirst(c, "src/views/", "")
		if strings.HasSuffix(c, ".vue") {
			c = strings.ReplaceAll(c, ".vue", "")
		}
		in.Component = &c
	}
	if in.Url != nil && *in.Url != "" {
		u := *in.Url
		if strings.HasSuffix(u, ".vue") {
			u = strings.ReplaceAll(u, ".vue", "")
		}
		if !strings.HasPrefix(u, "http") && !strings.HasPrefix(u, "/") &&
			!strings.HasPrefix(strings.TrimSpace(u), "{{") {
			u = "/" + u
		}
		in.Url = &u
	}
	if in.MenuType != nil && *in.MenuType == 0 && (in.Component == nil || *in.Component == "") {
		def := "layouts/RouteView"
		in.Component = &def
	}
	return in
}

// replaceFirst 等价 Java String.replaceFirst 的字面量场景（首个命中替换一次）。
func replaceFirst(s, old, replacement string) string {
	idx := strings.Index(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + replacement + s[idx+len(old):]
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

// parseQueryRule QueryGenerator 值规则纯判定（与 role 域同款，便于单测）：
// 值含 * → LIKE（* 替换为 %）；值含 , → IN；前缀 != → NE；其余 EQ
// （QueryGenerator.java L163-168 注释掉了「字符串默认全模糊」，V2.8 生效语义即 EQ）。
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

// buildMenuTree 由扁平菜单列表构建 SysPermissionTree 树（对齐 Java getTreeList 递归）。
func buildMenuTree(perms []entity.SysPermission) []permTreeModel {
	var out []permTreeModel
	appendMenuTree(&out, perms, nil)
	return out
}

// appendMenuTree 把 parent 的子节点追加进 dest（parent 为 nil 时收集根节点）。
func appendMenuTree(dest *[]permTreeModel, perms []entity.SysPermission, parent *permTreeModel) {
	for i := range perms {
		p := perms[i]
		node := newPermTreeModel(p)
		switch {
		case parent == nil && p.ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !last.IsLeaf {
				appendMenuTree(&last.Children, perms, last)
			}
		case parent != nil && p.ParentId == parent.Id:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !last.IsLeaf {
				appendMenuTree(&last.Children, perms, last)
			}
		}
	}
}

// newPermTreeModel 由 SysPermission 构造树节点
// （对齐 Java SysPermissionTree(SysPermission) 构造器）。
func newPermTreeModel(p entity.SysPermission) permTreeModel {
	node := permTreeModel{
		Id:                 p.Id,
		Key:                p.Id,
		Title:              p.Name,
		ParentId:           p.ParentId,
		Name:               p.Name,
		Perms:              p.Perms,
		PermsType:          p.PermsType,
		Icon:               p.Icon,
		Component:          p.Component,
		Url:                p.Url,
		Redirect:           p.Redirect,
		SortNo:             p.SortNo,
		MenuType:           p.MenuType,
		IsLeaf:             p.IsLeaf == 1,
		Route:              p.IsRoute == 1,
		KeepAlive:          p.KeepAlive == 1,
		AlwaysShow:         p.AlwaysShow == 1,
		Hidden:             p.Hidden == 1,
		InternalOrExternal: p.InternalOrExternal == 1,
		Status:             p.Status,
		Description:        p.Description,
		DelFlag:            p.DelFlag,
		CreateBy:           p.CreateBy,
		CreateTime:         p.CreateTime,
		UpdateBy:           p.UpdateBy,
		UpdateTime:         p.UpdateTime,
	}
	if !node.IsLeaf {
		node.Children = []permTreeModel{}
	}
	return node
}

// QueryListAsync 异步加载子菜单。
//
// 契约蓝本 SysRoleController#queryAsync（GET /sys/permission/queryListAsync?pid=）：
//   - pid 非空 → parent_id=?；pid 空 → parent_id IS NULL（对齐 XML choose 分支，
//     不含 parent_id=” 的行）；
//   - 无 del_flag 过滤、无排序（对齐 Java SQL）；空列表 → 「未找到角色信息」
//     （Java 原话术）；
//   - 偏差文档化：Java SQL 引用不存在的 leaf_flag 列必然报错被 catch 吞掉
//     （返回 success:true + result:null），Go 用 is_leaf 修正实现。
func (s *Service) QueryListAsync(ctx context.Context, parentID string) ([]asyncTreeModel, error) {
	cols := dao.SysPermission.Columns()
	m := dao.SysPermission.Ctx(ctx).
		Fields(cols.Id, cols.ParentId, cols.Name, cols.Icon, cols.IsLeaf)
	if parentID != "" {
		m = m.Where(cols.ParentId, parentID)
	} else {
		m = m.WhereNull(cols.ParentId)
	}
	var rows []entity.SysPermission
	if err := m.Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询子菜单失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrNoPermInfo
	}
	out := make([]asyncTreeModel, 0, len(rows))
	for _, p := range rows {
		out = append(out, asyncTreeModel{
			Key:      p.Id,
			Title:    p.Name,
			Icon:     p.Icon,
			ParentId: p.ParentId,
			IsLeaf:   p.IsLeaf == 1,
			Children: nil,
		})
	}
	return out, nil
}

// PermissionTreeList 菜单权限树（角色/部门授权弹窗数据源）。
//
// 契约蓝本 SysRoleController#queryTreeList（GET /sys/permission/queryTreeList）：
// del_flag=0 + sort_no ASC；result {treeList:[TreeModel], ids:[全部 id]}。
func (s *Service) PermissionTreeList(ctx context.Context) ([]menuTreeList, []string, error) {
	rows, err := s.listPermissions(ctx)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, p := range rows {
		ids = append(ids, p.Id)
	}
	tree := buildMenuTreeListModel(rows)
	return tree, ids, nil
}

// buildMenuTreeListModel 构建 TreeModel 形态树（对齐 Java getTreeModelList 递归）。
func buildMenuTreeListModel(perms []entity.SysPermission) []menuTreeList {
	var out []menuTreeList
	appendMenuTreeList(&out, perms, nil)
	return out
}

// appendMenuTreeList TreeModel 形态的递归挂载。
func appendMenuTreeList(dest *[]menuTreeList, perms []entity.SysPermission, parent *menuTreeList) {
	for i := range perms {
		p := perms[i]
		node := menuTreeList{
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
			node.Children = []menuTreeList{}
		}
		switch {
		case parent == nil && p.ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !last.IsLeaf {
				appendMenuTreeList(&last.Children, perms, last)
			}
		case parent != nil && p.ParentId == parent.Key:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !last.IsLeaf {
				appendMenuTreeList(&last.Children, perms, last)
			}
		}
	}
}

// QueryRolePermission 查询角色已授权的菜单 ID 列表。
//
// 契约蓝本 SysRoleController#queryRolePermission（GET /sys/permission/queryRolePermission?roleId=）：
// sys_role_permission WHERE role_id → permission_id 字符串数组。
func (s *Service) QueryRolePermission(ctx context.Context, roleID string) ([]string, error) {
	cols := dao.SysRolePermission.Columns()
	var rows []entity.SysRolePermission
	if err := dao.SysRolePermission.Ctx(ctx).
		Fields(cols.PermissionId).
		Where(cols.RoleId, roleID).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询角色授权失败: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.PermissionId)
	}
	return out, nil
}

// SaveRolePermission 保存角色授权（diff 增量）。
//
// 契约蓝本 SysRoleController#saveRolePermission（POST /sys/permission/saveRolePermission，
// body {roleId, permissionIds, lastpermissionIds}）+ SysRolePermissionServiceImpl：
//   - add = lastPermIds 有而 permIds 无 → 插入；(permIds, lastPermIds) 参数序与
//     Java getDiff(main, diff) 一致（getDiff 取 diff 中 main 没有的元素）；
//   - delete = permIds 有而 lastPermIds 无 → 逐个删除 (role_id, permission_id)；
//   - Java 无事务，Go 包单事务（文档化改进：半途失败不再留脏数据）。
func (s *Service) SaveRolePermission(ctx context.Context, roleID, permIDs, lastPermIDs string) error {
	addList := diffPermIDs(lastPermIDs, permIDs)
	delList := diffPermIDs(permIDs, lastPermIDs)
	return dao.SysRolePermission.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.SysRolePermission.Columns()
		for _, pid := range addList {
			if pid == "" {
				continue
			}
			if _, err := dao.SysRolePermission.Ctx(ctx).Data(do.SysRolePermission{
				Id:           guid.S(),
				RoleId:       roleID,
				PermissionId: pid,
			}).Insert(); err != nil {
				return fmt.Errorf("保存角色授权失败: %w", err)
			}
		}
		for _, pid := range delList {
			if pid == "" {
				continue
			}
			if _, err := dao.SysRolePermission.Ctx(ctx).
				Where(cols.RoleId, roleID).
				Where(cols.PermissionId, pid).
				Delete(); err != nil {
				return fmt.Errorf("删除角色授权失败: %w", err)
			}
		}
		return nil
	})
}

// diffPermIDs 纯函数（便于单测）：取 diff 中 main 没有的元素。
// 语义对齐 Java getDiff(main, diff)：diff 空 → nil；main 空 → diff 全量拆分；
// 否则按字符串精确匹配去重保留 diff 原序。
func diffPermIDs(main, diff string) []string {
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

// GetPermRuleListByPermId 菜单的数据规则列表。
//
// 契约蓝本 SysPermissionController#getPermRuleListByPermId（GET ?permissionId=）
// + SysPermissionDataRuleImpl#getPermRuleListByPermId：permission_id 精确 + create_time 倒序。
func (s *Service) GetPermRuleListByPermId(ctx context.Context, permissionID string) ([]entity.SysPermissionDataRule, error) {
	cols := dao.SysPermissionDataRule.Columns()
	var rows []entity.SysPermissionDataRule
	if err := dao.SysPermissionDataRule.Ctx(ctx).
		Where(cols.PermissionId, permissionID).
		OrderDesc(cols.CreateTime).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询数据规则失败: %w", err)
	}
	return rows, nil
}

// AddPermissionRule 添加数据规则。
//
// 契约蓝本 SysPermissionController#addPermissionRule + SysPermissionDataRuleImpl#savePermissionDataRule
// （@Transactional）：create_time 服务端写入；插入后若所属菜单 rule_flag 为空/0
// 则置 1。id 为空时服务端生成。
func (s *Service) AddPermissionRule(ctx context.Context, in DataRuleInput) error {
	return dao.SysPermissionDataRule.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		data := do.SysPermissionDataRule{
			Id:         in.Id,
			CreateTime: gtime.Now(),
		}
		if in.Id == "" {
			data.Id = guid.S()
		}
		if in.PermissionId != nil {
			data.PermissionId = *in.PermissionId
		}
		if in.RuleName != nil {
			data.RuleName = *in.RuleName
		}
		if in.RuleColumn != nil {
			data.RuleColumn = *in.RuleColumn
		}
		if in.RuleConditions != nil {
			data.RuleConditions = *in.RuleConditions
		}
		if in.RuleValue != nil {
			data.RuleValue = *in.RuleValue
		}
		if in.Status != nil {
			data.Status = *in.Status
		}
		if _, err := dao.SysPermissionDataRule.Ctx(ctx).Data(data).Insert(); err != nil {
			return fmt.Errorf("操作失败: %w", err)
		}
		return s.markPermissionRuleFlag(ctx, in.PermissionId)
	})
}

// EditPermissionRule 编辑数据规则。
//
// 契约蓝本 SysPermissionController#editPermissionRule（PUT/POST）：Java 为
// saveOrUpdate（id 空 → save；否则按 id 更新非空字段），不维护 rule_flag、
// 不写时间字段，Go 保持一致。
func (s *Service) EditPermissionRule(ctx context.Context, in DataRuleInput) error {
	data := do.SysPermissionDataRule{}
	if in.PermissionId != nil {
		data.PermissionId = *in.PermissionId
	}
	if in.RuleName != nil {
		data.RuleName = *in.RuleName
	}
	if in.RuleColumn != nil {
		data.RuleColumn = *in.RuleColumn
	}
	if in.RuleConditions != nil {
		data.RuleConditions = *in.RuleConditions
	}
	if in.RuleValue != nil {
		data.RuleValue = *in.RuleValue
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	cols := dao.SysPermissionDataRule.Columns()
	if in.Id == "" {
		data.Id = guid.S()
		if _, err := dao.SysPermissionDataRule.Ctx(ctx).Data(data).Insert(); err != nil {
			return fmt.Errorf("操作失败: %w", err)
		}
		return nil
	}
	if _, err := dao.SysPermissionDataRule.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// DeletePermissionRule 删除数据规则。
//
// 契约蓝本 SysPermissionController#deletePermissionRule + SysPermissionDataRuleImpl#deletePermissionDataRule
// （@Transactional）：删除后若该菜单再无规则，且菜单 rule_flag=1 则回落为 0。
func (s *Service) DeletePermissionRule(ctx context.Context, id string) error {
	return dao.SysPermissionDataRule.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.SysPermissionDataRule.Columns()
		var rule entity.SysPermissionDataRule
		if err := dao.SysPermissionDataRule.Ctx(ctx).Where(cols.Id, id).Scan(&rule); err != nil {
			return fmt.Errorf("查询数据规则失败: %w", err)
		}
		if rule.Id == "" {
			return nil
		}
		if _, err := dao.SysPermissionDataRule.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return fmt.Errorf("删除数据规则失败: %w", err)
		}
		cnt, err := dao.SysPermissionDataRule.Ctx(ctx).
			Where(cols.PermissionId, rule.PermissionId).Count()
		if err != nil {
			return fmt.Errorf("统计数据规则失败: %w", err)
		}
		if cnt == 0 {
			return s.resetPermissionRuleFlag(ctx, rule.PermissionId)
		}
		return nil
	})
}

// markPermissionRuleFlag 菜单 rule_flag 为空/0 时置 1（对齐 Java 判定：
// getRuleFlag()==null || ==RULE_FLAG_0；列可空但实体零值与 NULL 同路径）。
func (s *Service) markPermissionRuleFlag(ctx context.Context, permissionID *string) error {
	if permissionID == nil || *permissionID == "" {
		return nil
	}
	return s.updateRuleFlag(ctx, *permissionID, 1)
}

// resetPermissionRuleFlag 菜单 rule_flag=1 时回落为 0。
func (s *Service) resetPermissionRuleFlag(ctx context.Context, permissionID string) error {
	if permissionID == "" {
		return nil
	}
	return s.updateRuleFlag(ctx, permissionID, 0)
}

// updateRuleFlag 按 Java 语义更新 rule_flag（1：仅当当前为空/0；0：仅当当前为 1）。
func (s *Service) updateRuleFlag(ctx context.Context, permissionID string, flag int) error {
	cols := dao.SysPermission.Columns()
	var p entity.SysPermission
	if err := dao.SysPermission.Ctx(ctx).
		Fields(cols.Id, cols.RuleFlag).
		Where(cols.Id, permissionID).Scan(&p); err != nil {
		return fmt.Errorf("查询菜单失败: %w", err)
	}
	if p.Id == "" {
		return nil
	}
	if flag == 1 && p.RuleFlag != 0 {
		return nil
	}
	if flag == 0 && p.RuleFlag != 1 {
		return nil
	}
	if _, err := dao.SysPermission.Ctx(ctx).
		Where(cols.Id, permissionID).
		Data(do.SysPermission{RuleFlag: flag}).Update(); err != nil {
		return fmt.Errorf("维护菜单规则标记失败: %w", err)
	}
	return nil
}

// QueryPermissionRule 数据规则条件查询。
//
// 契约蓝本 SysPermissionController#queryPermissionRule + SysPermissionDataRuleImpl#queryPermissionRule
// （QueryGenerator.initQueryWrapper）：字符串字段走 parseQueryRule 语义
// （无前缀 EQ、* 通配 → LIKE、, → IN、!= → NE）。
func (s *Service) QueryPermissionRule(ctx context.Context, f DataRuleFilter) ([]entity.SysPermissionDataRule, error) {
	cols := dao.SysPermissionDataRule.Columns()
	m := dao.SysPermissionDataRule.Ctx(ctx)
	if f.PermissionId != "" {
		m = applyQueryRule(m, cols.PermissionId, f.PermissionId)
	}
	if f.RuleName != "" {
		m = applyQueryRule(m, cols.RuleName, f.RuleName)
	}
	if f.RuleColumn != "" {
		m = applyQueryRule(m, cols.RuleColumn, f.RuleColumn)
	}
	if f.RuleConditions != "" {
		m = applyQueryRule(m, cols.RuleConditions, f.RuleConditions)
	}
	if f.RuleValue != "" {
		m = applyQueryRule(m, cols.RuleValue, f.RuleValue)
	}
	if f.Status != "" {
		m = applyQueryRule(m, cols.Status, f.Status)
	}
	var rows []entity.SysPermissionDataRule
	if err := m.Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询数据规则失败: %w", err)
	}
	return rows, nil
}

// DataRuleFilter 数据规则查询过滤条件（QueryGenerator 消费面）。
type DataRuleFilter struct {
	PermissionId   string
	RuleName       string
	RuleColumn     string
	RuleConditions string
	RuleValue      string
	Status         string
}

// QueryDepartPermission 查询部门已授权的菜单 ID 列表。
//
// 契约蓝本 SysRoleController#queryDepartPermission（GET ?departId=）：
// sys_depart_permission WHERE depart_id → permission_id 字符串数组。
func (s *Service) QueryDepartPermission(ctx context.Context, departID string) ([]string, error) {
	cols := dao.SysDepartPermission.Columns()
	var rows []entity.SysDepartPermission
	if err := dao.SysDepartPermission.Ctx(ctx).
		Fields(cols.PermissionId).
		Where(cols.DepartId, departID).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门授权失败: %w", err)
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.PermissionId)
	}
	return out, nil
}

// SaveDepartPermission 保存部门授权（diff 增量）。
//
// 契约蓝本 SysRoleController#saveDepartPermission（POST，body
// {departId, permissionIds, lastPermissionIds}）+ SysDepartPermissionServiceImpl#saveDepartPermission：
//   - add → 插入 sys_depart_permission；
//   - delete → 删 (depart_id, permission_id)，并清理部门角色已授权的同菜单记录；
//   - 偏差文档化：Java 清理 sys_depart_role_permission 仅按 permission_id 过滤
//     （roleIds 查出未用，会跨部门误删），Go 收敛为 permission_id + 该部门角色
//     IN 范围，HTTP 契约一致且无跨部门副作用。
func (s *Service) SaveDepartPermission(ctx context.Context, departID, permIDs, lastPermIDs string) error {
	addList := diffPermIDs(lastPermIDs, permIDs)
	delList := diffPermIDs(permIDs, lastPermIDs)
	return dao.SysDepartPermission.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.SysDepartPermission.Columns()
		for _, pid := range addList {
			if pid == "" {
				continue
			}
			if _, err := dao.SysDepartPermission.Ctx(ctx).Data(do.SysDepartPermission{
				Id:           guid.S(),
				DepartId:     departID,
				PermissionId: pid,
			}).Insert(); err != nil {
				return fmt.Errorf("保存部门授权失败: %w", err)
			}
		}
		for _, pid := range delList {
			if pid == "" {
				continue
			}
			if _, err := dao.SysDepartPermission.Ctx(ctx).
				Where(cols.DepartId, departID).
				Where(cols.PermissionId, pid).
				Delete(); err != nil {
				return fmt.Errorf("删除部门授权失败: %w", err)
			}
			// 该部门名下角色对被移除菜单的授权一并清理（收敛 Java 的跨部门误删面）
			roleIDs, err := g.DB().Model("sys_depart_role").
				Fields("id").
				Where("depart_id", departID).
				Array()
			if err != nil {
				return fmt.Errorf("查询部门角色失败: %w", err)
			}
			if len(roleIDs) > 0 {
				if _, err := g.DB().Model("sys_depart_role_permission").
					Where("permission_id", pid).
					WhereIn("role_id", roleIDs).
					Delete(); err != nil {
					return fmt.Errorf("级联删除部门角色授权失败: %w", err)
				}
			}
		}
		return nil
	})
}
