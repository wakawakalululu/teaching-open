// Package teachingdaylog 提供班级每日教学记录域的 HTTP 处理器
// （契约蓝本：TeachingDepartDayLogController 的 10 个核心端点）。
package teachingdaylog

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcTDL "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingdaylog"
)

// Controller 班级日志控制器。
type Controller struct {
	svc *svcTDL.Service
}

// New 构造控制器。
func New(svc *svcTDL.Service) *Controller {
	return &Controller{svc: svc}
}

// dayLogReq POST/PUT 请求体（TeachingDepartDayLog 可编辑字段）。
type dayLogReq struct {
	Id                         string  `json:"id"`
	DepartId                   *string `json:"departId"`
	DepartName                 *string `json:"departName"`
	UnitOpenCount              *int    `json:"unitOpenCount"`
	CourseWorkAssignCount      *int    `json:"courseWorkAssignCount"`
	AdditionalWorkAssignCount  *int    `json:"additionalWorkAssignCount"`
	CourseWorkCorrectCount     *int    `json:"courseWorkCorrectCount"`
	AdditionalWorkCorrectCount *int    `json:"additionalWorkCorrectCount"`
	CourseWorkSubmitCount      *int    `json:"courseWorkSubmitCount"`
	AdditionalWorkSubmitCount  *int    `json:"additionalWorkSubmitCount"`
}

// toDayLogInput 请求体 → 服务层输入。
func toDayLogInput(in dayLogReq) svcTDL.DayLogInput {
	return svcTDL.DayLogInput{
		Id: in.Id, DepartId: in.DepartId, DepartName: in.DepartName,
		UnitOpenCount: in.UnitOpenCount, CourseWorkAssignCount: in.CourseWorkAssignCount,
		AdditionalWorkAssignCount:  in.AdditionalWorkAssignCount,
		CourseWorkCorrectCount:     in.CourseWorkCorrectCount,
		AdditionalWorkCorrectCount: in.AdditionalWorkCorrectCount,
		CourseWorkSubmitCount:      in.CourseWorkSubmitCount,
		AdditionalWorkSubmitCount:  in.AdditionalWorkSubmitCount,
	}
}

// parsePage 解析分页参数（默认 1/10）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// currentUser 当前登录用户（ID + userIdentity）。
func currentUser(r *ghttp.Request) (string, int) {
	return r.GetCtxVar(middleware.CtxKeyUserID).String(),
		0
}

// operatorID 当前登录用户 ID。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// writeError 统一错误映射。
func writeError(r *ghttp.Request, err error) {
	if errors.Is(err, svcTDL.ErrNotFound) {
		common.Fail(r, err.Error())
		return
	}
	common.Fail(r, "操作失败")
}

// UnitViewLog 单元打开计数（教师身份幂等计数）。
// 契约：GET /teaching/teachingDepartDayLog/unitViewLog?unitId=。
func (c *Controller) UnitViewLog(r *ghttp.Request) {
	if err := c.svc.UnitViewLog(r.Context(), operatorID(r), r.Get("unitId").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, nil)
}

// GetReport 每日教学记录聚合（单行；多行命中按 MP getOne throwEx 语义报错）。
// 契约：GET /teaching/teachingDepartDayLog/getReport?startTime=&endTime=。
func (c *Controller) GetReport(r *ghttp.Request) {
	res, err := c.svc.GetReport(r.Context(), operatorID(r),
		r.Get("startTime").String(), r.Get("endTime").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetReportGroupByDepart 按班级聚合（Java 忽略分页参数返回全量 List）。
// 契约：GET /teaching/teachingDepartDayLog/getReportGroupByDepart。
func (c *Controller) GetReportGroupByDepart(r *ghttp.Request) {
	res, err := c.svc.GetReportGroupByDepart(r.Context(), operatorID(r),
		r.Get("startTime").String(), r.Get("endTime").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetReportGroupByMonth 按月聚合。
// 契约：GET /teaching/teachingDepartDayLog/getReportGroupByMonth。
func (c *Controller) GetReportGroupByMonth(r *ghttp.Request) {
	res, err := c.svc.GetReportGroupByMonth(r.Context(), operatorID(r),
		r.Get("startTime").String(), r.Get("endTime").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// List 记录分页。
// 契约：GET /teaching/teachingDepartDayLog/list?pageNo=&pageSize=&departId=。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("departId").String(), r.Get("departName").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增记录。
// 契约：POST /teaching/teachingDepartDayLog/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in dayLogReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toDayLogInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑记录。
// 契约：PUT /teaching/teachingDepartDayLog/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in dayLogReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toDayLogInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除记录。
// 契约：DELETE /teaching/teachingDepartDayLog/delete?id=；话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除记录。
// 契约：DELETE /teaching/teachingDepartDayLog/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryByID 记录详情。
// 契约：GET /teaching/teachingDepartDayLog/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) QueryByID(r *ghttp.Request) {
	res, err := c.svc.QueryByID(r.Context(), r.Get("id").String())
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
