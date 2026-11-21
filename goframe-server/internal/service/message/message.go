// Package message 承载消息模板 + 消息发送记录域业务用例
// （契约蓝本：SysMessageTemplateController 8 个核心端点 + sendMsg +
// SysMessageController 6 个核心端点；exportXls/importExcel 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - sendMsg：按 templateCode 查模板，解析 test_data JSON 为参数 map，
//     渲染模板内容（${key} 替换）后落 sys_sms 发送记录；Go 简化落库即
//     「发送成功」（实际短信/邮件通道与 sms 域同一未接入状态，文档化）；
//   - 纯 MP CRUD。
package message

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// ErrNotFound 记录不存在（Java queryById error 原话术）。
var ErrNotFound = errors.New("未找到对应数据")

// TemplateInput 模板 add/edit 请求体。
type TemplateInput struct {
	Id               string
	TemplateName     *string
	TemplateCode     *string
	TemplateType     *string
	TemplateContent  *string
	TemplateTestJson *string
}

// SmsInput 消息记录 add/edit 请求体。
type SmsInput struct {
	Id           string
	EsTitle      *string
	EsType       *string
	EsReceiver   *string
	EsParam      *string
	EsContent    *string
	EsSendStatus *string
	EsSendNum    *int
	EsResult     *string
	Remark       *string
}

// Service 消息域业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// TemplateList 模板分页（QueryGenerator 消费面：templateCode/templateName 规则）。
func (s *Service) TemplateList(ctx context.Context, param page.Param, templateCode, templateName string) (*page.Result, error) {
	cols := dao.SysSmsTemplate.Columns()
	m := dao.SysSmsTemplate.Ctx(ctx)
	if templateCode != "" {
		m = applyQueryRule(m, cols.TemplateCode, templateCode)
	}
	if templateName != "" {
		m = applyQueryRule(m, cols.TemplateName, templateName)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计模板数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysSmsTemplate{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysSmsTemplate
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询模板列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// TemplateAdd 新增模板（id 为空时服务端生成）。
func (s *Service) TemplateAdd(ctx context.Context, in TemplateInput) error {
	data := do.SysSmsTemplate{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyTemplateInput(&data, in)
	if _, err := dao.SysSmsTemplate.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// TemplateEdit 编辑模板。
func (s *Service) TemplateEdit(ctx context.Context, in TemplateInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.SysSmsTemplate{}
	applyTemplateInput(&data, in)
	cols := dao.SysSmsTemplate.Columns()
	if _, err := dao.SysSmsTemplate.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// TemplateDelete 删除模板。
func (s *Service) TemplateDelete(ctx context.Context, id string) error {
	cols := dao.SysSmsTemplate.Columns()
	if _, err := dao.SysSmsTemplate.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}
	return nil
}

// TemplateDeleteBatch 批量删除模板。
func (s *Service) TemplateDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.SysSmsTemplate.Columns()
	if _, err := dao.SysSmsTemplate.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除模板失败: %w", err)
	}
	return nil
}

// TemplateQueryByID 模板详情。
func (s *Service) TemplateQueryByID(ctx context.Context, id string) (*entity.SysSmsTemplate, error) {
	var row entity.SysSmsTemplate
	if err := dao.SysSmsTemplate.Ctx(ctx).
		Where(dao.SysSmsTemplate.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// SendMsg 按模板编码发送消息。
//
// 契约蓝本 SysMessageTemplateController#sendMsg（body MsgParams
// {msgType, templateCode, testData, receiver}）+ pushMsgUtil.sendMessage：
//   - testData 为 JSON 对象串，逐键 ${key} 替换渲染模板内容；
//   - 渲染结果落 sys_sms 发送记录（es_type=msgType、es_receiver=receiver、
//     es_send_status=1 已发送、es_send_num=1）；实际短信/邮件通道与 sms 域
//     同一未接入状态（文档化）；
//   - 模板不存在 →「发送消息任务添加失败！」。
func (s *Service) SendMsg(ctx context.Context, msgType, templateCode, testData, receiver string) error {
	var tpl entity.SysSmsTemplate
	if err := dao.SysSmsTemplate.Ctx(ctx).
		Where(dao.SysSmsTemplate.Columns().TemplateCode, templateCode).
		Limit(1).Scan(&tpl); err != nil {
		return fmt.Errorf("查询模板失败: %w", err)
	}
	if tpl.Id == "" {
		return errors.New("发送消息任务添加失败！")
	}
	content := renderTemplate(tpl.TemplateContent, testData)
	if _, err := g.DB().Model("sys_sms").Data(map[string]interface{}{
		"id": guid.S(), "es_title": tpl.TemplateName, "es_type": msgType,
		"es_receiver": receiver, "es_param": testData, "es_content": content,
		"es_send_time":   time.Now().Format("2006-01-02 15:04:05"),
		"es_send_status": "1", "es_send_num": 1, "es_result": "发送成功",
		"create_by": "system",
	}).Insert(); err != nil {
		return fmt.Errorf("发送消息任务添加失败！: %w", err)
	}
	return nil
}

// renderTemplate ${key} 参数替换（对齐 pushMsgUtil 的占位符渲染）。
func renderTemplate(content, testDataJSON string) string {
	if testDataJSON == "" || content == "" {
		return content
	}
	var params map[string]string
	if err := gjsonUnmarshalHelper(testDataJSON, &params); err != nil {
		return content
	}
	for k, v := range params {
		content = strings.ReplaceAll(content, "${"+k+"}", v)
	}
	return content
}

// SmsList 消息记录分页。
func (s *Service) SmsList(ctx context.Context, param page.Param, esReceiver, esSendStatus string) (*page.Result, error) {
	cols := dao.SysSms.Columns()
	m := dao.SysSms.Ctx(ctx)
	if esReceiver != "" {
		m = m.WhereLike(cols.EsReceiver, "%"+esReceiver+"%")
	}
	if esSendStatus != "" {
		m = m.Where(cols.EsSendStatus, esSendStatus)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计消息数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysSms{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysSms
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// SmsAdd 新增消息记录。
func (s *Service) SmsAdd(ctx context.Context, in SmsInput) error {
	data := do.SysSms{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applySmsInput(&data, in)
	if _, err := dao.SysSms.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// SmsEdit 编辑消息记录。
func (s *Service) SmsEdit(ctx context.Context, in SmsInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.SysSms{}
	applySmsInput(&data, in)
	cols := dao.SysSms.Columns()
	if _, err := dao.SysSms.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// SmsDelete 删除消息记录。
func (s *Service) SmsDelete(ctx context.Context, id string) error {
	cols := dao.SysSms.Columns()
	if _, err := dao.SysSms.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除消息失败: %w", err)
	}
	return nil
}

// SmsDeleteBatch 批量删除消息记录。
func (s *Service) SmsDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.SysSms.Columns()
	if _, err := dao.SysSms.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除消息失败: %w", err)
	}
	return nil
}

// SmsQueryByID 消息详情。
func (s *Service) SmsQueryByID(ctx context.Context, id string) (*entity.SysSms, error) {
	var row entity.SysSms
	if err := dao.SysSms.Ctx(ctx).
		Where(dao.SysSms.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// applyTemplateInput 模板字段拷贝。
func applyTemplateInput(data *do.SysSmsTemplate, in TemplateInput) {
	if in.TemplateName != nil {
		data.TemplateName = *in.TemplateName
	}
	if in.TemplateCode != nil {
		data.TemplateCode = *in.TemplateCode
	}
	if in.TemplateType != nil {
		data.TemplateType = *in.TemplateType
	}
	if in.TemplateContent != nil {
		data.TemplateContent = *in.TemplateContent
	}
	if in.TemplateTestJson != nil {
		data.TemplateTestJson = *in.TemplateTestJson
	}
}

// applySmsInput 消息字段拷贝。
func applySmsInput(data *do.SysSms, in SmsInput) {
	if in.EsTitle != nil {
		data.EsTitle = *in.EsTitle
	}
	if in.EsType != nil {
		data.EsType = *in.EsType
	}
	if in.EsReceiver != nil {
		data.EsReceiver = *in.EsReceiver
	}
	if in.EsParam != nil {
		data.EsParam = *in.EsParam
	}
	if in.EsContent != nil {
		data.EsContent = *in.EsContent
	}
	if in.EsSendStatus != nil {
		data.EsSendStatus = *in.EsSendStatus
	}
	if in.EsSendNum != nil {
		data.EsSendNum = *in.EsSendNum
	}
	if in.EsResult != nil {
		data.EsResult = *in.EsResult
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}
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

// gjsonUnmarshalHelper JSON 反序列化 helper。
func gjsonUnmarshalHelper(data string, v interface{}) error {
	return gjson.Unmarshal([]byte(data), v)
}
