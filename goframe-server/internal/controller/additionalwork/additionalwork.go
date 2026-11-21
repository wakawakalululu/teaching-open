// Package additionalwork 提供附加作业域的 HTTP 处理器
// （契约蓝本：TeachingAdditionalWorkController 的 8 个端点）。
package additionalwork

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcAW "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/additionalwork"
)

// Controller 附加作业控制器。
type Controller struct {
	svc *svcAW.Service
}

// New 构造控制器。
func New(svc *svcAW.Service) *Controller {
	return &Controller{svc: svc}
}

// workReq POST/PUT 请求体（TeachingAdditionalWork 可编辑字段）。
type workReq struct {
	Id              string  `json:"id"`
	CodeType        *int    `json:"codeType"`
	WorkName        *string `json:"workName"`
	WorkDesc        *string `json:"workDesc"`
	WorkCover       *string `json:"workCover"`
	WorkUrl         *string `json:"workUrl"`
	WorkDept        *string `json:"workDept"`
	Status          *int    `json:"status"`
	WorkDocumentUrl *string `json:"workDocumentUrl"`
}

// toWorkInput 请求体 → 服务层输入。
func toWorkInput(in workReq) svcAW.WorkInput {
	return svcAW.WorkInput{
		Id: in.Id, CodeType: in.CodeType, WorkName: in.WorkName, WorkDesc: in.WorkDesc,
		WorkCover: in.WorkCover, WorkUrl: in.WorkUrl, WorkDept: in.WorkDept,
		Status: in.Status, WorkDocumentUrl: in.WorkDocumentUrl,
	}
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

// writeError 统一错误映射。
func writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcAW.ErrNotFound), errors.Is(err, svcAW.ErrNoDepartClass):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// GetWorkInfo 附加作业详情（恒 success，不存在 result=null）。
// 契约：GET /teaching/teachingAdditionalWork/getWorkInfo?id=。
func (c *Controller) GetWorkInfo(r *ghttp.Request) {
	res, err := c.svc.GetWorkInfo(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// List 附加作业分页（非 admin/dev 按负责部门 work_dept OR 链过滤）。
// 契约：GET /teaching/teachingAdditionalWork/list。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), operatorID(r), parsePage(r),
		r.Get("workName").String(), r.Get("status").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// TeacherList 老师附加作业分页（强制 create_by=当前用户）。
// 契约：GET /teaching/teachingAdditionalWork/teacherList。
func (c *Controller) TeacherList(r *ghttp.Request) {
	res, err := c.svc.TeacherList(r.Context(), operatorID(r), parsePage(r),
		r.Get("workName").String(), r.Get("status").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增附加作业（Redis 计数先于落库）。
// 契约：POST /teaching/teachingAdditionalWork/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in workReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toWorkInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑附加作业。
// 契约：PUT /teaching/teachingAdditionalWork/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in workReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toWorkInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除附加作业。
// 契约：DELETE /teaching/teachingAdditionalWork/delete?id=；话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除附加作业。
// 契约：DELETE /teaching/teachingAdditionalWork/deleteBatch?ids=；话术「批量删除成功！」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功！", nil)
}

// QueryByID 附加作业详情（管理端）。
// 契约：GET /teaching/teachingAdditionalWork/queryById?id=；不存在 →「未找到对应数据」。
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
