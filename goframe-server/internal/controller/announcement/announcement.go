// Package announcement 提供系统通告与发送记录域的 HTTP 处理器
// （契约蓝本：SysAnnouncementController 10 个核心端点 +
// SysAnnouncementSendController 9 个端点）。
package announcement

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcAnn "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/announcement"
)

// Controller 系统通告控制器。
type Controller struct {
	svc *svcAnn.Service
}

// New 构造控制器。
func New(svc *svcAnn.Service) *Controller {
	return &Controller{svc: svc}
}

// annReq POST/PUT 请求体（SysAnnouncement 可编辑字段；userIds 为 Java
// @TableField(exist=false) 的非表列入参）。
type annReq struct {
	Id          string  `json:"id"`
	Titile      *string `json:"titile"`
	MsgContent  *string `json:"msgContent"`
	StartTime   *string `json:"startTime"`
	EndTime     *string `json:"endTime"`
	Priority    *string `json:"priority"`
	MsgCategory *string `json:"msgCategory"`
	MsgType     *string `json:"msgType"`
	SendStatus  *string `json:"sendStatus"`
	BusType     *string `json:"busType"`
	BusId       *string `json:"busId"`
	OpenType    *string `json:"openType"`
	OpenPage    *string `json:"openPage"`
	MsgAbstract *string `json:"msgAbstract"`
	UserIds     *string `json:"userIds"`
}

// toAnnInput 请求体 → 服务层输入。
func toAnnInput(in annReq) svcAnn.AnnouncementInput {
	return svcAnn.AnnouncementInput{
		Id: in.Id, Titile: in.Titile, MsgContent: in.MsgContent,
		StartTime: in.StartTime, EndTime: in.EndTime, Priority: in.Priority,
		MsgCategory: in.MsgCategory, MsgType: in.MsgType, SendStatus: in.SendStatus,
		BusType: in.BusType, BusId: in.BusId, OpenType: in.OpenType,
		OpenPage: in.OpenPage, MsgAbstract: in.MsgAbstract, UserIds: in.UserIds,
	}
}

// sendReq POST/PUT 请求体（SysAnnouncementSend 可编辑字段）。
type sendReq struct {
	Id       string  `json:"id"`
	AnntId   *string `json:"anntId"`
	UserId   *string `json:"userId"`
	ReadFlag *string `json:"readFlag"`
}

// parsePage 解析分页参数（默认 1/10）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// operatorID 当前登录用户 ID。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// operatorName 当前登录用户名（doReleaseData sender 用）。
func operatorName(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUsername).String()
}

// writeError 统一错误映射。
func writeError(r *ghttp.Request, err error) {
	if errors.Is(err, svcAnn.ErrNotFound) {
		common.Fail(r, err.Error())
		return
	}
	common.Fail(r, "操作失败")
}

// List 通告分页（del_flag=0 + 排序）。
// 契约：GET /sys/annountCement/list。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("titile").String(), r.Get("msgCategory").String(),
		r.Get("msgType").String(), r.Get("sender").String(),
		r.Get("column").String(), r.Get("order").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增通告（msgType=USER 同步插入标记行）。
// 契约：POST /sys/annountCement/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in annReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toAnnInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑通告。
// 契约：PUT /sys/annountCement/edit；成功话术「修改成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in annReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toAnnInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "修改成功!", nil)
}

// Delete 逻辑删通告。
// 契约：DELETE /sys/annountCement/delete?id=；话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量逻辑删。
// 契约：DELETE /sys/annountCement/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryByID 通告详情。
// 契约：GET /sys/annountCement/queryById?id=；不存在 →「未找到对应实体」。
func (c *Controller) QueryByID(r *ghttp.Request) {
	res, err := c.svc.QueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// DoReleaseData 发布通告。
// 契约：GET /sys/annountCement/doReleaseData?id=；话术「该系统通知发布成功」。
func (c *Controller) DoReleaseData(r *ghttp.Request) {
	if err := c.svc.DoReleaseData(r.Context(), r.Get("id").String(), operatorName(r)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "该系统通知发布成功", nil)
}

// DoReovkeData 撤销发布。
// 契约：GET /sys/annountCement/doReovkeData?id=；话术「该系统通知撤销成功」。
func (c *Controller) DoReovkeData(r *ghttp.Request) {
	if err := c.svc.DoReovkeData(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "该系统通知撤销成功", nil)
}

// ListByUser 我的消息面板（标记行补插 + 未读消息分分类前 5 条）。
// 契约：GET /sys/annountCement/listByUser。
func (c *Controller) ListByUser(r *ghttp.Request) {
	res, err := c.svc.ListByUser(r.Context(), operatorID(r))
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// SyncNotic websocket 同步端点（websocket 未迁移，返回空成功——文档化差距）。
// 契约：GET /sys/annountCement/syncNotic?anntId=可选。
func (c *Controller) SyncNotic(r *ghttp.Request) {
	common.OK(r, nil)
}

// SendList 发送记录分页。
// 契约：GET /sys/sysAnnouncementSend/list。
func (c *Controller) SendList(r *ghttp.Request) {
	res, err := c.svc.SendList(r.Context(), parsePage(r),
		r.Get("anntId").String(), r.Get("userId").String(), r.Get("readFlag").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// SendAdd 新增发送记录。
// 契约：POST /sys/sysAnnouncementSend/add；成功话术「添加成功！」。
func (c *Controller) SendAdd(r *ghttp.Request) {
	var in sendReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SendAdd(r.Context(), svcAnn.SendInput{
		Id: in.Id, AnntId: in.AnntId, UserId: in.UserId, ReadFlag: in.ReadFlag,
	}); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// SendEdit 编辑发送记录。
// 契约：PUT /sys/sysAnnouncementSend/edit；成功话术「编辑成功!」。
func (c *Controller) SendEdit(r *ghttp.Request) {
	var in sendReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SendEdit(r.Context(), svcAnn.SendInput{
		Id: in.Id, AnntId: in.AnntId, UserId: in.UserId, ReadFlag: in.ReadFlag,
	}); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// SendDelete 删除发送记录。
// 契约：DELETE /sys/sysAnnouncementSend/delete?id=；话术「删除成功!」。
func (c *Controller) SendDelete(r *ghttp.Request) {
	if err := c.svc.SendDelete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// SendDeleteBatch 批量删除发送记录。
// 契约：DELETE /sys/sysAnnouncementSend/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) SendDeleteBatch(r *ghttp.Request) {
	if err := c.svc.SendDeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// SendQueryByID 发送记录详情。
// 契约：GET /sys/sysAnnouncementSend/queryById?id=；不存在 →「未找到对应实体」。
func (c *Controller) SendQueryByID(r *ghttp.Request) {
	res, err := c.svc.SendQueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// EditByAnntIdAndUserId 标记单条通告已读。
// 契约：PUT /sys/sysAnnouncementSend/editByAnntIdAndUserId，body {anntId}。
func (c *Controller) EditByAnntIdAndUserID(r *ghttp.Request) {
	var in struct {
		AnntId string `json:"anntId"`
	}
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.EditByAnntIdAndUserID(r.Context(), operatorID(r), in.AnntId); err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, nil)
}

// GetMyAnnouncementSend 我的消息分页（JOIN 查询）。
// 契约：GET /sys/sysAnnouncementSend/getMyAnnouncementSend?titile=&sender=&readFlag=&busType=。
func (c *Controller) GetMyAnnouncementSend(r *ghttp.Request) {
	res, err := c.svc.GetMyAnnouncementSend(r.Context(), operatorID(r), parsePage(r),
		r.Get("titile").String(), r.Get("sender").String(),
		r.Get("readFlag").String(), r.Get("busType").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// ReadAll 一键已读。
// 契约：PUT /sys/sysAnnouncementSend/readAll；话术「全部已读」。
func (c *Controller) ReadAll(r *ghttp.Request) {
	if err := c.svc.ReadAll(r.Context(), operatorID(r)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "全部已读", nil)
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
