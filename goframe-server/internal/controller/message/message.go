// Package message 提供消息模板与消息记录域的 HTTP 处理器
// （契约蓝本：SysMessageTemplateController + SysMessageController）。
package message

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcMSG "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/message"
)

// Controller 消息域控制器。
type Controller struct {
	svc *svcMSG.Service
}

// New 构造控制器。
func New(svc *svcMSG.Service) *Controller {
	return &Controller{svc: svc}
}

// templateReq POST/PUT 请求体（SysMessageTemplate 可编辑字段）。
type templateReq struct {
	Id               string  `json:"id"`
	TemplateName     *string `json:"templateName"`
	TemplateCode     *string `json:"templateCode"`
	TemplateType     *string `json:"templateType"`
	TemplateContent  *string `json:"templateContent"`
	TemplateTestJson *string `json:"templateTestJson"`
}

// smsReq POST/PUT 请求体（SysSms 可编辑字段）。
type smsReq struct {
	Id           string  `json:"id"`
	EsTitle      *string `json:"esTitle"`
	EsType       *string `json:"esType"`
	EsReceiver   *string `json:"esReceiver"`
	EsParam      *string `json:"esParam"`
	EsContent    *string `json:"esContent"`
	EsSendStatus *string `json:"esSendStatus"`
	EsSendNum    *int    `json:"esSendNum"`
	EsResult     *string `json:"esResult"`
	Remark       *string `json:"remark"`
}

// sendMsgReq POST /sendMsg 请求体（MsgParams）。
type sendMsgReq struct {
	MsgType      string `json:"msgType"`
	TemplateCode string `json:"templateCode"`
	TestData     string `json:"testData"`
	Receiver     string `json:"receiver"`
}

// parsePage 解析分页参数（默认 1/10）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// writeError 统一错误映射。
func writeError(r *ghttp.Request, err error) {
	if errors.Is(err, svcMSG.ErrNotFound) {
		common.Fail(r, err.Error())
		return
	}
	common.Fail(r, "操作失败")
}

// TemplateList 模板分页。
// 契约：GET /message/sysMessageTemplate/list。
func (c *Controller) TemplateList(r *ghttp.Request) {
	res, err := c.svc.TemplateList(r.Context(), parsePage(r),
		r.Get("templateCode").String(), r.Get("templateName").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// TemplateAdd 新增模板。
// 契约：POST /message/sysMessageTemplate/add；成功话术「添加成功！」。
func (c *Controller) TemplateAdd(r *ghttp.Request) {
	var in templateReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.TemplateAdd(r.Context(), svcMSG.TemplateInput{
		Id: in.Id, TemplateName: in.TemplateName, TemplateCode: in.TemplateCode,
		TemplateType: in.TemplateType, TemplateContent: in.TemplateContent,
		TemplateTestJson: in.TemplateTestJson,
	}); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// TemplateEdit 编辑模板。
// 契约：PUT /message/sysMessageTemplate/edit；成功话术「编辑成功!」。
func (c *Controller) TemplateEdit(r *ghttp.Request) {
	var in templateReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.TemplateEdit(r.Context(), svcMSG.TemplateInput{
		Id: in.Id, TemplateName: in.TemplateName, TemplateCode: in.TemplateCode,
		TemplateType: in.TemplateType, TemplateContent: in.TemplateContent,
		TemplateTestJson: in.TemplateTestJson,
	}); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// TemplateDelete 删除模板。
// 契约：DELETE /message/sysMessageTemplate/delete?id=；话术「删除成功!」。
func (c *Controller) TemplateDelete(r *ghttp.Request) {
	if err := c.svc.TemplateDelete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// TemplateDeleteBatch 批量删除模板。
// 契约：DELETE /message/sysMessageTemplate/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) TemplateDeleteBatch(r *ghttp.Request) {
	if err := c.svc.TemplateDeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// TemplateQueryByID 模板详情。
// 契约：GET /message/sysMessageTemplate/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) TemplateQueryByID(r *ghttp.Request) {
	res, err := c.svc.TemplateQueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// SendMsg 按模板发送消息（渲染占位符 + 落发送记录）。
// 契约：POST /message/sysMessageTemplate/sendMsg，body MsgParams；
// 成功话术「发送消息任务添加成功！」/失败「发送消息任务添加失败！」。
func (c *Controller) SendMsg(r *ghttp.Request) {
	var in sendMsgReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "解析Json出错！")
		return
	}
	if err := c.svc.SendMsg(r.Context(), in.MsgType, in.TemplateCode, in.TestData, in.Receiver); err != nil {
		if strings.Contains(err.Error(), "发送消息任务添加失败！") {
			common.Fail(r, "发送消息任务添加失败！")
			return
		}
		writeError(r, err)
		return
	}
	common.OKMsg(r, "发送消息任务添加成功！", nil)
}

// SmsList 消息记录分页。
// 契约：GET /message/sysMessage/list。
func (c *Controller) SmsList(r *ghttp.Request) {
	res, err := c.svc.SmsList(r.Context(), parsePage(r),
		r.Get("esReceiver").String(), r.Get("esSendStatus").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// SmsAdd 新增消息记录。
// 契约：POST /message/sysMessage/add；成功话术「添加成功！」。
func (c *Controller) SmsAdd(r *ghttp.Request) {
	var in smsReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SmsAdd(r.Context(), svcMSG.SmsInput{
		Id: in.Id, EsTitle: in.EsTitle, EsType: in.EsType, EsReceiver: in.EsReceiver,
		EsParam: in.EsParam, EsContent: in.EsContent, EsSendStatus: in.EsSendStatus,
		EsSendNum: in.EsSendNum, EsResult: in.EsResult, Remark: in.Remark,
	}); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// SmsEdit 编辑消息记录。
// 契约：PUT /message/sysMessage/edit；成功话术「编辑成功!」。
func (c *Controller) SmsEdit(r *ghttp.Request) {
	var in smsReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SmsEdit(r.Context(), svcMSG.SmsInput{
		Id: in.Id, EsTitle: in.EsTitle, EsType: in.EsType, EsReceiver: in.EsReceiver,
		EsParam: in.EsParam, EsContent: in.EsContent, EsSendStatus: in.EsSendStatus,
		EsSendNum: in.EsSendNum, EsResult: in.EsResult, Remark: in.Remark,
	}); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// SmsDelete 删除消息记录。
// 契约：DELETE /message/sysMessage/delete?id=；话术「删除成功!」。
func (c *Controller) SmsDelete(r *ghttp.Request) {
	if err := c.svc.SmsDelete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// SmsDeleteBatch 批量删除消息记录。
// 契约：DELETE /message/sysMessage/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) SmsDeleteBatch(r *ghttp.Request) {
	if err := c.svc.SmsDeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// SmsQueryByID 消息详情。
// 契约：GET /message/sysMessage/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) SmsQueryByID(r *ghttp.Request) {
	res, err := c.svc.SmsQueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// splitIDs 逗号拆分批量 ids 参数。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
