// Package course 承载课程域业务用例
// （契约蓝本：TeachingCourseController 的 10 个端点；exportXls/importExcel
// 有意未迁移；@PermissionData 数据权限切面暂未迁移，见 README 已知差距）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - 删除即物理删除，并清理 course_icon/course_cover/course_map 三个文件键
//     对应的 sys_file 记录（deleteByKeyWithFile：file_path 精确匹配、查不到即
//     成功；物理文件删除待文件域落地后补齐，当前先删记录）；
//   - del_flag 列存在但业务从不读写（Java 现状），所有查询不过滤；
//   - mineCourse：is_shared=1 OR 已开课授权部门（open_time < now()）；
//   - getHomeCourse 为 Shiro anon（登录页首页可调）；
//   - DictAspect 注入：IPage/List 响应附带 courseType_dictText/courseCategory_
//     dictText（sys_dict）与 departIds_dictText（sys_depart 多值逗号 join）；
//     @FileUrl 注入 _url 键（qiniu→qiniuDomain、local→staticDomain 拼 key）；
//     queryById 单实体不注入（Java 现状）。
package course

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// identRegexp 动态标识符白名单（排序列防注入，与既有域同一正则）。
var identRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// orderColumns 排序列白名单 = teaching_course 全量列。
var orderColumns = map[string]struct{}{
	"id": {}, "create_by": {}, "create_time": {}, "update_by": {}, "update_time": {},
	"sys_org_code": {}, "del_flag": {}, "course_name": {}, "course_desc": {},
	"course_icon": {}, "course_cover": {}, "show_type": {}, "course_map": {},
	"is_shared": {}, "show_home": {}, "order_num": {}, "depart_ids": {},
	"course_type": {}, "course_category": {},
}

// applyOrder 排序参数白名单应用（无默认排序，对齐 Java QueryGenerator）。
func applyOrder(m *gdb.Model, column, orderType string) *gdb.Model {
	column = strings.TrimSpace(column)
	if column == "" || strings.TrimSpace(orderType) == "" {
		return m
	}
	column = strings.TrimSuffix(column, "_dictText")
	column = strings.TrimSuffix(column, "_url")
	col := camelToUnderline(column)
	if !identRegexp.MatchString(col) {
		return m
	}
	if _, ok := orderColumns[col]; !ok {
		return m
	}
	if strings.Contains(strings.ToUpper(orderType), "ASC") {
		return m.OrderAsc(col)
	}
	return m.OrderDesc(col)
}

// camelToUnderline 驼峰转下划线。
func camelToUnderline(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
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

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrNotFound 课程不存在（Java queryById error 原话术）。
	ErrNotFound = errors.New("未找到对应数据")
)

// CourseInput add/edit 请求体（TeachingCourse 可编辑字段）。
//
// 指针对齐 MyBatis-Plus NOT_NULL 字段策略；Java add/edit 为纯 save/updateById，
// 不写审计列（body 未提供即落列默认值/不更新）。
type CourseInput struct {
	Id             string
	CourseName     *string
	CourseDesc     *string
	CourseIcon     *string
	CourseCover    *string
	ShowType       *int
	CourseMap      *string
	IsShared       *bool
	ShowHome       *bool
	OrderNum       *int
	DepartIds      *string
	CourseType     *string
	CourseCategory *string
	SysOrgCode     *string
	DelFlag        *int
}

// CourseFilter list 查询过滤条件（QueryGenerator 消费面）。
type CourseFilter struct {
	CourseName     string
	CourseType     string
	CourseCategory string
	IsShared       *bool
	ShowHome       *bool
	OrderNum       *int
	OrderColumn    string
	OrderType      string
}

// courseRecord 白名单输出行：TeachingCourse 全字段 + DictAspect/_url 注入键
// （对齐 Java IPage/List 响应的运行时增强）。
type courseRecord struct {
	entity.TeachingCourse
	CourseIconURL          string `json:"courseIcon_url"`
	CourseCoverURL         string `json:"courseCover_url"`
	CourseMapURL           string `json:"courseMap_url"`
	DepartIdsDictText      string `json:"departIds_dictText"`
	CourseTypeDictText     string `json:"courseType_dictText"`
	CourseCategoryDictText string `json:"courseCategory_dictText"`
}

// Service 课程业务服务。
type Service struct {
	staticDomain string
	qiniuDomain  string
	uploadType   string
}

// New 构造服务（站点静态域名/七牛域名/上传类型用于 _url 注入）。
func New(staticDomain, qiniuDomain, uploadType string) *Service {
	return &Service{staticDomain: staticDomain, qiniuDomain: qiniuDomain, uploadType: uploadType}
}

// MineCourse 我的课程列表。
//
// 契约蓝本 TeachingCourseController#mineCourse + TeachingCourseMapper.xml：
//
//	select * from teaching_course
//	where is_shared=1 or id in (
//	    select course_id from teaching_course_dept
//	    inner join sys_user_depart on sys_user_depart.dep_id = teaching_course_dept.dept_id
//	    where sys_user_depart.user_id = ? and open_time is not null and open_time < now())
//	order by order_num asc
//
// 参数为代码内常量 + 单占位符，无注入面。
func (s *Service) MineCourse(ctx context.Context, userID string) ([]entity.TeachingCourse, error) {
	const query = `SELECT * FROM teaching_course
WHERE is_shared = 1 OR id IN (
    SELECT course_id FROM teaching_course_dept
    INNER JOIN sys_user_depart ON sys_user_depart.dep_id = teaching_course_dept.dept_id
    WHERE sys_user_depart.user_id = ?
      AND teaching_course_dept.open_time IS NOT NULL
      AND teaching_course_dept.open_time < NOW()
)
ORDER BY order_num ASC`
	var rows []entity.TeachingCourse
	if err := g.DB().GetScan(ctx, &rows, query, userID); err != nil {
		return nil, fmt.Errorf("查询我的课程失败: %w", err)
	}
	return rows, nil
}

// GetHomeCourse 首页课程分页（Shiro anon）。
//
// 契约蓝本 TeachingCourseController#getHomeCourse（GET ?courseName=&courseType=
// &courseCategory=&pageNo=&pageSize=）：固定 show_home=1；courseName 非空 LIKE
// （Java 显式 like，区别于 /list 的 QueryGenerator EQ 语义）；courseType/
// courseCategory 提供时精确匹配（Java 参数为 Integer，MySQL 与 varchar 列
// 隐式转型，Go 直接按字符串等值）。
func (s *Service) GetHomeCourse(ctx context.Context, param page.Param, courseName, courseType, courseCategory string) (*page.Result, error) {
	cols := dao.TeachingCourse.Columns()
	m := dao.TeachingCourse.Ctx(ctx).Where(cols.ShowHome, 1)
	if courseName != "" {
		m = m.WhereLike(cols.CourseName, "%"+courseName+"%")
	}
	if courseType != "" {
		m = m.Where(cols.CourseType, courseType)
	}
	if courseCategory != "" {
		m = m.Where(cols.CourseCategory, courseCategory)
	}
	return s.pageCourses(ctx, m, param)
}

// List 课程管理分页列表。
//
// 契约蓝本 TeachingCourseController#list（GET /teaching/teachingCourse/list）：
//   - departId 参数：getParentDepartIds（org_code 每 3 位逐级取父）后
//     OR 链：depart_ids=” OR depart_ids LIKE %各父部门id%（无界 like，
//     Java 现状逐语义保留）；@PermissionData 数据权限切面暂未迁移；
//   - 实体字段过滤走 QueryGenerator 消费面；column/order 排序白名单。
func (s *Service) List(ctx context.Context, param page.Param, f CourseFilter, departID string) (*page.Result, error) {
	cols := dao.TeachingCourse.Columns()
	m := dao.TeachingCourse.Ctx(ctx)
	if departID != "" {
		parentIDs, err := s.getParentDepartIDs(ctx, departID)
		if err != nil {
			return nil, err
		}
		// (depart_ids = '' OR depart_ids LIKE '%id1%' OR ...) 单括号分组
		conds := []string{cols.DepartIds + " = ''"}
		args := make([]interface{}, 0, len(parentIDs))
		for _, pid := range parentIDs {
			conds = append(conds, cols.DepartIds+" LIKE ?")
			args = append(args, "%"+pid+"%")
		}
		m = m.Where("("+strings.Join(conds, " OR ")+")", args...)
	}
	if f.CourseName != "" {
		m = applyQueryRule(m, cols.CourseName, f.CourseName)
	}
	if f.CourseType != "" {
		m = applyQueryRule(m, cols.CourseType, f.CourseType)
	}
	if f.CourseCategory != "" {
		m = applyQueryRule(m, cols.CourseCategory, f.CourseCategory)
	}
	if f.IsShared != nil {
		m = m.Where(cols.IsShared, boolToInt(*f.IsShared))
	}
	if f.ShowHome != nil {
		m = m.Where(cols.ShowHome, boolToInt(*f.ShowHome))
	}
	if f.OrderNum != nil {
		m = m.Where(cols.OrderNum, *f.OrderNum)
	}
	m = applyOrder(m, f.OrderColumn, f.OrderType)
	return s.pageCourses(ctx, m, param)
}

// pageCourses 分页装配 + DictAspect/_url 注入。
func (s *Service) pageCourses(ctx context.Context, m *gdb.Model, param page.Param) (*page.Result, error) {
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计课程数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]courseRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingCourse
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询课程列表失败: %w", err)
	}
	records, err := s.Enrich(ctx, rows)
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// enrichCourses 批量注入 _dictText 与 _url 键（对齐 DictAspect + @FileUrl）。
func (s *Service) Enrich(ctx context.Context, rows []entity.TeachingCourse) ([]courseRecord, error) {
	typeTexts, err := s.dictTextMap(ctx, "course_type")
	if err != nil {
		return nil, err
	}
	categoryTexts, err := s.dictTextMap(ctx, "course_category")
	if err != nil {
		return nil, err
	}
	departNames, err := s.departNameMap(ctx, rows)
	if err != nil {
		return nil, err
	}
	out := make([]courseRecord, 0, len(rows))
	for _, r := range rows {
		rec := courseRecord{TeachingCourse: r}
		rec.CourseIconURL = s.fileURL(r.CourseIcon)
		rec.CourseCoverURL = s.fileURL(r.CourseCover)
		rec.CourseMapURL = s.fileURL(r.CourseMap)
		rec.CourseTypeDictText = typeTexts[r.CourseType]
		rec.CourseCategoryDictText = categoryTexts[r.CourseCategory]
		rec.DepartIdsDictText = joinDepartNames(r.DepartIds, departNames)
		out = append(out, rec)
	}
	return out, nil
}

// dictTextMap 读取普通字典的 键值→文本 映射（与 userrole 域同款实现）。
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

// departNameMap 收集课程授权部门 id 并加载名称映射（@Dict dictTable 注入）。
func (s *Service) departNameMap(ctx context.Context, rows []entity.TeachingCourse) (map[string]string, error) {
	idSet := make(map[string]struct{})
	for _, r := range rows {
		for _, id := range strings.Split(r.DepartIds, ",") {
			if id != "" {
				idSet[id] = struct{}{}
			}
		}
	}
	names := make(map[string]string)
	if len(idSet) == 0 {
		return names, nil
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	cols := dao.SysDepart.Columns()
	var departs []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).Fields(cols.Id, cols.DepartName).
		WhereIn(cols.Id, ids).Scan(&departs); err != nil {
		return nil, fmt.Errorf("查询部门名称失败: %w", err)
	}
	for _, d := range departs {
		names[d.Id] = d.DepartName
	}
	return names, nil
}

// joinDepartNames 逗号分隔部门 id 串 → 逗号分隔部门名串（DictAspect 多值语义）。
func joinDepartNames(departIDs string, names map[string]string) string {
	if departIDs == "" {
		return ""
	}
	parts := strings.Split(departIDs, ",")
	out := make([]string, 0, len(parts))
	for _, id := range parts {
		if id == "" {
			continue
		}
		out = append(out, names[id])
	}
	return strings.Join(out, ",")
}

// fileURL 文件展示地址（对齐 QiniuUtil#getFileUrl：空/"null" → ""；http 前缀
// 原样；local → staticDomain、七牛 → qiniuDomain 拼 "/" + key）。
func (s *Service) fileURL(key string) string {
	if key == "" || key == "null" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	domain := s.qiniuDomain
	if s.uploadType == "local" || domain == "" {
		domain = s.staticDomain
	}
	if domain == "" {
		return key
	}
	return strings.TrimSuffix(domain, "/") + "/" + key
}

// getParentDepartIds 部门自身 + org_code 逐级父部门 id（对齐
// SysDepartServiceImpl#getParentDepartIds：org_code 3 位/级，父级按
// org_code 精确 + del_flag=0 查询，查不到跳过——Java 加 null 再被控制器
// like 传入，Go 过滤 nil 语义等价）。
func (s *Service) getParentDepartIDs(ctx context.Context, departID string) ([]string, error) {
	cols := dao.SysDepart.Columns()
	var depart entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).Where(cols.Id, departID).Scan(&depart); err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	if depart.Id == "" {
		return []string{}, nil
	}
	ids := []string{depart.Id}
	orgCode := depart.OrgCode
	for i := 1; i < len(orgCode)/3; i++ {
		code := orgCode[:i*3]
		var parent entity.SysDepart
		if err := dao.SysDepart.Ctx(ctx).
			Where(cols.OrgCode, code).
			Where(cols.DelFlag, "0").
			Scan(&parent); err != nil {
			return nil, fmt.Errorf("查询父部门失败: %w", err)
		}
		if parent.Id != "" {
			ids = append(ids, parent.Id)
		}
	}
	return ids, nil
}

// Add 新增课程（纯 save，不写审计列，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in CourseInput) error {
	data := do.TeachingCourse{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyCourseInput(&data, in)
	if _, err := dao.TeachingCourse.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑课程（纯 updateById，MP 非空字段语义）。
func (s *Service) Edit(ctx context.Context, in CourseInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.TeachingCourse{}
	applyCourseInput(&data, in)
	cols := dao.TeachingCourse.Columns()
	if _, err := dao.TeachingCourse.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除课程（物理删 + 三个文件键的 sys_file 记录清理）。
//
// 契约蓝本 TeachingCourseController#delete：课程不存在同样返回「删除成功!」。
func (s *Service) Delete(ctx context.Context, id string) error {
	var course entity.TeachingCourse
	if err := dao.TeachingCourse.Ctx(ctx).
		Where(dao.TeachingCourse.Columns().Id, id).Scan(&course); err != nil {
		return fmt.Errorf("查询课程失败: %w", err)
	}
	if course.Id != "" {
		for _, key := range []string{course.CourseIcon, course.CourseCover, course.CourseMap} {
			if err := s.cleanupSysFileByKey(ctx, key); err != nil {
				return err
			}
		}
		cols := dao.TeachingCourse.Columns()
		if _, err := dao.TeachingCourse.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return fmt.Errorf("删除课程失败: %w", err)
		}
	}
	return nil
}

// DeleteBatch 批量删除课程（物理删 + 文件键清理）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingCourse.Columns()
	var courses []entity.TeachingCourse
	if err := dao.TeachingCourse.Ctx(ctx).WhereIn(cols.Id, ids).Scan(&courses); err != nil {
		return fmt.Errorf("查询课程失败: %w", err)
	}
	for _, course := range courses {
		for _, key := range []string{course.CourseIcon, course.CourseCover, course.CourseMap} {
			if err := s.cleanupSysFileByKey(ctx, key); err != nil {
				return err
			}
		}
	}
	if _, err := dao.TeachingCourse.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除课程失败: %w", err)
	}
	return nil
}

// cleanupSysFileByKey 按 file_path 精确匹配清理 sys_file 记录
// （对齐 SysFileServiceImpl#deleteByKeyWithFile：查不到即成功；
// file_location ∈ {1,2} 时删记录——物理文件删除待文件域补齐（文档化）；
// 其他 location 值 Java 静默保留记录，逐语义保留）。
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

// QueryById 课程详情（单实体，不注入 _dictText/_url，对齐 Java 现状）。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.TeachingCourse, error) {
	var row entity.TeachingCourse
	if err := dao.TeachingCourse.Ctx(ctx).
		Where(dao.TeachingCourse.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询课程失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// applyCourseInput 请求体字段拷贝进 DO。
func applyCourseInput(data *do.TeachingCourse, in CourseInput) {
	if in.CourseName != nil {
		data.CourseName = *in.CourseName
	}
	if in.CourseDesc != nil {
		data.CourseDesc = *in.CourseDesc
	}
	if in.CourseIcon != nil {
		data.CourseIcon = *in.CourseIcon
	}
	if in.CourseCover != nil {
		data.CourseCover = *in.CourseCover
	}
	if in.ShowType != nil {
		data.ShowType = *in.ShowType
	}
	if in.CourseMap != nil {
		data.CourseMap = *in.CourseMap
	}
	if in.IsShared != nil {
		data.IsShared = boolToInt(*in.IsShared)
	}
	if in.ShowHome != nil {
		data.ShowHome = boolToInt(*in.ShowHome)
	}
	if in.OrderNum != nil {
		data.OrderNum = *in.OrderNum
	}
	if in.DepartIds != nil {
		data.DepartIds = *in.DepartIds
	}
	if in.CourseType != nil {
		data.CourseType = *in.CourseType
	}
	if in.CourseCategory != nil {
		data.CourseCategory = *in.CourseCategory
	}
	if in.SysOrgCode != nil {
		data.SysOrgCode = *in.SysOrgCode
	}
	if in.DelFlag != nil {
		data.DelFlag = *in.DelFlag
	}
}

// boolToInt bool → tinyint 语义（1/0）。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
