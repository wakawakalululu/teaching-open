// Package depart 承载部门管理域业务用例
// （契约蓝本：SysDepartController 的 10 个核心端点；exportXls/importExcel
// 有意未迁移，与既有 Excel 决策一致；queryParentName 为前端死引用未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - sys_depart 无 @TableLogic，删除为物理删除 + 五表级联
//     （sys_user_depart/sys_depart_permission/sys_depart_role/
//     sys_depart_role_permission/sys_depart_role_user）；
//   - org_code 由 YouBian 流水规则服务端生成（3 字符/级），org_type = 父级 +1；
//   - edit 拒绝调整层级（parentId 变化 → 「暂不支持调整层级结构」），不重算编码；
//   - 树构建 isLeaf 由子节点存在性计算（sys_depart 无 is_leaf 列）；
//   - Java 的 @Cacheable 部门缓存依赖 Redis 显式失效，Go 版直查库（与 P0-1
//     字典域同一决策），无需失效面。
package depart

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// 业务错误（HTTP 层映射话术与 Java 保持一致）。
var (
	// ErrNotFound 实体不存在（Java edit/delete error500 原话术）。
	ErrNotFound = errors.New("未找到对应实体")
	// ErrParentChange 层级调整被拒绝（Java edit 原话术）。
	ErrParentChange = errors.New("暂不支持调整层级结构")
	// ErrParamInvalid 批量删除参数不识别（Java deleteBatch error500 原话术）。
	ErrParamInvalid = errors.New("参数不识别！")
)

// DepartInput add/edit 请求体（SysDepart 可编辑字段）。
//
// 指针字段对齐 MyBatis-Plus NOT_NULL 字段策略：JSON 未提供的字段不参与
// INSERT/UPDATE。orgCode/orgType 由 add 在服务端生成覆盖；edit 按 MP 语义
// 透传（body 携带则更新，与 Java updateById 行为一致）。
type DepartInput struct {
	Id             string
	ParentId       *string
	DepartName     *string
	DepartNameEn   *string
	DepartNameAbbr *string
	DepartOrder    *int
	Description    *string
	OrgCategory    *string
	OrgType        *string
	OrgCode        *string
	Mobile         *string
	Fax            *string
	Address        *string
	Memo           *string
	Status         *string
}

// departTreeModel 对齐 Java SysDepartTreeModel 的字段序列化
// （key/value=id、title=departName；isLeaf 由子节点存在性计算；
// 树路径叶子 children=null，扁平查询路径 children=[]）。
type departTreeModel struct {
	Key            string            `json:"key"`
	Value          string            `json:"value"`
	Title          string            `json:"title"`
	IsLeaf         bool              `json:"isLeaf"`
	Id             string            `json:"id"`
	ParentId       *string           `json:"parentId"`
	DepartName     string            `json:"departName"`
	DepartNameEn   string            `json:"departNameEn"`
	DepartNameAbbr string            `json:"departNameAbbr"`
	DepartOrder    int               `json:"departOrder"`
	Description    string            `json:"description"`
	OrgCategory    string            `json:"orgCategory"`
	OrgType        string            `json:"orgType"`
	OrgCode        string            `json:"orgCode"`
	Mobile         string            `json:"mobile"`
	Fax            string            `json:"fax"`
	Address        string            `json:"address"`
	Memo           string            `json:"memo"`
	Status         string            `json:"status"`
	DelFlag        string            `json:"delFlag"`
	CreateBy       string            `json:"createBy"`
	CreateTime     string            `json:"createTime"`
	UpdateBy       string            `json:"updateBy"`
	UpdateTime     string            `json:"updateTime"`
	Children       []departTreeModel `json:"children"`
}

// departIDModel 对齐 Java DepartIdModel（queryIdTree：仅 key/value/title/children）。
type departIDModel struct {
	Key      string          `json:"key"`
	Value    string          `json:"value"`
	Title    string          `json:"title"`
	Children []departIDModel `json:"children"`
}

// Service 部门管理业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// QueryTreeList 全量部门树。
//
// 契约蓝本 SysDepartController#queryTreeList（GET /sys/sysDepart/queryTreeList）：
// del_flag=0 + depart_order ASC → 内存拼树（parentId 空为根；叶子 children=null、
// isLeaf=true；孤儿节点丢弃，对齐 FindsDepartsChildrenUtil）。
func (s *Service) QueryTreeList(ctx context.Context) ([]departTreeModel, error) {
	rows, err := s.listDelFlag0(ctx)
	if err != nil {
		return nil, err
	}
	return buildDepartTree(rows), nil
}

// QueryMyDeptTreeList 我的部门树（上级用户首页）。
//
// 契约蓝本 SysDepartController#queryMyDeptTreeList（GET）：
//   - userIdentity=2（上级）：admin/dev 角色 → 负责部门重置为全部顶级部门 id
//     （getRootDepart 仅 parent_id=”，不过滤 del_flag）；否则用 sys_user.depart_ids；
//   - org_code 前缀 likeRight OR 链 + del_flag=0 + depart_order ASC；
//     命中前缀的部门 parentId 置空强制成根后拼树；
//   - 返回 (树, 身份"2")；userIdentity=1 → (nil, "1")，result 由控制器置 null；
//   - 偏差文档化：depart_ids 为空时 Java 走 NPE→catch（success=true result=null），
//     Go 返回空树 + 身份"2"（前端拿不到树，行为等价且更稳）。
func (s *Service) QueryMyDeptTreeList(ctx context.Context, operatorID string) ([]departTreeModel, string, error) {
	user, err := s.getUser(ctx, operatorID)
	if err != nil {
		return nil, "", err
	}
	if user.UserIdentity != 2 {
		return nil, "1", nil
	}
	privileged, err := s.hasAnyRoleCode(ctx, operatorID, "admin", "dev")
	if err != nil {
		return nil, "", err
	}
	departIDs := user.DepartIds
	if privileged {
		rootIDs, err := s.getRootDepartIDs(ctx)
		if err != nil {
			return nil, "", err
		}
		departIDs = strings.Join(rootIDs, ",")
	}
	tree, err := s.myDeptTree(ctx, departIDs)
	if err != nil {
		return nil, "", err
	}
	return tree, "2", nil
}

// myDeptTree 按负责部门 id 集合构建前缀过滤树（对齐 SysDepartServiceImpl#queryMyDeptTreeList）。
func (s *Service) myDeptTree(ctx context.Context, departIDs string) ([]departTreeModel, error) {
	if departIDs == "" {
		// Java：getMyDeptParentOrgCode → IN ('') 空集 → NPE 被 catch → result=null。
		// Go 返回空树（偏差文档化）。
		return []departTreeModel{}, nil
	}
	codes, err := s.getMyDeptParentOrgCodes(ctx, departIDs)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return []departTreeModel{}, nil
	}
	cols := dao.SysDepart.Columns()
	m := dao.SysDepart.Ctx(ctx)
	m = whereOrgCodePrefixes(m, cols.OrgCode, codes)
	m = m.Where(cols.DelFlag, "0").OrderAsc(cols.DepartOrder)
	var rows []entity.SysDepart
	if err := m.Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询我的部门失败: %w", err)
	}
	// 命中前缀的部门 parentId 置空强制成根（对齐 Java 循环 setParentId(null)）
	for i := range rows {
		for _, code := range codes {
			if rows[i].OrgCode == code {
				rows[i].ParentId = ""
				break
			}
		}
	}
	return buildDepartTree(rows), nil
}

// getMyDeptParentOrgCodes 负责部门编码前缀集合（对齐 Java getMyDeptParentOrgCode
// + getMyDeptParentNode，含其「直接拼接全部编码不做最小节点收敛」的现状——
// likeRight OR 链语义下冗余前缀不影响结果集，逐语义保留）。
func (s *Service) getMyDeptParentOrgCodes(ctx context.Context, departIDs string) ([]string, error) {
	ids := strings.Split(departIDs, ",")
	cols := dao.SysDepart.Columns()
	var rows []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Where(cols.DelFlag, "0").
		WhereIn(cols.Id, ids).
		OrderAsc(cols.OrgCode).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询负责部门失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	// 按 org_code 前 3 字符（公司级）分组 → 组内逗号拼接 → 全组拼接 → 拆分
	groups := make(map[string][]string)
	order := make([]string, 0, len(rows))
	for _, r := range rows {
		code := r.OrgCode
		if len(code) > 3 {
			code = code[:3]
		}
		if _, ok := groups[code]; !ok {
			order = append(order, code)
		}
		groups[code] = append(groups[code], r.OrgCode)
	}
	parts := make([]string, 0, len(order))
	for _, prefix := range order {
		parts = append(parts, strings.Join(groups[prefix], ","))
	}
	return strings.Split(strings.Join(parts, ","), ","), nil
}

// getRootDepartIDs 全部顶级部门 id（对齐 SysDepartMapper#getRootDepart 注解 SQL：
// parent_id=”，不过滤 del_flag）。
func (s *Service) getRootDepartIDs(ctx context.Context) ([]string, error) {
	cols := dao.SysDepart.Columns()
	var rows []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Fields(cols.Id).
		Where(cols.ParentId, "").
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询顶级部门失败: %w", err)
	}
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.Id)
	}
	return ids, nil
}

// QueryIdTree 部门 id/名称树（add/edit 弹窗选择父部门）。
//
// 契约蓝本 SysDepartController#queryIdTree（GET /sys/sysDepart/queryIdTree）：
// del_flag=0 + depart_order ASC → DepartIdModel 树（key/value=id、title=名称；
// 叶子 children=[]，无 setEmptyChildrenAsNull 等价处理，对齐 Java）。
func (s *Service) QueryIdTree(ctx context.Context) ([]departIDModel, error) {
	rows, err := s.listDelFlag0(ctx)
	if err != nil {
		return nil, err
	}
	nodes := make([]departIDModel, 0, len(rows))
	for _, r := range rows {
		nodes = append(nodes, departIDModel{
			Key:      r.Id,
			Value:    r.Id,
			Title:    r.DepartName,
			Children: []departIDModel{},
		})
	}
	var out []departIDModel
	attachIDChildren(&out, rows, nodes, nil)
	return out, nil
}

// attachIDChildren DepartIdModel 树递归挂载（parent 为 nil 时收集根节点，
// 语义同 Java getGrandChildren 的逐层 attach；孤儿节点丢弃）。
func attachIDChildren(dest *[]departIDModel, rows []entity.SysDepart, nodes []departIDModel, parent *departIDModel) {
	for i := range nodes {
		switch {
		case parent == nil && rows[i].ParentId == "":
			*dest = append(*dest, nodes[i])
			last := &(*dest)[len(*dest)-1]
			attachIDChildren(&last.Children, rows, nodes, last)
		case parent != nil && rows[i].ParentId == parent.Key:
			parent.Children = append(parent.Children, nodes[i])
			last := &parent.Children[len(parent.Children)-1]
			attachIDChildren(&last.Children, rows, nodes, last)
		}
	}
}

// whereOrgCodePrefixes 组装括号分组的 org_code 前缀 OR 链
// （对齐 Java query.or().likeRight(...) 连写后再 AND del_flag 的分组语义；
// 列名为代码内常量，值走占位符绑定）。
func whereOrgCodePrefixes(m *gdb.Model, column string, codes []string) *gdb.Model {
	conds := make([]string, 0, len(codes))
	args := make([]interface{}, 0, len(codes))
	for _, code := range codes {
		conds = append(conds, column+" LIKE ?")
		args = append(args, code+"%")
	}
	return m.Where("("+strings.Join(conds, " OR ")+")", args...)
}

// Add 新增部门。
//
// 契约蓝本 SysDepartController#add + SysDepartServiceImpl#saveDepartData：
//   - create_by 由控制器写入（与角色域不同，这里 Java 显式 setCreateBy）；
//   - parentId null → ""；id 为 UUID 32 hex；org_code/org_type 由
//     OrgCodeRule 生成（根级取同级最大 code 递增，子级 = 父码 + 子段）；
//     org_type：空表根级 "1"，否则继承同级根；子级 = 父 orgType + 1；
//   - create_time 服务端写入；del_flag="0" 固定。
func (s *Service) Add(ctx context.Context, operatorID string, in DepartInput) error {
	data := do.SysDepart{
		Id:         guid.S(),
		CreateBy:   operatorID,
		CreateTime: gtime.Now(),
		DelFlag:    "0",
	}
	if in.ParentId == nil {
		data.ParentId = ""
	} else {
		data.ParentId = *in.ParentId
	}
	parentID := ""
	if in.ParentId != nil {
		parentID = *in.ParentId
	}
	code, orgType, err := s.genOrgCodeAndType(ctx, parentID)
	if err != nil {
		return err
	}
	data.OrgCode = code
	data.OrgType = orgType
	applyDepartInput(&data, in)
	if _, err := dao.SysDepart.Ctx(ctx).Data(data).Insert(); err != nil {
		// 唯一键 uniq_depart_org_code 冲突等数据库错误 → Java catch 原话术
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// genOrgCodeAndType 生成部门编码与类型（对齐 OrgCodeRule#execute）。
func (s *Service) genOrgCodeAndType(ctx context.Context, parentID string) (string, string, error) {
	cols := dao.SysDepart.Columns()
	if parentID == "" {
		// 根级：parent_id='' OR parent_id IS NULL，按 org_code 倒序取首条
		var rows []entity.SysDepart
		if err := dao.SysDepart.Ctx(ctx).
			Where("parent_id = '' OR parent_id IS NULL").
			OrderDesc(cols.OrgCode).
			Scan(&rows); err != nil {
			return "", "", fmt.Errorf("查询同级部门失败: %w", err)
		}
		if len(rows) == 0 {
			return nextYouBianCode(""), "1", nil
		}
		return nextYouBianCode(rows[0].OrgCode), rows[0].OrgType, nil
	}
	// 子级：父部门 code + 子段；org_type = 父 orgType + 1
	var parent entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).Where(cols.Id, parentID).Scan(&parent); err != nil {
		return "", "", fmt.Errorf("查询父部门失败: %w", err)
	}
	if parent.Id == "" {
		// Java：getById(parentId).getOrgCode() → NPE → catch「操作失败」
		return "", "", fmt.Errorf("操作失败")
	}
	var siblings []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Where(cols.ParentId, parentID).
		OrderDesc(cols.OrgCode).
		Scan(&siblings); err != nil {
		return "", "", fmt.Errorf("查询同级部门失败: %w", err)
	}
	var code string
	if len(siblings) == 0 {
		code = subYouBianCode(parent.OrgCode, "")
	} else {
		code = subYouBianCode(parent.OrgCode, siblings[0].OrgCode)
	}
	orgType, err := strconv.Atoi(strings.TrimSpace(parent.OrgType))
	if err != nil {
		// Java Integer.valueOf(depart.getOrgType())+1 的异常路径
		return "", "", fmt.Errorf("操作失败")
	}
	return code, strconv.Itoa(orgType + 1), nil
}

// Edit 编辑部门。
//
// 契约蓝本 SysDepartController#edit + SysDepartServiceImpl#updateDepartDataById：
//   - 先查存在性（不存在 → 「未找到对应实体」）；
//   - parentId 非空且与旧值不同 → 「暂不支持调整层级结构」（orgCode 同步逻辑
//     已被注释，本版本禁用换父）；
//   - update_by（控制器写入 username）+ update_time 服务端维护；其余字段按
//     MP 语义透传（body 未提供的字段不更新）。
func (s *Service) Edit(ctx context.Context, operatorID string, in DepartInput) error {
	cols := dao.SysDepart.Columns()
	var old entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).Where(cols.Id, in.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询部门失败: %w", err)
	}
	if old.Id == "" {
		return ErrNotFound
	}
	if in.ParentId != nil && *in.ParentId != "" && *in.ParentId != old.ParentId {
		return ErrParentChange
	}
	data := do.SysDepart{
		UpdateBy:   operatorID,
		UpdateTime: gtime.Now(),
	}
	if in.ParentId != nil {
		data.ParentId = *in.ParentId
	}
	applyDepartInput(&data, in)
	if _, err := dao.SysDepart.Ctx(ctx).Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除部门（物理级联）。
//
// 契约蓝本 SysDepartController#delete（DELETE ?id=）+ SysDepartServiceImpl#delete
// （@Transactional）：存在性校验在控制器（null → 「未找到对应实体」）；级联
// 顺序对齐 Java：递归收集子部门 → 物理删部门 → 查部门角色 → 依次清
// sys_user_depart/sys_depart_permission/sys_depart_role →（有角色时）
// sys_depart_role_permission/sys_depart_role_user。
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.cascadeDelete(ctx, []string{id})
}

// DeleteBatch 批量删除部门（物理级联）。
//
// 契约蓝本 SysDepartController#deleteBatch：空 ids → 「参数不识别！」；
// 其余同单删级联（无存在性校验，不存在的 id 级联空跑）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return ErrParamInvalid
	}
	return s.cascadeDelete(ctx, ids)
}

// cascadeDelete 五表级联物理删除（单删/批删共用，单事务，顺序对齐 Java）。
func (s *Service) cascadeDelete(ctx context.Context, ids []string) error {
	return dao.SysDepart.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 递归收集子部门 id（对齐 checkChildrenExists）
		idList := make([]string, 0, len(ids))
		idList = append(idList, ids...)
		for _, id := range ids {
			if err := s.collectChildIDs(ctx, id, &idList); err != nil {
				return err
			}
		}
		cols := dao.SysDepart.Columns()
		if _, err := dao.SysDepart.Ctx(ctx).WhereIn(cols.Id, idList).Delete(); err != nil {
			return fmt.Errorf("删除部门失败: %w", err)
		}
		// 部门角色 id（供后两张表清理）
		roleCols := dao.SysDepartRole.Columns()
		var departRoles []entity.SysDepartRole
		if err := dao.SysDepartRole.Ctx(ctx).
			Fields(roleCols.Id).
			WhereIn(roleCols.DepartId, idList).
			Scan(&departRoles); err != nil {
			return fmt.Errorf("查询部门角色失败: %w", err)
		}
		// 用户与部门关系
		udCols := dao.SysUserDepart.Columns()
		if _, err := dao.SysUserDepart.Ctx(ctx).
			WhereIn(udCols.DepId, idList).Delete(); err != nil {
			return fmt.Errorf("级联删除用户部门关系失败: %w", err)
		}
		// 部门授权
		dpCols := dao.SysDepartPermission.Columns()
		if _, err := dao.SysDepartPermission.Ctx(ctx).
			WhereIn(dpCols.DepartId, idList).Delete(); err != nil {
			return fmt.Errorf("级联删除部门授权失败: %w", err)
		}
		// 部门角色
		if _, err := dao.SysDepartRole.Ctx(ctx).
			WhereIn(roleCols.DepartId, idList).Delete(); err != nil {
			return fmt.Errorf("级联删除部门角色失败: %w", err)
		}
		if len(departRoles) > 0 {
			roleIDs := make([]string, 0, len(departRoles))
			for _, r := range departRoles {
				roleIDs = append(roleIDs, r.Id)
			}
			drpCols := dao.SysDepartRolePermission.Columns()
			if _, err := dao.SysDepartRolePermission.Ctx(ctx).
				WhereIn(drpCols.RoleId, roleIDs).Delete(); err != nil {
				return fmt.Errorf("级联删除部门角色授权失败: %w", err)
			}
			druCols := dao.SysDepartRoleUser.Columns()
			if _, err := dao.SysDepartRoleUser.Ctx(ctx).
				WhereIn(druCols.DroleId, roleIDs).Delete(); err != nil {
				return fmt.Errorf("级联删除部门角色用户失败: %w", err)
			}
		}
		return nil
	})
}

// collectChildIDs 递归收集子部门 id（对齐 Java checkChildrenExists）。
func (s *Service) collectChildIDs(ctx context.Context, parentID string, out *[]string) error {
	cols := dao.SysDepart.Columns()
	var children []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Fields(cols.Id).
		Where(cols.ParentId, parentID).
		Scan(&children); err != nil {
		return fmt.Errorf("查询子部门失败: %w", err)
	}
	for _, child := range children {
		*out = append(*out, child.Id)
		if err := s.collectChildIDs(ctx, child.Id, out); err != nil {
			return err
		}
	}
	return nil
}

// SearchBy 部门关键词搜索。
//
// 契约蓝本 SysDepartController#searchBy（GET ?keyWord=&myDeptSearch=）+
// SysDepartServiceImpl#searhBy：
//   - myDeptSearch 非空（我的部门搜索）：depart_ids 为空 → 空结果；否则
//     org_code 前缀 OR 链 + del_flag=0；
//   - 普通搜索：不加 del_flag 过滤（对齐 Java）；
//   - 始终 depart_name LIKE %kw%；结果为扁平模型（children=[]，字段默认值）。
func (s *Service) SearchBy(ctx context.Context, keyWord, myDeptSearch, departIDs string) ([]departTreeModel, error) {
	cols := dao.SysDepart.Columns()
	m := dao.SysDepart.Ctx(ctx)
	if myDeptSearch != "" {
		if departIDs == "" {
			return []departTreeModel{}, nil
		}
		codes, err := s.getMyDeptParentOrgCodes(ctx, departIDs)
		if err != nil {
			return nil, err
		}
		if len(codes) == 0 {
			return []departTreeModel{}, nil
		}
		m = whereOrgCodePrefixes(m, cols.OrgCode, codes)
		m = m.Where(cols.DelFlag, "0")
	}
	m = m.WhereLike(cols.DepartName, "%"+keyWord+"%")
	var rows []entity.SysDepart
	if err := m.OrderAsc(cols.OrgCode).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	out := make([]departTreeModel, 0, len(rows))
	for _, r := range rows {
		out = append(out, newDepartTreeModel(r))
	}
	return out, nil
}

// ListAll 全量部门扁平列表。
//
// 契约蓝本 SysDepartController#listAll（GET ?id=可选，逗号分隔）：
// org_code ASC；可选 id IN；不过滤 del_flag（对齐 Java）。
func (s *Service) ListAll(ctx context.Context, id string) ([]entity.SysDepart, error) {
	cols := dao.SysDepart.Columns()
	m := dao.SysDepart.Ctx(ctx).OrderAsc(cols.OrgCode)
	if id != "" {
		m = m.WhereIn(cols.Id, strings.Split(id, ","))
	}
	var rows []entity.SysDepart
	if err := m.Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	return rows, nil
}

// RemoveAll 清空班级成员（GET /sys/sysDepart/removeAll?id=）。
//
// 契约蓝本 SysDepartController#removeAll：仅删 sys_user_depart 的 dep_id 关联，
// 不清理 sys_depart_role_user、不动部门本身（Java 现状，逐语义保留）；
// 成功 Result.ok()（code=200、message=「成功」）。
func (s *Service) RemoveAll(ctx context.Context, depID string) error {
	udCols := dao.SysUserDepart.Columns()
	if _, err := dao.SysUserDepart.Ctx(ctx).
		Where(udCols.DepId, depID).Delete(); err != nil {
		return fmt.Errorf("清空班级失败: %w", err)
	}
	return nil
}

// listDelFlag0 del_flag=0 + depart_order ASC 全量部门查询。
func (s *Service) listDelFlag0(ctx context.Context) ([]entity.SysDepart, error) {
	cols := dao.SysDepart.Columns()
	var rows []entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Where(cols.DelFlag, "0").
		OrderAsc(cols.DepartOrder).
		Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	return rows, nil
}

// CheckExists 部门存在性校验（对齐 Java delete/edit 的 getById 分支）。
func (s *Service) CheckExists(ctx context.Context, id string) error {
	if id == "" {
		return ErrNotFound
	}
	var row entity.SysDepart
	if err := dao.SysDepart.Ctx(ctx).
		Where(dao.SysDepart.Columns().Id, id).
		Scan(&row); err != nil {
		return fmt.Errorf("查询部门失败: %w", err)
	}
	if row.Id == "" {
		return ErrNotFound
	}
	return nil
}

// MyDepartIDs 当前用户的负责部门 id 串（userIdentity=2 时为 depart_ids，
// 否则空串，对齐 Java searchBy 的 departIds 取值条件）。
func (s *Service) MyDepartIDs(ctx context.Context, operatorID string) string {
	user, err := s.getUser(ctx, operatorID)
	if err != nil || user.UserIdentity != 2 {
		return ""
	}
	return user.DepartIds
}

// getUser 加载当前登录用户（queryMyDeptTreeList 需要 userIdentity/depart_ids）。
func (s *Service) getUser(ctx context.Context, operatorID string) (*entity.SysUser, error) {
	if operatorID == "" {
		return &entity.SysUser{}, nil
	}
	var u entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Id, operatorID).
		Scan(&u); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &u, nil
}

// hasAnyRoleCode 判断用户是否拥有任一指定角色编码（与 role 域同款实现）。
func (s *Service) hasAnyRoleCode(ctx context.Context, userID string, roleCodes ...string) (bool, error) {
	if userID == "" || len(roleCodes) == 0 {
		return false, nil
	}
	rec, err := g.DB().Model("sys_user_role ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Where("ur.user_id", userID).
		WhereIn("r.role_code", roleCodes).
		Limit(1).
		One()
	if err != nil {
		return false, fmt.Errorf("查询用户角色失败: %w", err)
	}
	return !rec.IsEmpty(), nil
}

// buildDepartTree 由扁平部门列表构建树（对齐 FindsDepartsChildrenUtil：
// 根 = parentId 空；叶子 children=null + isLeaf=true；非叶 isLeaf=false）。
func buildDepartTree(rows []entity.SysDepart) []departTreeModel {
	var out []departTreeModel
	appendDepartTree(&out, rows, nil)
	return out
}

// appendDepartTree 树递归挂载（parent 为 nil 时收集根节点，孤儿丢弃）。
func appendDepartTree(dest *[]departTreeModel, rows []entity.SysDepart, parent *departTreeModel) {
	for i := range rows {
		node := newDepartTreeModel(rows[i])
		switch {
		case parent == nil && rows[i].ParentId == "":
			*dest = append(*dest, node)
			last := &(*dest)[len(*dest)-1]
			appendDepartTree(&last.Children, rows, last)
			refreshLeafFlag(last)
		case parent != nil && rows[i].ParentId == parent.Id:
			parent.Children = append(parent.Children, node)
			last := &parent.Children[len(parent.Children)-1]
			appendDepartTree(&last.Children, rows, last)
			refreshLeafFlag(last)
		}
	}
}

// refreshLeafFlag 由 children 存在性回填 isLeaf（对齐 setEmptyChildrenAsNull：
// children 空 → null + isLeaf=true；非空 → isLeaf=false）。
func refreshLeafFlag(node *departTreeModel) {
	if len(node.Children) == 0 {
		node.Children = nil
		node.IsLeaf = true
	} else {
		node.IsLeaf = false
	}
}

// newDepartTreeModel 由 SysDepart 构造树节点（对齐 SysDepartTreeModel(SysDepart)
// 构造器：children 默认空数组，isLeaf 默认 false，由 refreshLeafFlag 修正）。
func newDepartTreeModel(r entity.SysDepart) departTreeModel {
	pid := r.ParentId
	return departTreeModel{
		Key:            r.Id,
		Value:          r.Id,
		Title:          r.DepartName,
		Id:             r.Id,
		ParentId:       &pid,
		DepartName:     r.DepartName,
		DepartNameEn:   r.DepartNameEn,
		DepartNameAbbr: r.DepartNameAbbr,
		DepartOrder:    r.DepartOrder,
		Description:    r.Description,
		OrgCategory:    r.OrgCategory,
		OrgType:        r.OrgType,
		OrgCode:        r.OrgCode,
		Mobile:         r.Mobile,
		Fax:            r.Fax,
		Address:        r.Address,
		Memo:           r.Memo,
		Status:         r.Status,
		DelFlag:        r.DelFlag,
		CreateBy:       r.CreateBy,
		CreateTime:     r.CreateTime,
		UpdateBy:       r.UpdateBy,
		UpdateTime:     r.UpdateTime,
		Children:       []departTreeModel{},
	}
}

// applyDepartInput 把请求体可编辑字段拷贝进 DO（ParentId 由 add/edit 单独
// 管理；Id/审计列亦由调用方管理）。
func applyDepartInput(data *do.SysDepart, in DepartInput) {
	if in.DepartName != nil {
		data.DepartName = *in.DepartName
	}
	if in.DepartNameEn != nil {
		data.DepartNameEn = *in.DepartNameEn
	}
	if in.DepartNameAbbr != nil {
		data.DepartNameAbbr = *in.DepartNameAbbr
	}
	if in.DepartOrder != nil {
		data.DepartOrder = *in.DepartOrder
	}
	if in.Description != nil {
		data.Description = *in.Description
	}
	if in.OrgCategory != nil {
		data.OrgCategory = *in.OrgCategory
	}
	if in.OrgType != nil {
		data.OrgType = *in.OrgType
	}
	if in.OrgCode != nil {
		data.OrgCode = *in.OrgCode
	}
	if in.Mobile != nil {
		data.Mobile = *in.Mobile
	}
	if in.Fax != nil {
		data.Fax = *in.Fax
	}
	if in.Address != nil {
		data.Address = *in.Address
	}
	if in.Memo != nil {
		data.Memo = *in.Memo
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
}
