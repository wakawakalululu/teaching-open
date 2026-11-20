// Package category 提供分类字典的 HTTP 处理器（契约蓝本：SysCategoryController）。
package category

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcCategory "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/category"
)

// Controller 分类字典控制器。
type Controller struct {
	svc *svcCategory.Service
}

// New 构造控制器。
func New(svc *svcCategory.Service) *Controller {
	return &Controller{svc: svc}
}

// writeError 统一错误映射。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcCategory.ErrEntityNotFound),
		errors.Is(err, svcCategory.ErrInvalidColumn),
		errors.Is(err, svcCategory.ErrLoadTreeParam),
		errors.Is(err, svcCategory.ErrCategoryCodeEmpty):
		common.Fail(r, err.Error())
	default:
		// loadOne/loadAllData 等业务级失败话术直接透传
		common.Fail(r, err.Error())
	}
}

// RootList 根列表分页查询。
func (c *Controller) RootList(r *ghttp.Request) {
	pageNo := r.Get("pageNo", 1).Int()
	pageSize := r.Get("pageSize", 10).Int()
	list, total, err := c.svc.RootList(r.Context(), r.Get("pid").String(), pageNo, pageSize)
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, page.New(list, total, pageNo, pageSize))
}

// ChildList 子列表查询。
func (c *Controller) ChildList(r *ghttp.Request) {
	res, err := c.svc.ChildList(r.Context(), r.Get("pid").String(), r.Get("code").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// QueryById 按 id 查询。
func (c *Controller) QueryById(r *ghttp.Request) {
	res, err := c.svc.QueryById(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增分类（code 由服务端生成）。
func (c *Controller) Add(r *ghttp.Request) {
	var in do.SysCategory
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑分类。
func (c *Controller) Edit(r *ghttp.Request) {
	var in do.SysCategory
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), gconv.String(in.Id), &in); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "修改成功!", nil)
}

// Delete 删除分类。
func (c *Controller) Delete(r *ghttp.Request) {
	if err := c.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除。
func (c *Controller) DeleteBatch(r *ghttp.Request) {
	ids := splitIDs(r.Get("ids").String())
	if len(ids) == 0 {
		common.Fail(r, "参数不识别！")
		return
	}
	if err := c.svc.DeleteBatch(r.Context(), ids); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// LoadOne 单条回显（field 白名单校验）。
func (c *Controller) LoadOne(r *ghttp.Request) {
	res, err := c.svc.LoadOne(r.Context(), r.Get("field").String(), r.Get("val").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// LoadTreeChildren 加载节点子数据。
func (c *Controller) LoadTreeChildren(r *ghttp.Request) {
	res, err := c.svc.LoadTreeChildren(r.Context(), r.Get("pid").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// LoadTreeRoot 加载一级节点（async=false 递归全量）。
func (c *Controller) LoadTreeRoot(r *ghttp.Request) {
	res, err := c.svc.LoadTreeRoot(r.Context(), r.Get("async").Bool(), r.Get("pcode").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// LoadDictItem 表单回显（ids/pcode）。
func (c *Controller) LoadDictItem(r *ghttp.Request) {
	res, err := c.svc.LoadDictItem(r.Context(), r.Get("ids").String(), r.Get("pcode").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// LoadAllData 列表页值替换数据。
func (c *Controller) LoadAllData(r *ghttp.Request) {
	res, err := c.svc.LoadAllData(r.Context(), r.Get("code").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// CheckCode 校验子分类编码前缀。
func (c *Controller) CheckCode(r *ghttp.Request) {
	if err := c.svc.CheckCode(r.Context(), r.Get("pid").String(), r.Get("code").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, nil)
}

// LoadTreeData 分类树控件加载节点。
func (c *Controller) LoadTreeData(r *ghttp.Request) {
	res, err := c.svc.LoadTreeData(r.Context(),
		r.Get("pid").String(), r.Get("pcode").String(), r.Get("condition").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// splitIDs 逗号拆分 ids 参数。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
