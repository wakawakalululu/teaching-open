// Package courseunit 提供课程单元域的 HTTP 处理器
// （契约蓝本：TeachingCourseUnitController 的 10 个核心端点）。
package courseunit

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcCU "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/courseunit"
)

// Controller 课程单元控制器。
type Controller struct {
	svc *svcCU.Service
}

// New 构造控制器。
func New(svc *svcCU.Service) *Controller {
	return &Controller{svc: svc}
}

// unitReq POST/PUT 请求体（TeachingCourseUnit 可编辑字段）。
type unitReq struct {
	Id                string  `json:"id"`
	UnitName          *string `json:"unitName"`
	UnitIntro         *string `json:"unitIntro"`
	UnitCover         *string `json:"unitCover"`
	CourseId          *string `json:"courseId"`
	CourseVideo       *string `json:"courseVideo"`
	CourseVideoSource *int    `json:"courseVideoSource"`
	ShowCourseVideo   *bool   `json:"showCourseVideo"`
	CourseCase        *string `json:"courseCase"`
	ShowCourseCase    *bool   `json:"showCourseCase"`
	CoursePpt         *string `json:"coursePpt"`
	ShowCoursePpt     *bool   `json:"showCoursePpt"`
	CourseWorkType    *int    `json:"courseWorkType"`
	CourseWork        *string `json:"courseWork"`
	CourseWorkAnswer  *string `json:"courseWorkAnswer"`
	CoursePlan        *string `json:"coursePlan"`
	ShowCoursePlan    *bool   `json:"showCoursePlan"`
	MapX              *int    `json:"mapX"`
	MapY              *int    `json:"mapY"`
	MediaContent      *string `json:"mediaContent"`
	OrderNum          *int    `json:"orderNum"`
	DelFlag           *int    `json:"delFlag"`
	// courseMapIcon 为前端冗余字段（Java 反序列化静默丢弃），Go 同样忽略
}

// toUnitInput 请求体 → 服务层输入。
func toUnitInput(in unitReq) svcCU.UnitInput {
	return svcCU.UnitInput{
		Id: in.Id, UnitName: in.UnitName, UnitIntro: in.UnitIntro,
		UnitCover: in.UnitCover, CourseId: in.CourseId, CourseVideo: in.CourseVideo,
		CourseVideoSource: in.CourseVideoSource, ShowCourseVideo: in.ShowCourseVideo,
		CourseCase: in.CourseCase, ShowCourseCase: in.ShowCourseCase,
		CoursePpt: in.CoursePpt, ShowCoursePpt: in.ShowCoursePpt,
		CourseWorkType: in.CourseWorkType, CourseWork: in.CourseWork,
		CourseWorkAnswer: in.CourseWorkAnswer, CoursePlan: in.CoursePlan,
		ShowCoursePlan: in.ShowCoursePlan, MapX: in.MapX, MapY: in.MapY,
		MediaContent: in.MediaContent, OrderNum: in.OrderNum, DelFlag: in.DelFlag,
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
	case errors.Is(err, svcCU.ErrNoCoursePermission), errors.Is(err, svcCU.ErrNotFound):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// MineUnit 我的课程单元分页（学生字段隐藏 + ow365 预览串）。
// 契约：GET /teaching/teachingCourseUnit/mineUnit?courseId=必填&pageNo=&pageSize=。
func (c *Controller) MineUnit(r *ghttp.Request) {
	res, err := c.svc.MineUnit(r.Context(), operatorID(r),
		r.Get("courseId").String(), parsePage(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// List 课程单元管理分页。
// 契约：GET /teaching/teachingCourseUnit/list?pageNo=&pageSize=&courseId=。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("courseId").String(), r.Get("unitName").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetUnitWorkInfo 单元作业信息（含 mineWorkId，历史 SQL 语义保留）。
// 契约：GET /teaching/teachingCourseUnit/getUnitWorkInfo?unitId=必填。
func (c *Controller) GetUnitWorkInfo(r *ghttp.Request) {
	res, err := c.svc.GetUnitWorkInfo(r.Context(), r.Get("unitId").String(), operatorID(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增单元。
// 契约：POST /teaching/teachingCourseUnit/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in unitReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toUnitInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑单元。
// 契约：PUT /teaching/teachingCourseUnit/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in unitReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toUnitInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// EditBatch 地图编辑器批量保存（mapX/mapY/orderNum）。
// 契约：PUT /teaching/teachingCourseUnit/editBatch，body 为单元数组。
func (c *Controller) EditBatch(r *ghttp.Request) {
	var in []unitReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	inputs := make([]svcCU.UnitInput, 0, len(in))
	for _, u := range in {
		inputs = append(inputs, toUnitInput(u))
	}
	if err := c.svc.EditBatch(r.Context(), inputs); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除单元（物理删 + 7 文件键清理）。
// 契约：DELETE /teaching/teachingCourseUnit/delete?id=；不存在同样「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除单元。
// 契约：DELETE /teaching/teachingCourseUnit/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryById 单元详情。
// 契约：GET /teaching/teachingCourseUnit/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
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
