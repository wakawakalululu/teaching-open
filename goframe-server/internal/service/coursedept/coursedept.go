// Package coursedept 承载班级课程授权域业务用例
// （契约蓝本：TeachingCourseDeptController 的 7 个核心端点；exportXls/
// importExcel 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - addOrUpdate：逐 (deptId,courseId) 对 upsert（存在则 updateById 复用 id，
//     否则 save）；Java 循环无事务 → Go 包单事务（文档化改进）；
//   - list 为三表联表（授权表 + course_name + depart_name）；
//   - 纯 MP CRUD（add/edit/delete/deleteBatch 无文件清理、无级联）。
package coursedept

import (
	"context"
	"errors"
	"fmt"
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

// DeptInput add/edit 请求体（TeachingCourseDept 可编辑字段）。
type DeptInput struct {
	Id       string
	DeptId   *string
	CourseId *string
	OpenTime *string
}

// deptRecord 联表白名单输出行（CourseDeptModel：实体全字段 + courseName +
// departName，对齐 Java getCourseDeptList SQL）。
type deptRecord struct {
	entity.TeachingCourseDept
	CourseName string `json:"courseName"`
	DepartName string `json:"departName"`
}

// Service 班级课程授权业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// List 授权分页列表（三表联表）。
//
// 契约蓝本 TeachingCourseDeptController#list + TeachingCourseDeptMapper.xml：
// teaching_course_dept d LEFT JOIN teaching_course LEFT JOIN sys_depart；
// QueryGenerator 消费面收敛为 courseName/departName（前端查询条件）。
func (s *Service) List(ctx context.Context, param page.Param, courseName, departName string) (*page.Result, error) {
	const baseQuery = `FROM teaching_course_dept d
LEFT JOIN teaching_course ON teaching_course.id = d.course_id
LEFT JOIN sys_depart ON sys_depart.id = d.dept_id`
	where := " WHERE 1=1"
	args := make([]interface{}, 0, 2)
	if courseName != "" {
		where += " AND teaching_course.course_name LIKE ?"
		args = append(args, "%"+courseName+"%")
	}
	if departName != "" {
		where += " AND sys_depart.depart_name LIKE ?"
		args = append(args, "%"+departName+"%")
	}
	total, err := g.DB().GetOne(ctx, "SELECT COUNT(*) AS n "+baseQuery+where, args...)
	if err != nil {
		return nil, fmt.Errorf("统计授权数量失败: %w", err)
	}
	if total.IsEmpty() || total["n"].Int() == 0 {
		return page.New([]deptRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	totalN := total["n"].Int()
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}
	var rows []deptRecord
	query := "SELECT d.*, teaching_course.course_name, sys_depart.depart_name " +
		baseQuery + where + " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)
	if err := g.DB().GetScan(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("查询授权列表失败: %w", err)
	}
	return page.New(rows, totalN, param.PageNo, param.PageSize), nil
}

// AddOrUpdate 批量授权 upsert。
//
// 契约蓝本 TeachingCourseDeptController#addOrUpdate（POST body
// {deptId, courseIdList:[...]}, DepartCourseVO）：逐对存在则复用 id 更新
// （open_time 等字段清空）、否则新增；Java 循环无事务 → Go 包单事务
// （文档化改进，HTTP 契约一致）。
func (s *Service) AddOrUpdate(ctx context.Context, deptID string, courseIDList []string) error {
	if deptID == "" || len(courseIDList) == 0 {
		return nil
	}
	return dao.TeachingCourseDept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.TeachingCourseDept.Columns()
		for _, courseID := range courseIDList {
			if courseID == "" {
				continue
			}
			var one entity.TeachingCourseDept
			if err := dao.TeachingCourseDept.Ctx(ctx).
				Where(cols.DeptId, deptID).
				Where(cols.CourseId, courseID).
				Scan(&one); err != nil {
				return fmt.Errorf("查询授权失败: %w", err)
			}
			if one.Id == "" {
				if _, err := dao.TeachingCourseDept.Ctx(ctx).Data(do.TeachingCourseDept{
					Id:       guid.S(),
					DeptId:   deptID,
					CourseId: courseID,
				}).Insert(); err != nil {
					return fmt.Errorf("保存授权失败: %w", err)
				}
				continue
			}
			// Java 语义：new TeachingCourseDept 仅含 courseId/deptId/id →
			// updateById 非空字段（open_time 不动），此处等价 no-op 更新
			if _, err := dao.TeachingCourseDept.Ctx(ctx).
				Where(cols.Id, one.Id).
				Data(do.TeachingCourseDept{DeptId: deptID, CourseId: courseID}).
				Update(); err != nil {
				return fmt.Errorf("更新授权失败: %w", err)
			}
		}
		return nil
	})
}

// Add 新增授权（纯 save，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in DeptInput) error {
	data := do.TeachingCourseDept{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyInput(&data, in)
	if _, err := dao.TeachingCourseDept.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑授权（纯 updateById）。
func (s *Service) Edit(ctx context.Context, in DeptInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.TeachingCourseDept{}
	applyInput(&data, in)
	cols := dao.TeachingCourseDept.Columns()
	if _, err := dao.TeachingCourseDept.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除授权（无级联，对齐 Java removeById）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.TeachingCourseDept.Columns()
	if _, err := dao.TeachingCourseDept.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除授权失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除授权（无级联）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingCourseDept.Columns()
	if _, err := dao.TeachingCourseDept.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除授权失败: %w", err)
	}
	return nil
}

// QueryById 授权详情。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.TeachingCourseDept, error) {
	var row entity.TeachingCourseDept
	if err := dao.TeachingCourseDept.Ctx(ctx).
		Where(dao.TeachingCourseDept.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询授权失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// applyInput 请求体字段拷贝进 DO。
func applyInput(data *do.TeachingCourseDept, in DeptInput) {
	if in.DeptId != nil {
		data.DeptId = *in.DeptId
	}
	if in.CourseId != nil {
		data.CourseId = *in.CourseId
	}
	if in.OpenTime != nil && *in.OpenTime != "" {
		data.OpenTime = *in.OpenTime
	}
}

// splitCourseIDs 逗号拆分（保留给控制器使用）。
func splitCourseIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
