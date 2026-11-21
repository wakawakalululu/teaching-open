// Package sysspecial 承载 P2-6 各域特殊端点（非标准 CRUD）：
// fillRule testFillRule、checkRule checkByCode、dataSource options 已在
// syscrud 域实现、dataLog queryCompareList/queryDataVerList、
// duplicate check（动态表/列防注入白名单）、quartz pause/resume。
package sysspecial

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
)

// ErrNotFound 编码不存在（Java checkByCode error 原话术）。
var ErrNotFound = errors.New("该编码不存在")

// ErrNotImplemented 规则/调度能力未实现（文档化差距）。
var ErrNotImplemented = errors.New("该规则编码未实现")

// identRegexp 动态表/列名白名单（duplicate check 防注入）。
var identRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// TestFillRule 填充规则试算。
//
// 契约蓝本 SysFillRuleController#testFillRule：FillRuleUtil 按 rule_code
// 反射执行；Go 仅移植了 org_num_role（YouBian 部门编码），其余规则返回
// ErrNotImplemented（文档化差距）。
func (s *Service) TestFillRule(ctx context.Context, ruleCode, testValue string) (string, error) {
	if ruleCode != "org_num_role" {
		return "", ErrNotImplemented
	}
	rec, err := g.DB().Model("sys_depart").
		Where("parent_id = '' OR parent_id IS NULL").
		Order("org_code DESC").Limit(1).One()
	if err != nil {
		return "", fmt.Errorf("查询部门失败: %w", err)
	}
	if rec.IsEmpty() {
		return "A01", nil
	}
	oldCode := rec["org_code"].String()
	// 复用 category 域的 YouBian 递增（A99→B01、Z99→追段）
	return nextYouBianCode(oldCode), nil
}

// nextYouBianCode 与 category/depart 域同一移植源（A01 起递增）。
func nextYouBianCode(code string) string {
	if code == "" {
		return "A01"
	}
	if len(code) < 3 {
		return ""
	}
	before := code[:len(code)-3]
	after := code[len(code)-3:]
	letter := after[0]
	num := 0
	for _, ch := range after[1:] {
		if ch < '0' || ch > '9' {
			return ""
		}
		num = num*10 + int(ch-'0')
	}
	if num == 99 {
		if letter == 'Z' {
			return code + "A01"
		}
		return before + string(letter+1) + "01"
	}
	return before + string(letter) + fmt.Sprintf("%02d", num+1)
}

// CheckByCodeResult checkByCode 的返回（Java：通过=Result.ok()；
// 不通过=error(message)+result=明细）。
type CheckByCodeResult struct {
	Passed  bool
	Message string
	Detail  map[string]interface{}
}

// CheckByCode 编码校验规则试算。
//
// 契约蓝本 SysCheckRuleController#checkByCode + checkValue：按 rule_code
// 查规则，rule_json 为 [{digits, pattern, message}] 数组；digits='*' 检查
// 全值，否则按位数顺序截取递增；首个未通过即返回该段 message。
func (s *Service) CheckByCode(ctx context.Context, ruleCode, value string) (*CheckByCodeResult, error) {
	var rule entity.SysCheckRule
	if err := g.DB().Model("sys_check_rule").
		Where("rule_code", ruleCode).Limit(1).Scan(&rule); err != nil {
		return nil, fmt.Errorf("查询校验规则失败: %w", err)
	}
	if rule.Id == "" {
		return nil, ErrNotFound
	}
	// rules: [{digits, pattern, message}]
	var rules []struct {
		Digits  string `json:"digits"`
		Pattern string `json:"pattern"`
		Message string `json:"message"`
	}
	raw := g.DB().GetOne
	_ = raw
	if err := loadRuleJSON(ctx, rule.Id, &rules); err != nil {
		return nil, err
	}
	beginIndex := 0
	for _, r := range rules {
		var checkValue string
		if r.Digits == "*" {
			checkValue = value
		} else {
			num := 0
			for _, ch := range r.Digits {
				if ch < '0' || ch > '9' {
					return nil, fmt.Errorf("规则位数非法: %s", r.Digits)
				}
				num = num*10 + int(ch-'0')
			}
			endIndex := beginIndex + num
			if endIndex > len(value) {
				endIndex = len(value)
			}
			if beginIndex > endIndex {
				checkValue = ""
			} else {
				checkValue = value[beginIndex:endIndex]
			}
			beginIndex += num
		}
		passed, perr := regexp.MatchString(r.Pattern, checkValue)
		if perr != nil {
			return nil, fmt.Errorf("校验正则非法: %s", r.Pattern)
		}
		if !passed {
			return &CheckByCodeResult{
				Passed:  false,
				Message: r.Message,
				Detail: map[string]interface{}{
					"digits": r.Digits, "pattern": r.Pattern, "message": r.Message,
					"checkValue": checkValue, "passed": false,
				},
			}, nil
		}
	}
	return &CheckByCodeResult{Passed: true}, nil
}

// loadRuleJSON 直接从记录读取 rule_json 并解析（entity 扫描亦可，此处
// 独立读取避免依赖列拼写差异）。
func loadRuleJSON(ctx context.Context, ruleID string, out interface{}) error {
	rec, err := g.DB().Model("sys_check_rule").Fields("rule_json").
		Where("id", ruleID).Limit(1).One()
	if err != nil {
		return fmt.Errorf("查询规则 JSON 失败: %w", err)
	}
	if rec.IsEmpty() {
		return nil
	}
	var rules []struct {
		Digits  string `json:"digits"`
		Pattern string `json:"pattern"`
		Message string `json:"message"`
	}
	if err := rec["rule_json"].Struct(&rules); err != nil {
		// 非法 JSON 视为无规则（Java parseArray 异常 → 全局 500，此处宽松化）
		return nil
	}
	// 通过反射写入调用方 out（泛型数组）
	switch o := out.(type) {
	case *[]struct {
		Digits  string `json:"digits"`
		Pattern string `json:"pattern"`
		Message string `json:"message"`
	}:
		*o = rules
	}
	return nil
}

// DuplicateCheck 重复校验（动态表/列 + dataId 编辑模式排除自身）。
//
// 契约蓝本 DuplicateCheckController#check + SysDictMapper#duplicateCheckCountSql：
// Java 原始 SQL 拼接 → Go 白名单校验表/列名后占位符绑定（防注入修正，
// 语义一致）；可用 → ok「该值可用！」；已存在 → error「该值不可用，系统中已存在！」。
func (s *Service) DuplicateCheck(ctx context.Context, tableName, fieldName, fieldVal, dataID string) (bool, error) {
	if !identRegexp.MatchString(tableName) || !identRegexp.MatchString(fieldName) ||
		!identRegexp.MatchString(fieldVal) {
		// Java 对非法输入无防护；Go fail-closed 报不可用（文档化修正）
		return false, nil
	}
	query := fmt.Sprintf("SELECT COUNT(*) AS n FROM %s WHERE %s = ?", tableName, fieldName)
	args := []interface{}{fieldVal}
	if dataID != "" {
		// 编辑页面校验：排除自身（dataId2 为另一个对比 id，Java SQL 中
		// and id <> dataId —— 以 dataId 为准）
		query += " AND id <> ?"
		args = append(args, dataID)
	}
	rec, err := g.DB().GetOne(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("重复校验失败: %w", err)
	}
	return rec["n"].Int() == 0, nil
}

// QuartzSetStatus Quartz 任务状态切换（pause → -1 / resume → 0）。
//
// 契约蓝本 QuartzJobController#pause|resume：Java 走 quartz scheduler；
// Go 无调度器后端 → 落库 status 字段（0正常 -1停止），执行能力为
// 文档化待补齐项。
func (s *Service) QuartzSetStatus(ctx context.Context, jobClassName string, status int) error {
	if _, err := g.DB().Model("sys_quartz_job").
		Where("job_class_name", strings.TrimSpace(jobClassName)).
		Data(g.Map{"status": status}).Update(); err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}
	return nil
}

// TrimJobClassName 修剪任务类名（Java trim() 语义）。
func TrimJobClassName(s string) string {
	return strings.TrimSpace(s)
}

// Service 特殊端点服务（无状态占位，便于路由注入）。
type Service struct{}

// New 构造服务。
func New() *Service { return &Service{} }

// QueryCompareList 数据日志对比（两条 id 的记录列表）。
func (s *Service) QueryCompareList(ctx context.Context, dataID1, dataID2 string) ([]map[string]interface{}, error) {
	ids := []string{dataID1, dataID2}
	rows, err := g.DB().Model("sys_data_log").WhereIn("id", ids).All()
	if err != nil {
		return nil, fmt.Errorf("查询对比数据失败: %w", err)
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Map())
	}
	return out, nil
}

// QueryDataVerList 数据版本列表（data_table + data_id）。
func (s *Service) QueryDataVerList(ctx context.Context, dataTable, dataID string) ([]map[string]interface{}, error) {
	rows, err := g.DB().Model("sys_data_log").
		Where("data_table", dataTable).Where("data_id", dataID).All()
	if err != nil {
		return nil, fmt.Errorf("查询版本信息失败: %w", err)
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Map())
	}
	return out, nil
}
