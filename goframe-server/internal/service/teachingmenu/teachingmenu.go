// Package teachingmenu 承载门户菜单域业务用例
// （契约蓝本：TeachingMenuController 的 11 个端点）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - getUserMenu 为 anon：hidden=0 + sort_no 升序 → 扁平路由 JSON
//     {id, route, title, url, needLogin, internalOrExternal, icon?}；
//   - add/edit/delete 与 sys permission 同构（menuType0 清父、is_leaf 维护、
//     物理级联递归删子树，teaching_menu 无关联表故无四表清理）；
//   - queryListAsync 的 mapper SQL 引用不存在的 leaf_flag 列（与 sys 权限域
//     同款死列缺陷）→ Go 用 is_leaf 修正实现（偏差文档化）。
package teachingmenu

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrMenuNotFound 菜单不存在（Java JeecgBootException 原话术）。
	ErrMenuNotFound = errors.New("未找到菜单信息")
	// ErrNoMenuInfo 异步子菜单为空（Java queryAsync error500 原话术）。
	ErrNoMenuInfo = errors.New("未找到角色信息")
)

// MenuInput add/edit 请求体（TeachingMenu 可编辑字段）。
type MenuInput struct {
	Id                 string
	ParentId           *string
	Name               *string
	Icon               *string
	Url                *string
	SortNo             *float64
	MenuType           *int
	Leaf               *bool
	Route              *bool
	Description        *string
	Hidden             *bool
	NeedLogin          *int
	InternalOrExternal *bool
	DelFlag            *int
}

// menuTreeModel 对齐 Java TeachingMenuTree 的字段序列化（/list 等树形响应）。
type menuTreeModel struct {
	Id                 string          `json:"id"`
	Key                string          `json:"key"`
	Title              string          `json:"title"`
	ParentId           string          `json:"parentId"`
	Name               string          `json:"name"`
	Icon               string          `json:"icon"`
	Url                string          `json:"url"`
	SortNo             float64         `json:"sortNo"`
	MenuType           int             `json:"menuType"`
	IsLeaf             bool            `json:"isLeaf"`
	Route              bool            `json:"route"`
	Description        string          `json:"description"`
	DelFlag            int             `json:"delFlag"`
	CreateBy           string          `json:"createBy"`
	CreateTime         string          `json:"createTime"`
	UpdateBy           string          `json:"updateBy"`
	UpdateTime         string          `json:"updateTime"`
	Hidden             bool            `json:"hidden"`
	NeedLogin          int             `json:"needLogin"`
	InternalOrExternal bool            `json:"internalOrExternal"`
	Children           []menuTreeModel `json:"children"`
}

// portalRoute getUserMenu 的路由 JSON（对齐 getPermissionJsonObject：扁平
// 字段而非 meta 包装）。
type portalRoute struct {
	Id                 string        `json:"id"`
	Route              bool          `json:"route"`
	Title              string        `json:"title"`
	Url                string        `json:"url"`
	NeedLogin          int           `json:"needLogin"`
	InternalOrExternal bool          `json:"internalOrExternal"`
	Icon               string        `json:"icon,omitempty"`
	Children           []portalRoute `json:"children,omitempty"`
}

// treeModel QueryTreeList/queryListAsync 的 TreeModel（同 sys 权限域形态）。
type treeModel struct {
	Key       string      `json:"key"`
	Title     string      `json:"title"`
	SlotTitle string      `json:"slotTitle"`
	Value     string      `json:"value"`
	Label     string      `json:"label"`
	ParentId  string      `json:"parentId"`
	Icon      string      `json:"icon"`
	IsLeaf    bool        `json:"isLeaf"`
	Children  []treeModel `json:"children"`
}

// Service 门户菜单业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// GetUserMenu 门户路由菜单（anon）。
//
// 契约蓝本 TeachingMenuController#getUserMenu：hidden=0 + sort_no 升序 →
// 路由树（menuType 0/1 出对象，按钮 2 出 null 跳过；children 组装）。
func (s *Service) GetUserMenu(ctx context.Context) ([]portalRoute, error) {
	cols := dao.TeachingMenu.Columns()
	var rows []entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).
		Where(cols.Hidden, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询门户菜单失败: %w", err)
	}
	var out []portalRoute
	appendPortalRoutes(&out, rows, nil)
	return out, nil
}

// appendPortalRoutes 门户路由树递归挂载（对齐 getPermissionJsonArray）。
func appendPortalRoutes(dest *[]portalRoute, rows []entity.TeachingMenu, parent *portalRoute) {
	for i := range rows {
		p := rows[i]
		if p.MenuType != 0 && p.MenuType != 1 {
			continue
		}
		node := portalRoute{
			Id:                 p.Id,
			Route:              p.Route,
			Title:              p.Name,
			Url:                p.Url,
			NeedLogin:          p.NeedLogin,
			InternalOrExternal: p.InternalOrExternal,
			Icon:               p.Icon,
		}
		switch {
		case parent == nil && p.ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !p.Leaf {
				appendPortalRoutes(&last.Children, rows, last)
			}
		case parent != nil && p.ParentId == parent.Id:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !p.Leaf {
				appendPortalRoutes(&last.Children, rows, last)
			}
		}
	}
}

// List 全量菜单树（del_flag=0 + sort_no 升序）。
func (s *Service) List(ctx context.Context) ([]menuTreeModel, error) {
	rows, err := s.listDelFlag0(ctx)
	if err != nil {
		return nil, err
	}
	var out []menuTreeModel
	appendMenuTree(&out, rows, nil)
	return out, nil
}

// GetSystemMenuList 一级菜单扁平列表。
func (s *Service) GetSystemMenuList(ctx context.Context) ([]menuTreeModel, error) {
	cols := dao.TeachingMenu.Columns()
	var rows []entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).
		Where(cols.MenuType, 0).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询一级菜单失败: %w", err)
	}
	out := make([]menuTreeModel, 0, len(rows))
	for _, r := range rows {
		out = append(out, newMenuTreeModel(r))
	}
	return out, nil
}

// GetSystemSubmenu 子菜单扁平列表。
func (s *Service) GetSystemSubmenu(ctx context.Context, parentID string) ([]menuTreeModel, error) {
	cols := dao.TeachingMenu.Columns()
	var rows []entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).
		Where(cols.ParentId, parentID).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询子菜单失败: %w", err)
	}
	out := make([]menuTreeModel, 0, len(rows))
	for _, r := range rows {
		out = append(out, newMenuTreeModel(r))
	}
	return out, nil
}

// GetSystemSubmenuBatch 批量子菜单 → {parentId: [子菜单]}。
func (s *Service) GetSystemSubmenuBatch(ctx context.Context, parentIDs []string) (map[string][]menuTreeModel, error) {
	if len(parentIDs) == 0 {
		return map[string][]menuTreeModel{}, nil
	}
	cols := dao.TeachingMenu.Columns()
	var rows []entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).
		WhereIn(cols.ParentId, parentIDs).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("批量查询子菜单失败: %w", err)
	}
	out := make(map[string][]menuTreeModel)
	for _, r := range rows {
		out[r.ParentId] = append(out[r.ParentId], newMenuTreeModel(r))
	}
	return out, nil
}

// Add 添加菜单（menuType0 → parentId 不写入；父置非叶；create_time/del_flag/
// leaf 固定；id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in MenuInput) error {
	data := do.TeachingMenu{
		Id:         in.Id,
		CreateTime: gtime.Now(),
		DelFlag:    0,
		Leaf:       1,
	}
	if in.Id == "" {
		data.Id = guid.S()
	}
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
	if _, err := dao.TeachingMenu.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑菜单（同 sys permission editPermission 语义：存在性校验 +
// menuType0 父清空 + 叶子双向维护）。
func (s *Service) Edit(ctx context.Context, in MenuInput) error {
	if in.Id == "" {
		return ErrMenuNotFound
	}
	cols := dao.TeachingMenu.Columns()
	var old entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).Where(cols.Id, in.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询菜单失败: %w", err)
	}
	if old.Id == "" {
		return ErrMenuNotFound
	}
	data := do.TeachingMenu{UpdateTime: gtime.Now()}
	if in.MenuType != nil && *in.MenuType == 0 {
		empty := ""
		data.ParentId = empty
	} else if in.ParentId != nil {
		data.ParentId = *in.ParentId
	}
	cnt, err := dao.TeachingMenu.Ctx(ctx).Where(cols.ParentId, in.Id).Count()
	if err != nil {
		return fmt.Errorf("统计子菜单失败: %w", err)
	}
	if cnt == 0 {
		data.Leaf = 1
	}
	applyMenuInput(&data, in)
	if _, err := dao.TeachingMenu.Ctx(ctx).Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
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
			if err := s.setMenuLeaf(ctx, newPid, 0); err != nil {
				return err
			}
		}
		cc, err := dao.TeachingMenu.Ctx(ctx).Where(cols.ParentId, old.ParentId).Count()
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

// Delete 删除菜单（物理级联递归删子树 + 父叶子维护；teaching_menu 无关联表）。
func (s *Service) Delete(ctx context.Context, id string) error {
	return dao.TeachingMenu.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.TeachingMenu.Columns()
		var p entity.TeachingMenu
		if err := dao.TeachingMenu.Ctx(ctx).Where(cols.Id, id).Scan(&p); err != nil {
			return fmt.Errorf("查询菜单失败: %w", err)
		}
		if p.Id == "" {
			return ErrMenuNotFound
		}
		if p.ParentId != "" {
			cnt, err := dao.TeachingMenu.Ctx(ctx).Where(cols.ParentId, p.ParentId).Count()
			if err != nil {
				return fmt.Errorf("统计兄弟节点失败: %w", err)
			}
			if cnt == 1 {
				if err := s.setMenuLeaf(ctx, p.ParentId, 1); err != nil {
					return err
				}
			}
		}
		if _, err := dao.TeachingMenu.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return fmt.Errorf("删除菜单失败: %w", err)
		}
		return s.removeChildren(ctx, id)
	})
}

// DeleteBatch 批量删除（异常分支「删除成功!」缺陷话术未照搬）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// removeChildren 递归删子树（对齐 Java removeChildrenBy，无关联表清理）。
func (s *Service) removeChildren(ctx context.Context, parentID string) error {
	cols := dao.TeachingMenu.Columns()
	var children []entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).
		Where(cols.ParentId, parentID).
		Scan(&children); err != nil {
		return fmt.Errorf("查询子菜单失败: %w", err)
	}
	if len(children) == 0 {
		return nil
	}
	if _, err := dao.TeachingMenu.Ctx(ctx).
		Where(cols.ParentId, parentID).Delete(); err != nil {
		return fmt.Errorf("删除子菜单失败: %w", err)
	}
	for _, child := range children {
		cnt, err := dao.TeachingMenu.Ctx(ctx).Where(cols.ParentId, child.Id).Count()
		if err != nil {
			return fmt.Errorf("统计下级菜单失败: %w", err)
		}
		if cnt > 0 {
			if err := s.removeChildren(ctx, child.Id); err != nil {
				return err
			}
		}
	}
	return nil
}

// setMenuLeaf 维护 is_leaf 列（对齐 TeachingMenuMapper#setMenuLeaf 注解 SQL）。
func (s *Service) setMenuLeaf(ctx context.Context, id string, leaf int) error {
	if id == "" {
		return nil
	}
	if _, err := dao.TeachingMenu.Ctx(ctx).
		Where(dao.TeachingMenu.Columns().Id, id).
		Data(do.TeachingMenu{Leaf: leaf}).Update(); err != nil {
		return fmt.Errorf("维护菜单叶子标记失败: %w", err)
	}
	return nil
}

// QueryTreeList 菜单树 + 全部 ids（{treeList, ids}）。
func (s *Service) QueryTreeList(ctx context.Context) ([]treeModel, []string, error) {
	rows, err := s.listDelFlag0(ctx)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	var out []treeModel
	appendTreeModelList(&out, rows, nil)
	return out, ids, nil
}

// QueryListAsync 异步子菜单（pid 空 → parent_id IS NULL；无 del_flag 过滤、
// 无排序；leaf_flag 死列缺陷已用 is_leaf 修正——偏差文档化）。
func (s *Service) QueryListAsync(ctx context.Context, parentID string) ([]treeModel, error) {
	cols := dao.TeachingMenu.Columns()
	m := dao.TeachingMenu.Ctx(ctx).
		Fields(cols.Id, cols.ParentId, cols.Name, cols.Icon, cols.Leaf)
	if parentID != "" {
		m = m.Where(cols.ParentId, parentID)
	} else {
		m = m.WhereNull(cols.ParentId)
	}
	var rows []entity.TeachingMenu
	if err := m.Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询子菜单失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrNoMenuInfo
	}
	out := make([]treeModel, 0, len(rows))
	for _, p := range rows {
		node := treeModel{
			Key:       p.Id,
			Title:     p.Name,
			SlotTitle: p.Name,
			Value:     p.Id,
			Label:     p.Name,
			ParentId:  p.ParentId,
			Icon:      p.Icon,
			IsLeaf:    p.Leaf,
		}
		if !node.IsLeaf {
			node.Children = []treeModel{}
		}
		out = append(out, node)
	}
	return out, nil
}

// listDelFlag0 del_flag=0 + sort_no 升序全量查询。
func (s *Service) listDelFlag0(ctx context.Context) ([]entity.TeachingMenu, error) {
	cols := dao.TeachingMenu.Columns()
	var rows []entity.TeachingMenu
	if err := dao.TeachingMenu.Ctx(ctx).
		Where(cols.DelFlag, 0).
		OrderAsc(cols.SortNo).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询菜单失败: %w", err)
	}
	return rows, nil
}

// appendMenuTree 菜单树递归挂载（叶子 children=null、非叶 isLeaf=false；
// isLeaf 取 is_leaf 列，对齐 TeachingMenuTree 构造器路径）。
func appendMenuTree(dest *[]menuTreeModel, rows []entity.TeachingMenu, parent *menuTreeModel) {
	for i := range rows {
		node := newMenuTreeModel(rows[i])
		switch {
		case parent == nil && rows[i].ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !last.IsLeaf {
				appendMenuTree(&last.Children, rows, last)
			}
		case parent != nil && rows[i].ParentId == parent.Id:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !last.IsLeaf {
				appendMenuTree(&last.Children, rows, last)
			}
		}
	}
}

// newMenuTreeModel 由 TeachingMenu 构造树节点（对齐 TeachingMenuTree 构造器）。
func newMenuTreeModel(r entity.TeachingMenu) menuTreeModel {
	node := menuTreeModel{
		Id:                 r.Id,
		Key:                r.Id,
		Title:              r.Name,
		ParentId:           r.ParentId,
		Name:               r.Name,
		Icon:               r.Icon,
		Url:                r.Url,
		SortNo:             r.SortNo,
		MenuType:           r.MenuType,
		IsLeaf:             r.Leaf,
		Route:              r.Route,
		Description:        r.Description,
		DelFlag:            r.DelFlag,
		CreateBy:           r.CreateBy,
		CreateTime:         r.CreateTime,
		UpdateBy:           r.UpdateBy,
		UpdateTime:         r.UpdateTime,
		Hidden:             r.Hidden,
		NeedLogin:          r.NeedLogin,
		InternalOrExternal: r.InternalOrExternal,
	}
	if !node.IsLeaf {
		node.Children = []menuTreeModel{}
	}
	return node
}

// appendTreeModelList TreeModel 形态递归挂载。
func appendTreeModelList(dest *[]treeModel, rows []entity.TeachingMenu, parent *treeModel) {
	for i := range rows {
		p := rows[i]
		node := treeModel{
			Key:       p.Id,
			Title:     p.Name,
			SlotTitle: p.Name,
			Value:     p.Id,
			Label:     p.Name,
			ParentId:  p.ParentId,
			Icon:      p.Icon,
			IsLeaf:    p.Leaf,
		}
		if !node.IsLeaf {
			node.Children = []treeModel{}
		}
		switch {
		case parent == nil && p.ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			if !last.IsLeaf {
				appendTreeModelList(&last.Children, rows, last)
			}
		case parent != nil && p.ParentId == parent.Key:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			if !last.IsLeaf {
				appendTreeModelList(&last.Children, rows, last)
			}
		}
	}
}

// applyMenuInput 请求体字段拷贝进 DO（ParentId 由 add/edit 单独管理）。
func applyMenuInput(data *do.TeachingMenu, in MenuInput) {
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Icon != nil {
		data.Icon = *in.Icon
	}
	if in.Url != nil {
		data.Url = *in.Url
	}
	if in.SortNo != nil {
		data.SortNo = *in.SortNo
	}
	if in.MenuType != nil {
		data.MenuType = *in.MenuType
	}
	if in.Leaf != nil {
		data.Leaf = boolToInt(*in.Leaf)
	}
	if in.Route != nil {
		data.Route = boolToInt(*in.Route)
	}
	if in.Description != nil {
		data.Description = *in.Description
	}
	if in.Hidden != nil {
		data.Hidden = boolToInt(*in.Hidden)
	}
	if in.NeedLogin != nil {
		data.NeedLogin = *in.NeedLogin
	}
	if in.InternalOrExternal != nil {
		data.InternalOrExternal = boolToInt(*in.InternalOrExternal)
	}
	if in.DelFlag != nil {
		data.DelFlag = *in.DelFlag
	}
}

// boolToInt bool → tinyint 语义。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
