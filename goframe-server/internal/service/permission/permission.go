// Package permission 承载菜单/权限业务用例（契约蓝本：SysPermissionController +
// SysPermissionMapper.queryByUser）。
//
// 输出形状逐字段对齐 Java getPermissionJsonObject：
//
//	id/route/path/name/hidden/alwaysShow/component/redirect
//	meta{keepAlive/internalOrExternal/title/icon/url}/children
//
// 已核对的 Java 版事实（勿重复分析）：
//   - menuType=2（按钮）在 getPermissionJsonObject 中返回 null 被 continue，
//     meta.permissionList 分支实际不可达，按钮权限全部走 auth/allAuth 输出；
//   - 顶层仅输出 parent_id 为空的节点，父节点缺失的孤儿节点静默丢弃；
//   - 外链 URL（http(s):// 或 {{ 开头）path 取 MD5（前端按此寻址）。
package permission

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// 菜单类型（CommonConstant.MENU_TYPE_*）。
const (
	menuTypeRoot   = 0 // 一级菜单
	menuTypeSub    = 1 // 子菜单
	menuTypeButton = 2 // 按钮权限
)

// indexComponent 首页组件标识（缺失时由服务端兜底注入）。
const indexComponent = "dashboard/Index"

// statusValid 按钮权限有效状态（CommonConstant.STATUS_1）。
const statusValid = "1"

// Service 菜单权限业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// QueryByUser 查询用户可见的菜单/按钮集合。
//
// SQL 移植自 SysPermissionMapper.queryByUser：角色授权 + 部门角色授权两路
// UNION，附带 /online 表单路由兜底，del_flag=0，按 sort_no 升序。
// 全部占位符绑定，无动态标识符。
func (s *Service) QueryByUser(ctx context.Context, username string) ([]entity.SysPermission, error) {
	const query = `SELECT * FROM (
		SELECT p.* FROM sys_permission p
		WHERE (exists(
			select a.id from sys_role_permission a
			join sys_role b on a.role_id = b.id
			join sys_user_role c on c.role_id = b.id
			join sys_user d on d.id = c.user_id
			where p.id = a.permission_id AND d.username = ?
		)
		or (p.url like '%:code' and p.url like '/online%' and p.hidden = 1)
		or p.url = '/online')
		and p.del_flag = 0
		UNION
		SELECT p.* FROM sys_permission p
		WHERE exists(
			select a.id from sys_depart_role_permission a
			join sys_depart_role b on a.role_id = b.id
			join sys_depart_role_user c on c.drole_id = b.id
			join sys_user d on d.id = c.user_id
			where p.id = a.permission_id AND d.username = ?
		)
		and p.del_flag = 0
	) h order by h.sort_no ASC`

	records, err := dao.SysPermission.DB().GetAll(ctx, query, username, username)
	if err != nil {
		return nil, fmt.Errorf("查询用户菜单失败: %w", err)
	}
	var list []entity.SysPermission
	if err := records.Structs(&list); err != nil {
		return nil, fmt.Errorf("解析菜单数据失败: %w", err)
	}
	return list, nil
}

// GetUserPermissionByToken 组装前端路由数据：{menu, auth, allAuth}。
//
// 与 Java 一致：菜单列表缺失首页（dashboard/Index）时，取库内该组件配置
// 兜底插入到首位。
func (s *Service) GetUserPermissionByToken(ctx context.Context, username string) (g.Map, error) {
	metaList, err := s.QueryByUser(ctx, username)
	if err != nil {
		return nil, err
	}
	if !hasIndexPage(metaList) {
		index, err := s.getIndexPage(ctx)
		if err != nil {
			return nil, err
		}
		if index != nil {
			metaList = append([]entity.SysPermission{*index}, metaList...)
		}
	}

	menu := buildMenuArray(metaList, "")
	auth := buildAuthArray(metaList)
	allAuth, err := s.buildAllAuthArray(ctx)
	if err != nil {
		return nil, err
	}
	return g.Map{
		"menu":    menu,
		"auth":    auth,
		"allAuth": allAuth,
	}, nil
}

// getIndexPage 取首页菜单配置（component=dashboard/Index，取第一条）。
func (s *Service) getIndexPage(ctx context.Context) (*entity.SysPermission, error) {
	cols := dao.SysPermission.Columns()
	records, err := dao.SysPermission.Ctx(ctx).
		Where(cols.Component, indexComponent).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询首页菜单失败: %w", err)
	}
	if records.IsEmpty() {
		return nil, nil
	}
	var data *entity.SysPermission
	if err := records.Struct(&data); err != nil {
		return nil, fmt.Errorf("解析首页菜单失败: %w", err)
	}
	return data, nil
}

// buildAllAuthArray 全部按钮权限集合（del_flag=0 且 menu_type=2，不过滤 status，同 Java）。
func (s *Service) buildAllAuthArray(ctx context.Context) ([]g.Map, error) {
	cols := dao.SysPermission.Columns()
	records, err := dao.SysPermission.Ctx(ctx).
		Where(cols.DelFlag, 0).
		Where(cols.MenuType, menuTypeButton).
		All()
	if err != nil {
		return nil, fmt.Errorf("查询全部按钮权限失败: %w", err)
	}
	out := make([]g.Map, 0, len(records))
	for _, rec := range records {
		out = append(out, g.Map{
			"action":   rec[cols.Perms].String(),
			"status":   rec[cols.Status].String(),
			"type":     rec[cols.PermsType].String(),
			"describe": rec[cols.Name].String(),
		})
	}
	return out, nil
}

// hasIndexPage 判断菜单列表是否已包含首页配置。
func hasIndexPage(list []entity.SysPermission) bool {
	for _, p := range list {
		if p.Component == indexComponent {
			return true
		}
	}
	return false
}

// buildMenuArray 组装菜单路由树。
//
// parentID 为空表示顶层调用：仅输出 parent_id 为空的节点；其余按
// parent_id 归入对应父节点的 children（孤儿节点静默丢弃，同 Java）。
func buildMenuArray(list []entity.SysPermission, parentID string) []g.Map {
	var out []g.Map
	for i := range list {
		p := &list[i]
		// 按钮不进入路由树（Java getPermissionJsonObject 对 type=2 返回 null）
		if p.MenuType == menuTypeButton {
			continue
		}
		node := menuNodeJSON(p)
		if node == nil {
			continue
		}
		if parentID == "" {
			if p.ParentId == "" {
				out = append(out, node)
				attachChildren(list, p, node)
			}
			continue
		}
		if p.ParentId == parentID {
			out = append(out, node)
			attachChildren(list, p, node)
		}
	}
	return out
}

// attachChildren 非叶子节点递归装载子菜单（children 仅在非空时输出）。
func attachChildren(list []entity.SysPermission, parent *entity.SysPermission, node g.Map) {
	if parent.IsLeaf == 1 {
		return
	}
	children := buildMenuArray(list, parent.Id)
	if len(children) > 0 {
		node["children"] = children
	}
}

// menuNodeJSON 单个菜单节点（移植 getPermissionJsonObject，type 0/1）。
func menuNodeJSON(p *entity.SysPermission) g.Map {
	node := g.Map{
		"id":    p.Id,
		"route": routeFlag(p.IsRoute),
	}
	// 外链 URL 的 path 为 MD5 摘要（前端按此生成占位路由）
	if isWWWHttpURL(p.Url) {
		node["path"] = md5Hex(p.Url)
	} else if p.Url != "" {
		node["path"] = p.Url
	}
	name := p.ComponentName
	if name == "" {
		name = urlToRouteName(p.Url)
	}
	if name != "" {
		node["name"] = name
	}
	if p.Hidden == 1 {
		node["hidden"] = true
	}
	if p.AlwaysShow == 1 {
		node["alwaysShow"] = true
	}
	if p.Component != "" {
		node["component"] = p.Component
	}
	meta := g.Map{
		"keepAlive":          p.KeepAlive == 1,
		"internalOrExternal": p.InternalOrExternal == 1,
		"title":              p.Name,
	}
	if p.Icon != "" {
		meta["icon"] = p.Icon
	}
	if isWWWHttpURL(p.Url) {
		meta["url"] = p.Url
	}
	node["meta"] = meta
	// 一级菜单输出跳转地址（Java 对空值不输出）
	if p.ParentId == "" && p.Redirect != "" {
		node["redirect"] = p.Redirect
	}
	return node
}

// routeFlag 是否生成路由（"1"/"0"，Java 契约为字符串）。
func routeFlag(isRoute int) string {
	if isRoute == 1 {
		return "1"
	}
	return "0"
}

// isWWWHttpURL 外链判断：http(s):// 或 {{ 动态片段开头（同 Java isWWWHttpUrl）。
func isWWWHttpURL(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "{{")
}

// urlToRouteName URL 转路由 name：去前导斜杠，"/"→"-"，":"→"@"（同 Java）。
func urlToRouteName(url string) string {
	if url == "" {
		return ""
	}
	url = strings.TrimPrefix(url, "/")
	url = strings.ReplaceAll(url, "/", "-")
	url = strings.ReplaceAll(url, ":", "@")
	return url
}

// md5Hex MD5 摘要（外链菜单 path）。
func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// buildAuthArray 用户拥有的按钮权限（menu_type=2 且 status 有效，同 Java getAuthJsonArray）。
func buildAuthArray(list []entity.SysPermission) []g.Map {
	out := make([]g.Map, 0, 8)
	for i := range list {
		p := &list[i]
		if p.MenuType == menuTypeButton && p.Status == statusValid {
			out = append(out, g.Map{
				"action":   p.Perms,
				"type":     p.PermsType,
				"describe": p.Name,
			})
		}
	}
	return out
}
