// Package rbac 提供跨业务域复用的角色查询与校验能力。
//
// 登录服务（service/login）与管理类用例（如租户配置保存）都需要
// 判断操作人角色，统一收敛到此处，避免各处 SQL 漂移。
package rbac

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// ListRoles 返回用户拥有的全部角色编码。
func ListRoles(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, nil
	}
	var codes []string
	err := g.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Fields("r.role_code").
		Where("ur.user_id", userID).
		Ctx(ctx).
		Scan(&codes)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// HasRole 判断用户是否拥有指定角色编码。
func HasRole(ctx context.Context, userID, roleCode string) (bool, error) {
	roles, err := ListRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		if role == roleCode {
			return true, nil
		}
	}
	return false, nil
}
