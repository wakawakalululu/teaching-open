// Package teachingdaylog 承载班级每日教学记录域业务用例
// （契约蓝本：TeachingDepartDayLogController 的 10 个核心端点；exportXls/
// importExcel 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - unitViewLog：仅教师（userIdentity=2）计数；单元 → 反查班级 →
//     departLog:unitView:{departId} set 幂等 → unit_open_count 日计数；
//   - getReport：聚合求和取单行（MP getOne 默认 throwEx=true——多行命中
//     即 500，逐语义保留）；数据范围：非 admin/dev 仅 depart_ids 列表内；
//   - getReportGroupByDepart：group by depart_id；getReportGroupByMonth：
//     group by date_format(create_time,'%Y-%m')；两者返回 List（pageNo/
//     pageSize 参数被 Java 忽略）；
//   - 计数行 upsert（addLog）：depart_id + create_time=yyyyMMdd 查行，
//     Java 格式错位导致实际恒新插（逐语义保留）。
package teachingdaylog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	twKV "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingwork"
)

// ErrNotFound 记录不存在（Java queryById error 原话术）。
var ErrNotFound = errors.New("未找到对应数据")

// DayLogInput add/edit 请求体（TeachingDepartDayLog 可编辑字段）。
type DayLogInput struct {
	Id                         string
	DepartId                   *string
	DepartName                 *string
	UnitOpenCount              *int
	CourseWorkAssignCount      *int
	AdditionalWorkAssignCount  *int
	CourseWorkCorrectCount     *int
	AdditionalWorkCorrectCount *int
	CourseWorkSubmitCount      *int
	AdditionalWorkSubmitCount  *int
}

// reportRecord 聚合报表行（sum 列 + 分组键）。
type reportRecord struct {
	DepartId                   string `json:"departId"`
	DepartName                 string `json:"departName"`
	CreateTime                 string `json:"createTime"`
	UnitOpenCount              int    `json:"unitOpenCount"`
	CourseWorkAssignCount      int    `json:"courseWorkAssignCount"`
	AdditionalWorkAssignCount  int    `json:"additionalWorkAssignCount"`
	CourseWorkCorrectCount     int    `json:"courseWorkCorrectCount"`
	AdditionalWorkCorrectCount int    `json:"additionalWorkCorrectCount"`
	CourseWorkSubmitCount      int    `json:"courseWorkSubmitCount"`
	AdditionalWorkSubmitCount  int    `json:"additionalWorkSubmitCount"`
}

// Service 班级日志业务服务。
type Service struct {
	kv twKV.KVStore
}

// New 构造服务。
func New(redisEnabled bool) *Service {
	return &Service{kv: twKV.NewKVStore(redisEnabled)}
}

// UnitViewLog 单元打开计数（教师身份）。
//
// 契约蓝本 TeachingDepartDayLogController#unitViewLog：单元反查班级 →
// departLog:unitView:{departId} set 幂等 → unit_open_count 计数；学生身份
// 仅返回成功不计数（对齐 Java userIdentity==2 分支）。
func (s *Service) UnitViewLog(ctx context.Context, operatorID, unitID string) error {
	// 教师身份判定（JWT claims 不含 userIdentity，查库对齐 Java LoginUser）
	var operator entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Fields(dao.SysUser.Columns().UserIdentity).
		Where(dao.SysUser.Columns().Id, operatorID).Scan(&operator); err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	userIdentity := operator.UserIdentity

	// 反查班级（getUserDepartIdByUnitId SQL）
	const query = `SELECT dept_id FROM teaching_course_dept
WHERE course_id = (SELECT course_id FROM teaching_course_unit WHERE id = ?)
 AND dept_id IN (SELECT dep_id FROM sys_user_depart WHERE user_id = ?) LIMIT 1`
	rec, err := g.DB().GetOne(ctx, query, unitID, operatorID)
	if err != nil {
		return fmt.Errorf("查询班级失败: %w", err)
	}
	departID := ""
	if !rec.IsEmpty() {
		departID = rec["dept_id"].String()
	}
	if departID == "" || userIdentity != 2 {
		return nil
	}
	key := fmt.Sprintf("departLog:unitView:%s", departID)
	if !s.kv.SHasKey(ctx, key, unitID) {
		s.kv.SAdd(ctx, key, unitID)
		incrementDayLog(ctx, departID, "unit_open_count")
	}
	return nil
}

// aggregateWhere 报表查询公共段：sum 列表 + 非 admin 部门范围 + 时间区间。
func aggregateWhere(ctx context.Context, operatorID string, startTime, endTime string) (string, []interface{}, error) {
	where := "1=1"
	args := make([]interface{}, 0, 3)
	privileged, err := hasAnyRoleCode(ctx, operatorID, "admin", "dev")
	if err != nil {
		return "", nil, err
	}
	if !privileged {
		departIDs, err := myDepartIDList(ctx, operatorID)
		if err != nil {
			return "", nil, err
		}
		if len(departIDs) == 0 {
			// Java：departIds.split(",") → IN ('') 永不命中 → 空报表
			where += " AND 1=0"
			return where, args, nil
		}
		where += " AND depart_id IN (" + placeholders(len(departIDs)) + ")"
		for _, id := range departIDs {
			args = append(args, id)
		}
	}
	if startTime != "" {
		where += " AND create_time >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		where += " AND create_time <= ?"
		args = append(args, endTime)
	}
	return where, args, nil
}

// sumSelect 聚合投影（对齐 Java select 子句）。
const sumSelect = `depart_id, depart_name,
 sum(unit_open_count) AS unit_open_count,
 sum(course_work_assign_count) AS course_work_assign_count,
 sum(additional_work_assign_count) AS additional_work_assign_count,
 sum(course_work_correct_count) AS course_work_correct_count,
 sum(additional_work_correct_count) AS additional_work_correct_count,
 sum(course_work_submit_count) AS course_work_submit_count,
 sum(additional_work_submit_count) AS additional_work_submit_count`

// GetReport 每日教学记录聚合（单行）。
//
// 契约蓝本 TeachingDepartDayLogController#getReport：MP getOne 默认
// throwEx=true——多行命中即全局异常（逐语义保留：len>1 返回「操作失败」）。
func (s *Service) GetReport(ctx context.Context, operatorID, startTime, endTime string) (*reportRecord, error) {
	where, args, err := aggregateWhere(ctx, operatorID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	var rows []reportRecord
	if err := g.DB().GetScan(ctx, &rows,
		"SELECT "+sumSelect+" FROM teaching_depart_day_log WHERE "+where, args...); err != nil {
		return nil, fmt.Errorf("查询报表失败: %w", err)
	}
	if len(rows) > 1 {
		// MP getOne(throwEx=true)：多行命中抛异常 → 全局 500 话术
		return nil, errors.New("操作失败")
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// GetReportGroupByDepart 按班级聚合（Java 忽略分页参数，返回全量 List）。
func (s *Service) GetReportGroupByDepart(ctx context.Context, operatorID, startTime, endTime string) ([]reportRecord, error) {
	where, args, err := aggregateWhere(ctx, operatorID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	var rows []reportRecord
	if err := g.DB().GetScan(ctx, &rows,
		"SELECT "+sumSelect+" FROM teaching_depart_day_log WHERE "+where+
			" GROUP BY depart_id", args...); err != nil {
		return nil, fmt.Errorf("查询班级聚合报表失败: %w", err)
	}
	return rows, nil
}

// GetReportGroupByMonth 按月聚合（group by date_format(create_time,'%Y-%m')）。
func (s *Service) GetReportGroupByMonth(ctx context.Context, operatorID, startTime, endTime string) ([]reportRecord, error) {
	where, args, err := aggregateWhere(ctx, operatorID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	var rows []reportRecord
	if err := g.DB().GetScan(ctx, &rows,
		"SELECT date_format(create_time, '%Y-%m') AS createTime, "+
			"sum(unit_open_count) AS unit_open_count, "+
			"sum(course_work_assign_count) AS course_work_assign_count, "+
			"sum(additional_work_assign_count) AS additional_work_assign_count, "+
			"sum(course_work_correct_count) AS course_work_correct_count, "+
			"sum(additional_work_correct_count) AS additional_work_correct_count, "+
			"sum(course_work_submit_count) AS course_work_submit_count, "+
			"sum(additional_work_submit_count) AS additional_work_submit_count "+
			"FROM teaching_depart_day_log WHERE "+where+
			" GROUP BY date_format(create_time, '%Y-%m')", args...); err != nil {
		return nil, fmt.Errorf("查询月度聚合报表失败: %w", err)
	}
	return rows, nil
}

// List 记录分页（QueryGenerator 消费面：departId/departName 过滤）。
func (s *Service) List(ctx context.Context, param page.Param, departID, departName string) (*page.Result, error) {
	cols := dao.TeachingDepartDayLog.Columns()
	m := dao.TeachingDepartDayLog.Ctx(ctx)
	if departID != "" {
		m = m.Where(cols.DepartId, departID)
	}
	if departName != "" {
		m = m.WhereLike(cols.DepartName, "%"+departName+"%")
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计记录数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.TeachingDepartDayLog{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingDepartDayLog
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询记录列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增记录（纯 save，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in DayLogInput) error {
	data := do.TeachingDepartDayLog{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyInput(&data, in)
	if _, err := dao.TeachingDepartDayLog.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑记录（纯 updateById）。
func (s *Service) Edit(ctx context.Context, in DayLogInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.TeachingDepartDayLog{}
	applyInput(&data, in)
	cols := dao.TeachingDepartDayLog.Columns()
	if _, err := dao.TeachingDepartDayLog.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除记录。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.TeachingDepartDayLog.Columns()
	if _, err := dao.TeachingDepartDayLog.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除记录失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除记录。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingDepartDayLog.Columns()
	if _, err := dao.TeachingDepartDayLog.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除记录失败: %w", err)
	}
	return nil
}

// QueryByID 记录详情。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.TeachingDepartDayLog, error) {
	var row entity.TeachingDepartDayLog
	if err := dao.TeachingDepartDayLog.Ctx(ctx).
		Where(dao.TeachingDepartDayLog.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// incrementDayLog 计数行 upsert（对齐 TeachingDepartDayLogServiceImpl#addLog：
// depart_id + create_time=yyyyMMdd 查行；Java 格式错位导致实际恒新插）。
// 供 teachingwork/additionalwork 域复用（此处为规范实现，后续波次收口）。
func incrementDayLog(ctx context.Context, departID, countColumn string) {
	today := gtime.Now().Format("Ymd")
	rec, err := g.DB().Model("teaching_depart_day_log").
		Where("depart_id", departID).
		Where("create_time", today).
		Limit(1).One()
	if err != nil {
		return
	}
	if rec.IsEmpty() {
		departName := ""
		var depart entity.SysDepart
		if err := dao.SysDepart.Ctx(ctx).
			Where(dao.SysDepart.Columns().Id, departID).Scan(&depart); err == nil {
			departName = depart.DepartName
		}
		init := g.Map{
			"id":                            guid.S(),
			"depart_id":                     departID,
			"depart_name":                   departName,
			"create_time":                   gtime.Now(),
			"unit_open_count":               0,
			"course_work_assign_count":      0,
			"additional_work_assign_count":  0,
			"course_work_correct_count":     0,
			"additional_work_correct_count": 0,
			"course_work_submit_count":      0,
			"additional_work_submit_count":  0,
			countColumn:                     1,
		}
		_, _ = g.DB().Model("teaching_depart_day_log").Data(init).Insert()
		return
	}
	_, _ = g.DB().Model("teaching_depart_day_log").
		Where("id", rec["id"].String()).
		Data(g.Map{countColumn: gdb.Raw(fmt.Sprintf("%s + 1", countColumn))}).
		Update()
}

// myDepartIDList 当前用户负责部门 id 列表（sys_user.depart_ids）。
func myDepartIDList(ctx context.Context, operatorID string) ([]string, error) {
	if operatorID == "" {
		return []string{}, nil
	}
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Id, operatorID).Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	out := make([]string, 0)
	for _, id := range strings.Split(user.DepartIds, ",") {
		if id != "" {
			out = append(out, id)
		}
	}
	return out, nil
}

// hasAnyRoleCode 判断用户是否拥有任一角色编码（与既有域同款实现）。
func hasAnyRoleCode(ctx context.Context, userID string, roleCodes ...string) (bool, error) {
	if userID == "" || len(roleCodes) == 0 {
		return false, nil
	}
	rec, err := g.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Where("ur.user_id", userID).
		WhereIn("r.role_code", roleCodes).
		Limit(1).One()
	if err != nil {
		return false, fmt.Errorf("查询用户角色失败: %w", err)
	}
	return !rec.IsEmpty(), nil
}

// placeholders n 个 "?" 占位符。
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// applyInput 请求体字段拷贝进 DO。
func applyInput(data *do.TeachingDepartDayLog, in DayLogInput) {
	if in.DepartId != nil {
		data.DepartId = *in.DepartId
	}
	if in.DepartName != nil {
		data.DepartName = *in.DepartName
	}
	if in.UnitOpenCount != nil {
		data.UnitOpenCount = *in.UnitOpenCount
	}
	if in.CourseWorkAssignCount != nil {
		data.CourseWorkAssignCount = *in.CourseWorkAssignCount
	}
	if in.AdditionalWorkAssignCount != nil {
		data.AdditionalWorkAssignCount = *in.AdditionalWorkAssignCount
	}
	if in.CourseWorkCorrectCount != nil {
		data.CourseWorkCorrectCount = *in.CourseWorkCorrectCount
	}
	if in.AdditionalWorkCorrectCount != nil {
		data.AdditionalWorkCorrectCount = *in.AdditionalWorkCorrectCount
	}
	if in.CourseWorkSubmitCount != nil {
		data.CourseWorkSubmitCount = *in.CourseWorkSubmitCount
	}
	if in.AdditionalWorkSubmitCount != nil {
		data.AdditionalWorkSubmitCount = *in.AdditionalWorkSubmitCount
	}
}
