// Package userdepart 承载用户-部门关联域业务用例（契约蓝本：
// SysUserController L413/L457/L951/L1001/L1016/L1062/L1091/L1127/L1157 +
// SysUserDepartServiceImpl / SysDepartServiceImpl / SysUserServiceImpl /
// SysDepartRoleUserServiceImpl / SysUserMapper.xml / SysDepartMapper.xml）。
//
// 逐端点契约（Java V2.8，路径/方法/参数/响应逐字段对齐）：
//
//  1. GET /sys/user/userDepartList              （SysUserController L413）
//     参数：userId（必填）；实现：SysUserDepartServiceImpl#queryDepartIdsOfUser；
//     result：List<DepartIdModel>{key,value,title,children}；有数据时
//     success=true+「查找成功」，无数据时 success=false+「查找失败」。
//     前端消费：web/src/views/system/modules/UserModal.vue（checkedDepartKeys/title）。
//
//  2. GET /sys/user/queryUserByDepId            （SysUserController L457）
//     参数：id（部门id，必填）、realname（可选，模糊匹配）；
//     实现：先取部门的 org_code，再 queryUserByDepCode（org_code 前缀 LIKE
//     圈定当前+下级部门 → sys_user_depart → sys_user）；
//     result：List<SysUser>，password/salt 不输出（Java WRITE_ONLY），
//     orgCode 临时借位存「用户部门名（逗号拼接）」。
//     前端消费：web/src/components/jeecgbiz/modal/JSelectUserByDepModal.vue
//     （id/username/realname/sex/phone/orgCode），不分页，客户端本地分页。
//
//  3. GET /sys/user/departUserList              （SysUserController L951）
//     参数：pageNo(默认1)/pageSize(默认10)/depId（可选）+ QueryGenerator
//     字符串规则（username/realname：* 通配 LIKE、逗号 IN、!= NE、其余 EQ）
//     + column/order 排序参数；depId 为空且当前用户 userIdentity=2 时按
//     其负责部门（depart_ids）取全部子部门；result：IPage 分页结构
//     （records/total/size/current/pages），dep 圈定为空时 result=null；
//     records 的 orgCode 被覆写为部门名（逗号拼接），并注入 sex_dictText
//     （Java DictAspect 对 IPage 的字典注入）。
//     前端消费：web/src/views/system/modules/DeptUserInfo.vue
//     （records/total + id/username/realname/orgCode/sex_dictText/phone）。
//
//  4. GET /sys/user/queryByOrgCode              （SysUserController L1001）
//     参数：pageNo/pageSize/orgCode（必填）+ realname/workNo（模糊）；
//     SQL 移植 SysUserMapper.getUserByOrgCode：三表 INNER JOIN，
//     sys_user.del_flag=0 AND org_code 前缀 LIKE，ORDER BY sys_depart.org_code ASC；
//     result：IPage<SysUserSysDepartModel>
//     {id,realname,workNo,post,telephone,email,phone,departId,departName}。
//     前端暂无调用方，按 Java 契约完整实现。
//
//  5. GET /sys/user/queryByOrgCodeForAddressList（SysUserController L1016）
//     参数：同 4，orgCode 可选（为空时全查）；
//     在 4 的基础上把同用户多部门记录按 userId 合并（departName 用 " | " 拼接，
//     id 更名为 userId）；total 仍为合并前数量（Java 保持一致）。
//     前端消费：web/src/views/system/modules/AddressListRight.vue
//     （records/total + userId/realname/workNo/departName/post/telephone/email）。
//
//  6. POST /sys/user/editSysDepartWithUser      （SysUserController L1062）
//     请求体：{depId, userIdList:[]}（Java SysDepartUsersVO）；
//     级联语义：逐个用户判重（dep_id+user_id），不存在才插入——增量补挂，
//     不删除已有关系（Java 保持一致，非先删后插）。
//     前端消费：web/src/views/system/modules/DeptUserInfo.vue（res.message）。
//
//  7. DELETE /sys/user/deleteUserInDepart       （SysUserController L1091）
//     参数：depId、userId（查询串）；先做 lessThanUserRoleLevel 角色等级
//     校验（当前用户等级 < 目标用户等级 → 权限不足），再物理删除
//     sys_user_depart 一行；删除成功后级联清理该部门角色下该用户的
//     sys_depart_role_user 记录（Java SysDepartController 同款级联）；
//     无关联时报「当前选中部门与用户无关联关系!」。
//
//  8. DELETE /sys/user/deleteUserInDepartBatch  （SysUserController L1127）
//     参数：depId、userIds（逗号分隔，前端带尾逗号）；逐个校验角色等级后
//     批量物理删除，删除有行才级联清理部门角色用户关系；无论行数多少
//     只要无异常均返回「删除成功!」（Java 保持一致）。
//
//  9. GET /sys/user/getCurrentUserDeparts       （SysUserController L1157）
//     无参数（取当前登录人）；实现：SysDepartMapper.queryUserDeparts
//     （sys_user_depart 反查 sys_depart）；result：{list:[SysDepart...],
//     orgCode:<当前用户 org_code>}。
//     前端消费：web/src/components/tools/DepartSelect.vue
//     （list[].id/orgCode/departName + result.orgCode）。
//
// 安全与一致性说明：
//   - 全部 SQL 占位符绑定；排序列过 ^[a-zA-Z0-9_]+$ 白名单 + sys_user 列集合
//     校验（Java QueryGenerator 仅黑名单过滤，Go 版升级为白名单）；
//   - del_flag 为 0/1 标志位（非时间戳）：本域所有涉及 sys_depart/sys_user 的
//     查询手动过滤 del_flag=0（Java 版部分 SQL 遗漏该过滤，Go 版按项目规范
//     统一补齐，属安全加固，见各方法注释）；sys_user_depart 无 del_flag；
//   - sys_depart_role / sys_depart_role_user 无对应 dao（文件所有权限定），
//     级联读写用占位符绑定的原生 SQL 完成（先例：service/permission）。
package userdepart

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 业务错误（HTTP 层映射为 Result{success:false}，话术与 Java 保持一致）。
var (
	// ErrPermissionDenied 角色等级不足（Java lessThanUserRoleLevel → error500("权限不足")）。
	ErrPermissionDenied = errors.New("权限不足")
	// ErrNoRelation 部门与用户无关联（Java deleteUserInDepart 的 else 分支话术）。
	ErrNoRelation = errors.New("当前选中部门与用户无关联关系!")
	// ErrDepartNotFound 部门不存在（Java 版此处为 NPE 500，Go 版给出明确话术）。
	ErrDepartNotFound = errors.New("部门不存在")
	// ErrParamRequired 必填参数缺失。
	ErrParamRequired = errors.New("参数不能为空")
)

// identRegexp 限定动态 SQL 标识符：仅字母/数字/下划线。
var identRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// 常量（对齐 Java CommonConstant / SysUserServiceImpl）。
const (
	// userIdentitySuperior 上级身份（CommonConstant.USER_IDENTITY_2）。
	userIdentitySuperior = 2
	// delFlagNormal 正常记录标志位。
	delFlagNormal = 0
	// sexDictCode 性别字典编码（Java SysUser.sex @Dict(dicCode="sex")）。
	sexDictCode = "sex"
)

// DepartIdModel 用户已关联部门节点（对应 Java org.jeecg.modules.system.model.DepartIdModel，
// convertByUserDepart：key/value=部门 id，title=部门名，children 恒为空数组）。
type DepartIdModel struct {
	Key      string          `json:"key"`
	Value    string          `json:"value"`
	Title    string          `json:"title"`
	Children []DepartIdModel `json:"children"`
}

// SysUserSysDepartModel 用户+部门联表模型（对应 Java SysUserSysDepartModel，
// JSON 字段名与 Java 一致）。
type SysUserSysDepartModel struct {
	Id         string `json:"id"`
	Realname   string `json:"realname"`
	WorkNo     string `json:"workNo"`
	Post       string `json:"post"`
	Telephone  string `json:"telephone"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	DepartId   string `json:"departId"`
	DepartName string `json:"departName"`
}

// SysUserSysDepartAddressModel 通讯录合并后的记录（Java queryByOrgCodeForAddressList：
// 原模型去掉 id，改以 userId 输出，并保留 departId/departName）。
type SysUserSysDepartAddressModel struct {
	Realname   string `json:"realname"`
	WorkNo     string `json:"workNo"`
	Post       string `json:"post"`
	Telephone  string `json:"telephone"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	DepartId   string `json:"departId"`
	DepartName string `json:"departName"`
	UserId     string `json:"userId"`
}

// DepartUserRecord 部门用户分页记录：SysUser 全字段 + DictAspect 注入的
// sex_dictText + 被覆写为部门名的 orgCode（Java 临时借位展示）。
type DepartUserRecord struct {
	entity.SysUser
	SexDictText string `json:"sex_dictText"`
}

// DepartUserListInput departUserList 查询入参。
type DepartUserListInput struct {
	PageNo      int
	PageSize    int
	DepId       string
	Username    string
	Realname    string
	OrderColumn string
	OrderType   string
}

// QueryUserByOrgCodeInput queryByOrgCode / queryByOrgCodeForAddressList 查询入参。
type QueryUserByOrgCodeInput struct {
	PageNo   int
	PageSize int
	OrgCode  string
	Realname string
	WorkNo   string
}

// PageParam/PageResult 复用分页模型。
type (
	PageParam  = page.Param
	PageResult = page.Result
)

// sysUserOrderColumns departUserList 排序列白名单（来自 dao.SysUser 列集合）。
var sysUserOrderColumns = buildColumnSet(dao.SysUser.Columns())

// buildColumnSet 反射收集列结构体的列名集合（先例：service/category）。
func buildColumnSet(cols interface{}) map[string]struct{} {
	set := make(map[string]struct{}, 32)
	v := reflect.ValueOf(cols)
	for i := 0; i < v.Type().NumField(); i++ {
		set[v.Field(i).String()] = struct{}{}
	}
	return set
}

// Service 部门关联域业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// QueryDepartIdsOfUser 按用户查其已关联的部门列表（端点 1）。
//
// Java queryDepartIdsOfUser：sys_user_depart 反查 dep_id 集合 → sys_depart
// 按 id IN 回查（Java 未过滤 del_flag，Go 版按规范补 del_flag=0）；
// 输出 DepartIdModel{key,value,title,children}，children 恒为空数组
// （Java 字段初始化为空 List）。
func (s *Service) QueryDepartIdsOfUser(ctx context.Context, userId string) ([]DepartIdModel, error) {
	if strings.TrimSpace(userId) == "" {
		return nil, ErrParamRequired
	}
	udCols := dao.SysUserDepart.Columns()
	var relations []entity.SysUserDepart
	if err := dao.SysUserDepart.Ctx(ctx).
		Where(udCols.UserId, userId).
		Scan(&relations); err != nil {
		return nil, fmt.Errorf("查询用户部门关系失败: %w", err)
	}
	if len(relations) == 0 {
		// Java：无关联时返回 null，控制器据此输出 success=false「查找失败」
		return nil, nil
	}
	depIds := make([]string, 0, len(relations))
	for _, rel := range relations {
		depIds = append(depIds, rel.DepId)
	}
	depCols := dao.SysDepart.Columns()
	var depts []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Where(depCols.DelFlag, delFlagNormal).
		WhereIn(depCols.Id, depIds).
		Scan(&depts); err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	list := make([]DepartIdModel, 0, len(depts))
	for _, dept := range depts {
		list = append(list, DepartIdModel{
			Key:      dept.Id,
			Value:    dept.Id,
			Title:    dept.DepartName,
			Children: []DepartIdModel{},
		})
	}
	return list, nil
}

// QueryUserByDepId 按部门查用户（端点 2，含下级部门）。
//
// Java queryUserByDepId：sys_depart getById → org_code 前缀 LIKE 圈定部门
// （Java likeRight，Go 版补 del_flag=0）→ sys_user_depart 收集 user_id →
// sys_user id IN + realname LIKE %v%（Java like）→ password/salt 置空
// （Go 版实体 json:"-" 天然不输出）。部门不存在时 Java 因 NPE 报 500，
// Go 版返回明确错误。
func (s *Service) QueryUserByDepId(ctx context.Context, depId, realname string) ([]entity.SysUser, error) {
	if strings.TrimSpace(depId) == "" {
		return nil, ErrParamRequired
	}
	depCols := dao.SysDepart.Columns()
	dep, err := dao.SysDepart.Ctx(ctx).
		Where(depCols.Id, depId).
		Where(depCols.DelFlag, delFlagNormal).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	if dep.IsEmpty() {
		return nil, ErrDepartNotFound
	}
	orgCode := dep[depCols.OrgCode].String()

	subDeps, err := dao.SysDepart.Ctx(ctx).
		Fields(depCols.Id).
		Where(depCols.DelFlag, delFlagNormal).
		WhereLike(depCols.OrgCode, orgCode+"%").
		All()
	if err != nil {
		return nil, fmt.Errorf("查询下级部门失败: %w", err)
	}
	if len(subDeps) == 0 {
		return []entity.SysUser{}, nil
	}
	depIds := make([]string, 0, len(subDeps))
	for _, rec := range subDeps {
		depIds = append(depIds, rec[depCols.Id].String())
	}

	udCols := dao.SysUserDepart.Columns()
	relations, err := dao.SysUserDepart.Ctx(ctx).
		Fields(udCols.UserId).
		WhereIn(udCols.DepId, depIds).
		All()
	if err != nil {
		return nil, fmt.Errorf("查询部门用户关系失败: %w", err)
	}
	if len(relations) == 0 {
		return []entity.SysUser{}, nil
	}
	userIds := make([]string, 0, len(relations))
	for _, rec := range relations {
		userIds = append(userIds, rec[udCols.UserId].String())
	}

	userCols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).
		Where(userCols.DelFlag, delFlagNormal).
		WhereIn(userCols.Id, userIds)
	if v := strings.TrimSpace(realname); v != "" {
		m = m.WhereLike(userCols.Realname, "%"+v+"%")
	}
	var list []entity.SysUser
	if err := m.Scan(&list); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return list, nil
}

// DepartUserList 部门用户分页（端点 3）。
//
// 级联语义：depId 为空且当前用户为上级（userIdentity=2）时，按其负责部门
// （sys_user.depart_ids）展开全部子部门；否则按 depId 取当前+下级部门
// （getSubDepIdsByDepId：del_flag=0 且 org_code 前缀 LIKE 自身）。
// 部门圈定为空 → 返回 nil（控制器输出 result=null + success=true，同 Java）。
// 记录的 orgCode 覆写为用户部门名（逗号拼接），并注入 sex_dictText。
func (s *Service) DepartUserList(ctx context.Context, currentUID string, in DepartUserListInput) (*PageResult, error) {
	subDepIds, err := s.resolveSubDepIds(ctx, currentUID, in.DepId)
	if err != nil {
		return nil, err
	}
	if len(subDepIds) == 0 {
		return nil, nil
	}

	userCols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx).Where(userCols.DelFlag, delFlagNormal)
	m = applyUserFilter(m, userCols.Username, in.Username)
	m = applyUserFilter(m, userCols.Realname, in.Realname)
	// Java：queryWrapper.inSql("sys_user.id", "select user_id from sys_user_depart where dep_id in ('...')")
	// —— Java 用字符串拼接部门 id，Go 版改为占位符绑定（同语义）。
	m = m.Where(fmt.Sprintf("%s IN (SELECT user_id FROM sys_user_depart WHERE dep_id IN (?))", userCols.Id), subDepIds)
	m = applyOrder(m, in.OrderColumn, in.OrderType)

	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计部门用户数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]DepartUserRecord{}, 0, in.PageNo, in.PageSize), nil
	}
	var users []entity.SysUser
	if err := m.Page(in.PageNo, in.PageSize).Scan(&users); err != nil {
		return nil, fmt.Errorf("查询部门用户列表失败: %w", err)
	}

	userIds := make([]string, 0, len(users))
	for i := range users {
		userIds = append(userIds, users[i].Id)
	}
	depNames, err := s.getDepNamesByUserIds(ctx, userIds)
	if err != nil {
		return nil, err
	}
	sexDict, err := s.sexDictTextMap(ctx)
	if err != nil {
		return nil, err
	}
	records := make([]DepartUserRecord, 0, len(users))
	for i := range users {
		// Java：item.setOrgCode(useDepNames.get(item.getId())) —— orgCode 借位存部门名
		users[i].OrgCode = depNames[users[i].Id]
		records = append(records, DepartUserRecord{
			SysUser:     users[i],
			SexDictText: sexDict[fmt.Sprintf("%d", users[i].Sex)],
		})
	}
	return page.New(records, total, in.PageNo, in.PageSize), nil
}

// QueryUserByOrgCode 按 orgCode 查用户（端点 4/5 共用核心，含下级部门）。
//
// SQL 移植 SysUserMapper.getUserByOrgCode + getUserByOrgCodeTotal：
//
//	SELECT 用户字段 + 部门字段 FROM sys_depart
//	INNER JOIN sys_user_depart ON dep_id = sys_depart.id
//	INNER JOIN sys_user ON sys_user.id = user_id
//	WHERE sys_user.del_flag = 0 [AND sys_depart.del_flag = 0（Go 版加固补齐）]
//	      AND sys_depart.org_code LIKE 'orgCode%'
//	      [AND realname LIKE %v%] [AND work_no LIKE %v%]
//	ORDER BY sys_depart.org_code ASC
//
// Java 用 ${orgCode}% 字符串插值（注入面），Go 版一律占位符绑定。
// 注意 total 统计的是「用户-部门对」行数（同 Java COUNT(1)，不去重）。
func (s *Service) QueryUserByOrgCode(ctx context.Context, in QueryUserByOrgCodeInput) ([]SysUserSysDepartModel, int, error) {
	const countSQL = `SELECT COUNT(1) AS total FROM sys_depart
INNER JOIN sys_user_depart ON sys_user_depart.dep_id = sys_depart.id
INNER JOIN sys_user ON sys_user.id = sys_user_depart.user_id
WHERE `
	const listSQL = `SELECT
	sys_user.id AS id,
	sys_user.realname AS realname,
	sys_user.work_no AS workNo,
	sys_user.post AS post,
	sys_user.telephone AS telephone,
	sys_user.email AS email,
	sys_user.phone AS phone,
	sys_depart.id AS departId,
	sys_depart.depart_name AS departName
FROM sys_depart
INNER JOIN sys_user_depart ON sys_user_depart.dep_id = sys_depart.id
INNER JOIN sys_user ON sys_user.id = sys_user_depart.user_id
WHERE `

	where := "sys_user.del_flag = 0 AND sys_depart.del_flag = 0 AND sys_depart.org_code LIKE ?"
	args := []interface{}{in.OrgCode + "%"}
	if v := strings.TrimSpace(in.Realname); v != "" {
		where += " AND sys_user.realname LIKE ?"
		args = append(args, "%"+v+"%")
	}
	if v := strings.TrimSpace(in.WorkNo); v != "" {
		where += " AND sys_user.work_no LIKE ?"
		args = append(args, "%"+v+"%")
	}

	one, err := dao.SysUser.DB().GetOne(ctx, countSQL+where, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("统计组织用户数量失败: %w", err)
	}
	total := one["total"].Int()

	offset := (in.PageNo - 1) * in.PageSize
	listArgs := append(append([]interface{}{}, args...), offset, in.PageSize)
	records, err := dao.SysUser.DB().GetAll(ctx,
		listSQL+where+" ORDER BY sys_depart.org_code ASC LIMIT ?, ?", listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询组织用户列表失败: %w", err)
	}
	list := make([]SysUserSysDepartModel, 0, len(records))
	if err := records.Structs(&list); err != nil {
		return nil, 0, fmt.Errorf("解析组织用户数据失败: %w", err)
	}
	return list, total, nil
}

// QueryByOrgCode orgCode 分页查询用户（端点 4，记录按「用户-部门对」输出）。
func (s *Service) QueryByOrgCode(ctx context.Context, in QueryUserByOrgCodeInput) (*PageResult, error) {
	if strings.TrimSpace(in.OrgCode) == "" {
		return nil, ErrParamRequired
	}
	list, total, err := s.QueryUserByOrgCode(ctx, in)
	if err != nil {
		return nil, err
	}
	return page.New(list, total, in.PageNo, in.PageSize), nil
}

// QueryByOrgCodeForAddressList 通讯录查询（端点 5）。
//
// 与 Java 一致：同一用户的多条部门记录仅在同一页内合并（跨页不合并，
// Java 也如此）；departName 用 " | " 拼接；id 更名为 userId；
// total 为合并前的行数（Java result.setRecords 前未改 total）。
// orgCode 为空时全查（Java：LIKE '%'）。
func (s *Service) QueryByOrgCodeForAddressList(ctx context.Context, in QueryUserByOrgCodeInput) (*PageResult, error) {
	list, total, err := s.QueryUserByOrgCode(ctx, in)
	if err != nil {
		return nil, err
	}
	return page.New(mergeAddressRecords(list), total, in.PageNo, in.PageSize), nil
}

// mergeAddressRecords 通讯录记录合并：同用户多部门仅保留首条，departName
// 用 " | " 拼接，id 更名为 userId（Java queryByOrgCodeForAddressList 的
// hasUser 合并逻辑）。
func mergeAddressRecords(list []SysUserSysDepartModel) []SysUserSysDepartAddressModel {
	merged := make([]SysUserSysDepartAddressModel, 0, len(list))
	indexByUser := make(map[string]int, len(list))
	for i := range list {
		if idx, ok := indexByUser[list[i].Id]; ok {
			merged[idx].DepartName += " | " + list[i].DepartName
			continue
		}
		indexByUser[list[i].Id] = len(merged)
		merged = append(merged, SysUserSysDepartAddressModel{
			Realname:   list[i].Realname,
			WorkNo:     list[i].WorkNo,
			Post:       list[i].Post,
			Telephone:  list[i].Telephone,
			Email:      list[i].Email,
			Phone:      list[i].Phone,
			DepartId:   list[i].DepartId,
			DepartName: list[i].DepartName,
			UserId:     list[i].Id,
		})
	}
	return merged
}

// EditSysDepartWithUser 把用户加入/调整部门（端点 6）。
//
// Java 级联语义（SysUserController#editSysDepartWithUser）：逐个用户按
// （dep_id, user_id）判重，不存在才插入——增量补挂、不删除已有关系，
// 也未更新 sys_user.org_code（Java 亦不改）；sys_user_depart 无时间列/
// del_flag，主键为服务端生成（Java ID_WORKER_STR，Go guid.S()）。
func (s *Service) EditSysDepartWithUser(ctx context.Context, depId string, userIdList []string) error {
	if strings.TrimSpace(depId) == "" {
		return ErrParamRequired
	}
	udCols := dao.SysUserDepart.Columns()
	for _, userId := range userIdList {
		if strings.TrimSpace(userId) == "" {
			continue
		}
		one, err := dao.SysUserDepart.Ctx(ctx).
			Where(udCols.DepId, depId).
			Where(udCols.UserId, userId).
			One()
		if err != nil {
			return fmt.Errorf("查询用户部门关系失败: %w", err)
		}
		if !one.IsEmpty() {
			continue
		}
		if _, err := dao.SysUserDepart.Ctx(ctx).Data(do.SysUserDepart{
			Id:     guid.S(),
			UserId: userId,
			DepId:  depId,
		}).Insert(); err != nil {
			return fmt.Errorf("保存用户部门关系失败: %w", err)
		}
	}
	return nil
}

// DeleteUserInDepart 删除指定部门的用户关系（端点 7）。
//
// Java 级联语义（deleteUserInDepart）：先删 sys_user_depart 一行（删除
// 0 行报「无关联关系」），成功后查 sys_depart_role（depart_id=depId）的
// 角色 id 集合，再删 sys_depart_role_user（user_id+角色 id 集合）。
// Go 版将删除与级联放入事务（Java 为两条独立语句，语句顺序一致，
// 事务仅加强原子性）。
func (s *Service) DeleteUserInDepart(ctx context.Context, depId, userId, currentUID string) error {
	if strings.TrimSpace(depId) == "" || strings.TrimSpace(userId) == "" {
		return ErrParamRequired
	}
	if err := s.checkUserRoleLevel(ctx, currentUID, userId); err != nil {
		return err
	}
	udCols := dao.SysUserDepart.Columns()
	if err := dao.SysUserDepart.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := dao.SysUserDepart.Ctx(ctx).
			Where(udCols.DepId, depId).
			Where(udCols.UserId, userId).
			Delete()
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrNoRelation
		}
		return s.removeDeptRoleUser(ctx, []string{userId}, depId)
	}); err != nil {
		if errors.Is(err, ErrNoRelation) {
			return err
		}
		return fmt.Errorf("删除用户部门关系失败: %w", err)
	}
	return nil
}

// DeleteUserInDepartBatch 批量删除指定部门的用户关系（端点 8）。
//
// Java 级联语义（deleteUserInDepartBatch）：先逐个校验目标用户角色等级，
// 再按 dep_id+user_id IN 批量物理删除，删除有行才级联清理部门角色用户
// 关系；无异常一律返回「删除成功!」（与删除行数无关）。
func (s *Service) DeleteUserInDepartBatch(ctx context.Context, depId string, userIds []string, currentUID string) error {
	if strings.TrimSpace(depId) == "" || len(userIds) == 0 {
		return ErrParamRequired
	}
	// Java：先全部校验再删除（任一目标等级高于当前用户即整体拒绝）
	for _, userId := range userIds {
		if err := s.checkUserRoleLevel(ctx, currentUID, userId); err != nil {
			return err
		}
	}
	udCols := dao.SysUserDepart.Columns()
	if err := dao.SysUserDepart.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		result, err := dao.SysUserDepart.Ctx(ctx).
			Where(udCols.DepId, depId).
			WhereIn(udCols.UserId, userIds).
			Delete()
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			// Java：b=false 时跳过级联，仍返回「删除成功!」
			return nil
		}
		return s.removeDeptRoleUser(ctx, userIds, depId)
	}); err != nil {
		return fmt.Errorf("批量删除用户部门关系失败: %w", err)
	}
	return nil
}

// GetCurrentUserDeparts 当前登录人所属部门（端点 9）。
//
// Java queryUserDeparts：sys_user_depart 反查 sys_depart（Java 未过滤
// del_flag，Go 版按规范补齐）；orgCode 取自当前用户 sys_user.org_code
// （Java 取自会话 LoginUser，Go 版以库内最新值为准）。
func (s *Service) GetCurrentUserDeparts(ctx context.Context, currentUID string) ([]entity.SysDepart, string, error) {
	if strings.TrimSpace(currentUID) == "" {
		return nil, "", ErrParamRequired
	}
	var list []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Where(dao.SysDepart.Columns().DelFlag, delFlagNormal).
		Where(fmt.Sprintf("%s IN (SELECT dep_id FROM sys_user_depart WHERE user_id = ?)",
			dao.SysDepart.Columns().Id), currentUID).
		Scan(&list); err != nil {
		return nil, "", fmt.Errorf("查询当前用户部门失败: %w", err)
	}
	user, err := s.getUserById(ctx, currentUID)
	if err != nil {
		return nil, "", err
	}
	orgCode := ""
	if user != nil {
		orgCode = user.OrgCode
	}
	return list, orgCode, nil
}

// resolveSubDepIds 解析部门圈定：depId 为空时按当前用户身份展开其负责部门
// 的全部子部门（Java departUserList 的 depId 缺省分支），否则按 depId 取
// 当前+下级部门。返回空集合表示「无可见部门」（控制器输出 result=null）。
func (s *Service) resolveSubDepIds(ctx context.Context, currentUID, depId string) ([]string, error) {
	if depId != "" {
		return s.getSubDepIdsByDepId(ctx, depId)
	}
	user, err := s.getUserById(ctx, currentUID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.UserIdentity != userIdentitySuperior {
		// Java：userIdentity != 2 时 subDepids 保持空 → result=null
		return nil, nil
	}
	return s.getMySubDepIdsByDepId(ctx, user.DepartIds)
}

// getSubDepIdsByDepId 按部门 id 取当前+下级部门 id（SQL 移植
// SysDepartMapper.getSubDepIdsByDepId：del_flag='0' 且 org_code LIKE
// 自身前缀，部门不存在时内层子查询为 NULL → 自然返回空集合）。
func (s *Service) getSubDepIdsByDepId(ctx context.Context, depId string) ([]string, error) {
	const query = `SELECT id FROM sys_depart
WHERE del_flag = '0'
  AND org_code LIKE CONCAT((SELECT org_code FROM sys_depart WHERE id = ?), '%')`
	records, err := dao.SysDepart.DB().GetAll(ctx, query, depId)
	if err != nil {
		return nil, fmt.Errorf("查询下级部门失败: %w", err)
	}
	ids := make([]string, 0, len(records))
	for _, rec := range records {
		ids = append(ids, rec["id"].String())
	}
	return ids, nil
}

// getMySubDepIdsByDepId 按用户负责部门（depart_ids，逗号分隔）展开子部门。
//
// Java 链路：getMyDeptParentOrgCode（del_flag=0 的部门，按 org_code 升序）
// → getMyDeptParentNode（按 3 位编码分组逗号拼接）→ getSubDepIdsByOrgCodes
// （org_code LIKE 每个编码前缀，del_flag=0）。分组拼接后拆分即全部负责
// 部门编码本身，Go 版直接以全部 org_code 为 LIKE 模式（等价，去重无副作用）。
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

	// getSubDepIdsByOrgCodes：del_flag='0' AND (org_code LIKE ? OR ...)
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

// getDepNamesByUserIds 批量查用户部门名（SQL 移植 SysUserMapper
// .getDepNamesByUserIds：sys_user_depart 与 sys_depart 按 dep_id 相连，
// 同用户多部门按行序「首个胜出、后续逗号追加」，与 Java 一致；
// Go 版补 sys_depart.del_flag=0 过滤）。
func (s *Service) getDepNamesByUserIds(ctx context.Context, userIds []string) (map[string]string, error) {
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

// sexDictTextMap 性别字典（dict_code=sex）的 value→text 映射，用于
// departUserList 注入 sex_dictText（Java DictAspect 对 IPage 记录的
// 字典注入；sys_dict/sys_dict_item 查询过滤 del_flag=0/status=1）。
func (s *Service) sexDictTextMap(ctx context.Context) (map[string]string, error) {
	dictCols := dao.SysDict.Columns()
	one, err := dao.SysDict.Ctx(ctx).
		Where(dictCols.DictCode, sexDictCode).
		Where(dictCols.DelFlag, delFlagNormal).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询性别字典失败: %w", err)
	}
	res := make(map[string]string, 4)
	if one.IsEmpty() {
		return res, nil
	}
	itemCols := dao.SysDictItem.Columns()
	records, err := dao.SysDictItem.Ctx(ctx).
		Fields(itemCols.ItemValue, itemCols.ItemText).
		Where(itemCols.DictId, one["id"].String()).
		All()
	if err != nil {
		return nil, fmt.Errorf("查询性别字典项失败: %w", err)
	}
	for _, rec := range records {
		res[rec[itemCols.ItemValue].String()] = rec[itemCols.ItemText].String()
	}
	return res, nil
}

// getUserRoleLevel 用户角色等级（SQL 移植 SysUserRoleMapper.getUserRoleLevel：
// max(role_level)，无角色时返回 -1，与 Java null→-1 一致）。
func (s *Service) getUserRoleLevel(ctx context.Context, userId string) (int, error) {
	const query = `SELECT MAX(role_level) AS level FROM sys_role
WHERE id IN (SELECT role_id FROM sys_user_role WHERE user_id = ?)`
	one, err := dao.SysUserDepart.DB().GetOne(ctx, query, userId)
	if err != nil {
		return 0, fmt.Errorf("查询用户角色等级失败: %w", err)
	}
	if one.IsEmpty() || one["level"].IsNil() {
		return -1, nil
	}
	return one["level"].Int(), nil
}

// checkUserRoleLevel 角色等级校验（Java BaseController.lessThanUserRoleLevel：
// 当前用户等级 < 目标用户等级 → 权限不足；当前用户缺失按拒绝处理）。
func (s *Service) checkUserRoleLevel(ctx context.Context, currentUID, targetUID string) error {
	if strings.TrimSpace(currentUID) == "" {
		return ErrPermissionDenied
	}
	currentLevel, err := s.getUserRoleLevel(ctx, currentUID)
	if err != nil {
		return err
	}
	targetLevel, err := s.getUserRoleLevel(ctx, targetUID)
	if err != nil {
		return err
	}
	if currentLevel < targetLevel {
		return ErrPermissionDenied
	}
	return nil
}

// removeDeptRoleUser 级联清理部门角色用户关系（Java
// SysDepartRoleUserServiceImpl#removeDeptRoleUser：按部门查角色 id 集合，
// 删除 sys_depart_role_user 中这些用户+角色的记录；Java 逐用户重复查询，
// Go 版合并为一次查询一次删除，语句效果等价）。
func (s *Service) removeDeptRoleUser(ctx context.Context, userIds []string, depId string) error {
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
	// sys_depart_role_user 无 dao（文件所有权限定），用占位符绑定的原生 SQL
	if _, err := dao.SysUserDepart.DB().Exec(ctx,
		"DELETE FROM sys_depart_role_user WHERE user_id IN (?) AND drole_id IN (?)",
		userIds, roleIds); err != nil {
		return fmt.Errorf("删除部门角色用户关系失败: %w", err)
	}
	return nil
}

// getUserById 按 id 查当前用户（del_flag=0；不存在返回 nil）。
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

// userFilterRule 单值过滤规则决策结果（Java QueryRuleEnum 子集）。
type userFilterRule struct {
	Op    string // "=" 精确 / "LIKE" 模糊 / "IN" 多值 / "<>" 不等于
	Value string // 非 IN 时的绑定值
}

// parseUserFilterRule QueryGenerator 字符串规则子集（Java convert2Rule +
// replaceValue）：值含 `*` 时 `*`→% 走 LIKE（前后缀位置由 * 决定）；
// 含逗号拆分走 IN；以 `!=` 开头走 NE（前缀剥离）；其余精确匹配。
func parseUserFilterRule(value string) userFilterRule {
	value = strings.TrimSpace(value)
	switch {
	case strings.Contains(value, "*"):
		return userFilterRule{Op: "LIKE", Value: strings.ReplaceAll(value, "*", "%")}
	case strings.Contains(value, ","):
		return userFilterRule{Op: "IN", Value: value}
	case strings.HasPrefix(value, "!="):
		return userFilterRule{Op: "<>", Value: value[2:]}
	default:
		return userFilterRule{Op: "=", Value: value}
	}
}

// applyUserFilter 将 parseUserFilterRule 的决策绑定到模型（全部占位符绑定）。
func applyUserFilter(m *gdb.Model, col, value string) *gdb.Model {
	value = strings.TrimSpace(value)
	if value == "" {
		return m
	}
	rule := parseUserFilterRule(value)
	switch rule.Op {
	case "LIKE":
		return m.WhereLike(col, rule.Value)
	case "IN":
		return m.WhereIn(col, SplitIDs(rule.Value))
	case "<>":
		return m.Where(fmt.Sprintf("%s <> ?", col), rule.Value)
	default:
		return m.Where(col, rule.Value)
	}
}

// applyOrder 排序处理（Java doMultiFieldsOrder：column/order 均非空才生效，
// column 依次剥离 _dictText/_url 后缀 → camelToUnderline → 按 order 是否含
// ASC 决定升/降序）。Go 版加固：列名须同时通过标识符正则与 sys_user 列集合
// 校验，否则静默忽略（Java 为黑名单过滤，失败抛 500）。
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
	if _, ok := sysUserOrderColumns[col]; !ok {
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

// SplitIDs 逗号拆分并去除空白项（前端 deleteUserInDepartBatch 传尾逗号）。
func SplitIDs(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
