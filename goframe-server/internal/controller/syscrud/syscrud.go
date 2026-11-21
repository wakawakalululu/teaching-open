// Package syscrud 提供批量标准 CRUD 域的泛型 HTTP 处理器
// （契约蓝本：各域 JeecgController 标准六件套 + 各域特殊端点）。
//
// 每个域的消息话术与 Java 一致：添加成功！/编辑成功!/删除成功!/批量删除成功！。
// list 成功响应为 Result.ok 形状（code=200，与 Java Result.ok(pageList) 一致）。
package syscrud

import (
	"context"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcSC "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/syscrud"
)

// Domain 单个 CRUD 域的处理器集合。
type Domain struct {
	svc      *svcSC.Engine
	notFound string // queryById 未找到话术
}

// NewDomain 构造域处理器。
func NewDomain(def svcSC.TableDef) *Domain {
	return &Domain{svc: svcSC.New(def), notFound: "未找到对应数据"}
}

// parseBody 解析请求体为列 map（白名单由引擎过滤）。
func parseBody(r *ghttp.Request) map[string]interface{} {
	var body map[string]interface{}
	if err := r.Parse(&body); err != nil {
		return map[string]interface{}{}
	}
	return body
}

// parsePage 解析分页参数（默认 1/10）。
func parsePage(r *ghttp.Request) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", 10).Int(),
	}
}

// List 分页列表（body 字段等值过滤由调用方经 query 传入）。
func (d *Domain) List(r *ghttp.Request, likeCol string) {
	filters := map[string]interface{}{
		"rule_name": r.Get("ruleName").String(), "rule_code": r.Get("ruleCode").String(),
		"code": r.Get("code").String(), "name": r.Get("name").String(),
		"log_type": r.Get("logType").Int(), "username": r.Get("username").String(),
		"depart_id": r.Get("departId").String(), "status": r.Get("status").String(),
		"msg_type": r.Get("msgType").String(), "data_table": r.Get("dataTable").String(),
		"data_id": r.Get("dataId").String(),
	}
	likeVal := ""
	switch likeCol {
	case "log_content":
		likeVal = r.Get("keyWord").String()
	case "work_name":
		likeVal = r.Get("workName").String()
	default:
		if likeCol != "" {
			likeVal = r.Get(likeCol).String()
		}
	}
	res, err := d.svc.List(r.Context(), parsePage(r), filters, likeCol, likeVal)
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}

// Add 新增（话术「添加成功！」）。
func (d *Domain) Add(r *ghttp.Request) {
	if _, err := d.svc.Add(r.Context(), parseBody(r)); err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OKMsg(r, "添加成功！", nil)
}

// Edit 编辑（话术「编辑成功!」）。
func (d *Domain) Edit(r *ghttp.Request) {
	var body map[string]interface{}
	_ = r.Parse(&body)
	id := ""
	if v, ok := body["id"].(string); ok {
		id = v
	}
	ok, err := d.svc.Edit(r.Context(), id, body)
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	// Java edit 对 updateById=false 不改写响应（仍 success=true）——逐语义保留
	_ = ok
	common.OKMsg(r, "编辑成功!", nil)
}

// Delete 删除（话术「删除成功!」）。
func (d *Domain) Delete(r *ghttp.Request) {
	if err := d.svc.Delete(r.Context(), r.Get("id").String()); err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// DeleteBatch 批量删除（话术「批量删除成功!」）。
func (d *Domain) DeleteBatch(r *ghttp.Request) {
	ids := r.Get("ids").String()
	if ids == "" || stringsAllSpace(ids) {
		common.Fail(r, "参数不识别！")
		return
	}
	if err := d.svc.DeleteBatch(r.Context(), svcSC.SplitIDs(ids)); err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OKMsg(r, "批量删除成功!", nil)
}

// QueryByID 详情（不存在 →「未找到对应数据」）。
func (d *Domain) QueryByID(r *ghttp.Request) {
	res, err := d.svc.QueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	if res == nil {
		common.Fail(r, d.notFound)
		return
	}
	common.OK(r, res)
}

// DeleteAll 全清（log 域 allclear）。
func (d *Domain) DeleteAll(r *ghttp.Request) {
	if err := d.svc.DeleteAll(r.Context()); err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OKMsg(r, "清除成功!", nil)
}

// stringsAllSpace 全空白判断。
func stringsAllSpace(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' {
			return false
		}
	}
	return true
}

// ListAllOptionsEngine 数据源 options（裸数组 [{value,label,text}]）——供
// register_p2_6 直接持有引擎调用。
func (d *Domain) ListAllOptionsEngine(ctx context.Context, name string) ([]map[string]string, error) {
	res, err := d.svc.ListAll(ctx, map[string]interface{}{"name": name})
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, 0, len(res))
	for _, row := range res {
		code, _ := row["code"].(string)
		name2, _ := row["name"].(string)
		out = append(out, map[string]string{"value": code, "label": name2, "text": name2})
	}
	return out, nil
}

// QueryByUserName 按 user_name 精确查询（SysUserAgentController#queryByUserName）。
func (d *Domain) QueryByUserName(r *ghttp.Request) {
	res, err := d.svc.QueryByUserName(r.Context(), r.Get("userName").String())
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	if res == nil {
		common.Fail(r, d.notFound)
		return
	}
	common.OK(r, res)
}
