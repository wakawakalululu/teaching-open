// Package depart 提供部门管理域的 HTTP 处理器
// （契约蓝本：SysDepartController 的 10 个核心端点）。
package depart

import (
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcDepart "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/depart"
)

// Controller 部门管理控制器。
type Controller struct {
	svc *svcDepart.Service
}

// New 构造控制器。
func New(svc *svcDepart.Service) *Controller {
	return &Controller{svc: svc}
}

// departAddEditReq POST/PUT /sys/sysDepart/add|edit 请求体。
// 指针字段对齐 MyBatis-Plus NOT_NULL 字段策略（未提供不参与写库）。
type departAddEditReq struct {
	Id             string  `json:"id"`
	ParentId       *string `json:"parentId"`
	DepartName     *string `json:"departName"`
	DepartNameEn   *string `json:"departNameEn"`
	DepartNameAbbr *string `json:"departNameAbbr"`
	DepartOrder    *int    `json:"departOrder"`
	Description    *string `json:"description"`
	OrgCategory    *string `json:"orgCategory"`
	OrgType        *string `json:"orgType"`
	OrgCode        *string `json:"orgCode"`
	Mobile         *string `json:"mobile"`
	Fax            *string `json:"fax"`
	Address        *string `json:"address"`
	Memo           *string `json:"memo"`
	Status         *string `json:"status"`
	// 前端可能携带的冗余字段（Java 反序列化静默丢弃，Go 同样忽略）
	DelFlag    *string `json:"delFlag"`
	CreateBy   *string `json:"createBy"`
	CreateTime *string `json:"createTime"`
}

// toDepartInput 请求体 → 服务层输入。
func toDepartInput(in departAddEditReq) svcDepart.DepartInput {
	return svcDepart.DepartInput{
		Id:             in.Id,
		ParentId:       in.ParentId,
		DepartName:     in.DepartName,
		DepartNameEn:   in.DepartNameEn,
		DepartNameAbbr: in.DepartNameAbbr,
		DepartOrder:    in.DepartOrder,
		Description:    in.Description,
		OrgCategory:    in.OrgCategory,
		OrgType:        in.OrgType,
		OrgCode:        in.OrgCode,
		Mobile:         in.Mobile,
		Fax:            in.Fax,
		Address:        in.Address,
		Memo:           in.Memo,
		Status:         in.Status,
	}
}

// operatorID 当前登录用户 ID（等价 Java JwtUtil.getUserNameByToken 链路的用户态）。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// splitIDs 逗号拆分批量 ids 参数（去除空白项与空段，解析健壮化）。
func splitIDs(ids string) []string {
	var out []string
	for _, part := range strings.Split(ids, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// writeError 统一错误映射：业务错误返回 Java 原话术，未知错误记录日志并返回
// 通用话术。
func (c *Controller) writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcDepart.ErrNotFound),
		errors.Is(err, svcDepart.ErrParentChange),
		errors.Is(err, svcDepart.ErrParamInvalid):
		common.Fail(r, err.Error())
	default:
		g.Log().Errorf(r.Context(), "部门域操作异常 err=%s", err.Error())
		common.Fail(r, "操作失败")
	}
}

// QueryTreeList 全量部门树。
// 契约：GET /sys/sysDepart/queryTreeList → result 为 SysDepartTreeModel 树。
func (c *Controller) QueryTreeList(r *ghttp.Request) {
	res, err := c.svc.QueryTreeList(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// QueryMyDeptTreeList 我的部门树（上级用户）。
// 契约：GET /sys/sysDepart/queryMyDeptTreeList；message 字段承载身份
// （"2"=上级/"1"=普通用户）；普通用户 result 为 null。
func (c *Controller) QueryMyDeptTreeList(r *ghttp.Request) {
	tree, identity, err := c.svc.QueryMyDeptTreeList(r.Context(), operatorID(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	if identity != "2" {
		// Java：result=null + message="1" + success=true
		r.Response.WriteJson(g.Map{
			"success":   true,
			"message":   identity,
			"code":      0,
			"result":    nil,
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}
	// Java：result=树 + message="2" + success=true（code 保持默认 0，
	// 仓库统一约定成功响应 code=200，与 P0-1 先例一致）
	common.OKMsg(r, identity, tree)
}

// QueryIdTree 部门 id/名称树。
// 契约：GET /sys/sysDepart/queryIdTree → result 为 DepartIdModel 树。
func (c *Controller) QueryIdTree(r *ghttp.Request) {
	res, err := c.svc.QueryIdTree(r.Context())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Add 新增部门。
// 契约：POST /sys/sysDepart/add；create_by=当前用户；orgCode/orgType 服务端生成；
// 成功话术「添加成功！」。
func (c *Controller) Add(r *ghttp.Request) {
	var in departAddEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Add(r.Context(), operatorID(r), toDepartInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑部门。
// 契约：PUT /sys/sysDepart/edit；拒绝换父（「暂不支持调整层级结构」）；
// update_by=当前用户；成功话术「修改成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in departAddEditReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), operatorID(r), toDepartInput(in)); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "修改成功!", nil)
}

// Delete 删除部门（物理级联）。
// 契约：DELETE /sys/sysDepart/delete?id=必填；不存在 →「未找到对应实体」；
// 成功话术「删除成功!」。
func (c *Controller) Delete(r *ghttp.Request) {
	id := r.Get("id").String()
	if err := c.checkDepartExists(r, id); err != nil {
		c.writeError(r, err)
		return
	}
	if err := c.svc.Delete(r.Context(), id); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除部门（物理级联）。
// 契约：DELETE /sys/sysDepart/deleteBatch?ids=逗号分隔；空 →「参数不识别！」；
// 成功话术「删除成功!」。
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

// SearchBy 部门关键词搜索。
// 契约：GET /sys/sysDepart/searchBy?keyWord=必填&myDeptSearch=可选；
// 无匹配 → success=false +「未查询匹配数据！」（code 保持默认）。
func (c *Controller) SearchBy(r *ghttp.Request) {
	res, err := c.svc.SearchBy(r.Context(),
		r.Get("keyWord").String(), r.Get("myDeptSearch").String(), c.myDepartIDs(r))
	if err != nil {
		c.writeError(r, err)
		return
	}
	if len(res) == 0 {
		r.Response.WriteJson(g.Map{
			"success":   false,
			"message":   "未查询匹配数据！",
			"code":      0,
			"result":    nil,
			"timestamp": time.Now().UnixMilli(),
		})
		return
	}
	common.OK(r, res)
}

// ListAll 全量部门扁平列表。
// 契约：GET /sys/sysDepart/listAll?id=可选（逗号分隔）。
func (c *Controller) ListAll(r *ghttp.Request) {
	res, err := c.svc.ListAll(r.Context(), r.Get("id").String())
	if err != nil {
		c.writeError(r, err)
		return
	}
	common.OK(r, res)
}

// RemoveAll 清空班级成员。
// 契约：GET /sys/sysDepart/removeAll?id=必填；成功 Result.ok()（message=「成功」）。
func (c *Controller) RemoveAll(r *ghttp.Request) {
	if err := c.svc.RemoveAll(r.Context(), r.Get("id").String()); err != nil {
		c.writeError(r, err)
		return
	}
	common.OKMsg(r, "成功", nil)
}

// checkDepartExists 删除前存在性校验（对齐 Java delete 的 getById 分支）。
func (c *Controller) checkDepartExists(r *ghttp.Request, id string) error {
	return c.svc.CheckExists(r.Context(), id)
}

// myDepartIDs 当前用户的负责部门 id 串（上级身份才有值，对齐 Java searchBy）。
func (c *Controller) myDepartIDs(r *ghttp.Request) string {
	return c.svc.MyDepartIDs(r.Context(), operatorID(r))
}
