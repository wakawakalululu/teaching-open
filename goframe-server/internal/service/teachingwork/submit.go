// submit.go 作业提交/发送/点赞/标签/班级日志计数（契约出处见各方法注释）。
package teachingwork

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// UserInfo 作品作者信息卡（anon；仅 realname/sex/avatar/school 四键）。
// 契约蓝本 TeachingWorkController#userInfo：用户不存在 →「参数错误」。
func (s *Service) UserInfo(ctx context.Context, userID string) (map[string]interface{}, error) {
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Id, userID).Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user.Id == "" {
		return nil, ErrUserNotFound
	}
	return map[string]interface{}{
		"realname": user.Realname,
		"sex":      user.Sex,
		"avatar":   user.Avatar,
		"school":   user.School,
	}, nil
}

// Submit 提交作业（四路匹配）。
//
// 契约蓝本 TeachingWorkController#submit：
//  1. 旧作匹配：body id（user_id+id）→ additionalId → courseId → 兜底
//     work_name+user_id+work_type（courseId 分支同课程多作业互相覆盖，现状保留）；
//  2. 强制 id 空、userId=当前用户；courseId → getUserDepartIdByUnitId 反查班级；
//  3. 命中旧作：复用旧 id + createTime=now + 写 sys_data_log →「更新成功！」；
//     否则「添加成功！」；统一 saveOrUpdate（id 空 insert / 非空 update）；
//  4. 班级日志 Redis 幂等键 departLog:addiWorkSubmit|courseWorkSubmit:{departId}。
func (s *Service) Submit(ctx context.Context, operatorID string, in WorkInput) (string, error) {
	// 四路匹配（优先级逐语义移植）
	oldID := ""
	if in.Id != "" {
		oldID = s.findOldWork(ctx, "id", operatorID, in.Id)
	} else if in.AdditionalId != nil && *in.AdditionalId != "" {
		oldID = s.findOldWork(ctx, "additional_id", operatorID, *in.AdditionalId)
	} else if in.CourseId != nil && *in.CourseId != "" {
		oldID = s.findOldWork(ctx, "course_id", operatorID, *in.CourseId)
	} else {
		oldID = s.findOldWorkByName(ctx, operatorID, derefStr(in.WorkName), derefStr(in.WorkType))
	}
	data := do.TeachingWork{UserId: operatorID}
	if in.DepartId != nil {
		data.DepartId = *in.DepartId
	}
	if in.CourseId != nil && *in.CourseId != "" {
		// courseId 分支：单元 → 课程授权 → 用户班级 反查（getUserDepartIdByUnitId）
		if departID, err := s.userDepartIDByUnitID(ctx, operatorID, *in.CourseId); err == nil && departID != "" {
			data.DepartId = departID
			in.DepartId = &departID
		}
	}
	applyWorkInput(&data, in)
	data.UserId = operatorID
	isUpdate := false
	if oldID != "" {
		data.Id = oldID
		data.CreateTime = gtime.Now()
		isUpdate = true
	} else {
		data.Id = guid.S()
	}
	// MP saveOrUpdate：id 空 → insert；非空 → updateById
	cols := dao.TeachingWork.Columns()
	var err error
	if isUpdate {
		_, err = dao.TeachingWork.Ctx(ctx).Where(cols.Id, oldID).Data(data).Update()
	} else {
		_, err = dao.TeachingWork.Ctx(ctx).Data(data).Insert()
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrInternal.Error(), err)
	}
	if isUpdate {
		// 保留原作品历史（sys_data_log 版本自增）
		if err := s.addDataLog(ctx, "teaching_work", oldID, workSnapshot(data)); err != nil {
			return "", err
		}
	}
	// 班级日志计数（Redis 幂等键 + 日计数行）
	departID := derefStr(in.DepartId)
	if in.AdditionalId != nil && *in.AdditionalId != "" && departID != "" {
		if err := s.departLogCount(ctx, departID, "addiWorkSubmit",
			"additional_work_submit_count", data.Id.(string)); err != nil {
			return "", err
		}
	}
	if in.CourseId != nil && *in.CourseId != "" && departID != "" {
		if err := s.departLogCount(ctx, departID, "courseWorkSubmit",
			"course_work_submit_count", data.Id.(string)); err != nil {
			return "", err
		}
	}
	if isUpdate {
		return "更新成功！", nil
	}
	return "添加成功！", nil
}

// findOldWork 按 (user_id + 列=值) 查旧作 id（selectByMap 等价，取首条）。
func (s *Service) findOldWork(ctx context.Context, column, userID, val string) string {
	col := map[string]string{"id": "id", "additional_id": "additional_id", "course_id": "course_id"}[column]
	rec, err := g.DB().Model("teaching_work").Fields("id").
		Where("user_id", userID).Where(col, val).Limit(1).One()
	if err != nil || rec.IsEmpty() {
		return ""
	}
	return rec["id"].String()
}

// findOldWorkByName 兜底匹配：work_name+user_id+work_type。
func (s *Service) findOldWorkByName(ctx context.Context, userID, workName, workType string) string {
	if workName == "" {
		return ""
	}
	m := g.DB().Model("teaching_work").Fields("id").
		Where("work_name", workName).Where("user_id", userID)
	if workType != "" {
		m = m.Where("work_type", workType)
	}
	rec, err := m.Limit(1).One()
	if err != nil || rec.IsEmpty() {
		return ""
	}
	return rec["id"].String()
}

// userDepartIDByUnitID 单元 → 课程授权 → 用户班级（getUserDepartIdByUnitId SQL）。
func (s *Service) userDepartIDByUnitID(ctx context.Context, userID, unitID string) (string, error) {
	const query = `SELECT dept_id FROM teaching_course_dept
WHERE course_id = (SELECT course_id FROM teaching_course_unit WHERE id = ?)
 AND dept_id IN (SELECT dep_id FROM sys_user_depart WHERE user_id = ?) LIMIT 1`
	rec, err := g.DB().GetOne(ctx, query, unitID, userID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrInternal.Error(), err)
	}
	if rec.IsEmpty() {
		return "", nil
	}
	return rec["dept_id"].String(), nil
}

// addDataLog 数据变更历史（对齐 SysDataLogServiceImpl#addDataLog：版本号
// = 现有最大版本 + 1，无则 0）。
func (s *Service) addDataLog(ctx context.Context, table, dataID, content string) error {
	rec, err := g.DB().Model("sys_data_log").
		Fields("MAX(data_version) AS v").
		Where("data_table", table).Where("data_id", dataID).One()
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal.Error(), err)
	}
	version := 0
	if !rec.IsEmpty() && rec["v"].String() != "" {
		version = rec["v"].Int()
	}
	if _, err := g.DB().Model("sys_data_log").Data(g.Map{
		"id":           guid.S(),
		"data_table":   table,
		"data_id":      dataID,
		"data_content": content,
		"data_version": fmt.Sprintf("%d", version+1),
	}).Insert(); err != nil {
		return fmt.Errorf("%s: %w", ErrInternal.Error(), err)
	}
	return nil
}

// workSnapshot data_log 快照（Java JSONObject.toJSONString(teachingWork)）。
func workSnapshot(data do.TeachingWork) string {
	return fmt.Sprintf(`{"id":"%v","userId":"%v","departId":"%v","courseId":"%v","workName":"%v"}`,
		toString(data.Id), toString(data.UserId), toString(data.DepartId),
		toString(data.CourseId), toString(data.WorkName))
}

// toString interface{} → string（nil → ""）。
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", v)
}

// departLogCount 班级日志计数（Redis set 幂等 + 日计数行自增）。
// typeID ∈ addiWorkSubmit|courseWorkSubmit|addiWorkCorrect|courseWorkCorrect|addiWorkAssign。
func (s *Service) departLogCount(ctx context.Context, departID, typeID, countColumn, workID string) error {
	key := fmt.Sprintf(keyDepartLogFmt, typeID, departID)
	if s.kv.SHasKey(ctx, key, workID) {
		return nil
	}
	s.kv.SAdd(ctx, key, workID)
	// 日计数行（对齐 TeachingDepartDayLogServiceImpl#addLog：depart_id +
	// create_time=yyyyMMdd 查行；Java 格式错位导致实际恒新插，逐语义保留）
	today := gtime.Now().Format("Ymd")
	rec, err := g.DB().Model("teaching_depart_day_log").
		Where("depart_id", departID).
		Where("create_time", today).
		Limit(1).One()
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInternal.Error(), err)
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
		if _, err := g.DB().Model("teaching_depart_day_log").Data(init).Insert(); err != nil {
			return fmt.Errorf("%s: %w", ErrInternal.Error(), err)
		}
		return nil
	}
	if _, err := g.DB().Model("teaching_depart_day_log").
		Where("id", rec["id"].String()).
		Data(g.Map{countColumn: gdb.Raw(fmt.Sprintf("%s + 1", countColumn))}).
		Update(); err != nil {
		return fmt.Errorf("%s: %w", ErrInternal.Error(), err)
	}
	return nil
}

// SendWork 发送作业给用户列表（逐用户复制，无事务、吞异常计数——Java 现状）。
//
// 契约蓝本 TeachingWorkServiceImpl#sendWork：源作不存在 → 0；复制时
// viewNum=0/delFlag=0/createBy=学生 username/createTime=now/update 清空，
// 按 work_name+user_id 覆盖旧作（不含 work_type）。
func (s *Service) SendWork(ctx context.Context, sendWorkID string, userIDList []string) (int, error) {
	var source entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, sendWorkID).Scan(&source); err != nil {
		return 0, fmt.Errorf("查询源作业失败: %w", err)
	}
	if source.Id == "" {
		return 0, nil
	}
	count := 0
	for _, userID := range userIDList {
		var user entity.SysUser
		if err := dao.SysUser.Ctx(ctx).
			Where(dao.SysUser.Columns().Id, userID).Scan(&user); err != nil || user.Id == "" {
			continue
		}
		// 按 work_name+user_id 覆盖旧作（不含 work_type，对齐 Java）
		newID := s.findOldWorkByName(ctx, userID, source.WorkName, "")
		isUpdate := newID != ""
		if !isUpdate {
			newID = guid.S()
		}
		data := do.TeachingWork{
			Id:         newID,
			UserId:     userID,
			DepartId:   source.DepartId,
			CourseId:   source.CourseId,
			WorkName:   source.WorkName,
			WorkType:   source.WorkType,
			WorkFile:   source.WorkFile,
			WorkCover:  source.WorkCover,
			WorkScene:  source.WorkScene,
			CreateBy:   user.Username,
			CreateTime: gtime.Now(),
			ViewNum:    0,
			DelFlag:    0,
		}
		cols := dao.TeachingWork.Columns()
		var err error
		if isUpdate {
			_, err = dao.TeachingWork.Ctx(ctx).Where(cols.Id, newID).Data(data).Update()
		} else {
			_, err = dao.TeachingWork.Ctx(ctx).Data(data).Insert()
		}
		if err == nil {
			count++
		}
	}
	return count, nil
}

// StarWork 点赞作品（anon；Redis starWork:{workId}{ip} 24h 幂等）。
// 返回话术：点赞成功 / 已经点赞过了 / 未找到对作业。
func (s *Service) StarWork(ctx context.Context, workID, clientIP string) (string, error) {
	var work entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, workID).Scan(&work); err != nil {
		return "", fmt.Errorf("查询作业失败: %w", err)
	}
	if work.Id == "" {
		return "", ErrWorkNotFound
	}
	key := fmt.Sprintf(keyStarFmt, workID, clientIP)
	if _, exists := s.kv.Get(ctx, key); exists {
		return "已经点赞过了", nil
	}
	if _, err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, workID).
		Data(do.TeachingWork{StarNum: work.StarNum + 1}).Update(); err != nil {
		return "", fmt.Errorf("更新点赞数失败: %w", err)
	}
	s.kv.SetEx(ctx, key, "1", starWorkTTL)
	return "点赞成功", nil
}

// GetWorkTags 当前用户的标签集合（work:tag:{uid} set）。
func (s *Service) GetWorkTags(ctx context.Context, operatorID string) []string {
	return s.kv.SGet(ctx, fmt.Sprintf(keyUserTagFmt, operatorID))
}

// SetWorkTag 设置作品标签（三键族联动；workTag 为空时仅清 string 键；
// oldTag != newTag 时从旧标签 set 移除——Java 现状含 oldTag 为 null 时
// 格式化出 "null" 键的行为，逐语义保留）。
func (s *Service) SetWorkTag(ctx context.Context, operatorID, workID, workTag string) error {
	var work entity.TeachingWork
	if err := dao.TeachingWork.Ctx(ctx).
		Where(dao.TeachingWork.Columns().Id, workID).Scan(&work); err != nil {
		return fmt.Errorf("查询作业失败: %w", err)
	}
	if work.Id == "" {
		return ErrWorkNotFound
	}
	keyID := fmt.Sprintf(keyWorkTagFmt, operatorID, workID)
	keyUserTag := fmt.Sprintf(keyUserTagFmt, operatorID)
	oldTag, _ := s.kv.Get(ctx, keyID)
	if workTag != "" {
		s.kv.Set(ctx, keyID, workTag)
		s.kv.SAdd(ctx, fmt.Sprintf(keyWorkTagFmt, operatorID, workTag), workID)
		s.kv.SAdd(ctx, keyUserTag, workTag)
	} else {
		s.kv.Del(ctx, keyID)
	}
	if workTag != oldTag {
		s.kv.SRemove(ctx, fmt.Sprintf(keyWorkTagFmt, operatorID, oldTag), workID)
	}
	return nil
}

// DelWorkTag 删除标签（force=false 且标签使用中 → ErrTagInUse）。
func (s *Service) DelWorkTag(ctx context.Context, operatorID, tag string, force bool) error {
	if !force {
		if members := s.kv.SGet(ctx, fmt.Sprintf(keyWorkTagFmt, operatorID, tag)); len(members) > 0 {
			return ErrTagInUse
		}
	}
	s.kv.SRemove(ctx, fmt.Sprintf(keyUserTagFmt, operatorID), tag)
	return nil
}

// additionalWorkRecord 我的附加作业输出行（AdditionalWorkModel 消费面）。
type additionalWorkRecord struct {
	Id               string `json:"id"`
	AdditionalWorkId string `json:"additionalWorkId"`
	CodeType         int    `json:"codeType"`
	WorkName         string `json:"workName"`
	WorkDesc         string `json:"workDesc"`
	WorkCover        string `json:"workCover"`
	WorkUrl          string `json:"workUrl"`
	WorkDocumentUrl  string `json:"workDocumentUrl"`
	WorkDept         string `json:"workDept"`
	Status           int    `json:"status"`
	CreateTime       string `json:"createTime"`
	CreateBy         string `json:"createBy"`
	MineWorkId       string `json:"mineWorkId"`
	MineWorkName     string `json:"mineWorkName"`
	MineWorkStatus   *int   `json:"mineWorkStatus"`
	MineWorkUrl      string `json:"mineWorkUrl"`
	MineWorkCover    string `json:"mineWorkCover"`
	DepartId         string `json:"departId"`
	DepartName       string `json:"departName"`
	Comment          string `json:"comment"`
	Score            *int   `json:"score"`
}

// MineAdditionalWork 我的附加作业列表。
//
// 契约蓝本 TeachingWorkController#mineAdditionalWork + userAdditionalWork SQL：
//   - departId 参数因 Java stream 无终结操作而不生效（现状保留）；
//   - submit=false 分支的 OR 优先级（括号缺失：`... and w.id is null or
//     w.work_status<1`）逐语义保留；
//   - mineWorkUrl/cover 为 sys_file id → 控制器换成 domain+"/"+file_path
//     （Java 双前缀 bug 下 _url 键与该值相同，Go 直接输出等价值）；
//   - workDocumentUrl 走 ow365 预览串（fileURLFn 注入）。
func (s *Service) MineAdditionalWork(ctx context.Context, operatorID, submit string, status int,
	owFileURLStr func(string) string) ([]additionalWorkRecord, error) {
	// 用户所属部门 id 集合（userAdditionalWork 的 departIds 入参）
	deptIDs, err := g.DB().Model("sys_user_depart").Fields("dep_id").
		Where("user_id", operatorID).Array()
	if err != nil {
		return nil, fmt.Errorf("查询用户部门失败: %w", err)
	}
	if len(deptIDs) == 0 {
		return []additionalWorkRecord{}, nil
	}
	orConds := make([]string, 0, len(deptIDs))
	args := make([]interface{}, 0, len(deptIDs)+2)
	for _, id := range deptIDs {
		orConds = append(orConds, "a.work_dept LIKE CONCAT('%',?,'%')")
		args = append(args, id.String())
	}
	where := "a.status = 1 AND (" + strings.Join(orConds, " OR ") + ")"
	switch submit {
	case "true":
		where += " AND w.work_status >= 1"
		if status > 0 || (status == 0 && submit == "true") {
			where += " AND w.work_status = ?"
			args = append(args, status)
		}
	case "false":
		// Java 原文：and w.id is null or w.work_status < 1（OR 优先级括号缺失，
		// 逐语义保留）
		where += " AND w.id IS NULL OR w.work_status < 1"
	}
	query := `SELECT a.*, a.id AS additionalWorkId,
 w.id AS mineWorkId, w.work_name AS mineWorkName, w.work_status AS mineWorkStatus,
 w.work_file AS mineWorkUrl, w.work_cover AS mineWorkCover,
 c.score, c.comment
FROM teaching_additional_work a
LEFT JOIN teaching_work w ON w.additional_id = a.id AND w.user_id = ?
LEFT JOIN teaching_work_correct c ON c.work_id = w.id
WHERE ` + where + " ORDER BY a.create_time DESC"
	allArgs := append([]interface{}{operatorID}, args...)
	var rows []struct {
		entity.TeachingAdditionalWork
		AdditionalWorkId string `json:"additionalWorkId"`
		MineWorkId       string `json:"mineWorkId"`
		MineWorkName     string `json:"mineWorkName"`
		MineWorkStatus   *int   `json:"mineWorkStatus"`
		MineWorkUrl      string `json:"mineWorkUrl"`
		MineWorkCover    string `json:"mineWorkCover"`
		Score            *int   `json:"score"`
		Comment          string `json:"comment"`
	}
	if err := g.DB().GetScan(ctx, &rows, query, allArgs...); err != nil {
		return nil, fmt.Errorf("查询我的附加作业失败: %w", err)
	}
	out := make([]additionalWorkRecord, 0, len(rows))
	for _, r := range rows {
		rec := additionalWorkRecord{
			Id:               r.Id,
			AdditionalWorkId: r.Id,
			CodeType:         r.CodeType,
			WorkName:         r.WorkName,
			WorkDesc:         r.WorkDesc,
			WorkCover:        r.WorkCover,
			WorkUrl:          r.WorkUrl,
			WorkDocumentUrl:  r.WorkDocumentUrl,
			WorkDept:         r.WorkDept,
			Status:           r.Status,
			CreateTime:       r.CreateTime,
			CreateBy:         r.CreateBy,
			MineWorkId:       r.MineWorkId,
			MineWorkName:     r.MineWorkName,
			MineWorkStatus:   r.MineWorkStatus,
			MineWorkUrl:      r.MineWorkUrl,
			MineWorkCover:    r.MineWorkCover,
			Comment:          r.Comment,
			Score:            r.Score,
		}
		// mineWorkUrl/cover 为 sys_file.id → 换 domain+"/"+file_path（对齐 Java）
		if rec.MineWorkUrl != "" {
			if fp := s.filePathByID(ctx, rec.MineWorkUrl); fp != "" {
				rec.MineWorkUrl = s.fileURL(fp)
			}
		}
		if rec.MineWorkCover != "" {
			if fp := s.filePathByID(ctx, rec.MineWorkCover); fp != "" {
				rec.MineWorkCover = s.fileURL(fp)
			}
		}
		if rec.WorkDocumentUrl != "" && owFileURLStr != nil {
			rec.WorkDocumentUrl = owFileURLStr(rec.WorkDocumentUrl)
		}
		out = append(out, rec)
	}
	return out, nil
}

// filePathByID 按 sys_file.id 查 file_path。
func (s *Service) filePathByID(ctx context.Context, id string) string {
	if id == "" {
		return ""
	}
	rec, err := g.DB().Model("sys_file").Fields("file_path").
		Where("id", id).Limit(1).One()
	if err != nil || rec.IsEmpty() {
		return ""
	}
	return rec["file_path"].String()
}
