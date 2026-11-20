// Package config 承载站点配置业务用例（契约蓝本：SysConfigController）。
//
// 语义对齐：
//   - 读取仅返回 config_enabled=1 的配置；
//   - saveConfig 为按 key 的 upsert（存在则更新，不存在则插入）；
//   - saveTenantConfig 仅 admin 角色可操作（Java @RequiresRoles("admin")）。
package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/service/rbac"
)

// SiteOptions 前端站点展示参数（getCurrentConfig 契约字段，
// 由环境变量注入，来源对齐 Java application.yml 的 jeecg.* 配置）。
type SiteOptions struct {
	UploadType   string // 上传方式（local/qiniu/...）
	StaticDomain string // 静态资源域名
	FilePreview  string // 文件预览方式（空/"ow365"）
	OwID         string // ow365 预览应用 id（filePreview=ow365 时输出）
}

// 业务错误。
var (
	// ErrForbidden 无权限操作（非 admin）。
	ErrForbidden = errors.New("无权限操作，需要 admin 角色")
	// ErrKeyRequired 配置键为空。
	ErrKeyRequired = errors.New("配置键不能为空")
)

// Service 站点配置业务服务。
type Service struct {
	site SiteOptions
}

// New 构造服务。
func New(site SiteOptions) *Service {
	return &Service{site: site}
}

// GetConfig 按配置键取单条（仅 enabled=1，多条时取第一条，同 Java last("limit 1")）。
func (s *Service) GetConfig(ctx context.Context, key string) (*entity.SysConfig, error) {
	if key == "" {
		return nil, nil
	}
	cols := dao.SysConfig.Columns()
	one, err := dao.SysConfig.Ctx(ctx).
		Where(cols.ConfigKey, key).
		Where(cols.ConfigEnabled, 1).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}
	if one.IsEmpty() {
		return nil, nil
	}
	var data *entity.SysConfig
	if err := one.Struct(&data); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return data, nil
}

// GetAllConfigList 全部生效配置列表。
func (s *Service) GetAllConfigList(ctx context.Context) ([]entity.SysConfig, error) {
	var list []entity.SysConfig
	if err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().ConfigEnabled, 1).
		Scan(&list); err != nil {
		return nil, fmt.Errorf("查询配置列表失败: %w", err)
	}
	return list, nil
}

// GetAllConfigMap 全部生效配置的 key→value 映射。
func (s *Service) GetAllConfigMap(ctx context.Context) (map[string]string, error) {
	list, err := s.GetAllConfigList(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(list))
	for _, item := range list {
		m[item.ConfigKey] = item.ConfigValue
	}
	return m, nil
}

// GetCurrentConfig 对外公开配置：库内配置 + 服务端站点参数（同 Java 追加 jeecg.* 字段）。
func (s *Service) GetCurrentConfig(ctx context.Context) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	kv, err := s.GetAllConfigMap(ctx)
	if err != nil {
		return nil, err
	}
	for k, v := range kv {
		result[k] = v
	}
	result["uploadType"] = s.site.UploadType
	result["qiniuDomain"] = ""
	result["qiniuArea"] = ""
	result["staticDomain"] = s.site.StaticDomain
	if s.site.FilePreview != "" {
		result["filePreview"] = s.site.FilePreview
	}
	if s.site.FilePreview == "ow365" {
		result["owId"] = s.site.OwID
	}
	return result, nil
}

// SaveTenantConfig 保存租户配置（admin 专属）：逐 key upsert。
func (s *Service) SaveTenantConfig(ctx context.Context, operatorID string, kv map[string]interface{}) error {
	isAdmin, err := rbac.HasRole(ctx, operatorID, "admin")
	if err != nil {
		return fmt.Errorf("校验操作人角色失败: %w", err)
	}
	if !isAdmin {
		return ErrForbidden
	}
	for key, raw := range kv {
		value := gconv.String(raw)
		if err := s.saveConfig(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

// saveConfig 单键 upsert（语义同 SysConfigServiceImpl#saveConfig）。
func (s *Service) saveConfig(ctx context.Context, key, value string) error {
	if key == "" {
		return ErrKeyRequired
	}
	cols := dao.SysConfig.Columns()
	one, err := dao.SysConfig.Ctx(ctx).
		Where(cols.ConfigKey, key).
		Where(cols.ConfigEnabled, 1).
		One()
	if err != nil {
		return fmt.Errorf("查询配置失败: %w", err)
	}
	if one.IsEmpty() {
		_, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Id:            guid.S(),
			ConfigKey:     key,
			ConfigValue:   value,
			ConfigEnabled: 1,
		}).Insert()
		if err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
		return nil
	}
	if _, err := dao.SysConfig.Ctx(ctx).
		Where(cols.Id, one[cols.Id].String()).
		Data(do.SysConfig{ConfigValue: value, ConfigEnabled: 1}).
		Update(); err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}
	return nil
}
