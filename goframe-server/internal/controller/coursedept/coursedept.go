// Package coursedept 提供班级课程授权域的 HTTP 处理器
// （契约蓝本：TeachingCourseDeptController 的 7 个核心端点）。
package coursedept

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcCD "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/coursedept"
)

// Controller 班级课程授权控制器。
type Controller struct {
	svc *svcCD.Service
}

// New 构造控制器。
func New(svc *svcCD.Service) *Controller {
	return &Controller{svc: svc}
}

// deptReq POST/PUT 请求体（TeachingCourseDept 可编辑字段）。
type deptReq struct {
	Id       string  `json:"id"`
	DeptId   *string `json:"deptId"`
	CourseId *string `json:"courseId"`
	OpenTime *string `json:"openTime"`
}

// addOrUpdateReq POST /teaching/teachingCourseDept/addOrUpdate 请求体
// （DepartCourseVO）。
type addOrUpdateReq struct {
	DeptId       string   `json:"deptId"`
	CourseIdList []string `json:"courseIdList"`
}

// parsePage 解析分页参数（默认 1/10）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// writeError 统一错误映射。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcCD.ErrNotFound):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// List 授权分页列表（三表联表 + courseName/departName 查询）。
// 契约：GET /teaching/teachingCourseDept/list。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("courseName").String(), r.Get("departName").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// AddOrUpdate 批量授权 upsert。
// 契约：POST /teaching/teachingCourseDept/addOrUpdate，body {deptId, courseIdList}。
func (c *Controller) AddOrUpdate(r *ghttp.Request) {
	var in addOrUpdateReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.AddOrUpdate(r.Context(), in.DeptId, in.CourseIdList); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "操作成功！", nil)
}

// Add 新增授权。
// 契约：POST /teaching/teachingCourseDept/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in deptReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), svcCD.DeptInput{
		Id: in.Id, DeptId: in.DeptId, CourseId: in.CourseId, OpenTime: in.OpenTime,
	}); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑授权（开课时间在此维护）。
// 契约：PUT /teaching/teachingCourseDept/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in deptReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), svcCD.DeptInput{
		Id: in.Id, DeptId: in.DeptId, CourseId: in.CourseId, OpenTime: in.OpenTime,
	}); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除授权。
// 契约：DELETE /teaching/teachingCourseDept/delete?id=；成功话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除授权。
// 契约：DELETE /teaching/teachingCourseDept/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryById 授权详情。
// 契约：GET /teaching/teachingCourseDept/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// splitIDs 逗号拆分批量 ids 参数（去除空白项与空段）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
