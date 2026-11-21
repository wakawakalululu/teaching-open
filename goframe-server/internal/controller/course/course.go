// Package course 提供课程域的 HTTP 处理器
// （契约蓝本：TeachingCourseController 的 8 个核心端点；exportXls/importExcel
// 有意未迁移）。
package course

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcCourse "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/course"
)

// Controller 课程控制器。
type Controller struct {
	svc *svcCourse.Service
}

// New 构造控制器。
func New(svc *svcCourse.Service) *Controller {
	return &Controller{svc: svc}
}

// addEditReq POST/PUT /teaching/teachingCourse/add|edit 请求体。
type addEditReq struct {
	Id             string  `json:"id"`
	CourseName     *string `json:"courseName"`
	CourseDesc     *string `json:"courseDesc"`
	CourseIcon     *string `json:"courseIcon"`
	CourseCover    *string `json:"courseCover"`
	ShowType       *int    `json:"showType"`
	CourseMap      *string `json:"courseMap"`
	IsShared       *bool   `json:"isShared"`
	ShowHome       *bool   `json:"showHome"`
	OrderNum       *int    `json:"orderNum"`
	DepartIds      *string `json:"departIds"`
	CourseType     *string `json:"courseType"`
	CourseCategory *string `json:"courseCategory"`
	SysOrgCode     *string `json:"sysOrgCode"`
	DelFlag        *int    `json:"delFlag"`
	// courseMapIcon 为前端冗余字段（Java 反序列化静默丢弃），Go 同样忽略
}

// toCourseInput 请求体 → 服务层输入。
func toCourseInput(in addEditReq) svcCourse.CourseInput {
	return svcCourse.CourseInput{
		Id:             in.Id,
		CourseName:     in.CourseName,
		CourseDesc:     in.CourseDesc,
		CourseIcon:     in.CourseIcon,
		CourseCover:    in.CourseCover,
		ShowType:       in.ShowType,
		CourseMap:      in.CourseMap,
		IsShared:       in.IsShared,
		ShowHome:       in.ShowHome,
		OrderNum:       in.OrderNum,
		DepartIds:      in.DepartIds,
		CourseType:     in.CourseType,
		CourseCategory: in.CourseCategory,
		SysOrgCode:     in.SysOrgCode,
		DelFlag:        in.DelFlag,
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
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcCourse.ErrNotFound):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// MineCourse 我的课程列表。
// 契约：GET /teaching/teachingCourse/mineCourse → result 为注入了
// _dictText/_url 的课程数组。
func (c *Controller) MineCourse(r *ghttp.Request) {
	rows, err := c.svc.MineCourse(r.Context(), operatorID(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	records, err := c.svc.Enrich(r.Context(), rows)
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, records)
}

// GetHomeCourse 首页课程分页（Shiro anon）。
// 契约：GET /teaching/teachingCourse/getHomeCourse（公开）。
func (c *Controller) GetHomeCourse(r *ghttp.Request) {
	res, err := c.svc.GetHomeCourse(r.Context(), parsePage(r),
		r.Get("courseName").String(), r.Get("courseType").String(), r.Get("courseCategory").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// List 课程管理分页列表。
// 契约：GET /teaching/teachingCourse/list?pageNo=&pageSize=&departId=&courseName=...
func (c *Controller) List(r *ghttp.Request) {
	isShared := boolPtr(r.Get("isShared"))
	showHome := boolPtr(r.Get("showHome"))
	orderNum := intPtr(r.Get("orderNum"))
	res, err := c.svc.List(r.Context(), parsePage(r), svcCourse.CourseFilter{
		CourseName:     r.Get("courseName").String(),
		CourseType:     r.Get("courseType").String(),
		CourseCategory: r.Get("courseCategory").String(),
		IsShared:       isShared,
		ShowHome:       showHome,
		OrderNum:       orderNum,
		OrderColumn:    r.Get("column").String(),
		OrderType:      r.Get("order").String(),
	}, r.Get("departId").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增课程。
// 契约：POST /teaching/teachingCourse/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toCourseInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑课程。
// 契约：PUT /teaching/teachingCourse/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in addEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toCourseInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除课程（物理删 + sys_file 记录清理）。
// 契约：DELETE /teaching/teachingCourse/delete?id=；课程不存在同样「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除课程。
// 契约：DELETE /teaching/teachingCourse/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryById 课程详情。
// 契约：GET /teaching/teachingCourse/queryById?id=；不存在 →「未找到对应数据」。
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

// boolPtr 查询参数 → *bool（未提供时 nil）。
func boolPtr(v *gvar.Var) *bool {
	if v == nil || v.IsNil() {
		return nil
	}
	b := v.Bool()
	return &b
}

// intPtr 查询参数 → *int（未提供时 nil）。
func intPtr(v *gvar.Var) *int {
	if v == nil || v.IsNil() {
		return nil
	}
	i := v.Int()
	return &i
}
