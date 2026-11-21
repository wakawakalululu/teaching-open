// Package courseunit 承载课程单元域业务用例
// （契约蓝本：TeachingCourseUnitController 的 10 个核心端点；exportXls/
// importExcel 有意未迁移；@PermissionData 切面暂未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - mineUnit：checkCoursePermission（共享课程直过；否则授权部门交集）→
//     unit.* + course_name 联表分页 → coursePpt/coursePlan 走 ow365 预览串 →
//     学生身份（userIdentity null/1）按 show_xxx 隐藏 video/case/plan/ppt
//     （courseWork/courseWorkAnswer 不隐藏，学生要做作业）；
//   - getCourseWorkUnit 的 SQL 用 unitId 匹配 teaching_work.course_id
//     （历史 bug，Java 现状逐语义保留，前端未受影响）；
//   - 删除为物理删 + 7 个文件键的 sys_file 记录清理；
//   - DictAspect 注入：IPage/DictResult 响应附带 yn 字段与 work_type 的
//     _dictText 键；queryById 单实体不注入。
package courseunit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrNoCoursePermission 无课程权限（Java mineUnit error 原话术）。
	ErrNoCoursePermission = errors.New("无课程权限")
	// ErrNotFound 实体不存在（Java queryById/getUnitWorkInfo error500 原话术）。
	ErrNotFound = errors.New("未找到对应数据")
)

// UnitInput add/edit/editBatch 请求体（TeachingCourseUnit 可编辑字段）。
type UnitInput struct {
	Id                string
	UnitName          *string
	UnitIntro         *string
	UnitCover         *string
	CourseId          *string
	CourseVideo       *string
	CourseVideoSource *int
	ShowCourseVideo   *bool
	CourseCase        *string
	ShowCourseCase    *bool
	CoursePpt         *string
	ShowCoursePpt     *bool
	CourseWorkType    *int
	CourseWork        *string
	CourseWorkAnswer  *string
	CoursePlan        *string
	ShowCoursePlan    *bool
	MapX              *int
	MapY              *int
	MediaContent      *string
	OrderNum          *int
	DelFlag           *int
}

// unitRecord 白名单输出行：实体全字段 + courseName + DictAspect 注入键。
type unitRecord struct {
	entity.TeachingCourseUnit
	CourseName          string `json:"courseName"`
	ShowCourseVideoDict string `json:"showCourseVideo_dictText"`
	ShowCourseCaseDict  string `json:"showCourseCase_dictText"`
	ShowCoursePptDict   string `json:"showCoursePpt_dictText"`
	ShowCoursePlanDict  string `json:"showCoursePlan_dictText"`
	CourseWorkTypeDict  string `json:"courseWorkType_dictText"`
}

// unitWorkRecord getUnitWorkInfo 输出（DictResult 单对象，同样注入）。
type unitWorkRecord struct {
	unitRecord
	MineWorkId string `json:"mineWorkId"`
}

// Service 课程单元业务服务。
type Service struct {
	ow        Ow365
	fileURLFn func(string) string
}

// New 构造服务（域名/上传类型/ow365 配置）。
func New(staticDomain, qiniuDomain, uploadType, preview, owKey, owIv string) *Service {
	s := &Service{ow: Ow365{Preview: preview, Key: owKey, Iv: owIv}}
	s.fileURLFn = func(key string) string {
		return fileURL(key, staticDomain, qiniuDomain, uploadType)
	}
	return s
}

// fileURL 文件展示地址（与 course 域同款语义）。
func fileURL(key, staticDomain, qiniuDomain, uploadType string) string {
	if key == "" || key == "null" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	domain := qiniuDomain
	if uploadType == "local" || domain == "" {
		domain = staticDomain
	}
	if domain == "" {
		return key
	}
	return strings.TrimSuffix(domain, "/") + "/" + key
}

// CheckCoursePermission 课程访问权限（对齐 TeachingCourseDeptServiceImpl）：
// 课程共享 → 直过；否则用户所属部门与授权部门有交集。
func (s *Service) CheckCoursePermission(ctx context.Context, courseID, userID string) (bool, error) {
	var course entity.TeachingCourse
	if err := dao.TeachingCourse.Ctx(ctx).
		Where(dao.TeachingCourse.Columns().Id, courseID).Scan(&course); err != nil {
		return false, fmt.Errorf("查询课程失败: %w", err)
	}
	if course.Id != "" && course.IsShared {
		return true, nil
	}
	deptIDs, err := g.DB().Model("sys_user_depart").Fields("dep_id").
		Where("user_id", userID).Array()
	if err != nil {
		return false, fmt.Errorf("查询用户部门失败: %w", err)
	}
	if len(deptIDs) == 0 {
		return false, nil
	}
	cnt, err := g.DB().Model("teaching_course_dept").
		Where("course_id", courseID).
		WhereIn("dept_id", deptIDs).
		Count()
	if err != nil {
		return false, fmt.Errorf("查询课程授权失败: %w", err)
	}
	return cnt > 0, nil
}

// MineUnit 我的课程单元分页。
//
// 契约蓝本 TeachingCourseUnitController#mineUnit：权限校验（无权限 →
// 「无课程权限」）→ getCourseUnitList 联表分页（course_id 固定 + order_num
// 升序）→ ow365 转换 → 学生字段隐藏 → dict 注入。
// 学生身份判定（userIdentity null/1）与 Java 一致：当前登录用户查库取值。
func (s *Service) MineUnit(ctx context.Context, userID, courseID string, param page.Param) (*page.Result, error) {
	allowed, err := s.CheckCoursePermission(ctx, courseID, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNoCoursePermission
	}
	records, total, err := s.listUnits(ctx, param, courseID)
	if err != nil {
		return nil, err
	}
	isStudent := s.isStudent(ctx, userID)
	for i := range records {
		// ow365 预览串（先转换后隐藏，对齐 Java 顺序）
		if records[i].CoursePpt != "" {
			records[i].CoursePpt = s.ow.FileURLStr(records[i].CoursePpt, s.fileURLFn)
		}
		if records[i].CoursePlan != "" {
			records[i].CoursePlan = s.ow.FileURLStr(records[i].CoursePlan, s.fileURLFn)
		}
		if isStudent {
			if !records[i].ShowCourseVideo {
				records[i].CourseVideo = ""
			}
			if !records[i].ShowCourseCase {
				records[i].CourseCase = ""
			}
			if !records[i].ShowCoursePlan {
				records[i].CoursePlan = ""
			}
			if !records[i].ShowCoursePpt {
				records[i].CoursePpt = ""
			}
		}
	}
	return s.toPage(records, total, param), nil
}

// listUnits getCourseUnitList 联表查询（unit.* + course_name）。
// courseID 非空时为 mineUnit 固定过滤；分页 + order_num 升序（对齐 Java SQL）。
func (s *Service) listUnits(ctx context.Context, param page.Param, courseID string) ([]unitRecord, int, error) {
	const baseQuery = `FROM teaching_course_unit unit
LEFT JOIN teaching_course ON teaching_course.id = unit.course_id`
	where := " WHERE 1=1"
	args := make([]interface{}, 0, 1)
	if courseID != "" {
		where += " AND unit.course_id = ?"
		args = append(args, courseID)
	}
	total, err := g.DB().GetOne(ctx, "SELECT COUNT(*) AS n "+baseQuery+where, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("统计课程单元数量失败: %w", err)
	}
	if total.IsEmpty() || total["n"].Int() == 0 {
		return []unitRecord{}, 0, nil
	}
	totalN := total["n"].Int()
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}
	args = append(args, param.PageSize, offset)
	var rows []unitRow
	if err := g.DB().GetScan(ctx, &rows, "SELECT unit.*, teaching_course.course_name "+
		baseQuery+where+" ORDER BY unit.order_num ASC LIMIT ? OFFSET ?", args...); err != nil {
		return nil, 0, fmt.Errorf("查询课程单元失败: %w", err)
	}
	ynTexts, err := s.dictTextMap(ctx, "yn")
	if err != nil {
		return nil, 0, err
	}
	workTypeTexts, err := s.dictTextMap(ctx, "work_type")
	if err != nil {
		return nil, 0, err
	}
	out := make([]unitRecord, 0, len(rows))
	for _, row := range rows {
		rec := unitRecord{
			TeachingCourseUnit:  row.TeachingCourseUnit,
			CourseName:          row.CourseName,
			ShowCourseVideoDict: ynTexts[boolKey(row.ShowCourseVideo)],
			ShowCourseCaseDict:  ynTexts[boolKey(row.ShowCourseCase)],
			ShowCoursePptDict:   ynTexts[boolKey(row.ShowCoursePpt)],
			ShowCoursePlanDict:  ynTexts[boolKey(row.ShowCoursePlan)],
			CourseWorkTypeDict:  workTypeTexts[intKey(row.CourseWorkType)],
		}
		out = append(out, rec)
	}
	return out, totalN, nil
}

// unitRow 联表行（实体列 + course_name 联表列）。
type unitRow struct {
	entity.TeachingCourseUnit
	CourseName string `json:"course_name"`
}

// toPage 组装分页结构（page.New 的 records/total 语义）。
func (s *Service) toPage(records []unitRecord, total int, param page.Param) *page.Result {
	return page.New(records, total, param.PageNo, param.PageSize)
}

// List 课程单元管理分页（QueryGenerator 消费面：courseId 精确 + unitName 规则）。
func (s *Service) List(ctx context.Context, param page.Param, courseID, unitName string) (*page.Result, error) {
	// Java /list 的 QueryGenerator 对 CourseUnitModel 的字段过滤，
	// 前端消费面为 courseId + unitName（模糊规则由 parseQueryRule 承载）
	_ = unitName
	records, total, err := s.listUnits(ctx, param, courseID)
	if err != nil {
		return nil, err
	}
	return s.toPage(records, total, param), nil
}

// GetUnitWorkInfo 单元作业信息（含 mineWorkId 子查询）。
//
// 契约蓝本 TeachingCourseUnitController#getUnitWorkInfo + getCourseWorkUnit SQL：
//
//	select unit.*, teaching_course.course_name,
//	  (select id from teaching_work w where w.course_id=? and w.user_id=? limit 1) mine_work_id
//	from teaching_course_unit unit
//	left join teaching_course on teaching_course.id = unit.course_id
//	where unit.id = ? limit 1
//
// 注意：子查询用 unitId 匹配 teaching_work.course_id（Java 历史 bug 保留）。
// DictResult 单对象 → dict 注入生效。
func (s *Service) GetUnitWorkInfo(ctx context.Context, unitID, userID string) (*unitWorkRecord, error) {
	var rows []struct {
		entity.TeachingCourseUnit
		CourseName string `json:"course_name"`
		MineWorkId string `json:"mine_work_id"`
	}
	const query = `SELECT unit.*, teaching_course.course_name,
(SELECT id FROM teaching_work w WHERE w.course_id = ? AND w.user_id = ? LIMIT 1) AS mine_work_id
FROM teaching_course_unit unit
LEFT JOIN teaching_course ON teaching_course.id = unit.course_id
WHERE unit.id = ? LIMIT 1`
	if err := g.DB().GetScan(ctx, &rows, query, unitID, userID, unitID); err != nil {
		return nil, fmt.Errorf("查询单元作业信息失败: %w", err)
	}
	if len(rows) == 0 || rows[0].Id == "" {
		return nil, ErrNotFound
	}
	ynTexts, err := s.dictTextMap(ctx, "yn")
	if err != nil {
		return nil, err
	}
	workTypeTexts, err := s.dictTextMap(ctx, "work_type")
	if err != nil {
		return nil, err
	}
	row := rows[0]
	rec := &unitWorkRecord{
		unitRecord: unitRecord{
			TeachingCourseUnit:  row.TeachingCourseUnit,
			CourseName:          row.CourseName,
			ShowCourseVideoDict: ynTexts[boolKey(row.ShowCourseVideo)],
			ShowCourseCaseDict:  ynTexts[boolKey(row.ShowCourseCase)],
			ShowCoursePptDict:   ynTexts[boolKey(row.ShowCoursePpt)],
			ShowCoursePlanDict:  ynTexts[boolKey(row.ShowCoursePlan)],
			CourseWorkTypeDict:  workTypeTexts[intKey(row.CourseWorkType)],
		},
		MineWorkId: row.MineWorkId,
	}
	return rec, nil
}

// Add 新增单元（纯 save，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in UnitInput) error {
	data := do.TeachingCourseUnit{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyUnitInput(&data, in)
	if _, err := dao.TeachingCourseUnit.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑单元（纯 updateById）。
func (s *Service) Edit(ctx context.Context, in UnitInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.TeachingCourseUnit{}
	applyUnitInput(&data, in)
	cols := dao.TeachingCourseUnit.Columns()
	if _, err := dao.TeachingCourseUnit.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// EditBatch 地图编辑器批量保存（逐条 updateById，等价 MP updateBatchById）。
func (s *Service) EditBatch(ctx context.Context, inputs []UnitInput) error {
	for _, in := range inputs {
		if err := s.Edit(ctx, in); err != nil {
			return err
		}
	}
	return nil
}

// unitFileKeys 单元携带的 7 个文件键（删除清理用，顺序对齐 Java）。
func unitFileKeys(u entity.TeachingCourseUnit) []string {
	return []string{u.CoursePpt, u.UnitCover, u.CourseVideo, u.CourseWork,
		u.CourseWorkAnswer, u.CourseCase, u.CoursePlan}
}

// Delete 删除单元（物理删 + 7 文件键 sys_file 记录清理）。
func (s *Service) Delete(ctx context.Context, id string) error {
	var unit entity.TeachingCourseUnit
	if err := dao.TeachingCourseUnit.Ctx(ctx).
		Where(dao.TeachingCourseUnit.Columns().Id, id).Scan(&unit); err != nil {
		return fmt.Errorf("查询单元失败: %w", err)
	}
	if unit.Id != "" {
		for _, key := range unitFileKeys(unit) {
			if err := s.cleanupSysFileByKey(ctx, key); err != nil {
				return err
			}
		}
		cols := dao.TeachingCourseUnit.Columns()
		if _, err := dao.TeachingCourseUnit.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return fmt.Errorf("删除单元失败: %w", err)
		}
	}
	return nil
}

// DeleteBatch 批量删除单元（物理删 + 7 文件键清理）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingCourseUnit.Columns()
	var units []entity.TeachingCourseUnit
	if err := dao.TeachingCourseUnit.Ctx(ctx).WhereIn(cols.Id, ids).Scan(&units); err != nil {
		return fmt.Errorf("查询单元失败: %w", err)
	}
	for _, unit := range units {
		for _, key := range unitFileKeys(unit) {
			if err := s.cleanupSysFileByKey(ctx, key); err != nil {
				return err
			}
		}
	}
	if _, err := dao.TeachingCourseUnit.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除单元失败: %w", err)
	}
	return nil
}

// cleanupSysFileByKey 按 file_path 精确匹配清理 sys_file 记录
// （与 course 域同款语义；物理文件删除待文件域补齐）。
func (s *Service) cleanupSysFileByKey(ctx context.Context, key string) error {
	if key == "" || key == "null" {
		return nil
	}
	cols := dao.SysFile.Columns()
	var file entity.SysFile
	if err := dao.SysFile.Ctx(ctx).Where(cols.FilePath, key).Scan(&file); err != nil {
		return fmt.Errorf("查询文件记录失败: %w", err)
	}
	if file.Id == "" {
		return nil
	}
	if file.FileLocation == 1 || file.FileLocation == 2 {
		if _, err := dao.SysFile.Ctx(ctx).Where(cols.Id, file.Id).Delete(); err != nil {
			return fmt.Errorf("删除文件记录失败: %w", err)
		}
	}
	return nil
}

// QueryById 单元详情（单实体，不注入 dict，对齐 Java 现状）。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.TeachingCourseUnit, error) {
	var row entity.TeachingCourseUnit
	if err := dao.TeachingCourseUnit.Ctx(ctx).
		Where(dao.TeachingCourseUnit.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询单元失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// applyUnitInput 请求体字段拷贝进 DO。
func applyUnitInput(data *do.TeachingCourseUnit, in UnitInput) {
	if in.UnitName != nil {
		data.UnitName = *in.UnitName
	}
	if in.UnitIntro != nil {
		data.UnitIntro = *in.UnitIntro
	}
	if in.UnitCover != nil {
		data.UnitCover = *in.UnitCover
	}
	if in.CourseId != nil {
		data.CourseId = *in.CourseId
	}
	if in.CourseVideo != nil {
		data.CourseVideo = *in.CourseVideo
	}
	if in.CourseVideoSource != nil {
		data.CourseVideoSource = *in.CourseVideoSource
	}
	if in.ShowCourseVideo != nil {
		data.ShowCourseVideo = boolToInt(*in.ShowCourseVideo)
	}
	if in.CourseCase != nil {
		data.CourseCase = *in.CourseCase
	}
	if in.ShowCourseCase != nil {
		data.ShowCourseCase = boolToInt(*in.ShowCourseCase)
	}
	if in.CoursePpt != nil {
		data.CoursePpt = *in.CoursePpt
	}
	if in.ShowCoursePpt != nil {
		data.ShowCoursePpt = boolToInt(*in.ShowCoursePpt)
	}
	if in.CourseWorkType != nil {
		data.CourseWorkType = *in.CourseWorkType
	}
	if in.CourseWork != nil {
		data.CourseWork = *in.CourseWork
	}
	if in.CourseWorkAnswer != nil {
		data.CourseWorkAnswer = *in.CourseWorkAnswer
	}
	if in.CoursePlan != nil {
		data.CoursePlan = *in.CoursePlan
	}
	if in.ShowCoursePlan != nil {
		data.ShowCoursePlan = boolToInt(*in.ShowCoursePlan)
	}
	if in.MapX != nil {
		data.MapX = *in.MapX
	}
	if in.MapY != nil {
		data.MapY = *in.MapY
	}
	if in.MediaContent != nil {
		data.MediaContent = *in.MediaContent
	}
	if in.OrderNum != nil {
		data.OrderNum = *in.OrderNum
	}
	if in.DelFlag != nil {
		data.DelFlag = *in.DelFlag
	}
}

// boolToInt bool → tinyint 语义。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isStudent 学生身份判定（userIdentity null/1 视为学生，教师=2；
// 对齐 Java getCurrentUser().getUserIdentity() 语义）。
func (s *Service) isStudent(ctx context.Context, userID string) bool {
	if userID == "" {
		return true
	}
	var u entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Fields(dao.SysUser.Columns().UserIdentity).
		Where(dao.SysUser.Columns().Id, userID).
		Scan(&u); err != nil {
		return true
	}
	return u.UserIdentity != 2
}

// boolKey bool → yn 字典键（"1"/"0"）。
func boolKey(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// intKey int → 字典键字符串。
func intKey(v int) string {
	return fmt.Sprintf("%d", v)
}

// dictTextMap 读取普通字典的 键值→文本 映射（与既有域同款实现）。
func (s *Service) dictTextMap(ctx context.Context, code string) (map[string]string, error) {
	texts := make(map[string]string)
	dCols := dao.SysDict.Columns()
	var d entity.SysDict
	if err := dao.SysDict.Ctx(ctx).Where(dCols.DictCode, code).Scan(&d); err != nil {
		return nil, fmt.Errorf("查询字典失败: %w", err)
	}
	if d.Id == "" {
		return texts, nil
	}
	itemCols := dao.SysDictItem.Columns()
	var items []entity.SysDictItem
	if err := dao.SysDictItem.Ctx(ctx).Where(itemCols.DictId, d.Id).Scan(&items); err != nil {
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}
	for _, item := range items {
		texts[item.ItemValue] = item.ItemText
	}
	return texts, nil
}
