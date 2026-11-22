// Package dict 提供字典与字典项的 HTTP 处理器（契约蓝本：SysDictController/SysDictItemController）。
package dict

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcDict "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/dict"
)

// Controller 字典控制器。
type Controller struct {
	svc *svcDict.Service
}

// New 构造控制器。
func New(svc *svcDict.Service) *Controller {
	return &Controller{svc: svc}
}

// ok 内部便捷成功响应。
func ok(r *ghttp.Request, data interface{}) {
	common.OK(r, data)
}

// writeError 统一错误映射：业务错误返回原话术，未知错误记录日志并返回通用话术。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcDict.ErrDictCodeInvalid),
		errors.Is(err, svcDict.ErrEntityNotFound),
		errors.Is(err, svcDict.ErrIDsRequired),
		errors.Is(err, svcDict.ErrTableDictFilterUnsupported):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "字典操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// parsePage 解析分页参数（默认 1/10，同 Java @RequestParam 默认值）。
func parsePage(r *ghttp.Request) svcDict.PageParam {
	return svcDict.PageParam{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// splitIDs 逗号拆分 ids 参数（去除空白项）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// QueryAllDictItems 一次性返回全部字典（登录后初始化强依赖）。
func (c *Controller) QueryAllDictItems(r *ghttp.Request) {
	res, err := c.svc.QueryAllDictItems(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// GetDictItems 按字典 Code 获取字典项（普通字典/表字典二态）。
func (c *Controller) GetDictItems(r *ghttp.Request) {
	dictCode := r.GetRouter("dictCode").String()
	if dictCode == "" {
		dictCode = r.Get("dictCode").String()
	}
	res, err := c.svc.GetDictItems(r.Context(), dictCode)
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// GetDictText 按字典 Code 与键值取文本。
func (c *Controller) GetDictText(r *ghttp.Request) {
	text, err := c.svc.GetDictText(r.Context(),
		r.GetRouter("dictCode").String(), r.GetRouter("key").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, text)
}

// LoadDict 表字典异步搜索（文本列模糊匹配）。
func (c *Controller) LoadDict(r *ghttp.Request) {
	res, err := c.svc.LoadDict(r.Context(),
		r.GetRouter("dictCode").String(), r.Get("keyword").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// LoadDictItem 表字典回显（按值列多个键取文本）。
func (c *Controller) LoadDictItem(r *ghttp.Request) {
	keys := splitIDs(r.Get("key").String())
	res, err := c.svc.LoadDictItem(r.Context(), r.GetRouter("dictCode").String(), keys)
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// List 字典分页列表。
func (c *Controller) List(r *ghttp.Request) {
	res, err := c.svc.ListPage(r.Context(), parsePage(r),
		r.Get("dictName").String(), r.Get("dictCode").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// TreeList 树形字典列表。
func (c *Controller) TreeList(r *ghttp.Request) {
	res, err := c.svc.TreeList(r.Context(), r.Get("dictName").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// QueryTableData 表字典分页查询（契约：GET /sys/dict/queryTableData）。
func (c *Controller) QueryTableData(r *ghttp.Request) {
	res, err := c.svc.QueryTableData(r.Context(),
		r.Get("table").String(), r.Get("text").String(), r.Get("code").String(),
		r.Get("condition").String(), page.Param{
			PageNo:   r.Get("pageNo", 1).Int(),
			PageSize: r.Get("pageSize", 10).Int(),
		})
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// LoadTreeData 通用表树加载（契约：GET /sys/dict/loadTreeData）。
func (c *Controller) LoadTreeData(r *ghttp.Request) {
	res, err := c.svc.LoadTreeData(r.Context(),
		r.Get("pid").String(), r.Get("pidField").String(), r.Get("tableName").String(),
		r.Get("text").String(), r.Get("code").String(), r.Get("hasChildField").String(),
		r.Get("condition").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// GetByCode 按字典编码查询。
func (c *Controller) GetByCode(r *ghttp.Request) {
	res, err := c.svc.GetByCode(r.Context(), r.Get("dicCode").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// Add 新增字典。
func (c *Controller) Add(r *ghttp.Request) {
	var in do.SysDict
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "保存成功！", nil)
}

// Edit 编辑字典。
func (c *Controller) Edit(r *ghttp.Request) {
	var in do.SysDict
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), gconv.String(in.Id), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 软删除字典（回收站）。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量软删除。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteList 回收站列表。
func (c *Controller) DeleteList(r *ghttp.Request) {
	res, err := c.svc.DeleteList(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// DeletePhysic 物理删除字典并级联清理字典项。
func (c *Controller) DeletePhysic(r *ghttp.Request) {
	if err := c.svc.DeletePhysic(r.Context(), r.GetRouter("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// Back 回收站恢复。
func (c *Controller) Back(r *ghttp.Request) {
	if err := c.svc.Back(r.Context(), r.GetRouter("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "操作成功!", nil)
}

// RefleshCache 刷新字典缓存。
//
// Go 版字典直查库（无 Redis 缓存层），此接口保留以兼容前端管理页按钮，
// 返回成功即可。
func (c *Controller) RefleshCache(r *ghttp.Request) {
	ok(r, nil)
}

// ItemList 字典项分页列表。
func (c *Controller) ItemList(r *ghttp.Request) {
	res, err := c.svc.ItemList(r.Context(), parsePage(r),
		r.Get("dictId").String(), r.Get("itemText").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	ok(r, res)
}

// ItemAdd 新增字典项。
func (c *Controller) ItemAdd(r *ghttp.Request) {
	var in do.SysDictItem
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.ItemAdd(r.Context(), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "保存成功！", nil)
}

// ItemEdit 编辑字典项。
func (c *Controller) ItemEdit(r *ghttp.Request) {
	var in do.SysDictItem
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.ItemEdit(r.Context(), gconv.String(in.Id), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// ItemDelete 删除字典项。
func (c *Controller) ItemDelete(r *ghttp.Request) {
	if err := c.svc.ItemDelete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// ItemDeleteBatch 批量删除字典项。
func (c *Controller) ItemDeleteBatch(r *ghttp.Request) {
	if err := c.svc.ItemDeleteBatch(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}
