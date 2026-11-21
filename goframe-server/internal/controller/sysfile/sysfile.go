// Package sysfile 提供文件记录域的 HTTP 处理器
// （契约蓝本：SysFileController 的 7 个核心端点，基路径 /system/sysFile）。
package sysfile

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcSF "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/sysfile"
)

// Controller 文件记录控制器。
type Controller struct {
	svc *svcSF.Service
}

// New 构造控制器。
func New(svc *svcSF.Service) *Controller {
	return &Controller{svc: svc}
}

// fileReq POST/PUT 请求体（SysFile 可编辑字段）。
type fileReq struct {
	Id           string  `json:"id"`
	FileType     *int    `json:"fileType"`
	FileName     *string `json:"fileName"`
	FilePath     *string `json:"filePath"`
	FileLocation *int    `json:"fileLocation"`
	FileTag      *string `json:"fileTag"`
	SysOrgCode   *string `json:"sysOrgCode"`
	DelFlag      *int    `json:"delFlag"`
}

// toFileInput 请求体 → 服务层输入。
func toFileInput(in fileReq) svcSF.FileInput {
	return svcSF.FileInput{
		Id: in.Id, FileType: in.FileType, FileName: in.FileName,
		FilePath: in.FilePath, FileLocation: in.FileLocation,
		FileTag: in.FileTag, SysOrgCode: in.SysOrgCode, DelFlag: in.DelFlag,
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
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcSF.ErrNotFound):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// List 文件记录分页。
// 契约：GET /system/sysFile/list?pageNo=&pageSize=&fileName=&filePath=&fileTag=。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("fileName").String(), r.Get("filePath").String(), r.Get("fileTag").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增文件记录（返回回写 id 的实体，JUpload 依赖 res.result.id）。
// 契约：POST /system/sysFile/add。
func (c *Controller) Add(r *ghttp.Request) {
	var in fileReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	res, err := c.svc.Add(r.Context(), toFileInput(in))
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Edit 编辑文件记录。
// 契约：PUT /system/sysFile/edit；成功话术「编辑成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in fileReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toFileInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除文件记录 + 物理文件。
// 契约：DELETE /system/sysFile/delete?id=；记录不存在同样「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteByPath 按 file_path 精确匹配删除（恒成功，JUpload 自动清理依赖）。
// 契约：DELETE /system/sysFile/deleteByPath?filePath=。
func (c *Controller) DeleteByPath(r *ghttp.Request) {
	if err := c.svc.DeleteByPath(r.Context(), r.Get("filePath").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除。
// 契约：DELETE /system/sysFile/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryById 文件记录详情。
// 契约：GET /system/sysFile/queryById?id=；不存在 →「未找到对应数据」。
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
