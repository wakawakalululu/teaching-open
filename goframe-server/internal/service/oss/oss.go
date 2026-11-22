// Package oss 承载 OSS 文件域业务用例（契约蓝本：OSSFileController 的 4 个端点）。
//
// Java 版通过 MinioUtil 直连 MinIO；本部署 UPLOAD_TYPE=local，因此以本地存储
// 作为后端：上传落盘 {STORAGE_LOCAL_ROOT}/oss/{uuid}.{ext}，url 记录
// {staticDomain}/{path}；删除同时清理记录与物理文件。接口形状与语义对齐 Java，
// 后续接入 MinIO 时只需替换本包的上传/删除实现。
package oss

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// ErrNotFound 文件不存在（Java 话术「文件不存在!」）。
var ErrNotFound = errors.New("文件不存在!")

// Service OSS 文件业务服务。
type Service struct {
	localRoot    string
	staticDomain string
}

// New 构造服务（staticDomain 用于拼接可访问 URL）。
func New(localRoot, staticDomain string) *Service {
	return &Service{localRoot: localRoot, staticDomain: strings.TrimRight(staticDomain, "/")}
}

// List 文件分页（Java 消费面：fileName 模糊 + createBy 精确，createTime 倒序）。
func (s *Service) List(ctx context.Context, param page.Param, fileName, createBy string) (*page.Result, error) {
	cols := dao.OssFile.Columns()
	m := dao.OssFile.Ctx(ctx)
	if fileName != "" {
		m = m.WhereLike(cols.FileName, "%"+fileName+"%")
	}
	if createBy != "" {
		m = m.Where(cols.CreateBy, createBy)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计文件数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.OssFile{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.OssFile
	if err := m.Page(param.PageNo, param.PageSize).OrderDesc(cols.CreateTime).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询文件列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Upload 上传文件到本地 OSS 根目录并登记记录。
// 返回形如 /sys/common/static/oss/xxx.ext 的相对 URL（与 Java 返回 url 字段一致）。
func (s *Service) Upload(ctx context.Context, fileName, tmpPath, ext, operator string) (*entity.OssFile, error) {
	relPath := "oss/" + guid.S() + strings.ToLower(ext)
	absPath := filepath.Join(s.localRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, fmt.Errorf("创建 OSS 目录失败: %w", err)
	}
	src, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败: %w", err)
	}
	if err := os.WriteFile(absPath, src, 0o644); err != nil {
		return nil, fmt.Errorf("写入 OSS 文件失败: %w", err)
	}
	// 静态预览走 /api/sys/common/static/*path；无 staticDomain 时返回相对路径
	url := relPath
	if s.staticDomain != "" {
		url = s.staticDomain + "/" + relPath
	}
	row := &entity.OssFile{
		Id:         guid.S(),
		FileName:   fileName,
		Url:        url,
		CreateBy:   operator,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	if _, err := dao.OssFile.Ctx(ctx).Data(do.OssFile{
		Id: row.Id, FileName: row.FileName, Url: row.Url,
		CreateBy: row.CreateBy, CreateTime: row.CreateTime,
	}).Insert(); err != nil {
		return nil, fmt.Errorf("登记 OSS 文件失败: %w", err)
	}
	return row, nil
}

// Delete 删除记录并清理物理文件（本地 URL 才删盘）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.OssFile.Columns()
	var row entity.OssFile
	if err := dao.OssFile.Ctx(ctx).Where(cols.Id, id).Scan(&row); err != nil {
		return fmt.Errorf("查询文件失败: %w", err)
	}
	if row.Id == "" {
		return ErrNotFound
	}
	if _, err := dao.OssFile.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除文件记录失败: %w", err)
	}
	// 本地文件：url 形如 {domain}/oss/xxx 或 oss/xxx
	rel := row.Url
	if idx := strings.Index(rel, "oss/"); idx >= 0 {
		rel = rel[idx:]
	}
	if strings.HasPrefix(rel, "oss/") {
		abs := filepath.Join(s.localRoot, filepath.FromSlash(rel))
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			g.Log().Warningf(ctx, "OSS 物理文件删除失败 id=%s err=%v", id, err)
		}
	}
	return nil
}

// QueryByID 按 id 查询（不存在 → ErrNotFound）。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.OssFile, error) {
	var row entity.OssFile
	if err := dao.OssFile.Ctx(ctx).Where(dao.OssFile.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询文件失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}
