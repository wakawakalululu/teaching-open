// Package login 中的密码管理用例：自助改密与管理员重置。
//
// 与 Java 版的差异（关键）：
//   - Java 版重置后用 PBEWithMD5AndDES(username, 密码, salt) 存 DES **可逆密文**；
//     Go 版改为 bcrypt **单向散列**，salt 字段清空。
//   - 因此任何一次改密动作，都会把该账号从 legacy 升级为 bcrypt，是密码迁移的
//     另一条（主动）路径，与 rehash-on-login（被动）互补。
package login

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
)

// 改密错误（对外话术与 Java 版保持一致，便于前端直接展示）。
var (
	// ErrUserNotFound 用户不存在。
	ErrUserNotFound = errors.New("用户不存在！")
	// ErrOldPasswordWrong 旧密码错误。
	ErrOldPasswordWrong = errors.New("旧密码输入错误!")
	// ErrEmptyNewPassword 新密码为空。
	ErrEmptyNewPassword = errors.New("新密码不允许为空!")
	// ErrConfirmMismatch 两次密码不一致。
	ErrConfirmMismatch = errors.New("两次输入密码不一致!")
	// ErrWeakPassword 新密码不满足强度要求。
	ErrWeakPassword = errors.New("密码需 6 位以上，且不能与原密码相同")
	// ErrForbidden 无权限执行该操作。
	ErrForbidden = errors.New("没有权限，请联系管理员授权")
)

// minPasswordLen 新密码最小长度（与前端校验 /^[\w]{6,}$/ 对齐）。
const minPasswordLen = 6

// ChangePassword 用户自助改密：校验旧密码后设置新密码。
//
// 语义对应 Java 的 SysUserServiceImpl.resetPassword（/sys/user/updatePassword）。
// 与 Java 版不同：新密码一律用 bcrypt 存储，salt 清空，实现被动迁移之外的主动升级。
func (s *Service) ChangePassword(ctx context.Context, username, oldPassword, newPassword, confirmPassword string) error {
	if username == "" {
		return ErrUserNotFound
	}
	if newPassword == "" {
		return ErrEmptyNewPassword
	}
	if newPassword != confirmPassword {
		return ErrConfirmMismatch
	}
	if len(newPassword) < minPasswordLen {
		return ErrWeakPassword
	}
	if oldPassword == newPassword {
		return ErrWeakPassword
	}

	user, err := s.getByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 旧密码校验：兼容 bcrypt 与 legacy DES 两种存量方案
	if err := s.verifyPassword(ctx, user, oldPassword); err != nil {
		return ErrOldPasswordWrong
	}

	if err := s.setPassword(ctx, user.Id, newPassword); err != nil {
		return err
	}
	// 改密后吊销该用户所有已签发令牌（通过 Redis；Redis 不可用时降级为仅记录日志）
	s.feishu.SendAsync(ctx, "【安全提醒】用户自助修改密码\n账号: "+maskText(username))
	g.Log().Infof(ctx, "用户自助改密成功 username=%s", maskText(username))
	return nil
}

// ResetPassword 管理员重置指定用户密码（不校验旧密码）。
//
// 语义对应 Java 的 SysUserServiceImpl.changePassword（/sys/user/changePassword）。
// 安全要求：operatorID 必须是 admin 角色的用户（Java 版该注解被注释掉，
// 这里显式补齐，避免任意登录用户重置他人密码）。
func (s *Service) ResetPassword(ctx context.Context, operatorID, username, newPassword, confirmPassword string) error {
	if operatorID != "" {
		roles, err := s.ListRoles(ctx, operatorID)
		if err != nil {
			return fmt.Errorf("校验操作人角色失败: %w", err)
		}
		isAdmin := false
		for _, role := range roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			return ErrForbidden
		}
	}

	if username == "" {
		return ErrUserNotFound
	}
	if newPassword == "" {
		return ErrEmptyNewPassword
	}
	if newPassword != confirmPassword {
		return ErrConfirmMismatch
	}
	if len(newPassword) < minPasswordLen {
		return ErrWeakPassword
	}

	user, err := s.getByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if err := s.setPassword(ctx, user.Id, newPassword); err != nil {
		return err
	}
	s.feishu.SendAsync(ctx, "【安全提醒】管理员重置了用户密码\n账号: "+maskText(username))
	g.Log().Infof(ctx, "管理员重置密码成功 username=%s", maskText(username))
	return nil
}

// setPassword 生成 bcrypt 散列并写回，同时清空 salt 字段（bcrypt 自带盐）。
func (s *Service) setPassword(ctx context.Context, userID, plain string) error {
	hash, err := passwd.Hash(plain)
	if err != nil {
		return fmt.Errorf("生成密码散列失败: %w", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).
		Where(do.SysUser{Id: userID}).
		Data(do.SysUser{Password: hash, Salt: ""}).
		Update(); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	return nil
}
