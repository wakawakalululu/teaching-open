// Package additionalwork 承载附加作业域业务用例
// （契约蓝本：TeachingAdditionalWorkController 的 8 个端点；该控制器无
// exportXls/importExcel 端点——前端 URL 配置为死链）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - list：非 admin/dev 按负责部门 work_dept LIKE OR 链过滤（单括号分组），
//     无负责班级 →「您没有负责的班级」；teacherList 强制 create_by=当前用户；
//   - add/edit：workDept 逐班写 Redis departLog:addiWorkAssign:{departId}
//     （set 幂等）+ 班级日志计数列 course_work_assign_count（键名与计数列
//     错位的 Java 现状逐语义保留）；add 落库后「添加成功！」、edit 先计数
//     后更新「编辑成功!」；workDept 为空时 Java split NPE → 全局 500，
//     Go 跳过计数循环（文档化偏差）；
//   - getWorkInfo：SQL 仅 select * from teaching_additional_work where id，
//     MineAdditionalWorkModel 的 mine* 扩展字段恒 null；result 恒 success
//     （不存在也 result=null 不报错）；
//   - DictAspect 注入（IPage）：codeType_dictText（work_type）、
//     workDept_dictText（sys_depart 多值逗号 join）、status_dictText
//     （additional_work_status）。
package additionalwork

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

// ErrNotFound 附加作业不存在（Java queryById error 原话术）。
var ErrNotFound = errors.New("未找到对应数据")

// ErrNoDepartClass 非管理员无负责班级（Java /list error 原话术）。
var ErrNoDepartClass = errors.New("您没有负责的班级")

// WorkInput add/edit 请求体（TeachingAdditionalWork 可编辑字段）。
type WorkInput struct {
	Id              string
	CodeType        *int
	WorkName        *string
	WorkDesc        *string
	WorkCover       *string
	WorkUrl         *string
	WorkDept        *string
	Status          *int
	WorkDocumentUrl *string
}

// workRecord IPage 白名单输出行（实体全字段 + DictAspect 注入键）。
type workRecord struct {
	entity.TeachingAdditionalWork
	CodeTypeDictText string `json:"codeType_dictText"`
	WorkDeptDictText string `json:"workDept_dictText"`
	StatusDictText   string `json:"status_dictText"`
}

// Service 附加作业业务服务。
type Service struct {
	kv twKV.KVStore
}

// New 构造服务。
func New(redisEnabled bool) *Service {
	return &Service{kv: twKV.NewKVStore(redisEnabled)}
}

// List 附加作业分页。
//
// 契约蓝本 TeachingAdditionalWorkController#list：非 admin/dev 按负责部门
// work_dept LIKE OR 链（单括号分组，对齐 queryWrapper.and(q -> or like...)）；
// 实体字段过滤走 QueryGenerator 消费面。
func (s *Service) List(ctx context.Context, operatorID string, param page.Param, workName, status string) (*page.Result, error) {
	cols := dao.TeachingAdditionalWork.Columns()
	m := dao.TeachingAdditionalWork.Ctx(ctx)
	privileged, privErr := s.hasAnyRoleCode(ctx, operatorID, "admin", "dev")
	if privErr != nil {
		return nil, privErr
	}
	if !privileged {
		myIDs, err := s.mySubDepartIDs(ctx, operatorID)
		if err != nil {
			return nil, err
		}
		if len(myIDs) == 0 {
			return nil, ErrNoDepartClass
		}
		conds := make([]string, 0, len(myIDs))
		args := make([]interface{}, 0, len(myIDs))
		for _, id := range myIDs {
			conds = append(conds, cols.WorkDept+" LIKE ?")
			args = append(args, "%"+id+"%")
		}
		m = m.Where("("+strings.Join(conds, " OR ")+")", args...)
	}
	if workName != "" {
		m = applyQueryRule(m, cols.WorkName, workName)
	}
	if status != "" {
		m = applyQueryRule(m, cols.Status, status)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计附加作业数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]workRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingAdditionalWork
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询附加作业失败: %w", err)
	}
	records, err := s.enrich(ctx, rows)
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// TeacherList 老师附加作业分页（强制 create_by=当前用户）。
func (s *Service) TeacherList(ctx context.Context, operatorID string, param page.Param, workName, status string) (*page.Result, error) {
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Id, operatorID).Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	cols := dao.TeachingAdditionalWork.Columns()
	m := dao.TeachingAdditionalWork.Ctx(ctx).Where(cols.CreateBy, user.Username)
	if workName != "" {
		m = applyQueryRule(m, cols.WorkName, workName)
	}
	if status != "" {
		m = applyQueryRule(m, cols.Status, status)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计附加作业数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]workRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingAdditionalWork
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询附加作业失败: %w", err)
	}
	records, err := s.enrich(ctx, rows)
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// GetWorkInfo 附加作业详情。
//
// 契约蓝本 TeachingAdditionalWorkController#getWorkInfo +
// TeachingAdditionalWorkServiceImpl#getWorkInfo：Result 恒 success，
// result 为实体或 null（不存在不报错）；SQL 仅查主表，mine* 扩展字段恒 null。
func (s *Service) GetWorkInfo(ctx context.Context, id string) (*entity.TeachingAdditionalWork, error) {
	var row entity.TeachingAdditionalWork
	if err := dao.TeachingAdditionalWork.Ctx(ctx).
		Where(dao.TeachingAdditionalWork.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询附加作业失败: %w", err)
	}
	if row.Id == "" {
		return nil, nil
	}
	return &row, nil
}

// Add 新增附加作业。
//
// 契约蓝本 TeachingAdditionalWorkController#add：Redis 计数先于落库（Java
// 现状逐语义保留）；落库 →「添加成功！」；id 为空时服务端生成。
func (s *Service) Add(ctx context.Context, in WorkInput) error {
	assignDepartLog(ctx, s.kv, derefStr(in.WorkDept), in.Id)
	data := do.TeachingAdditionalWork{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyInput(&data, in)
	if _, err := dao.TeachingAdditionalWork.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑附加作业（Redis 计数先于更新——对齐 Java 顺序）。
func (s *Service) Edit(ctx context.Context, in WorkInput) error {
	assignDepartLog(ctx, s.kv, derefStr(in.WorkDept), in.Id)
	if in.Id == "" {
		return nil
	}
	data := do.TeachingAdditionalWork{}
	applyInput(&data, in)
	cols := dao.TeachingAdditionalWork.Columns()
	if _, err := dao.TeachingAdditionalWork.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除附加作业。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.TeachingAdditionalWork.Columns()
	if _, err := dao.TeachingAdditionalWork.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除附加作业失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除附加作业。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingAdditionalWork.Columns()
	if _, err := dao.TeachingAdditionalWork.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除附加作业失败: %w", err)
	}
	return nil
}

// QueryByID 附加作业详情（管理端）。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.TeachingAdditionalWork, error) {
	var row entity.TeachingAdditionalWork
	if err := dao.TeachingAdditionalWork.Ctx(ctx).
		Where(dao.TeachingAdditionalWork.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询附加作业失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// assignDepartLog add/edit 的逐班 Redis 计数（departLog:addiWorkAssign:
// {departId} set 幂等 + 日计数行 course_work_assign_count——键名与计数列
// 错位为 Java 现状，逐语义保留；workDept 为空跳过，Java 为 NPE→500 的
// 文档化偏差）。
func assignDepartLog(ctx context.Context, kv twKV.KVStore, workDept, workID string) {
	if workDept == "" || workID == "" {
		return
	}
	for _, departID := range strings.Split(workDept, ",") {
		departID = strings.TrimSpace(departID)
		if departID == "" {
			continue
		}
		key := fmt.Sprintf("departLog:addiWorkAssign:%s", departID)
		if kv.SHasKey(ctx, key, workID) {
			continue
		}
		kv.SAdd(ctx, key, workID)
		incrementDayLog(ctx, departID, "course_work_assign_count")
	}
}

// incrementDayLog 班级日志日计数行自增（对齐 TeachingDepartDayLogServiceImpl#
// addLog：depart_id + create_time=yyyyMMdd 查行；Java 格式错位导致实际恒
// 新插，逐语义保留）。
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

// enrich 批量注入 codeType/status 字典与 workDept 部门名（多值逗号 join）。
func (s *Service) enrich(ctx context.Context, rows []entity.TeachingAdditionalWork) ([]workRecord, error) {
	typeTexts, err := s.dictTextMap(ctx, "work_type")
	if err != nil {
		return nil, err
	}
	statusTexts, err := s.dictTextMap(ctx, "additional_work_status")
	if err != nil {
		return nil, err
	}
	idSet := make(map[string]struct{})
	for _, r := range rows {
		for _, id := range strings.Split(r.WorkDept, ",") {
			if id != "" {
				idSet[id] = struct{}{}
			}
		}
	}
	departNames := make(map[string]string)
	if len(idSet) > 0 {
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
			departNames[d.Id] = d.DepartName
		}
	}
	out := make([]workRecord, 0, len(rows))
	for _, r := range rows {
		rec := workRecord{
			TeachingAdditionalWork: r,
			CodeTypeDictText:       typeTexts[intKey(r.CodeType)],
			StatusDictText:         statusTexts[intKey(r.Status)],
		}
		parts := strings.Split(r.WorkDept, ",")
		names := make([]string, 0, len(parts))
		for _, id := range parts {
			if id == "" {
				continue
			}
			names = append(names, departNames[id])
		}
		rec.WorkDeptDictText = strings.Join(names, ",")
		out = append(out, rec)
	}
	return out, nil
}

// mySubDepartIDs 负责部门及其子部门 id（与 teachingwork 域同款实现）。
func (s *Service) mySubDepartIDs(ctx context.Context, operatorID string) ([]string, error) {
	if operatorID == "" {
		return []string{}, nil
	}
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Id, operatorID).Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user.DepartIds == "" {
		return []string{}, nil
	}
	var codes []string
	for _, id := range strings.Split(user.DepartIds, ",") {
		if id == "" {
			continue
		}
		var d entity.SysDepart
		if err := dao.SysDepart.Ctx(ctx).
			Where(dao.SysDepart.Columns().Id, id).Scan(&d); err != nil {
			return nil, fmt.Errorf("查询负责部门失败: %w", err)
		}
		if d.Id != "" && d.OrgCode != "" {
			codes = append(codes, d.OrgCode)
		}
	}
	if len(codes) == 0 {
		return []string{}, nil
	}
	conds := make([]string, 0, len(codes))
	args := make([]interface{}, 0, len(codes))
	for _, code := range codes {
		conds = append(conds, "org_code LIKE ?")
		args = append(args, code+"%")
	}
	ids, err := g.DB().Model("sys_depart").Fields("id").
		Where("("+strings.Join(conds, " OR ")+")", args...).Array()
	if err != nil {
		return nil, fmt.Errorf("查询子部门失败: %w", err)
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out, nil
}

// hasAnyRoleCode 判断用户是否拥有任一角色编码（与既有域同款实现）。
func (s *Service) hasAnyRoleCode(ctx context.Context, userID string, roleCodes ...string) (bool, error) {
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

// applyInput 请求体字段拷贝进 DO。
func applyInput(data *do.TeachingAdditionalWork, in WorkInput) {
	if in.CodeType != nil {
		data.CodeType = *in.CodeType
	}
	if in.WorkName != nil {
		data.WorkName = *in.WorkName
	}
	if in.WorkDesc != nil {
		data.WorkDesc = *in.WorkDesc
	}
	if in.WorkCover != nil {
		data.WorkCover = *in.WorkCover
	}
	if in.WorkUrl != nil {
		data.WorkUrl = *in.WorkUrl
	}
	if in.WorkDept != nil {
		data.WorkDept = *in.WorkDept
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.WorkDocumentUrl != nil {
		data.WorkDocumentUrl = *in.WorkDocumentUrl
	}
}

// derefStr *string → string。
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
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

// parseQueryRule QueryGenerator 值规则纯判定（与既有域同款）。
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
