// Package teachinguser 承载门户用户上下文域业务用例
// （契约蓝本：TeachingUserController 的 3 个端点）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - /info 返回当前用户完整实体（Go entity password/salt 已 json:"-" 隔离）；
//   - /getUserRootDept：org_code 前 3 位（公司级）→ org_code 精确查部门
//     （无 del_flag 过滤，对齐 Java getOne）；查不到部门 → result=null；
//     用户 org_code 为空 → success=false「获取不到当前用户的组织机构」；
//   - /edit：body 全字段透传更新（MP 非空字段语义），密码强制保留旧值，
//     update_time 服务端写入；Go 额外隔离 salt（安全偏差文档化）。
package teachinguser

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/os/gtime"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误。
var (
	// ErrNoOrgCode 用户组织机构缺失（Java 原话术）。
	ErrNoOrgCode = errors.New("获取不到当前用户的组织机构")
	// ErrNotFound 用户不存在（Java edit error500 原话术）。
	ErrNotFound = errors.New("未找到对应实体")
)

// EditInput /teaching/user/edit 请求体（SysUser 可编辑字段的透传集合）。
//
// Java 为 JSON → SysUser 全字段反序列化后 updateById（非空字段更新）；
// password 强制保留旧值。指针字段对齐 MP NOT_NULL 语义。
type EditInput struct {
	Id           string
	Username     *string
	Realname     *string
	Avatar       *string
	Birthday     *string
	Sex          *int
	Email        *string
	Phone        *string
	OrgCode      *string
	Status       *int
	WorkNo       *string
	Post         *string
	Telephone    *string
	School       *string
	UserIdentity *int
	ActivitiSync *int
}

// Service 门户用户上下文业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// Info 当前用户完整实体。
func (s *Service) Info(ctx context.Context, operatorID string) (*entity.SysUser, error) {
	return s.getUser(ctx, operatorID)
}

// getUser 按 ID 加载用户。
func (s *Service) getUser(ctx context.Context, userID string) (*entity.SysUser, error) {
	if userID == "" {
		return &entity.SysUser{}, nil
	}
	var u entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Where(dao.SysUser.Columns().Id, userID).Scan(&u); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &u, nil
}

// UserRootDept 用户根部门（org_code 前 3 位公司级）。
//
// 契约蓝本 TeachingUserController#getUserRootDept：userId 缺省取当前用户；
// org_code 为空 → ErrNoOrgCode；前 3 位精确查 sys_depart（无 del_flag 过滤，
// 未命中 → result=null 仍 success）。
func (s *Service) UserRootDept(ctx context.Context, operatorID, userID string) (*entity.SysDepart, error) {
	orgCode := ""
	if userID == "" {
		user, err := s.getUser(ctx, operatorID)
		if err != nil {
			return nil, err
		}
		orgCode = user.OrgCode
	} else {
		user, err := s.getUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user.Id == "" {
			// Java getById(userId).getOrgCode() → NPE → 全局异常
			return nil, errors.New("操作失败")
		}
		orgCode = user.OrgCode
	}
	if orgCode == "" {
		return nil, ErrNoOrgCode
	}
	rootCode := orgCode
	if len(orgCode) > 3 {
		rootCode = orgCode[:3]
	}
	cols := dao.SysDepart.Columns()
	var depart entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).Where(cols.OrgCode, rootCode).Scan(&depart); err != nil {
		return nil, fmt.Errorf("查询根部门失败: %w", err)
	}
	if depart.Id == "" {
		return nil, nil
	}
	return &depart, nil
}

// Edit 修改用户资料（门户自助）。
//
// 契约蓝本 TeachingUserController#edit：存在性校验（「未找到对应实体」）；
// password 强制保留旧值（Go 额外不透传 salt，文档化偏差）；update_time 服务端。
func (s *Service) Edit(ctx context.Context, in EditInput) error {
	cols := dao.SysUser.Columns()
	var old entity.SysUser
	if err := dao.SysUser.Ctx(ctx).Where(cols.Id, in.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if old.Id == "" {
		return ErrNotFound
	}
	data := do.SysUser{UpdateTime: gtime.Now()}
	if in.Username != nil {
		data.Username = *in.Username
	}
	if in.Realname != nil {
		data.Realname = *in.Realname
	}
	if in.Avatar != nil {
		data.Avatar = *in.Avatar
	}
	if in.Birthday != nil {
		data.Birthday = *in.Birthday
	}
	if in.Sex != nil {
		data.Sex = *in.Sex
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.OrgCode != nil {
		data.OrgCode = *in.OrgCode
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.WorkNo != nil {
		data.WorkNo = *in.WorkNo
	}
	if in.Post != nil {
		data.Post = *in.Post
	}
	if in.Telephone != nil {
		data.Telephone = *in.Telephone
	}
	if in.School != nil {
		data.School = *in.School
	}
	if in.UserIdentity != nil {
		data.UserIdentity = *in.UserIdentity
	}
	if in.ActivitiSync != nil {
		data.ActivitiSync = *in.ActivitiSync
	}
	if _, err := dao.SysUser.Ctx(ctx).Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// AppUserList APP 端用户分页（activitiSync=1 + del_flag=0）。
//
// 契约蓝本 SysUserController#appUserList：username 精确优先，否则
// username/realname 对 keyword 模糊（OR 链）。
func (s *Service) AppUserList(ctx context.Context, param page.Param, keyword, username string) (*page.Result, error) {
	cols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).
		Where(cols.ActivitiSync, 1).
		Where(cols.DelFlag, 0)
	if username != "" {
		m = m.Where(cols.Username, username)
	} else if keyword != "" {
		m = m.Where("("+stringLike(cols.Username)+" OR "+stringLike(cols.Realname)+")",
			"%"+keyword+"%", "%"+keyword+"%")
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计用户数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysUser{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysUser
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// stringLike 占位符 LIKE 片段（列名为代码内常量）。
func stringLike(col string) string {
	return col + " LIKE ?"
}
