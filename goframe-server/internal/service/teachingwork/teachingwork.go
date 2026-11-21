// Package teachingwork 承载作业域业务用例
// （契约蓝本：TeachingWorkController 的 22 个核心端点；exportXls/importExcel
// 有意未迁移；@PermissionData 切面暂未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - submit 四路匹配优先级：body id（user_id+id）→ additionalId → courseId →
//     work_name+user_id+work_type 兜底；命中旧作复用旧 id + createTime=now +
//     写 sys_data_log；courseId 分支经 getUserDepartIdByUnitId 反查班级；
//   - Redis 键格式与 Java 完全一致（work:tag:%s:%s / work:tag:%s /
//     starWork:{workId}{ip} / departLog:*:{departId}），兼容灰度双写；
//     REDIS_ENABLED=false 时走进程内存兜底（sms 域先例）；
//   - sendWork 无事务逐用户复制（按 work_name+user_id 覆盖旧作），吞异常计数；
//   - listWorkModel 的 DISTINCT + 批改表 JOIN（work_id 唯一键保证不重）；
//   - leaderboard 的 orderBy switch 仅认 view/time/star（未知值无排序，
//     前端已传 create_time 的行为逐语义保留）；
//   - DictAspect 注入：IPage/DictResult 响应附带 @Dict 表字典与 work_type/
//     work_status/_url 键；queryById/submit/add/edit 返回不注入。
package teachingwork

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// Redis 键格式（与 Java CacheConstant/IPUtils 组装逐字符一致）。
const (
	keyWorkTagFmt   = "work:tag:%s:%s"  // WORK_TAG：uid+workId → 标签值 / uid+tag → 作品id集合
	keyUserTagFmt   = "work:tag:%s"     // USER_WORK_TAG：uid → 标签集合
	keyStarFmt      = "starWork:%s%s"   // workId + ip（无冒号分隔，对齐 Java）
	keyDepartLogFmt = "departLog:%s:%s" // 计数类型:departId
	starWorkTTL     = 24 * time.Hour
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrUserNotFound 用户不存在（Java userInfo error 原话术）。
	ErrUserNotFound = errors.New("参数错误")
	// ErrWorkNotFound 作业不存在（Java starWork/studentWorkInfo/setWorkTag 原话术）。
	ErrWorkNotFound = errors.New("未找到对作业")
	// ErrNotFound 数据不存在（Java edit/queryById error 原话术）。
	ErrNotFound = errors.New("未找到对应数据")
	// ErrNoDepartClass 非管理员无负责班级（Java /list error 原话术）。
	ErrNoDepartClass = errors.New("您没有负责的班级")
	// ErrTagInUse 标签使用中（Java delWorkTag error 原话术）。
	ErrTagInUse = errors.New("标签正在使用中,是否确认删除？")
	// ErrInternal 系统内部错误（Java submit catch 原话术）。
	ErrInternal = errors.New("系统内部错误")
)

// WorkInput submit/add/edit 请求体（TeachingWork 可编辑字段）。
type WorkInput struct {
	Id           string
	DepartId     *string
	CourseId     *string
	AdditionalId *string
	WorkName     *string
	WorkType     *string
	WorkFile     *string
	WorkCover    *string
	WorkStatus   *int
	WorkScene    *string
	HasCloudData *bool
	ViewNum      *int
	StarNum      *int
	CollectNum   *int
	DelFlag      *int
}

// CorrectInput 批改子表输入。
type CorrectInput struct {
	Id      string
	Score   *int
	Comment *string
}

// CommentInput 评论子表输入。
type CommentInput struct {
	Id      string
	Comment *string
	UserId  string
}

// WorkPageInput /add /edit 请求体（TeachingWorkPage：主表 + 两个子表数组）。
type WorkPageInput struct {
	Work                    WorkInput
	TeachingWorkCorrectList []CorrectInput
	TeachingWorkCommentList []CommentInput
}

// workRow listWorkModel/studentWorkInfo 联表行（实体列 + 注入列）。
type workRow struct {
	entity.TeachingWork
	Username       string `json:"username"`
	Realname       string `json:"realname"`
	Avatar         string `json:"avatar"`
	Sex            int    `json:"sex"`
	WorkFileKey    string `json:"workFileKey"`
	CoverFileKey   string `json:"coverFileKey"`
	TeacherComment string `json:"teacherComment"`
	Score          *int   `json:"score"`
	WorkId         string `json:"workId"`
}

// workModel IPage/DictResult 白名单输出行（实体 + 联表列 + DictAspect 注入键
// + workTag 回填）。
type workModel struct {
	entity.TeachingWork
	Username             string `json:"username"`
	Realname             string `json:"realname"`
	Avatar               string `json:"avatar"`
	Sex                  int    `json:"sex"`
	WorkFileKey          string `json:"workFileKey"`
	CoverFileKey         string `json:"coverFileKey"`
	TeacherComment       string `json:"teacherComment"`
	Score                *int   `json:"score"`
	WorkId               string `json:"workId"`
	WorkTag              string `json:"workTag"`
	UserIdDictText       string `json:"userId_dictText"`
	DepartIdDictText     string `json:"departId_dictText"`
	CourseIdDictText     string `json:"courseId_dictText"`
	AdditionalIdDictText string `json:"additionalId_dictText"`
	WorkTypeDictText     string `json:"workType_dictText"`
	WorkStatusDictText   string `json:"workStatus_dictText"`
	WorkFileKeyURL       string `json:"workFileKey_url"`
	CoverFileKeyURL      string `json:"coverFileKey_url"`
	AvatarURL            string `json:"avatar_url"`
}

// commentRow getWorkComments 输出行（c.* + username/realname/avatar +
// avatar_url 注入）。
type commentRow struct {
	entity.TeachingWorkComment
	Username  string `json:"username"`
	Realname  string `json:"realname"`
	Avatar    string `json:"avatar"`
	AvatarURL string `json:"avatar_url"`
}

// Service 作业域业务服务。
type Service struct {
	kv           KVStore
	staticDomain string
	qiniuDomain  string
	uploadType   string
}

// New 构造服务。
func New(redisEnabled bool, staticDomain, qiniuDomain, uploadType string) *Service {
	return &Service{
		kv:           NewKVStore(redisEnabled),
		staticDomain: staticDomain,
		qiniuDomain:  qiniuDomain,
		uploadType:   uploadType,
	}
}

// fileURL 文件展示地址（与既有域同款语义）。
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

// listWorkModelSQL listWorkModel 查询（对齐 TeachingWorkMapper#listWorkModel）。
const listWorkModelSQL = `SELECT DISTINCT teaching_work.id AS workId, teaching_work.*,
 c.comment AS teacherComment, c.score,
 sys_user.username, sys_user.realname, sys_user.avatar,
 (SELECT file_path FROM sys_file WHERE sys_file.id = teaching_work.work_file) AS workFileKey,
 (SELECT file_path FROM sys_file WHERE sys_file.id = teaching_work.work_cover) AS coverFileKey
FROM teaching_work
LEFT JOIN sys_user ON teaching_work.user_id = sys_user.id
LEFT JOIN teaching_work_correct c ON c.work_id = teaching_work.id`

// pageWorkRows 分页执行 listWorkModel（where 由调用方组装，列名为代码内
// 常量、值全占位符；deptIDs 非空时加部门 JOIN）。
func (s *Service) pageWorkRows(ctx context.Context, param page.Param, where string, args []interface{}, deptIDs []string, order string) ([]workRow, int, error) {
	join := ""
	joinArgs := make([]interface{}, 0, len(deptIDs))
	if len(deptIDs) > 0 {
		ph := make([]string, 0, len(deptIDs))
		for _, id := range deptIDs {
			ph = append(ph, "?")
			joinArgs = append(joinArgs, id)
		}
		join = " JOIN sys_user_depart ud ON sys_user.id = ud.user_id AND ud.dep_id IN (" +
			strings.Join(ph, ",") + ")"
	}
	fullWhere := " WHERE 1=1"
	if where != "" {
		fullWhere += " AND (" + where + ")"
	}
	fullArgs := append(append([]interface{}{}, args...), joinArgs...)
	total, err := g.DB().GetOne(ctx,
		"SELECT COUNT(DISTINCT teaching_work.id) AS n FROM teaching_work "+
			"LEFT JOIN sys_user ON teaching_work.user_id = sys_user.id "+
			"LEFT JOIN teaching_work_correct c ON c.work_id = teaching_work.id"+join+fullWhere,
		fullArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("统计作业数量失败: %w", err)
	}
	if total.IsEmpty() || total["n"].Int() == 0 {
		return []workRow{}, 0, nil
	}
	totalN := total["n"].Int()
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}
	query := listWorkModelSQL + join + fullWhere
	if order != "" {
		query += " ORDER BY " + order
	}
	query += " LIMIT ? OFFSET ?"
	fullArgs = append(fullArgs, param.PageSize, offset)
	var rows []workRow
	if err := g.DB().GetScan(ctx, &rows, query, fullArgs...); err != nil {
		return nil, 0, fmt.Errorf("查询作业列表失败: %w", err)
	}
	return rows, totalN, nil
}

// toWorkModels 联表行 → 输出模型（DictAspect 注入 + workTag 回填）。
func (s *Service) toWorkModels(ctx context.Context, rows []workRow, operatorID string) ([]workModel, error) {
	typeTexts, err := s.dictTextMap(ctx, "work_type")
	if err != nil {
		return nil, err
	}
	statusTexts, err := s.dictTextMap(ctx, "work_status")
	if err != nil {
		return nil, err
	}
	userNames, err := s.dictTableMap(ctx, "sys_user", "id", "realname", collectIDs(rows,
		func(r workRow) string { return r.UserId }))
	if err != nil {
		return nil, err
	}
	departNames, err := s.dictTableMap(ctx, "sys_depart", "id", "depart_name", collectIDs(rows,
		func(r workRow) string { return r.DepartId }))
	if err != nil {
		return nil, err
	}
	unitNames, err := s.dictTableMap(ctx, "teaching_course_unit", "id", "unit_name", collectIDs(rows,
		func(r workRow) string { return r.CourseId }))
	if err != nil {
		return nil, err
	}
	additionalNames, err := s.dictTableMap(ctx, "teaching_additional_work", "id", "work_name", collectIDs(rows,
		func(r workRow) string { return r.AdditionalId }))
	if err != nil {
		return nil, err
	}
	out := make([]workModel, 0, len(rows))
	for _, r := range rows {
		m := workModel{
			TeachingWork:         r.TeachingWork,
			Username:             r.Username,
			Realname:             r.Realname,
			Avatar:               r.Avatar,
			Sex:                  r.Sex,
			WorkFileKey:          r.WorkFileKey,
			CoverFileKey:         r.CoverFileKey,
			TeacherComment:       r.TeacherComment,
			Score:                r.Score,
			WorkId:               r.WorkId,
			UserIdDictText:       userNames[r.UserId],
			DepartIdDictText:     departNames[r.DepartId],
			CourseIdDictText:     unitNames[r.CourseId],
			AdditionalIdDictText: additionalNames[r.AdditionalId],
			WorkTypeDictText:     typeTexts[r.WorkType],
			WorkStatusDictText:   statusTexts[intKey(r.WorkStatus)],
			WorkFileKeyURL:       s.fileURL(r.WorkFileKey),
			CoverFileKeyURL:      s.fileURL(r.CoverFileKey),
			AvatarURL:            s.fileURL(r.Avatar),
		}
		if tag, ok := s.kv.Get(ctx, fmt.Sprintf(keyWorkTagFmt, operatorID, r.Id)); ok {
			m.WorkTag = tag
		}
		out = append(out, m)
	}
	return out, nil
}

// Mine 我的作业分页（user_id=当前用户 + create_time 倒序 + workTag 过滤）。
func (s *Service) Mine(ctx context.Context, operatorID string, param page.Param,
	workName, workType, workStatus, workScene, workTag string) (*page.Result, error) {
	where := "teaching_work.user_id = ?"
	args := []interface{}{operatorID}
	if workName != "" {
		where += " AND teaching_work.work_name LIKE ?"
		args = append(args, "%"+workName+"%")
	}
	if workType != "" {
		where += " AND teaching_work.work_type = ?"
		args = append(args, workType)
	}
	if workStatus != "" {
		where += " AND teaching_work.work_status = ?"
		args = append(args, workStatus)
	}
	if workScene != "" {
		where += " AND teaching_work.work_scene = ?"
		args = append(args, workScene)
	}
	if workTag != "" {
		tagIDs := s.kv.SGet(ctx, fmt.Sprintf(keyWorkTagFmt, operatorID, workTag))
		if len(tagIDs) > 0 {
			where += " AND teaching_work.id IN (" + placeholders(len(tagIDs)) + ")"
			for _, id := range tagIDs {
				args = append(args, id)
			}
		}
	}
	rows, total, err := s.pageWorkRows(ctx, param, where, args, nil,
		"teaching_work.create_time DESC")
	if err != nil {
		return nil, err
	}
	records, err := s.toWorkModels(ctx, rows, operatorID)
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// List 作业管理分页（教师端）。
func (s *Service) List(ctx context.Context, operatorID string, param page.Param,
	username, workName, realname, workScene, workTag, begin, end, workStatus string) (*page.Result, error) {
	where := "1=1"
	args := make([]interface{}, 0, 8)
	if username != "" {
		where += " AND teaching_work.create_by = ?"
		args = append(args, username)
	}
	if workName != "" {
		where += " AND teaching_work.work_name LIKE ?"
		args = append(args, "%"+workName+"%")
	}
	if realname != "" {
		where += " AND sys_user.realname LIKE ?"
		args = append(args, "%"+realname+"%")
	}
	if workScene != "" {
		where += " AND teaching_work.work_scene = ?"
		args = append(args, workScene)
	}
	if workStatus != "" {
		where += " AND teaching_work.work_status = ?"
		args = append(args, workStatus)
	}
	if begin != "" {
		where += " AND teaching_work.create_time >= ?"
		args = append(args, begin)
	}
	if end != "" {
		where += " AND teaching_work.create_time <= ?"
		args = append(args, end)
	}
	var deptIDs []string
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
		deptIDs = myIDs
	}
	if workTag != "" {
		tagIDs := s.kv.SGet(ctx, fmt.Sprintf(keyWorkTagFmt, operatorID, workTag))
		if len(tagIDs) > 0 {
			where += " AND teaching_work.id IN (" + placeholders(len(tagIDs)) + ")"
			for _, id := range tagIDs {
				args = append(args, id)
			}
		}
	}
	rows, total, err := s.pageWorkRows(ctx, param, where, args, deptIDs, "")
	if err != nil {
		return nil, err
	}
	records, err := s.toWorkModels(ctx, rows, operatorID)
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// GreatWork 精选作品分页（固定 work_status=2，anon）。
func (s *Service) GreatWork(ctx context.Context, param page.Param) (*page.Result, error) {
	rows, total, err := s.pageWorkRows(ctx, param, "teaching_work.work_status = ?", []interface{}{2}, nil, "")
	if err != nil {
		return nil, err
	}
	records, err := s.toWorkModels(ctx, rows, "")
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// Leaderboard 作品排行分页（固定 work_status>=3；orderBy 仅认 view/time/star，
// 未知值无排序——Java switch 现状逐语义保留）。
func (s *Service) Leaderboard(ctx context.Context, param page.Param,
	orderBy, workStatus, userID string) (*page.Result, error) {
	where := "teaching_work.work_status >= 3"
	args := make([]interface{}, 0, 2)
	if userID != "" {
		where += " AND teaching_work.user_id = ?"
		args = append(args, userID)
	}
	if workStatus != "" {
		where += " AND teaching_work.work_status = ?"
		args = append(args, workStatus)
	}
	order := leaderboardOrder(orderBy)
	rows, total, err := s.pageWorkRows(ctx, param, where, args, nil, order)
	if err != nil {
		return nil, err
	}
	records, err := s.toWorkModels(ctx, rows, "")
	if err != nil {
		return nil, err
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// StudentWorkInfo 作品详情（anon；浏览数 +1；DictResult 单对象注入）。
func (s *Service) StudentWorkInfo(ctx context.Context, workID string) (*workModel, error) {
	const query = `SELECT teaching_work.*, sys_user.username, sys_user.realname,
 sys_user.avatar, sys_user.sex,
 (SELECT file_path FROM sys_file WHERE sys_file.id = teaching_work.work_file) AS workFileKey,
 (SELECT file_path FROM sys_file WHERE sys_file.id = teaching_work.work_cover) AS coverFileKey
FROM teaching_work
LEFT JOIN sys_user ON teaching_work.user_id = sys_user.id
WHERE sys_user.del_flag = 0 AND teaching_work.id = ?`
	var rows []workRow
	if err := g.DB().GetScan(ctx, &rows, query, workID); err != nil {
		return nil, fmt.Errorf("查询作品详情失败: %w", err)
	}
	if len(rows) == 0 || rows[0].Id == "" {
		return nil, ErrWorkNotFound
	}
	if _, err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, workID).
		Data(do.TeachingWork{ViewNum: rows[0].ViewNum + 1}).Update(); err != nil {
		return nil, fmt.Errorf("更新浏览数失败: %w", err)
	}
	models, err := s.toWorkModels(ctx, rows, "")
	if err != nil {
		return nil, err
	}
	return &models[0], nil
}

// GetWorkComments 作品评论分页（anon；create_time 倒序）。
func (s *Service) GetWorkComments(ctx context.Context, workID string, page, pageSize int) ([]commentRow, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	const query = `SELECT c.*, u.username, u.realname, u.avatar
FROM teaching_work_comment c
LEFT JOIN sys_user u ON u.id = c.user_id
WHERE work_id = ?
ORDER BY c.create_time DESC
LIMIT ? OFFSET ?`
	var rows []commentRow
	if err := g.DB().GetScan(ctx, &rows, query, workID, pageSize, offset); err != nil {
		return nil, fmt.Errorf("查询作品评论失败: %w", err)
	}
	for i := range rows {
		rows[i].AvatarURL = s.fileURL(rows[i].Avatar)
	}
	return rows, nil
}

// SaveComment 保存作品评论（仅取 workId+comment，user_id=当前用户）。
func (s *Service) SaveComment(ctx context.Context, operatorID, workID, comment string) error {
	if _, err := dao.TeachingWorkComment.Ctx(ctx).Data(do.TeachingWorkComment{
		Id:      guid.S(),
		WorkId:  workID,
		Comment: comment,
		UserId:  operatorID,
	}).Insert(); err != nil {
		return fmt.Errorf("保存评论失败: %w", err)
	}
	return nil
}

// QueryByID 作业详情（单实体，不注入 dict）。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.TeachingWork, error) {
	var row entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询作业失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// QueryCorrectsByMainID 批改列表（非 IPage，不注入）。
func (s *Service) QueryCorrectsByMainID(ctx context.Context, workID string) ([]entity.TeachingWorkCorrect, error) {
	var rows []entity.TeachingWorkCorrect
	if err := dao.TeachingWorkCorrect.Ctx(ctx).
		Where(dao.TeachingWorkCorrect.Columns().WorkId, workID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询批改失败: %w", err)
	}
	return rows, nil
}

// QueryCommentsByMainID 评论列表（非 IPage，不注入）。
func (s *Service) QueryCommentsByMainID(ctx context.Context, workID string) ([]entity.TeachingWorkComment, error) {
	var rows []entity.TeachingWorkComment
	if err := dao.TeachingWorkComment.Ctx(ctx).
		Where(dao.TeachingWorkComment.Columns().WorkId, workID).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}
	return rows, nil
}

// AddMain 新增作业（主表 + 子表外键回填，单事务对齐 saveMain）。
func (s *Service) AddMain(ctx context.Context, in WorkPageInput) (*entity.TeachingWork, error) {
	data := do.TeachingWork{Id: in.Work.Id}
	if in.Work.Id == "" {
		data.Id = guid.S()
	}
	applyWorkInput(&data, in.Work)
	idStr := data.Id.(string)
	err := dao.TeachingWork.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.TeachingWork.Ctx(ctx).Data(data).Insert(); err != nil {
			return fmt.Errorf("操作失败: %w", err)
		}
		return s.insertChildren(ctx, idStr, in)
	})
	if err != nil {
		return nil, err
	}
	var row entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, idStr).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询作业失败: %w", err)
	}
	return &row, nil
}

// EditMain 编辑作业（先删两子表再重插——入参为空也删，对齐 updateMain；
// 班级日志批改计数）。
func (s *Service) EditMain(ctx context.Context, operatorID string, in WorkPageInput) error {
	cols := dao.TeachingWork.Columns()
	var old entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).Where(cols.Id, in.Work.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询作业失败: %w", err)
	}
	if old.Id == "" {
		return ErrNotFound
	}
	data := do.TeachingWork{}
	applyWorkInput(&data, in.Work)
	departID := old.DepartId
	if in.Work.DepartId != nil {
		departID = *in.Work.DepartId
	}
	if err := dao.TeachingWork.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.TeachingWork.Ctx(ctx).
			Where(cols.Id, in.Work.Id).Data(data).Update(); err != nil {
			return fmt.Errorf("操作失败: %w", err)
		}
		if _, err := dao.TeachingWorkCorrect.Ctx(ctx).
			Where(dao.TeachingWorkCorrect.Columns().WorkId, in.Work.Id).Delete(); err != nil {
			return fmt.Errorf("清空批改失败: %w", err)
		}
		if _, err := dao.TeachingWorkComment.Ctx(ctx).
			Where(dao.TeachingWorkComment.Columns().WorkId, in.Work.Id).Delete(); err != nil {
			return fmt.Errorf("清空评论失败: %w", err)
		}
		return s.insertChildren(ctx, in.Work.Id, in)
	}); err != nil {
		return err
	}
	if departID != "" {
		if in.Work.AdditionalId != nil && *in.Work.AdditionalId != "" {
			if err := s.departLogCount(ctx, departID, "addiWorkCorrect",
				"additional_work_correct_count", in.Work.Id); err != nil {
				return err
			}
		}
		if in.Work.CourseId != nil && *in.Work.CourseId != "" {
			if err := s.departLogCount(ctx, departID, "courseWorkCorrect",
				"course_work_correct_count", in.Work.Id); err != nil {
				return err
			}
		}
	}
	return nil
}

// insertChildren 子表插入（外键回填 workId）。
func (s *Service) insertChildren(ctx context.Context, workID string, in WorkPageInput) error {
	for _, c := range in.TeachingWorkCorrectList {
		cid := c.Id
		if cid == "" {
			cid = guid.S()
		}
		if _, err := dao.TeachingWorkCorrect.Ctx(ctx).Data(do.TeachingWorkCorrect{
			Id: cid, WorkId: workID, Score: derefInt(c.Score), Comment: derefStr(c.Comment),
		}).Insert(); err != nil {
			return fmt.Errorf("保存批改失败: %w", err)
		}
	}
	for _, c := range in.TeachingWorkCommentList {
		cid := c.Id
		if cid == "" {
			cid = guid.S()
		}
		if _, err := dao.TeachingWorkComment.Ctx(ctx).Data(do.TeachingWorkComment{
			Id: cid, WorkId: workID, Comment: derefStr(c.Comment), UserId: c.UserId,
		}).Insert(); err != nil {
			return fmt.Errorf("保存评论失败: %w", err)
		}
	}
	return nil
}

// DeleteMain 删除作业（子表清理 + sys_file 记录清理 + 物理删主表，单事务）。
func (s *Service) DeleteMain(ctx context.Context, id string) error {
	var work entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, id).Scan(&work); err != nil {
		return fmt.Errorf("查询作业失败: %w", err)
	}
	if work.Id == "" {
		return nil
	}
	return dao.TeachingWork.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.TeachingWorkCorrect.Ctx(ctx).
			Where(dao.TeachingWorkCorrect.Columns().WorkId, id).Delete(); err != nil {
			return fmt.Errorf("删除批改失败: %w", err)
		}
		if _, err := dao.TeachingWorkComment.Ctx(ctx).
			Where(dao.TeachingWorkComment.Columns().WorkId, id).Delete(); err != nil {
			return fmt.Errorf("删除评论失败: %w", err)
		}
		for _, key := range []string{work.WorkCover, work.WorkFile} {
			if err := s.cleanupSysFileByKey(ctx, key); err != nil {
				return err
			}
		}
		cols := dao.TeachingWork.Columns()
		if _, err := dao.TeachingWork.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
			return fmt.Errorf("删除作业失败: %w", err)
		}
		return nil
	})
}

// DeleteBatchMain 批量删除作业（逐条 delMain，单事务对齐 delBatchMain）。
func (s *Service) DeleteBatchMain(ctx context.Context, ids []string) error {
	return dao.TeachingWork.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, id := range ids {
			if err := s.DeleteMain(ctx, id); err != nil {
				return err
			}
		}
		return nil
	})
}

// cleanupSysFileByKey 按 file_path 精确匹配清理 sys_file 记录（与 course 域
// 同款语义；物理文件删除待文件域补齐）。
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

// applyWorkInput 请求体字段拷贝进 DO（UserId/CreateTime/UpdateTime 由调用方管理）。
func applyWorkInput(data *do.TeachingWork, in WorkInput) {
	if in.DepartId != nil {
		data.DepartId = *in.DepartId
	}
	if in.CourseId != nil {
		data.CourseId = *in.CourseId
	}
	if in.AdditionalId != nil {
		data.AdditionalId = *in.AdditionalId
	}
	if in.WorkName != nil {
		data.WorkName = *in.WorkName
	}
	if in.WorkType != nil {
		data.WorkType = *in.WorkType
	}
	if in.WorkFile != nil {
		data.WorkFile = *in.WorkFile
	}
	if in.WorkCover != nil {
		data.WorkCover = *in.WorkCover
	}
	if in.WorkStatus != nil {
		data.WorkStatus = *in.WorkStatus
	}
	if in.WorkScene != nil {
		data.WorkScene = *in.WorkScene
	}
	if in.HasCloudData != nil {
		data.HasCloudData = boolToInt(*in.HasCloudData)
	}
	if in.ViewNum != nil {
		data.ViewNum = *in.ViewNum
	}
	if in.StarNum != nil {
		data.StarNum = *in.StarNum
	}
	if in.CollectNum != nil {
		data.CollectNum = *in.CollectNum
	}
	if in.DelFlag != nil {
		data.DelFlag = *in.DelFlag
	}
}

// derefStr *string → string（nil → ""）。
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// derefInt *int → int（nil → 0）。
func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// boolToInt bool → tinyint 语义。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intKey int → 字典键字符串。
func intKey(v int) string {
	return fmt.Sprintf("%d", v)
}

// placeholders n 个 "?, " 占位符。
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// collectIDs 按行提取非空 id 集合。
func collectIDs(rows []workRow, get func(workRow) string) []string {
	set := make(map[string]struct{})
	for _, r := range rows {
		if v := get(r); v != "" {
			set[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
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

// dictTableMap 表字典 键值→文本 映射（@Dict dictTable 注入语义）。
func (s *Service) dictTableMap(ctx context.Context, table, codeCol, textCol string, ids []string) (map[string]string, error) {
	out := make(map[string]string)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := g.DB().Model(table).Fields(codeCol, textCol).WhereIn(codeCol, ids).All()
	if err != nil {
		return nil, fmt.Errorf("查询表字典失败: %w", err)
	}
	for _, row := range rows {
		out[row[codeCol].String()] = row[textCol].String()
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

// mySubDepartIDs 负责部门及其子部门 id（对齐 SysDepartService#getMySubDepIdsByDepId：
// 按负责部门 org_code 前缀命中子部门集合）。
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
	prefixes := strings.Split(user.DepartIds, ",")
	// 负责部门 id 串实际存部门 id：取其 org_code 作为前缀集合
	var codes []string
	for _, id := range prefixes {
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
	ids, err := g.DB().Model("sys_depart").
		Fields("id").
		Where("("+strings.Join(conds, " OR ")+")", args...).
		Array()
	if err != nil {
		return nil, fmt.Errorf("查询子部门失败: %w", err)
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out, nil
}

// leaderboardOrder 排行榜排序纯判定（便于单测）：仅认 view/time/star，
// 未知值无排序（Java switch 现状逐语义保留）。
func leaderboardOrder(orderBy string) string {
	switch orderBy {
	case "view":
		return "teaching_work.view_num DESC"
	case "time":
		return "teaching_work.create_time DESC"
	case "star":
		return "teaching_work.star_num DESC"
	}
	return ""
}
