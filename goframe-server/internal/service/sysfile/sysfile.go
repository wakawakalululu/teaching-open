// Package sysfile 承载文件记录域业务用例
// （契约蓝本：SysFileController 的 7 个核心端点，基路径 /system/sysFile；
// exportXls/importExcel 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - deleteWithFile(id)/deleteByKeyWithFile(key)：file_path 精确匹配；
//     查不到即静默成功；file_location=1（本地）删物理文件（返回值未校验）+
//     删记录；=2（七牛）删远端对象成功后删记录（Go 当前无七牛后端，删除远端
//     以日志记录代替并照常删记录——偏差文档化）；其他 location 值静默保留；
//   - add：Java 纯 save（ID_WORKER_STR 雪花 id），返回回写 id 的实体。
package sysfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// ErrNotFound 实体不存在（Java queryById error 原话术）。
var ErrNotFound = errors.New("未找到对应数据")

// FileInput add/edit 请求体（SysFile 可编辑字段）。
type FileInput struct {
	Id           string
	FileType     *int
	FileName     *string
	FilePath     *string
	FileLocation *int
	FileTag      *string
	SysOrgCode   *string
	DelFlag      *int
}

// Service 文件记录业务服务。
type Service struct {
	localRoot string
}

// New 构造服务（localRoot 为本地存储根目录，本地物理删除用）。
func New(localRoot string) *Service {
	return &Service{localRoot: localRoot}
}

// List 文件记录分页（QueryGenerator 消费面：fileName/filePath/fileTag 规则）。
func (s *Service) List(ctx context.Context, param page.Param, fileName, filePath, fileTag string) (*page.Result, error) {
	cols := dao.SysFile.Columns()
	m := dao.SysFile.Ctx(ctx)
	if fileName != "" {
		m = applyQueryRule(m, cols.FileName, fileName)
	}
	if filePath != "" {
		m = applyQueryRule(m, cols.FilePath, filePath)
	}
	if fileTag != "" {
		m = applyQueryRule(m, cols.FileTag, fileTag)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计文件数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysFile{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysFile
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询文件列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增文件记录（id 为空时服务端生成，返回回写 id 的实体）。
func (s *Service) Add(ctx context.Context, in FileInput) (*entity.SysFile, error) {
	data := do.SysFile{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyInput(&data, in)
	if _, err := dao.SysFile.Ctx(ctx).Data(data).Insert(); err != nil {
		return nil, fmt.Errorf("操作失败: %w", err)
	}
	var row entity.SysFile
	if err := dao.SysFile.Ctx(ctx).
		Where(dao.SysFile.Columns().Id, data.Id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询文件记录失败: %w", err)
	}
	return &row, nil
}

// Edit 编辑文件记录（纯 updateById）。
func (s *Service) Edit(ctx context.Context, in FileInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.SysFile{}
	applyInput(&data, in)
	cols := dao.SysFile.Columns()
	if _, err := dao.SysFile.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除文件记录 + 物理文件（deleteWithFile）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.SysFile.Columns()
	var file entity.SysFile
	if err := dao.SysFile.Ctx(ctx).Where(cols.Id, id).Scan(&file); err != nil {
		return fmt.Errorf("查询文件记录失败: %w", err)
	}
	if file.Id == "" {
		return nil
	}
	return s.removeWithFile(ctx, file)
}

// DeleteByPath 按 file_path 精确匹配删除（记录不存在也返回成功）。
func (s *Service) DeleteByPath(ctx context.Context, filePath string) error {
	cols := dao.SysFile.Columns()
	var file entity.SysFile
	if err := dao.SysFile.Ctx(ctx).Where(cols.FilePath, filePath).Scan(&file); err != nil {
		return fmt.Errorf("查询文件记录失败: %w", err)
	}
	if file.Id == "" {
		return nil
	}
	return s.removeWithFile(ctx, file)
}

// DeleteBatch 批量删除（循环 deleteWithFile）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// QueryById 文件记录详情。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.SysFile, error) {
	var row entity.SysFile
	if err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询文件记录失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// removeWithFile 物理删除 + 记录删除（对齐 SysFileServiceImpl#deleteWithFile）。
func (s *Service) removeWithFile(ctx context.Context, file entity.SysFile) error {
	switch file.FileLocation {
	case 1:
		// 本地：物理文件删除失败不阻断（Java File.delete() 返回值未校验）
		if s.localRoot != "" && file.FilePath != "" {
			_ = os.Remove(filepath.Join(s.localRoot, filepath.FromSlash(file.FilePath)))
		}
		return s.removeByID(ctx, file.Id)
	case 2:
		// 七牛：当前无远端删除后端（Go 版无七牛 SDK），远端对象删除为待补齐项
		//（偏差文档化），记录照常删除
		g.Log().Infof(ctx, "文件记录删除（七牛远端对象删除待补齐）key=%s", file.FilePath)
		return s.removeByID(ctx, file.Id)
	default:
		// 其他 location 值：Java 静默保留记录，逐语义保留
		return nil
	}
}

// removeByID 删除单条记录。
func (s *Service) removeByID(ctx context.Context, id string) error {
	cols := dao.SysFile.Columns()
	if _, err := dao.SysFile.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除文件记录失败: %w", err)
	}
	return nil
}

// applyInput 请求体字段拷贝进 DO。
func applyInput(data *do.SysFile, in FileInput) {
	if in.FileType != nil {
		data.FileType = *in.FileType
	}
	if in.FileName != nil {
		data.FileName = *in.FileName
	}
	if in.FilePath != nil {
		data.FilePath = *in.FilePath
	}
	if in.FileLocation != nil {
		data.FileLocation = *in.FileLocation
	}
	if in.FileTag != nil {
		data.FileTag = *in.FileTag
	}
	if in.SysOrgCode != nil {
		data.SysOrgCode = *in.SysOrgCode
	}
	if in.DelFlag != nil {
		data.DelFlag = *in.DelFlag
	}
}

// applyQueryRule 按 QueryGenerator 值规则应用过滤条件（与既有域同款）。
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

// parseQueryRule QueryGenerator 值规则纯判定：* → LIKE、, → IN、!= → NE、其余 EQ。
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
