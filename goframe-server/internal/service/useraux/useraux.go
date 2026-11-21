// Package useraux 承载用户管理「用户辅助域」业务用例（契约蓝本：
// api/jeecg-boot-module-system/.../system/controller/SysUserController.java）。
//
// 本域以 sys_user 单表为主（dao.SysUser），全部占位符绑定；回收站的两个
// 端点按 Java 契约联表/级联：recycleBin 联表 sys_user_depart + sys_depart
// 把 orgCode 覆写为部门名（queryLogicDeleted 后处理），deleteRecycleBin
// 在事务内级联物理删除 sys_user_depart + sys_user_role（removeLogicDeleted）。
//
// del_flag 语义（Java SysUser.delFlag 带 @TableLogic）：MyBatis-Plus 内置方法
// （getById/list/listByIds）自动追加 del_flag=0 过滤，回收站专用 SQL 手写
// del_flag=1。Go 版等价实现：普通查询显式 Where(DelFlag, 0)，回收站查询
// Where(DelFlag, 1)，不启用 GoFrame 软删特性（sys_user 无 deleted_at 列）。
package useraux

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// 业务错误（HTTP 层映射为 Result{success:false}，话术与 Java 契约一致）。
var (
	// ErrEntityNotFound 未找到对应实体（Java queryById error500 话术）。
	ErrEntityNotFound = errors.New("未找到对应实体")
	// ErrUserExists 用户账号已存在（Java checkOnlyUser 冲突话术）。
	ErrUserExists = errors.New("用户账号已存在")
	// ErrVerifyFailed 验证失败（Java querySysUser 未命中话术）。
	ErrVerifyFailed = errors.New("验证失败")
)

// SysUserBrief querySysUser 输出白名单（对齐 Java 返回 Map 的键集合）：
// 只暴露 username 与脱敏后的 phone，绝不输出 password/salt 等敏感字段。
type SysUserBrief struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

// Service 用户辅助域业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// QueryById 按 id 查用户详情。
//
// 契约（SysUserController.java L326-337）：
//   - GET /sys/user/queryById?id={id}；
//   - Java sysUserService.getById(id)：@TableLogic 自动追加 del_flag=0；
//   - 未找到 → error500("未找到对应实体")；找到 → result=SysUser 全字段。
//
// 输出形状：entity.SysUser（password/salt 与 Java @JsonProperty(WRITE_ONLY)
// 一致不序列化，Go 侧以 json:"-" 达成同一效果）。
// 前端消费：web/src 未发现直接调用方（任务书提到 DeptUserInfo.vue 实际使用
// departUserList 等接口）；本端点照 Java 契约保留，供外部/后续页面使用。
func (s *Service) QueryById(ctx context.Context, id string) (*entity.SysUser, error) {
	cols := dao.SysUser.Columns()
	var one entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(cols.Id, id).
		Where(cols.DelFlag, 0).
		Scan(&one); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if one.Id == "" {
		return nil, ErrEntityNotFound
	}
	return &one, nil
}

// CheckOnlyUser 用户名/手机号/邮箱唯一性校验。
//
// 契约（SysUserController.java L364-388）：
//   - GET /sys/user/checkOnlyUser?username=|email=|phone=（query 参数绑定
//     SysUser bean，三个字段均可独立校验，可同时传入取交集）；
//   - Java 按非空字段逐个 eq 构造查询（全部为空时退化为查全表），@TableLogic
//     追加 del_flag=0，命中任意记录即视为「已存在」。
//
// 任务书提示的 checkType/dataId「编辑时排除自身」语义在本仓库 V2.8 代码中
// 不存在（全库检索 checkOnlyUser 仅 SysUserController 一处实现，无 dataId
// 参数）；web/src 消费方 Register.vue/Step1.vue 也只在注册/找回密码场景传
// 单字段且只读 res.success。因此 Go 版照实迁移，不实现排除自身。
//
// 返回 exists=true 表示账号已存在（HTTP 层据此返回 success:false）。
func (s *Service) CheckOnlyUser(ctx context.Context, username, email, phone string) (bool, error) {
	cols := dao.SysUser.Columns()
	m := dao.SysUser.Ctx(ctx)
	// 与 Java LambdaQueryWrapper.eq(isNotBlank(...), ...) 一致：非空才加条件；
	// 三字段全空时无条件查询（等价 Java 查全表），存在任何启用用户即判已存在。
	if phone != "" {
		m = m.Where(cols.Phone, phone)
	}
	if email != "" {
		m = m.Where(cols.Email, email)
	}
	if username != "" {
		m = m.Where(cols.Username, username)
	}
	m = m.Where(cols.DelFlag, 0)

	count, err := m.Count()
	if err != nil {
		return false, fmt.Errorf("校验用户唯一性失败: %w", err)
	}
	return count > 0, nil
}

// GenerateUserId 生成用户 ID。
//
// 契约（SysUserController.java L441-449）：
//   - GET /sys/user/generateUserId；
//   - Java UUID.randomUUID().toString().replace("-", "")，即不带连字符的
//     32 位小写十六进制 UUID v4；
//   - 前端 UserModal.vue/DepartWindow.vue 用该 id 作为新增用户主键并联动
//     部门绑定，故长度/字符集必须与 Java 完全一致（32 字符）。
//
// Go 版用已引入的 google/uuid 生成 v4 UUID 后去连字符，等价实现。
func (s *Service) GenerateUserId() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("生成用户ID失败: %w", err)
	}
	return strings.ReplaceAll(id.String(), "-", ""), nil
}

// QueryByIds 按 ids 批量查用户。
//
// 契约（SysUserController.java L813-822）：
//   - GET /sys/user/queryByIds?userIds=id1,id2,...（逗号分隔）；
//   - Java listByIds：@TableLogic 自动追加 del_flag=0（已删除用户不出现在
//     结果中），无排序语义（按存储引擎返回顺序）。
//
// 输出形状：entity.SysUser 数组（password/salt 不序列化）。
// 前端消费：SysAnnouncementModal.vue 读取 result[i].realname 并把整行传给
// SelectUserListModal（列：username/realname/sex_dictText/phone/email/
// status_dictText；后两个 dictText 字段 Java 版本就不存在，前端本就渲染为空，
// Go 版保持一致的原始行输出）。
func (s *Service) QueryByIds(ctx context.Context, ids []string) ([]entity.SysUser, error) {
	list := make([]entity.SysUser, 0)
	if len(ids) == 0 {
		return list, nil
	}
	cols := dao.SysUser.Columns()
	if err := dao.SysUser.Ctx(ctx).
		WhereIn(cols.Id, ids).
		Where(cols.DelFlag, 0).
		Scan(&list); err != nil {
		return nil, fmt.Errorf("批量查询用户失败: %w", err)
	}
	return list, nil
}

// QuerySysUser 按用户名或手机号查询用户（改密/找回密码前校验）。
//
// 契约（SysUserController.java L1274-1303，CVE-2021-37305 修复后版本）：
//   - GET /sys/user/querySysUser?phone=... 或 ?username=...；
//   - phone 非空优先按手机号查（getUserByPhone：WHERE phone=? AND del_flag=0），
//     否则按用户名查（getUserByName：WHERE username=? AND del_flag=0）；
//   - 命中 → result={username, phone(脱敏)}；未命中 → success:false "验证失败"。
//
// 脱敏规则（Java desensitizePhone L1310-1315）：长度 ≥7 时保留前 3 位与后
// 4 位，中间以 **** 替代；不足 7 位原样返回。输出白名单仅 username/phone
// 两个字段，password/salt 等敏感字段天然不会出现。
//
// 前端消费：Step1.vue（忘记密码第一步）读取 result.username/result.phone。
func (s *Service) QuerySysUser(ctx context.Context, username, phone string) (*SysUserBrief, error) {
	cols := dao.SysUser.Columns()
	var (
		one   entity.SysUser
		found bool
	)
	if phone != "" {
		if err := dao.SysUser.Ctx(ctx).
			Where(cols.Phone, phone).
			Where(cols.DelFlag, 0).
			Scan(&one); err != nil {
			return nil, fmt.Errorf("按手机号查询用户失败: %w", err)
		}
		found = one.Id != ""
	}
	if !found && username != "" {
		if err := dao.SysUser.Ctx(ctx).
			Where(cols.Username, username).
			Where(cols.DelFlag, 0).
			Scan(&one); err != nil {
			return nil, fmt.Errorf("按用户名查询用户失败: %w", err)
		}
		found = one.Id != ""
	}
	if !found {
		return nil, ErrVerifyFailed
	}
	return &SysUserBrief{
		Username: one.Username,
		Phone:    maskPhone(one.Phone),
	}, nil
}

// maskPhone 手机号脱敏：保留前 3 位与后 4 位，中间以 **** 替代；
// 长度不足 7 原样返回（对齐 Java desensitizePhone 的 substring 语义，
// 手机号为 ASCII 数字串，字节切片与 Java UTF-16 切片等价）。
func maskPhone(phone string) string {
	if len(phone) >= 7 {
		return phone[:3] + "****" + phone[len(phone)-4:]
	}
	return phone
}

// RecycleBin 回收站用户列表（del_flag=1，无分页）。
//
// 契约（SysUserController.java L1454-1466 + SysUserServiceImpl
// queryLogicDeleted L383-389 + SysUserMapper.selectLogicDeleted）：
//   - GET /sys/user/recycleBin，无请求参数、无分页（Java 注释明示「无分页」）；
//   - 查询条件：del_flag = '1'（仅恰好为 1 的记录）；
//   - 查出后 Java 调用 getDepNamesByUserIds（联表 sys_user_depart + sys_depart）
//     把用户的 orgCode 字段覆写为部门名称（orgCode 临时借位展示，前端
//     UserRecycleBinModal.vue 的「部门」列即消费该字段），无部门的用户
//     Java 写 null，Go 版为空串（渲染同为空白）。
func (s *Service) RecycleBin(ctx context.Context) ([]entity.SysUser, error) {
	list := make([]entity.SysUser, 0)
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().DelFlag, 1).
		Scan(&list); err != nil {
		return nil, fmt.Errorf("查询回收站用户失败: %w", err)
	}
	if len(list) == 0 {
		return list, nil
	}
	userIds := make([]string, 0, len(list))
	for i := range list {
		userIds = append(userIds, list[i].Id)
	}
	depNames, err := s.depNamesByUserIds(ctx, userIds)
	if err != nil {
		return nil, err
	}
	for i := range list {
		// Java：item.setOrgCode(useDepNames.get(item.getId()))
		list[i].OrgCode = depNames[list[i].Id]
	}
	return list, nil
}

// depNamesByUserIds 批量查用户部门名（SQL 移植 SysUserMapper.getDepNamesByUserIds
// 的 Java 聚合语义：同用户多部门按行序「首个胜出、后续逗号追加」；Java 原查询
// 不过滤 sys_depart.del_flag，Go 版补 del_flag=0，与 userdepart/usercore 域
// 同款加固）。
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

// PutRecycleBin 从回收站还原用户（del_flag 1 → 0）。
//
// 契约（SysUserController.java L1474-1484 + SysUserMapper.revertLogicDeleted
// L144-154）：
//   - PUT /sys/user/putRecycleBin，body JSON {"userIds":"id1,id2"}；
//   - Java 端点名叫 put 但语义是「还原」（前端 UserRecycleBinModal.vue 的
//     handleRevert 调用它，提示语「还原成功」；任务书中「移入回收站」为笔误，
//     以 Java 代码 + 前端行为为准）；
//   - UPDATE sys_user SET del_flag=0, update_by=?, update_time=? WHERE
//     del_flag=1 AND id IN (...)：只还原 del_flag=1 的行；
//   - update_by 取当前登录用户名（Go 版从 JWT 中间件上下文取，等价
//     JwtUtil.getUserNameByToken）；update_time 由服务端写 gtime.Now()；
//   - userIds 为空时 Java 仍返回 ok("还原成功")（空操作），Go 版等价处理。
func (s *Service) PutRecycleBin(ctx context.Context, ids []string, updateBy string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.SysUser.Columns()
	if _, err := dao.SysUser.Ctx(ctx).
		Where(cols.DelFlag, 1).
		WhereIn(cols.Id, ids).
		Data(do.SysUser{
			DelFlag:    0,
			UpdateBy:   updateBy,
			UpdateTime: gtime.Now(),
		}).Update(); err != nil {
		return fmt.Errorf("还原用户失败: %w", err)
	}
	return nil
}

// DeleteRecycleBin 从回收站彻底删除用户（物理删除）。
//
// 契约（SysUserController.java L1492-1498 + SysUserServiceImpl
// removeLogicDeleted L397-408 + SysUserMapper.deleteLogicDeleted L157-159）：
//   - DELETE /sys/user/deleteRecycleBin?userIds=id1,id2（query 参数）；
//   - Java removeLogicDeleted 带 @Transactional，顺序为：物理删除 sys_user
//     （仅 del_flag=1）→ 按 user_id 物理删除 sys_user_depart 关联行 →
//     按 user_id 物理删除 sys_user_role 关联行；
//   - userIds 为空时 Java 仍返回 ok("删除成功")（空操作），Go 版等价处理。
//
// Go 版在单事务内按同一语句顺序执行三段物理删除（Java 为独立语句提交，
// 事务仅加强原子性，成功路径一致）。
func (s *Service) DeleteRecycleBin(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.SysUser.Columns()
	udCols := dao.SysUserDepart.Columns()
	urCols := dao.SysUserRole.Columns()
	if err := dao.SysUser.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 物理删除用户行（仅已进回收站的 del_flag=1 行）。
		if _, err := dao.SysUser.Ctx(ctx).
			Where(cols.DelFlag, 1).
			WhereIn(cols.Id, ids).
			Delete(); err != nil {
			return fmt.Errorf("彻底删除用户失败: %w", err)
		}
		// 2. 删除用户部门关系。
		if _, err := dao.SysUserDepart.Ctx(ctx).
			WhereIn(udCols.UserId, ids).
			Delete(); err != nil {
			return fmt.Errorf("删除用户部门关系失败: %w", err)
		}
		// 3. 删除用户角色关系。
		if _, err := dao.SysUserRole.Ctx(ctx).
			WhereIn(urCols.UserId, ids).
			Delete(); err != nil {
			return fmt.Errorf("删除用户角色关系失败: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
