// Package userrole 承载用户管理之「角色分配域」业务用例
// （契约蓝本：SysUserController 的 queryUserRole/userRoleList/addSysUserRole/
// deleteUserRole/deleteUserRoleBatch 五个端点）。
//
// 全部数据库操作走 dao/do 占位符绑定；sys_user_role 为纯关联表
// （仅 id/user_id/role_id，无 del_flag/create_time），删除为物理 DELETE，
// id 生成沿用仓库 guid.S() 先例（32 位 hex、无连字符，varchar(32) 兼容）。
package userrole

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误（HTTP 层映射为 Result{success:false}，话术与 Java 保持一致）。
var (
	// ErrPermissionDenied 操作人角色等级低于目标用户（Java lessThanUserRoleLevel）。
	ErrPermissionDenied = errors.New("权限不足")
	// ErrNoUserRole 用户无任何角色（Java queryUserRole error500 原话术）。
	ErrNoUserRole = errors.New("未找到用户相关角色信息")
	// ErrUserRequired 必填参数 userId 缺失（Java @RequestParam required=true）。
	ErrUserRequired = errors.New("用户ID不能为空")
)

// userRecord userRoleList 列表白名单输出行。
//
// 前端消费核对（web/src/views/system/RoleUserList.vue）：
//   - 表格列：id（rowKey）/username/realname/status_dictText；
//   - 行编辑 handleEdit2 把整行传给 UserModal.edit(record)，UserModal 消费
//     id/avatar/birthday/username/sex/realname/email/phone/activitiSync/workNo/
//     telephone/post/school/userIdentity/departIds，并经 Object.assign(model)
//     原样回传 PUT /sys/user/edit；
//   - 分页仅消费 result.records 与 result.total。
//
// 据此白名单输出下述字段；password/salt/createTime 等非消费列一律不透出。
type userRecord struct {
	Id             string `json:"id"`
	Username       string `json:"username"`
	Realname       string `json:"realname"`
	Avatar         string `json:"avatar"`
	Birthday       string `json:"birthday"`
	Sex            int    `json:"sex"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Status         int    `json:"status"`
	StatusDictText string `json:"status_dictText"`
	OrgCode        string `json:"orgCode"`
	ActivitiSync   int    `json:"activitiSync"`
	WorkNo         string `json:"workNo"`
	Post           string `json:"post"`
	School         string `json:"school"`
	Telephone      string `json:"telephone"`
	UserIdentity   int    `json:"userIdentity"`
	DepartIds      string `json:"departIds"`
}

// Service 用户角色分配业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// QueryUserRole 查询某用户已分配的角色 ID 列表。
//
// 契约蓝本 SysUserController#queryUserRole（GET /sys/user/queryUserRole）：
//   - 参数：userid（必填，用户 ID）；
//   - result：["角色ID", ...]（仅 role_id 字符串数组）；
//   - 用户无任何角色时 Java 走 result.error500("未找到用户相关角色信息")，
//     即 HTTP 200 + success:false；前端 UserModal#loadUserRoles 对该失败仅
//     console.log，不弹窗。
func (s *Service) QueryUserRole(ctx context.Context, userid string) ([]string, error) {
	cols := dao.SysUserRole.Columns()
	var rows []entity.SysUserRole
	if err := dao.SysUserRole.Ctx(ctx).
		Where(cols.UserId, userid).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, ErrNoUserRole
	}
	roleIds := make([]string, 0, len(rows))
	for _, row := range rows {
		roleIds = append(roleIds, row.RoleId)
	}
	return roleIds, nil
}

// UserRoleList 角色下用户分页列表（即「该角色已有的用户」，Java 契约无
// selected 标记字段，前端也未消费任何 selected，见 RoleUserList.vue）。
//
// 契约蓝本 SysUserController#userRoleList（GET /sys/user/userRoleList）：
//   - 参数：pageNo（默认 1）、pageSize（默认 10）、roleId、username（精确匹配）；
//   - Java SQL（SysUserMapper#getUserByRoleId）：
//     select * from sys_user where del_flag = 0
//     and id in (select user_id from sys_user_role where role_id=#{roleId})
//     [and username = #{username}]
//   - result：IPage 形状 {records,total,size,current,pages}；
//   - status_dictText 对齐 Java DictAspect 对 @Dict(dicCode="user_status") 的注入。
func (s *Service) UserRoleList(ctx context.Context, param page.Param, roleId, username string) (*page.Result, error) {
	if roleId == "" {
		// Java：role_id 为空时子查询 role_id = NULL 永不命中 → 空页。
		return page.New([]userRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	urCols := dao.SysUserRole.Columns()
	var userIds []string
	if err := dao.SysUserRole.Ctx(ctx).
		Fields(urCols.UserId).
		Where(urCols.RoleId, roleId).
		Scan(&userIds); err != nil {
		return nil, fmt.Errorf("查询角色下用户失败: %w", err)
	}
	if len(userIds) == 0 {
		return page.New([]userRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	userCols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).
		Where(userCols.DelFlag, 0).
		WhereIn(userCols.Id, userIds)
	if username != "" {
		m = m.Where(userCols.Username, username)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计角色用户数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]userRecord{}, 0, param.PageNo, param.PageSize), nil
	}
	var users []entity.SysUser
	if err := m.Page(param.PageNo, param.PageSize).Scan(&users); err != nil {
		return nil, fmt.Errorf("查询角色用户列表失败: %w", err)
	}
	statusTexts, err := s.dictTextMap(ctx, "user_status")
	if err != nil {
		return nil, err
	}
	records := make([]userRecord, 0, len(users))
	for _, u := range users {
		records = append(records, newUserRecord(u, statusTexts))
	}
	return page.New(records, total, param.PageNo, param.PageSize), nil
}

// AddSysUserRole 给指定角色添加用户（幂等：已存在的关系跳过）。
//
// 契约蓝本 SysUserController#addSysUserRole（POST /sys/user/addSysUserRole）：
//   - 请求体（Java SysUserRoleVO）：{"roleId": "角色ID", "userIdList": ["用户ID", ...]}；
//   - 语义：校验操作人角色等级不低于每个目标用户（lessThanUserRoleLevel），
//     已存在 (role_id,user_id) 关系的跳过，其余新增；成功话术「添加成功!」；
//   - Go 决策：权限校验前置为「先全量校验后写入」。Java 为逐个边校验边写，
//     中途失败会留下部分写入；前置校验无部分状态，HTTP 契约完全一致；
//   - id 生成：Java @TableId(IdType.ID_WORKER_STR) 雪花字符串，Go 侧沿用仓库
//     先例 guid.S()（32 位 hex、无连字符），列宽 varchar(32) 兼容；
//   - 入参 userIdList 为空时 Java 循环零次直接成功，此处一致；roleId 缺失属
//     前端异常，按无操作成功处理（Java 会插入 role_id=NULL 脏数据，明确不照搬）。
func (s *Service) AddSysUserRole(ctx context.Context, operatorID, roleId string, userIdList []string) error {
	if roleId == "" || len(userIdList) == 0 {
		return nil
	}
	// 权限前置校验：操作人等级不得低于任一目标用户。
	level, err := s.userRoleLevel(ctx, operatorID)
	if err != nil {
		return err
	}
	for _, uid := range userIdList {
		target, err := s.userRoleLevel(ctx, uid)
		if err != nil {
			return err
		}
		if level < target {
			return ErrPermissionDenied
		}
	}
	// 幂等写入：先查已存在关系，仅插入缺失项（Java 逐条 getOne→save 的集合化等价）。
	cols := dao.SysUserRole.Columns()
	var existed []string
	if err := dao.SysUserRole.Ctx(ctx).
		Fields(cols.UserId).
		Where(cols.RoleId, roleId).
		WhereIn(cols.UserId, userIdList).
		Scan(&existed); err != nil {
		return fmt.Errorf("查询已有用户角色失败: %w", err)
	}
	existedSet := make(map[string]struct{}, len(existed))
	for _, uid := range existed {
		existedSet[uid] = struct{}{}
	}
	for _, uid := range userIdList {
		if _, ok := existedSet[uid]; ok {
			continue
		}
		if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
			Id:     guid.S(),
			UserId: uid,
			RoleId: roleId,
		}).Insert(); err != nil {
			return fmt.Errorf("保存用户角色失败: %w", err)
		}
		// 标记本次已插入，入参重复项不再重复插入（对齐 Java 逐条 getOne 幂等）。
		existedSet[uid] = struct{}{}
	}
	return nil
}

// DeleteUserRole 删除单个用户的指定角色关系。
//
// 契约蓝本 SysUserController#deleteUserRole（DELETE /sys/user/deleteUserRole）：
//   - 参数：roleId（必填）、userId（必填），前端经 query string 传递；
//   - lessThanUserRoleLevel(userId) 为真时返回「权限不足」（HTTP 200 + success:false）；
//   - 物理 DELETE sys_user_role WHERE role_id=? AND user_id=?（该表无软删列）；
//   - 关系不存在也返回「删除成功!」（Java remove() 布尔返回值被忽略）。
func (s *Service) DeleteUserRole(ctx context.Context, operatorID, roleId, userId string) error {
	if userId == "" {
		return ErrUserRequired
	}
	if err := s.checkOperatorLevel(ctx, operatorID, userId); err != nil {
		return err
	}
	cols := dao.SysUserRole.Columns()
	if _, err := dao.SysUserRole.Ctx(ctx).
		Where(cols.RoleId, roleId).
		Where(cols.UserId, userId).
		Delete(); err != nil {
		return fmt.Errorf("删除用户角色失败: %w", err)
	}
	return nil
}

// DeleteUserRoleBatch 批量删除指定角色下的用户关系。
//
// 契约蓝本 SysUserController#deleteUserRoleBatch（DELETE /sys/user/deleteUserRoleBatch）：
//   - 参数：roleId（必填）、userIds（逗号分隔字符串，必填）；
//   - Java：userIds.split(",") 逐个校验等级后 remove WHERE role_id=? AND user_id IN (...)；
//   - Go 决策：解析健壮化——逗号拆分、去空白、丢弃空段（Java 原样 split，
//     空段进入 IN (”) 恒不命中，丢弃后语义一致且更稳）；解析后为空集时与
//     Java「IN (”) 删除 0 行」的行为对齐，直接返回成功（无操作）。
func (s *Service) DeleteUserRoleBatch(ctx context.Context, operatorID, roleId string, userIds []string) error {
	if len(userIds) == 0 {
		return nil
	}
	for _, uid := range userIds {
		if err := s.checkOperatorLevel(ctx, operatorID, uid); err != nil {
			return err
		}
	}
	cols := dao.SysUserRole.Columns()
	if _, err := dao.SysUserRole.Ctx(ctx).
		Where(cols.RoleId, roleId).
		WhereIn(cols.UserId, userIds).
		Delete(); err != nil {
		return fmt.Errorf("批量删除用户角色失败: %w", err)
	}
	return nil
}

// checkOperatorLevel 校验操作人角色等级不低于目标用户。
//
// 语义对齐 BaseController#lessThanUserRoleLevel：
// getUserRoleLevel(operator) < getUserRoleLevel(target) 即权限不足；
// 操作人上下文缺失等价 Java getCurrentUser()==null，直接判权限不足（fail-closed）。
func (s *Service) checkOperatorLevel(ctx context.Context, operatorID, targetID string) error {
	if operatorID == "" {
		return ErrPermissionDenied
	}
	level, err := s.userRoleLevel(ctx, operatorID)
	if err != nil {
		return err
	}
	target, err := s.userRoleLevel(ctx, targetID)
	if err != nil {
		return err
	}
	if level < target {
		return ErrPermissionDenied
	}
	return nil
}

// userRoleLevel 查询用户角色等级。
//
// 语义对齐 SysUserRoleMapper#getUserRoleLevel：
//
//	select max(role_level) from sys_role
//	  where id in (select role_id from sys_user_role where user_id = #{userId})
//
// 无任何角色时 Java 返回 -1（Integer null 兜底），此处一致。
// sys_role 无独立 dao，跨表聚合沿用 rbac 包的裸模型 + 占位符先例。
func (s *Service) userRoleLevel(ctx context.Context, userID string) (int, error) {
	rec, err := g.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Fields("MAX(r.role_level) AS role_level").
		Where("ur.user_id", userID).
		One()
	if err != nil {
		return -1, fmt.Errorf("查询用户角色等级失败: %w", err)
	}
	if rec.IsEmpty() || rec["role_level"].IsNil() {
		return -1, nil
	}
	return rec["role_level"].Int(), nil
}

// dictTextMap 读取一个普通字典的 键值→文本 映射（一次性载入，页内复用）。
//
// 语义对齐 DictAspect 使用的 SysDictMapper#queryDictTextByKey：
//
//	select s.item_text from sys_dict_item s
//	  where s.dict_id = (select id from sys_dict where dict_code = #{code})
//	    and s.item_value = #{key}
//
// 该 XML SQL 无 del_flag/status 过滤，此处保持一致（未命中返回空映射）。
func (s *Service) dictTextMap(ctx context.Context, code string) (map[string]string, error) {
	dictCols := dao.SysDict.Columns()
	var d entity.SysDict
	if err := dao.SysDict.Ctx(ctx).
		Where(dictCols.DictCode, code).
		Scan(&d); err != nil {
		return nil, fmt.Errorf("查询字典失败: %w", err)
	}
	texts := make(map[string]string)
	if d.Id == "" {
		return texts, nil
	}
	itemCols := dao.SysDictItem.Columns()
	var items []entity.SysDictItem
	if err := dao.SysDictItem.Ctx(ctx).
		Where(itemCols.DictId, d.Id).
		Scan(&items); err != nil {
		return nil, fmt.Errorf("查询字典项失败: %w", err)
	}
	for _, item := range items {
		texts[item.ItemValue] = item.ItemText
	}
	return texts, nil
}

// newUserRecord 把 sys_user 实体转换为白名单输出行。
// birthday 与 Java @JsonFormat(pattern = "yyyy-MM-dd") 对齐，仅保留日期部分；
// status_dictText 查 user_status 字典，未命中为空串（Java 为 null，前端展示同为空白）。
func newUserRecord(u entity.SysUser, statusTexts map[string]string) userRecord {
	birthday := u.Birthday
	if len(birthday) > 10 {
		birthday = birthday[:10]
	}
	return userRecord{
		Id:             u.Id,
		Username:       u.Username,
		Realname:       u.Realname,
		Avatar:         u.Avatar,
		Birthday:       birthday,
		Sex:            u.Sex,
		Email:          u.Email,
		Phone:          u.Phone,
		Status:         u.Status,
		StatusDictText: statusTexts[strconv.Itoa(u.Status)],
		OrgCode:        u.OrgCode,
		ActivitiSync:   u.ActivitiSync,
		WorkNo:         u.WorkNo,
		Post:           u.Post,
		School:         u.School,
		Telephone:      u.Telephone,
		UserIdentity:   u.UserIdentity,
		DepartIds:      u.DepartIds,
	}
}
