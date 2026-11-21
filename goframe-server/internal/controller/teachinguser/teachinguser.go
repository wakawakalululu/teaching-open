// Package teachinguser 提供门户用户上下文域的 HTTP 处理器
// （契约蓝本：TeachingUserController 的 3 个端点）。
package teachinguser

import (
	"errors"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	svcTU "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachinguser"
)

// Controller 门户用户上下文控制器。
type Controller struct {
	svc *svcTU.Service
}

// New 构造控制器。
func New(svc *svcTU.Service) *Controller {
	return &Controller{svc: svc}
}

// editReq PUT /teaching/user/edit 请求体（JSON 全字段透传更新）。
type editReq struct {
	Id           string  `json:"id"`
	Username     *string `json:"username"`
	Realname     *string `json:"realname"`
	Avatar       *string `json:"avatar"`
	Birthday     *string `json:"birthday"`
	Sex          *int    `json:"sex"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	OrgCode      *string `json:"orgCode"`
	Status       *int    `json:"status"`
	WorkNo       *string `json:"workNo"`
	Post         *string `json:"post"`
	Telephone    *string `json:"telephone"`
	School       *string `json:"school"`
	UserIdentity *int    `json:"userIdentity"`
	ActivitiSync *int    `json:"activitiSync"`
}

// operatorID 当前登录用户 ID。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// Info 当前用户完整实体。
// 契约：ANY /teaching/user/info（登录态）。
func (c *Controller) Info(r *ghttp.Request) {
	res, err := c.svc.Info(r.Context(), operatorID(r))
	if err != nil {
		common.Fail(r, "操作失败")
		return
	}
	common.OK(r, res)
}

// UserRootDept 用户根部门（org_code 前 3 位公司级）。
// 契约：ANY /teaching/user/getUserRootDept?userId=可选。
func (c *Controller) UserRootDept(r *ghttp.Request) {
	res, err := c.svc.UserRootDept(r.Context(), operatorID(r), r.Get("userId").String())
	if err != nil {
		switch {
		case errors.Is(err, svcTU.ErrNoOrgCode), errors.Is(err, svcTU.ErrNotFound):
			common.Fail(r, err.Error())
		default:
			common.Fail(r, "操作失败")
		}
		return
	}
	common.OK(r, res)
}

// Edit 修改用户资料（门户自助）。
// 契约：PUT /teaching/user/edit；密码保留旧值；成功话术「修改成功!」。
func (c *Controller) Edit(r *ghttp.Request) {
	var in editReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.Edit(r.Context(), svcTU.EditInput{
		Id: in.Id, Username: in.Username, Realname: in.Realname, Avatar: in.Avatar,
		Birthday: in.Birthday, Sex: in.Sex, Email: in.Email, Phone: in.Phone,
		OrgCode: in.OrgCode, Status: in.Status, WorkNo: in.WorkNo, Post: in.Post,
		Telephone: in.Telephone, School: in.School, UserIdentity: in.UserIdentity,
		ActivitiSync: in.ActivitiSync,
	}); err != nil {
		switch {
		case errors.Is(err, svcTU.ErrNotFound):
			common.Fail(r, err.Error())
		default:
			common.Fail(r, "操作失败")
		}
		return
	}
	common.OKMsg(r, "修改成功!", nil)
}
