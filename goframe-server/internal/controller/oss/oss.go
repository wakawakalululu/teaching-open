// Package oss 提供 OSS 文件域的 HTTP 处理器（契约蓝本：OSSFileController）。
package oss

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcOSS "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/oss"
)

// Controller OSS 文件控制器。
type Controller struct {
	svc *svcOSS.Service
}

// New 构造控制器。
func New(svc *svcOSS.Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) writeError(r *ghttp.Request, err error) {
	if errors.Is(err, svcOSS.ErrNotFound) {
		common.Fail(r, err.Error())
		return
	}
	common.Fail(r, "操作失败")
}

// List 文件分页。
// 契约：GET /oss/file/list（pageNo/pageSize/fileName/createBy）。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}, r.Get("fileName").String(), r.Get("createBy").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Upload 上传文件。
// 契约：POST /oss/file/upload（multipart file），返回记录（含 url）。
func (c *Controller) Upload(r *ghttp.Request) {
	file := r.GetUploadFile("file")
	if file == nil {
		common.Fail(r, "上传文件不能为空!")
		return
	}
	saved, err := file.Save("/tmp", true)
	if err != nil || saved == "" {
		common.Fail(r, "上传失败")
		return
	}
	row, err := c.svc.Upload(r.Context(), file.Filename,
		saved, filepath.Ext(file.Filename), operatorOf(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "上传成功", row)
}

// Delete 删除文件（记录 + 物理文件）。
// 契约：DELETE /oss/file/delete?id=
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// QueryByID 文件详情。
// 契约：GET /oss/file/queryById?id=
func (c *Controller) QueryByID(r *ghttp.Request) {
	res, err := c.svc.QueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// operatorOf 取当前操作人（鉴权中间件写入的上下文键）。
func operatorOf(r *ghttp.Request) string {
	if v := r.GetCtxVar("uname"); v != nil && !v.IsNil() {
		return v.String()
	}
	return "system"
}

var _ = strings.TrimSpace
