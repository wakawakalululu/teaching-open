// Package teachingnews 提供资讯域的 HTTP 处理器
// （契约蓝本：TeachingNewsController 的 8 个核心端点）。
package teachingnews

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcTN "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingnews"
)

// Controller 资讯控制器。
type Controller struct {
	svc *svcTN.Service
}

// New 构造控制器。
func New(svc *svcTN.Service) *Controller {
	return &Controller{svc: svc}
}

// newsReq POST/PUT 请求体（TeachingNews 可编辑字段）。
type newsReq struct {
	Id          string  `json:"id"`
	NewsTitle   *string `json:"newsTitle"`
	NewsContent *string `json:"newsContent"`
	NewsStatus  *int    `json:"newsStatus"`
}

// toNewsInput 请求体 → 服务层输入。
func toNewsInput(in newsReq) svcTN.NewsInput {
	return svcTN.NewsInput{
		Id: in.Id, NewsTitle: in.NewsTitle, NewsContent: in.NewsContent, NewsStatus: in.NewsStatus,
	}
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
	if errors.Is(err, svcTN.ErrNotFound) {
		common.Fail(r, err.Error())
		return
	}
	common.Fail(r, "操作失败")
}

// NewsList 已发布资讯分页（anon，列表不回大字段）。
// 契约：GET /teaching/teachingNews/newsList。
func (c *Controller) NewsList(r *ghttp.Request) {
	res, err := c.svc.NewsList(r.Context(), parsePage(r), r.Get("newsTitle").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// NewsDetail 资讯详情（anon，全量含 content）。
// 契约：GET /teaching/teachingNews/newsDetail?id=；status<=0 →「未找到对应数据」。
func (c *Controller) NewsDetail(r *ghttp.Request) {
	res, err := c.svc.NewsDetail(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// List 资讯管理分页。
// 契约：GET /teaching/teachingNews/list。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r), r.Get("newsTitle").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增资讯。
// 契约：POST /teaching/teachingNews/add；成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in newsReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toNewsInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑资讯。
// 契约：PUT /teaching/teachingNews/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in newsReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toNewsInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除资讯。
// 契约：DELETE /teaching/teachingNews/delete?id=；成功话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除资讯。
// 契约：DELETE /teaching/teachingNews/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryByID 资讯详情（管理端）。
// 契约：GET /teaching/teachingNews/queryById?id=；不存在 →「未找到对应数据」。
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
