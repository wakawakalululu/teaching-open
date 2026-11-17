// Package store 封装数据访问。全部 SQL 使用占位符绑定，禁止字符串拼接。
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ErrNotFound 记录不存在。
var ErrNotFound = errors.New("记录不存在")

// User 对应 sys_user 表（仅列出服务端实际使用的列）。
type User struct {
	ID         string         `db:"id"`
	Username   string         `db:"username"`
	Realname   string         `db:"realname"`
	Password   string         `db:"password"`
	Salt       string         `db:"salt"`
	Avatar     string         `db:"avatar"`
	Email      sql.NullString `db:"email"`
	Phone      sql.NullString `db:"phone"`
	Status     int            `db:"status"`
	DelFlag    int            `db:"del_flag"`
	OrgCode    sql.NullString `db:"org_code"`
	CreateTime time.Time      `db:"create_time"`
}

// UserRepo 用户仓储。
type UserRepo struct {
	db *sqlx.DB
}

// NewUserRepo 构造仓储。
func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

// userColumns 显式列出列名，避免 SELECT * 在表结构变更时把新敏感列带出。
const userColumns = "id, username, realname, password, salt, avatar, email, phone, status, del_flag, org_code, create_time"

// GetByUsername 按用户名查询未删除用户。
//
// 说明：列名与表名全部为代码内常量，用户输入只出现在 ? 占位符位置，
// 从根上杜绝 SQL 注入（对应 Java 版 Mapper 中 ${} 拼接的风险点）。
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := "SELECT " + userColumns + " FROM sys_user WHERE username = ? AND del_flag = 0 LIMIT 1"
	var u User
	if err := r.db.GetContext(ctx, &u, query, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &u, nil
}

// GetByID 按主键查询未删除用户。
func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	query := "SELECT " + userColumns + " FROM sys_user WHERE id = ? AND del_flag = 0 LIMIT 1"
	var u User
	if err := r.db.GetContext(ctx, &u, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &u, nil
}

// ListRoles 查询用户的角色编码。
func (r *UserRepo) ListRoles(ctx context.Context, userID string) ([]string, error) {
	const query = `
		SELECT r.role_code
		FROM sys_user_role ur
		JOIN sys_role r ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.del_flag = 0`
	roles := make([]string, 0, 4)
	if err := r.db.SelectContext(ctx, &roles, query, userID); err != nil {
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	return roles, nil
}

// UpdatePassword 更新密码散列（用于迁移期的 rehash-on-login）。
func (r *UserRepo) UpdatePassword(ctx context.Context, userID, hash, salt string) error {
	const query = "UPDATE sys_user SET password = ?, salt = ?, update_time = NOW() WHERE id = ?"
	if _, err := r.db.ExecContext(ctx, query, hash, salt, userID); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}
	return nil
}

// GetByPhone 按手机号查询未删除用户（用于手机号登录与找回密码）。
func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*User, error) {
	query := "SELECT " + userColumns + " FROM sys_user WHERE phone = ? AND del_flag = 0 LIMIT 1"
	var u User
	if err := r.db.GetContext(ctx, &u, query, phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &u, nil
}
