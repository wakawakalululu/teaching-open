// Package usercore 承载用户管理之「核心 CRUD 域」业务用例（契约蓝本：
// api/jeecg-boot-module-system/.../system/controller/SysUserController.java
// + service/impl/SysUserServiceImpl.java + mapper/xml/SysUserMapper.xml
// + service/impl/SysDepartServiceImpl.java + aspect/DictAspect.java）。
//
// 逐端点契约（Java V2.8，路径/方法/参数/响应逐字段对齐）：
//
//  1. GET /sys/user/list（SysUserController L92-167 queryPageList）
//     参数：pageNo(默认1)/pageSize(默认10) + departId（逗号分隔部门 id，走
//     sys_user_depart 子查询）+ roleCode/roleId（roleId 支持逗号分隔，走
//     sys_user_role 子查询；roleCode 仅在 roleId 为空时经 sys_role 反查）+
//     QueryGenerator.installMplus 的 bean 过滤（username/realname/sex/status 等，
//     * 通配 LIKE、逗号 IN、!= NE、其余 EQ；createTime_begin/_end 走 GE/LE）+
//     column/order 排序。
//     注意：任务书提到的「自定义 XML getUserDepartList」在本仓库 V2.8 代码中
//     不存在（全库检索无该 id），实际列表 SQL 为 SysUserMapper.getUserList
//     （select sys_user.* from sys_user ${ew.customSqlSegment}），departId 过滤
//     是 Controller 里手拼的 inSql 子查询，此处按实际代码迁移。
//     Java 的 area 参数（provinceId/cityId → sys_user.province/city）在
//     teachingopen2.8.sql 的 sys_user 表中无对应列，且前端无调用方传 area，
//     属一旦触发即 SQL 报错的死代码，Go 版明确拒绝迁移（见报告）。
//     数据范围：非 admin/dev 角色只能看到自己负责部门（sys_user.depart_ids）
//     展开的子部门用户，无负责部门报「您没有负责的班级」。
//     result：IPage 形状 + DictAspect 注入（sex_dictText/status_dictText/
//     departIds_dictText）+ Controller 后处理（orgCodeTxt=部门名逗号拼接、
//     roleTxt=角色名逗号拼接）。password/salt 不输出（Java WRITE_ONLY）。
//
//  2. POST /sys/user/add（SysUserController L169-210 add）
//     请求体：SysUser 字段 + selectedroles（逗号分隔角色 id）+ selecteddeparts
//     （逗号分隔部门 id）；角色为空时默认 student 角色；逐角色校验
//     role_level 不高于操作人等级，否则「权限不足，无法分配所选角色」；
//     写 sys_user（status=1、del_flag=0、create_time=now）+ sys_user_role +
//     sys_user_depart 三表；话术「添加成功！」。
//
//  3. PUT /sys/user/edit（SysUserController L212-262 edit）
//     先查旧记录（@TableLogic 自动 del_flag=0），无则「未找到对应实体」；
//     lessThanUserRoleLevel 校验「权限不足」；角色校验同 add；更新 sys_user
//     （MyBatis-Plus updateById 仅写非 NULL 字段，密码保持库内旧值不变），
//     sys_user_role 先删后插，sys_user_depart 先清已弃部门的
//     sys_depart_role_user 再先删后插；最后 updateNullPhoneEmail 把空串
//     email/phone 置 NULL；话术「修改成功!」。
//
//  4. DELETE /sys/user/delete（SysUserController L268-276 delete）
//     lessThanUserRoleLevel 校验；deleteUser → removeById → del_flag 逻辑删
//     （@TableLogic：UPDATE sys_user SET del_flag=1 WHERE id=? AND del_flag=0，
//     不级联 sys_user_role/sys_user_depart，不更新 update_time）；
//     话术「删除用户成功」。
//
//  5. DELETE /sys/user/deleteBatch（SysUserController L282-292 deleteBatch）
//     逐 id 校验等级后 removeByIds（同 4 的逻辑删批量版）；话术「批量删除用户成功」。
//
//  6. PUT /sys/user/frozenBatch（SysUserController L300-324 frozenBatch）
//     请求体 {ids, status}；逐 id 校验等级；UPDATE sys_user SET status=? WHERE
//     id=? AND del_flag=0（status：1 正常 / 2 冻结）；话术「操作成功!」。
//     Java 在 catch 分支 error500 后仍会执行 result.success("操作成功!") 覆盖
//     （异常也返回成功，属 Java 缺陷），Go 版拒绝照搬：非法 status 等异常
//     返回 HTTP 200 + success:false「操作失败」，见 Add/Edit/FrozenBatch 注释。
//
// 安全与一致性说明：
//   - 全部 SQL 占位符绑定；Java 版在 list 的 departId/roleId 过滤里用字符串
//     拼接构造 IN 子查询（注入面），Go 版改为占位符绑定（同语义）；
//   - 排序列过 ^[a-zA-Z0-9_]+$ 正则 + sys_user 全量列集合校验（白名单再排除
//     password/salt），Java 为 SqlInjectionUtil 黑名单 + 不校验列存在性；
//   - del_flag 为 0/1 标志位：查询手动过滤 del_flag=0，逻辑删显式写 del_flag=1，
//     不启用 GoFrame 软删特性（sys_user 无 deleted_at 列）；
//   - 密码：新增用户散列用 internal/passwd 的 bcrypt（Java 版为可逆 DES，
//     属既有迁移策略：新账号一律 bcrypt；bcrypt 自带盐，sys_user.salt 留空）；
//   - sys_role/sys_depart_role/sys_depart_role_user 无对应 dao（文件所有权
//     限定），跨表读写用占位符绑定的裸 SQL（先例：service/rbac、service/userrole、
//     service/userdepart）；
//   - Java 端点里的 sysBaseAPI.addLog 日志埋点属日志域（阶段 3），不迁移。
package usercore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/passwd"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/service/rbac"
)

// 业务错误（HTTP 层映射为 Result{success:false}，话术与 Java 契约逐字一致）。
var (
	// ErrPermissionDenied 操作人角色等级低于目标用户（Java lessThanUserRoleLevel）。
	ErrPermissionDenied = errors.New("权限不足")
	// ErrEntityNotFound 编辑目标用户不存在（Java edit 的 getById 为 null 分支）。
	ErrEntityNotFound = errors.New("未找到对应实体")
	// ErrNoDeptScope 非 admin/dev 操作人无负责部门（Java list 的 error500 话术）。
	ErrNoDeptScope = errors.New("您没有负责的班级")
	// ErrRoleForbidden 分配高于操作人等级的角色（Java add/edit 的 error500 话术）。
	ErrRoleForbidden = errors.New("权限不足，无法分配所选角色")
	// ErrPasswordRequired 新增用户缺密码（Java 会把空串 DES 加密入库形成不可登录
	// 账号，Go 版前置拦截，属安全加固偏差，见 Add 注释）。
	ErrPasswordRequired = errors.New("密码不能为空")
)

// 常量（对齐 Java CommonConstant / SysUserServiceImpl）。
const (
	delFlagNormal   = 0 // sys_user.del_flag 正常
	delFlagDeleted  = 1 // sys_user.del_flag 已删除（回收站）
	statusActive    = 1 // Java add 强制 status=1
	studentRoleCode = "student"
	sexDictCode     = "sex"         // SysUser.sex @Dict(dicCode="sex")
	userStatusDict  = "user_status" // SysUser.status @Dict(dicCode="user_status")
)

// identRegexp 限定动态 SQL 标识符：仅字母/数字/下划线。
var identRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// sysUserColumns 从实体 orm 标签反解的 sys_user 全量列名集合。
// dao/internal 的列结构是子集（无 birthday/sex/work_no 等），列集合以
// teachingopen2.8.sql（即实体 orm 标签）为准。
var sysUserColumns = buildEntityColumnSet(entity.SysUser{})

// orderColumns 排序列白名单 = 全量列 − password/salt（敏感列禁止作为排序键，
// Java 版对此无限制，属 Go 版安全加固）。
var orderColumns = buildOrderColumns()

// userFilterColumns list 端点允许的字符串型 bean 过滤字段（Java
// QueryGenerator.installMplus 对 SysUser bean 属性的通用过滤，Go 版以固定
// 白名单替代反射全属性遍历；del_flag 有意不在白名单内，防止客户端改写删除标记）。
var userFilterColumns = map[string]string{
	"id":        "id",
	"username":  "username",
	"realname":  "realname",
	"avatar":    "avatar",
	"email":     "email",
	"phone":     "phone",
	"orgCode":   "org_code",
	"workNo":    "work_no",
	"post":      "post",
	"school":    "school",
	"telephone": "telephone",
	"departIds": "depart_ids",
	"thirdId":   "third_id",
	"thirdType": "third_type",
}

// userFilterIntColumns 数值型过滤字段（值须可解析为整数，解析失败按 Java
// installMplus 逐字段 try/catch 的语义静默跳过该条件）。
var userFilterIntColumns = map[string]string{
	"sex":          "sex",
	"status":       "status",
	"userIdentity": "user_identity",
	"activitiSync": "activiti_sync",
}

// buildEntityColumnSet 反射收集实体 orm 标签首段的列名集合。
func buildEntityColumnSet(v entity.SysUser) map[string]struct{} {
	set := make(map[string]struct{}, 32)
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("orm")
		if tag == "" {
			continue
		}
		name := strings.SplitN(tag, ",", 2)[0]
		if name != "" {
			set[name] = struct{}{}
		}
	}
	return set
}

// buildOrderColumns 组装排序白名单（全量列排除 password/salt）。
func buildOrderColumns() map[string]struct{} {
	set := make(map[string]struct{}, len(sysUserColumns))
	for k := range sysUserColumns {
		set[k] = struct{}{}
	}
	delete(set, "password")
	delete(set, "salt")
	return set
}

// ListInput list 查询入参（bean 过滤字段保留原始规则串，规则解析在服务层）。
type ListInput struct {
	PageNo          int
	PageSize        int
	Id              string
	Username        string
	Realname        string
	Phone           string
	Email           string
	Sex             string
	Status          string
	DepartId        string // 逗号分隔部门 id（sys_user_depart 子查询）
	DepartIds       string // sys_user.depart_ids 列过滤（QueryGenerator bean 过滤）
	RoleId          string // 逗号分隔角色 id（sys_user_role 子查询）
	RoleCode        string // roleId 为空时经 sys_role.role_code 反查
	CreateTimeBegin string
	CreateTimeEnd   string
	OrderColumn     string
	OrderType       string
}

// UserPayload add/edit 共享请求体。指针字段镜像 MyBatis-Plus updateById 的
// 非 NULL 更新策略：仅客户端显式传入（JSON 非 null）的字段参与写库，
// 未传字段保持库内旧值。selectedroles/selecteddeparts 非 SysUser 属性，
// Java 用 jsonObject.getString 单独取出，这里同样独立承载。
type UserPayload struct {
	Id           *string `json:"id"`
	Username     *string `json:"username"`
	Password     *string `json:"password"`
	Realname     *string `json:"realname"`
	Avatar       *string `json:"avatar"`
	Birthday     *string `json:"birthday"`
	Sex          *int    `json:"sex"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	OrgCode      *string `json:"orgCode"`
	Status       *int    `json:"status"`
	ActivitiSync *int    `json:"activitiSync"`
	WorkNo       *string `json:"workNo"`
	Post         *string `json:"post"`
	School       *string `json:"school"`
	Telephone    *string `json:"telephone"`
	UserIdentity *int    `json:"userIdentity"`
	DepartIds    *string `json:"departIds"`
	ThirdId      *string `json:"thirdId"`
	ThirdType    *string `json:"thirdType"`

	SelectedRoles   string `json:"selectedroles"`
	SelectedDeparts string `json:"selecteddeparts"`
}

// listRecord list 端点白名单输出行：SysUser 全字段 + DictAspect 注入的
// 三个 *_dictText + Controller 后处理的 orgCodeTxt/roleTxt。
//
// 前端消费核对（逐一阅读）：
//   - UserList.vue：表格列 id/username/realname/avatar/sex_dictText/birthday/
//     phone/school/userIdentity/roleTxt/orgCodeTxt/departIds_dictText/
//     status_dictText/createTime；行编辑把整行传给 UserModal.edit(record)；
//   - SelectUserModal.vue：username/realname/phone + records/total（roleNames/
//     departNames 列在 Java 版即不存在、恒显示 "--"，Go 版保持一致不造假字段）；
//   - SelectUserListModal.vue：username/realname/sex/phone/email/status + records/total；
//   - JSelectMultiUser.vue（JSelectBizComponent）：realname/username/phone/birthday，
//     回显查询发 id=逗号拼接 → IN 过滤。
//   - UserModal.edit 的回填消费：id/avatar/birthday/username/sex/realname/email/
//     phone/activitiSync/workNo/telephone/post/school/userIdentity/departIds。
//
// password/salt 永不输出（Java @JsonProperty(WRITE_ONLY) 等价）。
type listRecord struct {
	Id                string `json:"id"`
	Username          string `json:"username"`
	Realname          string `json:"realname"`
	Avatar            string `json:"avatar"`
	Birthday          string `json:"birthday"`
	Sex               int    `json:"sex"`
	SexDictText       string `json:"sex_dictText"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	OrgCode           string `json:"orgCode"`
	OrgCodeTxt        string `json:"orgCodeTxt"`
	RoleTxt           string `json:"roleTxt"`
	Status            int    `json:"status"`
	StatusDictText    string `json:"status_dictText"`
	DelFlag           int    `json:"delFlag"`
	WorkNo            string `json:"workNo"`
	Post              string `json:"post"`
	School            string `json:"school"`
	Telephone         string `json:"telephone"`
	CreateBy          string `json:"createBy"`
	CreateTime        string `json:"createTime"`
	UpdateBy          string `json:"updateBy"`
	UpdateTime        string `json:"updateTime"`
	ActivitiSync      int    `json:"activitiSync"`
	UserIdentity      int    `json:"userIdentity"`
	DepartIds         string `json:"departIds"`
	DepartIdsDictText string `json:"departIds_dictText"`
	ThirdId           string `json:"thirdId"`
	ThirdType         string `json:"thirdType"`
}

// Service 用户核心 CRUD 业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// List 用户分页列表（端点 1）。
//
// 级联语义：先做数据范围圈定（非 admin/dev 仅本人负责部门展开的子部门用户），
// 再叠加 departId/roleId 子查询过滤与 QueryGenerator bean 过滤；查完按 Java
// controller 的后处理注入 orgCodeTxt（部门名逗号拼接）、roleTxt（角色名逗号
// 拼接）与 DictAspect 的 sex_dictText/status_dictText/departIds_dictText。
func (s *Service) List(ctx context.Context, operatorID string, in ListInput) (*page.Result, error) {
	// 数据范围：Java getCurrentUser() 为 null 会 NPE，Go 版 fail-closed。
	if strings.TrimSpace(operatorID) == "" {
		return nil, ErrPermissionDenied
	}
	cols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).Where(cols.DelFlag, delFlagNormal)

	// 非 admin/dev 角色只显示自己管理的部门下的用户（Java L136-146）。
	isAdmin, err := rbac.HasRole(ctx, operatorID, "admin")
	if err != nil {
		return nil, fmt.Errorf("查询操作人角色失败: %w", err)
	}
	isDev, err := rbac.HasRole(ctx, operatorID, "dev")
	if err != nil {
		return nil, fmt.Errorf("查询操作人角色失败: %w", err)
	}
	if !isAdmin && !isDev {
		user, err := s.getUserById(ctx, operatorID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			// Java：库内无此用户时 getCurrentUser 取 token 快照，Go 版以库内
			// 最新值为准；查不到等价于无负责部门，按「您没有负责的班级」拒绝。
			return nil, ErrNoDeptScope
		}
		myDepIds, err := s.getMySubDepIdsByDepId(ctx, user.DepartIds)
		if err != nil {
			return nil, err
		}
		if len(myDepIds) == 0 {
			return nil, ErrNoDeptScope
		}
		// Java 用字符串拼接 inSql，Go 版占位符绑定（同语义）。
		m = m.Where(fmt.Sprintf("%s IN (SELECT user_id FROM sys_user_depart WHERE dep_id IN (?))",
			cols.Id), myDepIds)
	}

	// departId 批量查询条件（Java L118-122 的 inSql 子查询）。
	if depIds := SplitIDs(in.DepartId); len(depIds) > 0 {
		m = m.Where(fmt.Sprintf("%s IN (SELECT user_id FROM sys_user_depart WHERE dep_id IN (?))",
			cols.Id), depIds)
	}

	// roleCode 查询条件：仅 roleId 为空时反查（Java L124-127）。
	roleId := in.RoleId
	if roleId == "" && in.RoleCode != "" {
		roleId, err = s.roleIDByCode(ctx, in.RoleCode)
		if err != nil {
			return nil, err
		}
	}
	// roleId 批量查询条件（Java L129-132 的 inSql 子查询）。
	if ids := SplitIDs(roleId); len(ids) > 0 {
		m = m.Where(fmt.Sprintf("%s IN (SELECT user_id FROM sys_user_role WHERE role_id IN (?))",
			cols.Id), ids)
	}

	// QueryGenerator bean 过滤（字符串型规则：* LIKE / 逗号 IN / != NE / 其余 EQ）。
	for name, col := range userFilterColumns {
		m = applyStringFilter(m, col, filterValueByName(in, name))
	}
	// 数值型过滤（解析失败跳过，对齐 installMplus 的逐字段异常吞并）。
	for name, col := range userFilterIntColumns {
		m = applyIntFilter(m, col, filterValueByName(in, name))
	}
	// 注册时间区间（Java _begin/_end → GE/LE；仅日期串时 Java 拼全天边界）。
	if v := strings.TrimSpace(in.CreateTimeBegin); v != "" {
		if len(v) == 10 {
			v += " 00:00:00"
		}
		m = m.Where(fmt.Sprintf("%s >= ?", cols.CreateTime), v)
	}
	if v := strings.TrimSpace(in.CreateTimeEnd); v != "" {
		if len(v) == 10 {
			v += " 23:59:59"
		}
		m = m.Where(fmt.Sprintf("%s <= ?", cols.CreateTime), v)
	}

	m = applyOrder(m, in.OrderColumn, in.OrderType)

	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计用户数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]listRecord{}, 0, in.PageNo, in.PageSize), nil
	}
	var users []entity.SysUser
	if err := m.Page(in.PageNo, in.PageSize).Scan(&users); err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	records, err := s.enrichRecords(ctx, users)
	if err != nil {
		return nil, err
	}
	return page.New(records, total, in.PageNo, in.PageSize), nil
}

// enrichRecords Java controller 后处理 + DictAspect 注入的 Go 实现：
// orgCodeTxt ← getDepNamesByUserIds（部门名逗号拼接）、roleTxt ←
// getRoleNamesByUserIds、sex_dictText ← sex 字典、status_dictText ←
// user_status 字典、departIds_dictText ← sys_depart 表字典（@Dict(dictTable)）。
func (s *Service) enrichRecords(ctx context.Context, users []entity.SysUser) ([]listRecord, error) {
	userIds := make([]string, 0, len(users))
	for i := range users {
		userIds = append(userIds, users[i].Id)
	}
	depNames, err := s.depNamesByUserIds(ctx, userIds)
	if err != nil {
		return nil, err
	}
	roleNames, err := s.roleNamesByUserIds(ctx, userIds)
	if err != nil {
		return nil, err
	}
	sexDict, err := s.dictTextMap(ctx, sexDictCode)
	if err != nil {
		return nil, err
	}
	statusDict, err := s.dictTextMap(ctx, userStatusDict)
	if err != nil {
		return nil, err
	}
	// departIds_dictText（表字典 sys_depart：dicCode=id，dicText=depart_name）。
	depNameByIds, err := s.departNameByIds(ctx, collectDepartIds(users))
	if err != nil {
		return nil, err
	}

	records := make([]listRecord, 0, len(users))
	for i := range users {
		u := users[i]
		records = append(records, listRecord{
			Id:                u.Id,
			Username:          u.Username,
			Realname:          u.Realname,
			Avatar:            u.Avatar,
			Birthday:          formatDate(u.Birthday),
			Sex:               u.Sex,
			SexDictText:       sexDict[strconv.Itoa(u.Sex)],
			Email:             u.Email,
			Phone:             u.Phone,
			OrgCode:           u.OrgCode,
			OrgCodeTxt:        depNames[u.Id],
			RoleTxt:           roleNames[u.Id],
			Status:            u.Status,
			StatusDictText:    statusDict[strconv.Itoa(u.Status)],
			DelFlag:           u.DelFlag,
			WorkNo:            u.WorkNo,
			Post:              u.Post,
			School:            u.School,
			Telephone:         u.Telephone,
			CreateBy:          u.CreateBy,
			CreateTime:        formatDateTime(u.CreateTime),
			UpdateBy:          u.UpdateBy,
			UpdateTime:        formatDateTime(u.UpdateTime),
			ActivitiSync:      u.ActivitiSync,
			UserIdentity:      u.UserIdentity,
			DepartIds:         u.DepartIds,
			DepartIdsDictText: translateTableDict(u.DepartIds, depNameByIds),
			ThirdId:           u.ThirdId,
			ThirdType:         u.ThirdType,
		})
	}
	return records, nil
}

// Add 新增用户（端点 2）。
//
// Java add 未做用户名唯一性校验（重复账号由 sys_user.uniq_sys_user_username
// 唯一键在写库时报错 → catch →「操作失败」），Go 版等价（写库错误走通用失败）。
//
// 有意偏差（安全加固，非契约破坏）：
//  1. 密码散列用 bcrypt（internal/passwd.Hash），Java 为可逆 DES + 8 位随机盐；
//     bcrypt 自带盐，sys_user.salt 留空。空密码 Java 会照常入库形成不可登录
//     账号，Go 版前置拦截返回「密码不能为空」。
//  2. create_by 由服务端写操作人用户名（Java add 不写 create_by）；按仓库
//     硬性规范「create_by/create_time 由服务端写」执行。
//  3. id 缺失时服务端生成 guid.S()（32 位 hex）；Java 依赖前端 generateUserId
//     回传（UserModal 提交必带 id），兜底沿用仓库先例。
//  4. Java 的 save/addUserWithRole/addUserWithDepart 是三个独立事务（中途失败
//     会留下半写状态），Go 版合并为单事务，仅加强原子性，成功路径一致。
func (s *Service) Add(ctx context.Context, operatorID, operatorName string, in *UserPayload) error {
	roles, err := s.resolveRoles(ctx, operatorID, in.SelectedRoles)
	if err != nil {
		return err
	}
	if in.Password == nil || strings.TrimSpace(*in.Password) == "" {
		return ErrPasswordRequired
	}
	hash, err := passwd.Hash(*in.Password)
	if err != nil {
		return fmt.Errorf("生成密码散列失败: %w", err)
	}

	id := ""
	if in.Id != nil {
		id = strings.TrimSpace(*in.Id)
	}
	if id == "" {
		id = guid.S()
	}

	data := do.SysUser{
		Id:         id,
		Password:   hash,
		CreateBy:   operatorName,
		CreateTime: gtime.Now(),
	}
	applyPayload(&data, in, false)
	// Java add 在 JSON 解析后强制 setStatus(1)/setDelFlag(0)（客户端传值被
	// 覆盖），此处放在 applyPayload 之后以保持同样的强制语义。
	data.Status = statusActive
	data.DelFlag = delFlagNormal

	if err := dao.SysUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.SysUser.Ctx(ctx).Data(data).Insert(); err != nil {
			return fmt.Errorf("保存用户失败: %w", err)
		}
		if err := s.insertUserRoles(ctx, id, roles); err != nil {
			return err
		}
		return s.insertUserDeparts(ctx, id, SplitIDs(in.SelectedDeparts))
	}); err != nil {
		return err
	}
	return nil
}

// Edit 编辑用户（端点 3）。
//
// 有意偏差：
//  1. 密码不参与更新：Java 把库内旧密码回填到更新实体再 updateById（等效于
//     不变），Go 版直接从更新集剔除 password，效果一致且少写一次该列；
//  2. del_flag/create_by/create_time 不参与更新：Java 会用客户端回传的旧值
//     原样写回（值不变），Go 版不写这些列，最终状态一致且杜绝客户端改写
//     删除标记（如回传 delFlag=1 会被 Java 忠实执行，属缺陷，明确不照搬）；
//  3. update_by 由服务端写操作人用户名（Java 保留客户端回传旧值）；
//  4. 空串 email/phone 直接写 NULL（Java 经 updateNullPhoneEmail 全表
//     UPDATE ... WHERE email=” 置 NULL，对被编辑行等效；全表清扫的副作用
//     不复刻，见注释）；
//  5. editUserWithRole/editUserWithDepart/updateNullPhoneEmail 在 Java 是
//     三个独立事务，Go 版合并为单事务，成功路径一致。
func (s *Service) Edit(ctx context.Context, operatorID, operatorName string, in *UserPayload) error {
	id := ""
	if in.Id != nil {
		id = strings.TrimSpace(*in.Id)
	}
	old, err := s.getUserById(ctx, id)
	if err != nil {
		return err
	}
	if old == nil {
		return ErrEntityNotFound
	}
	// Java lessThanUserRoleLevel(sysUser.getId())：操作人等级低于目标用户即拒绝。
	if err := s.checkOperatorLevel(ctx, operatorID, old.Id); err != nil {
		return err
	}
	roles, err := s.resolveRoles(ctx, operatorID, in.SelectedRoles)
	if err != nil {
		return err
	}

	// 更新集：仅客户端显式传入的字段（指针非 nil），服务端维护列见函数注释。
	data := do.SysUser{
		UpdateTime: gtime.Now(),
		UpdateBy:   operatorName,
	}
	applyPayload(&data, in, true)

	if err := dao.SysUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.SysUser.Ctx(ctx).
			Where(dao.SysUser.Columns().Id, old.Id).
			Where(dao.SysUser.Columns().DelFlag, delFlagNormal).
			Data(data).Update(); err != nil {
			return fmt.Errorf("更新用户失败: %w", err)
		}
		// sys_user_role 先删后插（Java editUserWithRole）。
		if err := s.replaceUserRoles(ctx, old.Id, roles); err != nil {
			return err
		}
		// sys_user_depart：先清理被移除部门的部门角色用户关系，再先删后插
		// （Java editUserWithDepart）。
		if err := s.replaceUserDeparts(ctx, old.Id, SplitIDs(in.SelectedDeparts)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Delete 删除用户（端点 4，del_flag 逻辑删，非物理删）。
//
// Java deleteUser 仅 removeById（@TableLogic 逻辑删），不级联清理
// sys_user_role/sys_user_depart（关联行成为孤儿，由回收站彻底删除时清理），
// Go 版保持一致不照搬「想当然的级联」；也不更新 update_time（MyBatis-Plus
// 逻辑删不触碰其他列）。
func (s *Service) Delete(ctx context.Context, operatorID, id string) error {
	if err := s.checkOperatorLevel(ctx, operatorID, id); err != nil {
		return err
	}
	return s.logicalDelete(ctx, []string{id})
}

// DeleteBatch 批量删除用户（端点 5）：逐 id 校验等级后批量逻辑删。
func (s *Service) DeleteBatch(ctx context.Context, operatorID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if err := s.checkOperatorLevel(ctx, operatorID, id); err != nil {
			return err
		}
	}
	return s.logicalDelete(ctx, ids)
}

// logicalDelete 逻辑删（对齐 MyBatis-Plus removeById/removeByIds）：
// UPDATE sys_user SET del_flag=1 WHERE del_flag=0 AND id IN (...)。
func (s *Service) logicalDelete(ctx context.Context, ids []string) error {
	cols := dao.SysUser.Columns()
	if _, err := dao.SysUser.Ctx(ctx).
		Where(cols.DelFlag, delFlagNormal).
		WhereIn(cols.Id, ids).
		Data(do.SysUser{DelFlag: delFlagDeleted}).
		Update(); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

// FrozenBatch 冻结/解冻用户（端点 6）。
//
// status 语义：1 正常 / 2 冻结（前端 UserList handleFrozen 分别传 1/2）；
// 无角色等级之外的校验，仅有 lessThanUserRoleLevel（操作人等级低于目标用户
// 即「权限不足」）。Java 对非法 status 会抛异常又被 result.success 覆盖成
// 「操作成功!」（缺陷），Go 版拒绝照搬：解析失败返回错误由控制器映射为
// success:false「操作失败」。
func (s *Service) FrozenBatch(ctx context.Context, operatorID string, ids []string, status string) error {
	if len(ids) == 0 {
		// Java：ids 空串 split 出 [""]，循环内 isNotEmpty 跳过，直接成功。
		return nil
	}
	statusValue, err := strconv.Atoi(strings.TrimSpace(status))
	if err != nil {
		return fmt.Errorf("status 参数非法: %s", status)
	}
	for _, id := range ids {
		if err := s.checkOperatorLevel(ctx, operatorID, id); err != nil {
			return err
		}
	}
	cols := dao.SysUser.Columns()
	if _, err := dao.SysUser.Ctx(ctx).
		Where(cols.DelFlag, delFlagNormal).
		WhereIn(cols.Id, ids).
		Data(do.SysUser{Status: statusValue}).
		Update(); err != nil {
		return fmt.Errorf("冻结/解冻用户失败: %w", err)
	}
	return nil
}

// ---------- add/edit 共用私有逻辑 ----------

// resolveRoles 解析 selectedroles：非空时逐角色校验 role_level 不高于操作人
// 等级（Java add/edit 的循环校验），为空时默认 student 角色（Java L183-189/
// L244-250）。角色不存在时 Java 因 NPE 被 catch 成「操作失败」，Go 版返回
// 普通错误走控制器的通用失败分支，话术一致。
func (s *Service) resolveRoles(ctx context.Context, operatorID, selectedRoles string) ([]string, error) {
	roles := SplitIDs(selectedRoles)
	if len(roles) == 0 {
		studentID, err := s.roleIDByCode(ctx, studentRoleCode)
		if err != nil {
			return nil, err
		}
		if studentID == "" {
			return nil, nil
		}
		return []string{studentID}, nil
	}
	level, err := s.userRoleLevel(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	for _, roleId := range roles {
		roleLevel, err := s.roleLevelByID(ctx, roleId)
		if err != nil {
			return nil, err
		}
		if roleLevel > level {
			return nil, ErrRoleForbidden
		}
	}
	return roles, nil
}

// applyPayload 把客户端显式传入的字段写入 DO（指针非 nil 才写）。
//
// 生日为空串时跳过（Java fastjson 把空串解析为 null → 不写该列，直接写空串
// 反而会被 datetime 列拒绝）；email/phone 的空串处理分模式：
//   - add：原样写 ”（Java add 的 updateById/insert 把非 null 字段全部落库）；
//   - edit：写 NULL（Java edit 写 ” 后再跑 updateNullPhoneEmail 全表
//     UPDATE sys_user SET email=NULL WHERE email=”，对被编辑行净效果为 NULL；
//     全表清扫的副作用不复刻，见 Edit 注释）。
//
// 始终不写：password（add 单独散列、edit 保持库内旧值）、del_flag、create_by、
// create_time（Java 把客户端回传的旧值原样写回，值不变，Go 版不写这些列，
// 最终状态一致且杜绝客户端改写删除标记）。
func applyPayload(data *do.SysUser, in *UserPayload, editMode bool) {
	if in.Username != nil {
		data.Username = *in.Username
	}
	if in.Realname != nil {
		data.Realname = *in.Realname
	}
	if in.Avatar != nil {
		data.Avatar = *in.Avatar
	}
	if in.Birthday != nil && strings.TrimSpace(*in.Birthday) != "" {
		data.Birthday = *in.Birthday
	}
	if in.Sex != nil {
		data.Sex = *in.Sex
	}
	if in.Email != nil {
		data.Email = emptyStringToNull(*in.Email, editMode)
	}
	if in.Phone != nil {
		data.Phone = emptyStringToNull(*in.Phone, editMode)
	}
	if in.OrgCode != nil {
		data.OrgCode = *in.OrgCode
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.ActivitiSync != nil {
		data.ActivitiSync = *in.ActivitiSync
	}
	if in.WorkNo != nil {
		data.WorkNo = *in.WorkNo
	}
	if in.Post != nil {
		data.Post = *in.Post
	}
	if in.School != nil {
		data.School = *in.School
	}
	if in.Telephone != nil {
		data.Telephone = *in.Telephone
	}
	if in.UserIdentity != nil {
		data.UserIdentity = *in.UserIdentity
	}
	if in.DepartIds != nil {
		data.DepartIds = *in.DepartIds
	}
	if in.ThirdId != nil {
		data.ThirdId = *in.ThirdId
	}
	if in.ThirdType != nil {
		data.ThirdType = *in.ThirdType
	}
}

// emptyStringToNull edit 模式下空串转 NULL（Java updateNullPhoneEmail 的
// UPDATE sys_user SET email=NULL WHERE email=”），add 模式原样写。
func emptyStringToNull(v string, editMode bool) interface{} {
	if editMode && strings.TrimSpace(v) == "" {
		return gdb.Raw("NULL")
	}
	return v
}

// insertUserRoles 写 sys_user_role 关联行（Java addUserWithRole：
// sys_user_role 无时间列/删除标记，主键服务端生成，仓库先例 guid.S()）。
func (s *Service) insertUserRoles(ctx context.Context, userId string, roleIds []string) error {
	for _, roleId := range roleIds {
		if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
			Id:     guid.S(),
			UserId: userId,
			RoleId: roleId,
		}).Insert(); err != nil {
			return fmt.Errorf("保存用户角色失败: %w", err)
		}
	}
	return nil
}

// replaceUserRoles sys_user_role 先删后插（Java editUserWithRole）。
func (s *Service) replaceUserRoles(ctx context.Context, userId string, roleIds []string) error {
	cols := dao.SysUserRole.Columns()
	if _, err := dao.SysUserRole.Ctx(ctx).
		Where(cols.UserId, userId).
		Delete(); err != nil {
		return fmt.Errorf("清理用户角色失败: %w", err)
	}
	return s.insertUserRoles(ctx, userId, roleIds)
}

// insertUserDeparts 写 sys_user_depart 关联行（Java addUserWithDepart）。
func (s *Service) insertUserDeparts(ctx context.Context, userId string, depIds []string) error {
	for _, depId := range depIds {
		if _, err := dao.SysUserDepart.Ctx(ctx).Data(do.SysUserDepart{
			Id:     guid.S(),
			UserId: userId,
			DepId:  depId,
		}).Insert(); err != nil {
			return fmt.Errorf("保存用户部门失败: %w", err)
		}
	}
	return nil
}

// replaceUserDeparts sys_user_depart 先删后插 + 被移除部门的
// sys_depart_role_user 级联清理（Java editUserWithDepart 的顺序：
// 查旧关联 → 对不在新集合里的旧部门清理部门角色用户 → 删全部旧关联 →
// 插新关联；selecteddeparts 为空时全部旧部门视作被移除）。
func (s *Service) replaceUserDeparts(ctx context.Context, userId string, newDepIds []string) error {
	udCols := dao.SysUserDepart.Columns()
	var relations []entity.SysUserDepart
	if err := dao.SysUserDepart.Ctx(ctx).
		Where(udCols.UserId, userId).
		Scan(&relations); err != nil {
		return fmt.Errorf("查询用户部门关系失败: %w", err)
	}
	kept := make(map[string]struct{}, len(newDepIds))
	for _, depId := range newDepIds {
		kept[depId] = struct{}{}
	}
	for _, rel := range relations {
		if _, ok := kept[rel.DepId]; ok {
			continue
		}
		if err := s.removeDepartRoleUser(ctx, userId, rel.DepId); err != nil {
			return err
		}
	}
	if _, err := dao.SysUserDepart.Ctx(ctx).
		Where(udCols.UserId, userId).
		Delete(); err != nil {
		return fmt.Errorf("清理用户部门关系失败: %w", err)
	}
	return s.insertUserDeparts(ctx, userId, newDepIds)
}

// removeDepartRoleUser 级联清理部门角色用户关系（Java editUserWithDepart 内
// 嵌逻辑：按部门查 sys_depart_role 的角色 id 集合，删 sys_depart_role_user 中
// 该用户 + 这些角色的行；两表无 dao，用占位符绑定的裸 SQL，先例 userdepart）。
func (s *Service) removeDepartRoleUser(ctx context.Context, userId, depId string) error {
	const query = `SELECT id FROM sys_depart_role WHERE depart_id = ?`
	records, err := dao.SysUserDepart.DB().GetAll(ctx, query, depId)
	if err != nil {
		return fmt.Errorf("查询部门角色失败: %w", err)
	}
	if len(records) == 0 {
		return nil
	}
	roleIds := make([]string, 0, len(records))
	for _, rec := range records {
		roleIds = append(roleIds, rec["id"].String())
	}
	if _, err := dao.SysUserDepart.DB().Exec(ctx,
		"DELETE FROM sys_depart_role_user WHERE user_id = ? AND drole_id IN (?)",
		userId, roleIds); err != nil {
		return fmt.Errorf("删除部门角色用户关系失败: %w", err)
	}
	return nil
}

// ---------- 通用查询/校验辅助 ----------

// getUserById 按 id 查用户（del_flag=0；Java getById 的 @TableLogic 等价）。
// password/salt 在实体上不序列化，仅服务端持有。
func (s *Service) getUserById(ctx context.Context, userId string) (*entity.SysUser, error) {
	cols := dao.SysUser.Columns()
	var user entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(cols.Id, userId).
		Where(cols.DelFlag, delFlagNormal).
		Scan(&user); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user.Id == "" {
		return nil, nil
	}
	return &user, nil
}

// userRoleLevel 用户角色等级（SQL 移植 SysUserRoleMapper.getUserRoleLevel：
// max(role_level)，无角色时 Java Integer null → -1，此处一致）。
func (s *Service) userRoleLevel(ctx context.Context, userId string) (int, error) {
	const query = `SELECT MAX(role_level) AS level FROM sys_role
WHERE id IN (SELECT role_id FROM sys_user_role WHERE user_id = ?)`
	one, err := dao.SysUser.DB().GetOne(ctx, query, userId)
	if err != nil {
		return -1, fmt.Errorf("查询用户角色等级失败: %w", err)
	}
	if one.IsEmpty() || one["level"].IsNil() {
		return -1, nil
	}
	return one["level"].Int(), nil
}

// checkOperatorLevel 角色等级校验（Java BaseController.lessThanUserRoleLevel：
// 操作人等级 < 目标用户等级 → 权限不足；操作人缺失按拒绝处理，fail-closed）。
func (s *Service) checkOperatorLevel(ctx context.Context, operatorID, targetID string) error {
	if strings.TrimSpace(operatorID) == "" {
		return ErrPermissionDenied
	}
	operatorLevel, err := s.userRoleLevel(ctx, operatorID)
	if err != nil {
		return err
	}
	targetLevel, err := s.userRoleLevel(ctx, targetID)
	if err != nil {
		return err
	}
	if operatorLevel < targetLevel {
		return ErrPermissionDenied
	}
	return nil
}

// roleLevelByID 按 id 查角色等级（Java sysRoleService.getById(roleId) 后取
// getRoleLevel()；角色不存在 Java NPE → catch「操作失败」，Go 返回普通错误
// 走控制器通用失败分支）。
func (s *Service) roleLevelByID(ctx context.Context, roleId string) (int, error) {
	one, err := dao.SysUser.DB().GetOne(ctx,
		"SELECT role_level FROM sys_role WHERE id = ?", roleId)
	if err != nil {
		return 0, fmt.Errorf("查询角色失败: %w", err)
	}
	if one.IsEmpty() {
		return 0, fmt.Errorf("角色不存在 roleId=%s", roleId)
	}
	return one["role_level"].Int(), nil
}

// roleIDByCode 按 role_code 查角色 id（Java sysRoleService.getRoleByCode）；
// 未命中返回空串（Java 返回 null 后跳过 roleId 条件，语义一致）。
func (s *Service) roleIDByCode(ctx context.Context, roleCode string) (string, error) {
	one, err := dao.SysUser.DB().GetOne(ctx,
		"SELECT id FROM sys_role WHERE role_code = ?", roleCode)
	if err != nil {
		return "", fmt.Errorf("查询角色失败: %w", err)
	}
	if one.IsEmpty() {
		return "", nil
	}
	return one["id"].String(), nil
}

// getMySubDepIdsByDepId 按用户负责部门（depart_ids，逗号分隔）展开全部子部门
// （SQL 移植 SysDepartServiceImpl#getMyDeptParentOrgCode + SysDepartMapper
// .getSubDepIdsByOrgCodes：del_flag=0 的负责部门按 org_code 升序取编码，再对
// 每个编码做 org_code LIKE 前缀匹配。Java 的 3 位编码分组拼接后拆分即全部
// 编码本身，Go 版直接以全部编码为 LIKE 模式，等价去重无副作用；
// 先例：service/userdepart#getMySubDepIdsByDepId）。
func (s *Service) getMySubDepIdsByDepId(ctx context.Context, departIds string) ([]string, error) {
	ids := SplitIDs(departIds)
	if len(ids) == 0 {
		return nil, nil
	}
	depCols := dao.SysDepart.Columns()
	records, err := dao.SysDepart.Ctx(ctx).
		Fields(depCols.OrgCode).
		Where(depCols.DelFlag, delFlagNormal).
		WhereIn(depCols.Id, ids).
		Order(depCols.OrgCode + " ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("查询负责部门失败: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	orgCodes := make([]string, 0, len(records))
	for _, rec := range records {
		if code := rec[depCols.OrgCode].String(); code != "" {
			orgCodes = append(orgCodes, code)
		}
	}
	if len(orgCodes) == 0 {
		return nil, nil
	}
	conditions := make([]string, 0, len(orgCodes))
	args := make([]interface{}, 0, len(orgCodes))
	for _, code := range orgCodes {
		conditions = append(conditions, "org_code LIKE ?")
		args = append(args, code+"%")
	}
	query := "SELECT id FROM sys_depart WHERE del_flag = '0' AND (" + strings.Join(conditions, " OR ") + ")"
	subRecords, err := dao.SysDepart.DB().GetAll(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询负责部门子部门失败: %w", err)
	}
	subIds := make([]string, 0, len(subRecords))
	for _, rec := range subRecords {
		subIds = append(subIds, rec["id"].String())
	}
	return subIds, nil
}

// depNamesByUserIds 批量查用户部门名（SQL 移植 SysUserMapper.getDepNamesByUserIds
// 的 Java 聚合语义：同用户多部门按行序「首个胜出、后续逗号追加」；Java 原查询
// 不过滤 sys_depart.del_flag，Go 版补 del_flag=0，与 userdepart 域同款加固）。
func (s *Service) depNamesByUserIds(ctx context.Context, userIds []string) (map[string]string, error) {
	res := make(map[string]string, len(userIds))
	if len(userIds) == 0 {
		return res, nil
	}
	const query = `SELECT ud.user_id AS user_id, d.depart_name AS depart_name
FROM sys_user_depart ud
INNER JOIN sys_depart d ON d.id = ud.dep_id AND d.del_flag = 0
WHERE ud.user_id IN (?)`
	records, err := dao.SysUserDepart.DB().GetAll(ctx, query, userIds)
	if err != nil {
		return nil, fmt.Errorf("查询用户部门名失败: %w", err)
	}
	for _, rec := range records {
		uid := rec["user_id"].String()
		name := rec["depart_name"].String()
		if _, ok := res[uid]; !ok {
			res[uid] = name
			continue
		}
		res[uid] = res[uid] + "," + name
	}
	return res, nil
}

// roleNamesByUserIds 批量查用户角色名（SQL 移植 SysUserMapper.getRoleNamesByUserIds：
// sys_user_role 与 sys_role 按 role_id 相连，DISTINCT r.id；聚合语义同部门名
// 「首个胜出、后续逗号追加」。sys_role 无 del_flag 列，不加该过滤）。
func (s *Service) roleNamesByUserIds(ctx context.Context, userIds []string) (map[string]string, error) {
	res := make(map[string]string, len(userIds))
	if len(userIds) == 0 {
		return res, nil
	}
	const query = `SELECT DISTINCT r.id, r.role_name, ur.user_id
FROM sys_user_role ur, sys_role r
WHERE r.id = ur.role_id AND ur.user_id IN (?)`
	records, err := dao.SysUser.DB().GetAll(ctx, query, userIds)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色名失败: %w", err)
	}
	for _, rec := range records {
		uid := rec["user_id"].String()
		name := rec["role_name"].String()
		if _, ok := res[uid]; !ok {
			res[uid] = name
			continue
		}
		res[uid] = res[uid] + "," + name
	}
	return res, nil
}

// departNameByIds 按 id 批量查部门名（表字典 queryTableDictTextByKey 的
// SELECT text FROM table WHERE code = ? 等价；Java 原查询无 del_flag 过滤，
// Go 版补 del_flag=0，与 getDepNamesByUserIds 同款加固）。
func (s *Service) departNameByIds(ctx context.Context, depIds []string) (map[string]string, error) {
	res := make(map[string]string, len(depIds))
	if len(depIds) == 0 {
		return res, nil
	}
	depCols := dao.SysDepart.Columns()
	records, err := dao.SysDepart.Ctx(ctx).
		Fields(depCols.Id, depCols.DepartName).
		Where(depCols.DelFlag, delFlagNormal).
		WhereIn(depCols.Id, depIds).
		All()
	if err != nil {
		return nil, fmt.Errorf("查询部门名失败: %w", err)
	}
	for _, rec := range records {
		res[rec[depCols.Id].String()] = rec[depCols.DepartName].String()
	}
	return res, nil
}

// dictTextMap 读取一个普通字典的 键值→文本 映射（一次性载入，页内复用）。
//
// 语义对齐 DictAspect 使用的 SysDictMapper#queryDictTextByKey：
//
//	select s.item_text from sys_dict_item s
//	  where s.dict_id = (select id from sys_dict where dict_code = #{code})
//	    and s.item_value = #{key}
//
// 该 XML SQL 无 del_flag/status 过滤，此处保持一致（先例：service/userrole）。
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

// ---------- 纯函数辅助（可单测） ----------

// filterValueByName 按过滤字段名取原始查询串（与 ListInput 字段一一对应）。
func filterValueByName(in ListInput, name string) string {
	switch name {
	case "id":
		return in.Id
	case "username":
		return in.Username
	case "realname":
		return in.Realname
	case "phone":
		return in.Phone
	case "email":
		return in.Email
	case "sex":
		return in.Sex
	case "status":
		return in.Status
	case "departIds":
		return in.DepartIds
	default:
		return ""
	}
}

// filterRule 单值过滤规则决策结果（Java QueryRuleEnum 子集）。
type filterRule struct {
	Op    string // "=" 精确 / "LIKE" 全模糊 / "LEFT_LIKE" 后缀 / "RIGHT_LIKE" 前缀 / "IN" 多值 / "<>" 不等于
	Value string
}

// parseFilterRule QueryGenerator 字符串规则子集（convert2Rule + replaceValue）：
//   - 值含 `*`：`*x*` 全模糊、`*x` 左模糊（后缀匹配）、`x*` 右模糊（前缀匹配）；
//   - 含逗号拆分走 IN；
//   - 以 `!` 开头走 NE（Java NOT_EQUAL="!"，剥离前缀）；
//   - 其余精确匹配。
//
// Java 的操作符前缀规则（">= x" 等要求操作符后跟空格）与 ",a,b," 环绕多 LIKE
// 分支在 /sys/user/list 的真实消费方中不存在，不迁移（见报告）。
func parseFilterRule(value string) filterRule {
	value = strings.TrimSpace(value)
	switch {
	case strings.Contains(value, "*"):
		switch {
		case strings.HasPrefix(value, "*") && strings.HasSuffix(value, "*") && len(value) >= 2:
			return filterRule{Op: "LIKE", Value: value[1 : len(value)-1]}
		case strings.HasPrefix(value, "*"):
			return filterRule{Op: "LEFT_LIKE", Value: value[1:]}
		default:
			return filterRule{Op: "RIGHT_LIKE", Value: strings.TrimSuffix(value, "*")}
		}
	case strings.Contains(value, ","):
		return filterRule{Op: "IN", Value: value}
	case strings.HasPrefix(value, "!"):
		return filterRule{Op: "<>", Value: value[1:]}
	default:
		return filterRule{Op: "=", Value: value}
	}
}

// applyStringFilter 将字符串过滤规则绑定到模型（全部占位符绑定；空值跳过）。
func applyStringFilter(m *gdb.Model, col, value string) *gdb.Model {
	value = strings.TrimSpace(value)
	if value == "" {
		return m
	}
	rule := parseFilterRule(value)
	switch rule.Op {
	case "LIKE":
		return m.WhereLike(col, "%"+rule.Value+"%")
	case "LEFT_LIKE":
		return m.WhereLike(col, "%"+rule.Value)
	case "RIGHT_LIKE":
		return m.WhereLike(col, rule.Value+"%")
	case "IN":
		return m.WhereIn(col, SplitIDs(rule.Value))
	case "<>":
		return m.Where(fmt.Sprintf("%s <> ?", col), rule.Value)
	default:
		return m.Where(col, rule.Value)
	}
}

// applyIntFilter 数值型过滤：解析为整数生效，失败跳过（对齐 installMplus
// 逐字段 try/catch 吞异常的语义）。
func applyIntFilter(m *gdb.Model, col, value string) *gdb.Model {
	value = strings.TrimSpace(value)
	if value == "" {
		return m
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return m
	}
	return m.Where(col, n)
}

// applyOrder 排序处理（Java doMultiFieldsOrder：column/order 均非空才生效，
// column 依次剥离 _dictText/_url 后缀 → camelToUnderline → 按 order 是否含
// ASC 决定升/降序）。Go 版加固：列名须同时通过标识符正则与 sys_user 列集合
// 校验（并排除 password/salt），否则静默忽略（Java 为黑名单过滤，非法列会
// 触发 SQL 异常）。
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

// camelToUnderline 驼峰转下划线（同 Java oConvertUtils.camelToUnderline）。
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

// SplitIDs 逗号拆分并去除空白项（前端 departId/ids 普遍带尾逗号；Java split
// 的尾空段在 IN 列表里恒不命中，丢弃后语义一致）。
func SplitIDs(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// collectDepartIds 汇总页内全部用户的 depart_ids 引用（供表字典回译）。
func collectDepartIds(users []entity.SysUser) []string {
	seen := make(map[string]struct{})
	var out []string
	for i := range users {
		for _, id := range SplitIDs(users[i].DepartIds) {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

// translateTableDict 表字典回译：逗号分隔的部门 id → 逗号拼接的部门名
// （对齐 DictAspect.translateDictValue：按 "," 拆分、跳过空白段、未命中的键
// 不拼接；全空时 Java 输出 null，Go 版输出空串，前端渲染同为空白）。
func translateTableDict(departIds string, names map[string]string) string {
	ids := SplitIDs(departIds)
	if len(ids) == 0 {
		return ""
	}
	translated := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := names[id]; ok {
			translated = append(translated, name)
		}
	}
	return strings.Join(translated, ",")
}

// formatDate 生日仅保留日期部分（Java @JsonFormat(pattern="yyyy-MM-dd")）。
func formatDate(v string) string {
	if len(v) > 10 {
		return v[:10]
	}
	return v
}

// formatDateTime 时间仅保留到秒（Java DictAspect 对无 @JsonFormat 的 Date
// 字段统一转 "yyyy-MM-dd HH:mm:ss"）。
func formatDateTime(v string) string {
	if len(v) > 19 {
		return v[:19]
	}
	return v
}
