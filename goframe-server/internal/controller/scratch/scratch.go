// Package scratch 提供素材库与背包域的 HTTP 处理器
// （契约蓝本：TeachingScratchAssetsController 7 个核心端点 +
// TeachingScratchBackpackController 4 个端点）。
package scratch

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcSC "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/scratch"
)

// Controller Scratch 素材库与背包控制器。
type Controller struct {
	svc *svcSC.Service
}

// New 构造控制器。
func New(svc *svcSC.Service) *Controller {
	return &Controller{svc: svc}
}

// assetReq POST/PUT 请求体（TeachingScratchAssets 可编辑字段）。
type assetReq struct {
	Id        string  `json:"id"`
	AssetType *int    `json:"assetType"`
	AssetName *string `json:"assetName"`
	AssetData *string `json:"assetData"`
	Md5Ext    *string `json:"md5Ext"`
	Tags      *string `json:"tags"`
	DelFlag   *int    `json:"delFlag"`
}

// toAssetInput 请求体 → 服务层输入。
func toAssetInput(in assetReq) svcSC.AssetInput {
	return svcSC.AssetInput{
		Id: in.Id, AssetType: in.AssetType, AssetName: in.AssetName,
		AssetData: in.AssetData, Md5Ext: in.Md5Ext, Tags: in.Tags, DelFlag: in.DelFlag,
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
	if errors.Is(err, svcSC.ErrNotFound) {
		common.Fail(r, err.Error())
		return
	}
	common.Fail(r, "操作失败")
}

// GetScratchAssets 素材库列表（裸 JSONArray，非 Result 包装）。
// 契约：GET /teaching/teachingScratchAssets/getScratchAssets?assetType=1背景 2声音 3造型 4角色。
func (c *Controller) GetScratchAssets(r *ghttp.Request) {
	res, err := c.svc.GetScratchAssets(r.Context(), r.Get("assetType").Int())
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	r.Response.WriteJson(res)
}

// List 素材分页。
// 契约：GET /teaching/teachingScratchAssets/list。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), parsePage(r),
		r.Get("assetType").String(), r.Get("assetName").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// AssetAdd 新增素材。
// 契约：POST /teaching/teachingScratchAssets/add；成功话术「添加成功！」。
func (c *Controller) AssetAdd(r *ghttp.Request) {
	var in assetReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), toAssetInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// AssetEdit 编辑素材。
// 契约：PUT /teaching/teachingScratchAssets/edit；成功话术「编辑成功!」。
func (c *Controller) AssetEdit(r *ghttp.Request) {
	var in assetReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), toAssetInput(in)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// AssetDelete 删除素材（assetType==4 角色清理为 TODO——Java 现状）。
// 契约：DELETE /teaching/teachingScratchAssets/delete?id=；话术「删除成功!」。
func (c *Controller) AssetDelete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// AssetDeleteBatch 批量删除素材。
// 契约：DELETE /teaching/teachingScratchAssets/deleteBatch?ids=；话术「批量删除成功!」。
func (c *Controller) AssetDeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// AssetQueryByID 素材详情。
// 契约：GET /teaching/teachingScratchAssets/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) AssetQueryByID(r *ghttp.Request) {
	res, err := c.svc.QueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// BackpackGetFile 背包 body 遗留键读取（恒空响应——Java 现状保留）。
// 契约：GET /teaching/scratch/backpack/getFile/{filename}。
func (c *Controller) BackpackGetFile(r *ghttp.Request) {
	raw := c.svc.BackpackGetFile(r.Context(), r.Get("filename").String())
	r.Response.Header().Set("Content-Type", "application/octet-stream")
	r.Response.Write(raw)
}

// BackpackContent 背包列表。
// 契约：GET /teaching/scratch/backpack/content?limit=20&offset=0。
func (c *Controller) BackpackContent(r *ghttp.Request) {
	res, err := c.svc.BackpackContent(r.Context(), operatorID(r),
		r.Get("limit", 20).Int(), r.Get("offset", 0).Int())
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}

// backpackSaveReq POST /save 请求体。
type backpackSaveReq struct {
	Body      string `json:"body"`
	Thumbnail string `json:"thumbnail"`
	Type      string `json:"type"`
	Mime      string `json:"mime"`
	Name      string `json:"name"`
}

// BackpackSave 保存到背包。
// 契约：POST /teaching/scratch/backpack/save → Result.ok(条目)。
func (c *Controller) BackpackSave(r *ghttp.Request) {
	var in backpackSaveReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	res, err := c.svc.BackpackSave(r, r.Context(), operatorID(r),
		in.Body, in.Thumbnail, in.Type, in.Mime, in.Name)
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}

// BackpackDelete 删除背包条目。
// 契约：DELETE /teaching/scratch/backpack/delete?id=。
func (c *Controller) BackpackDelete(r *ghttp.Request) {
	if err := c.svc.BackpackDelete(r.Context(), operatorID(r), r.Get("id").String()); err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, nil)
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
