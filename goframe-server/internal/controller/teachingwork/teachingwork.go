// Package teachingwork 提供作业域的 HTTP 处理器
// （契约蓝本：TeachingWorkController 的 22 个核心端点）。
package teachingwork

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/controller/common"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/middleware"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
	svcTW "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingwork"
)

// Controller 作业域控制器。
type Controller struct {
	svc *svcTW.Service
}

// New 构造控制器。
func New(svc *svcTW.Service) *Controller {
	return &Controller{svc: svc}
}

// workReq submit/add/edit 请求体（TeachingWork/TeachingWorkPage 主表字段）。
type workReq struct {
	Id           string  `json:"id"`
	DepartId     *string `json:"departId"`
	CourseId     *string `json:"courseId"`
	AdditionalId *string `json:"additionalId"`
	WorkName     *string `json:"workName"`
	WorkType     *string `json:"workType"`
	WorkFile     *string `json:"workFile"`
	WorkCover    *string `json:"workCover"`
	WorkStatus   *int    `json:"workStatus"`
	WorkScene    *string `json:"workScene"`
	HasCloudData *bool   `json:"hasCloudData"`
	ViewNum      *int    `json:"viewNum"`
	StarNum      *int    `json:"starNum"`
	CollectNum   *int    `json:"collectNum"`
	DelFlag      *int    `json:"delFlag"`
}

// correctReq 批改子表输入。
type correctReq struct {
	Id      string  `json:"id"`
	Score   *int    `json:"score"`
	Comment *string `json:"comment"`
}

// workCommentReq 评论子表输入（/saveComment 只消费 workId+comment）。
type workCommentReq struct {
	Id      string  `json:"id"`
	WorkId  string  `json:"workId"`
	Comment *string `json:"comment"`
	UserId  string  `json:"userId"`
}

// workPageReq POST/PUT /add|edit 请求体（TeachingWorkPage）。
type workPageReq struct {
	Work                    workReq          `json:"work"`
	TeachingWorkCorrectList []correctReq     `json:"teachingWorkCorrectList"`
	TeachingWorkCommentList []workCommentReq `json:"teachingWorkCommentList"`
	// 兼容前端把主表字段平铺在顶层（BeanUtils 语义由键名决定）
}

// sendWorkReq POST /sendWork 请求体（StudentWorkSendVO）。
type sendWorkReq struct {
	SendWorkId string   `json:"sendWorkId"`
	UserIdList []string `json:"userIdList"`
}

// toWorkInput 请求体 → 服务层输入。
func toWorkInput(in workReq) svcTW.WorkInput {
	return svcTW.WorkInput{
		Id: in.Id, DepartId: in.DepartId, CourseId: in.CourseId,
		AdditionalId: in.AdditionalId, WorkName: in.WorkName, WorkType: in.WorkType,
		WorkFile: in.WorkFile, WorkCover: in.WorkCover, WorkStatus: in.WorkStatus,
		WorkScene: in.WorkScene, HasCloudData: in.HasCloudData, ViewNum: in.ViewNum,
		StarNum: in.StarNum, CollectNum: in.CollectNum, DelFlag: in.DelFlag,
	}
}

// toWorkPageReq 兼容平铺/嵌套两种请求体（前端实际把主表字段平铺顶层）。
func (c *Controller) parseWorkPage(r *ghttp.Request) svcTW.WorkPageInput {
	var flat workReq
	_ = r.Parse(&flat)
	var nested struct {
		Work                    workReq          `json:"work"`
		TeachingWorkCorrectList []correctReq     `json:"teachingWorkCorrectList"`
		TeachingWorkCommentList []workCommentReq `json:"teachingWorkCommentList"`
	}
	_ = r.Parse(&nested)
	in := nested
	if nested.Work.Id == "" && nested.Work.WorkName == nil {
		in.Work = flat
	}
	return svcTW.WorkPageInput{
		Work:                    toWorkInput(in.Work),
		TeachingWorkCorrectList: toCorrectInputs(in.TeachingWorkCorrectList),
		TeachingWorkCommentList: toCommentInputs(in.TeachingWorkCommentList),
	}
}

// toCorrectInputs 子表输入转换。
func toCorrectInputs(list []correctReq) []svcTW.CorrectInput {
	out := make([]svcTW.CorrectInput, 0, len(list))
	for _, c := range list {
		out = append(out, svcTW.CorrectInput{Id: c.Id, Score: c.Score, Comment: c.Comment})
	}
	return out
}

// toCommentInputs 子表输入转换。
func toCommentInputs(list []workCommentReq) []svcTW.CommentInput {
	out := make([]svcTW.CommentInput, 0, len(list))
	for _, c := range list {
		out = append(out, svcTW.CommentInput{Id: c.Id, Comment: c.Comment, UserId: c.UserId})
	}
	return out
}

// parsePage 解析分页参数（mine 默认 pageSize=999，其余 10）。
func parsePage(r *ghttp.Request, defaultSize int) page.Param {
	return page.Param{
		PageNo:   r.Get("pageNo", 1).Int(),
		PageSize: r.Get("pageSize", defaultSize).Int(),
	}
}

// operatorID 当前登录用户 ID。
func operatorID(r *ghttp.Request) string {
	return r.GetCtxVar(middleware.CtxKeyUserID).String()
}

// writeError 统一错误映射（话术与 Java 保持一致）。
func writeError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, svcTW.ErrUserNotFound),
		errors.Is(err, svcTW.ErrWorkNotFound),
		errors.Is(err, svcTW.ErrNotFound),
		errors.Is(err, svcTW.ErrNoDepartClass),
		errors.Is(err, svcTW.ErrTagInUse):
		common.Fail(r, err.Error())
	default:
		common.Fail(r, "操作失败")
	}
}

// GetUserInfo 作品作者信息卡（anon）。
// 契约：GET /teaching/teachingWork/userInfo?userId=。
func (c *Controller) GetUserInfo(r *ghttp.Request) {
	res, err := c.svc.UserInfo(r.Context(), r.Get("userId").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Mine 我的作业分页。
// 契约：GET /teaching/teachingWork/mine?pageNo=&pageSize=999&workName=...&workTag=。
func (c *Controller) Mine(r *ghttp.Request) {
	res, err := c.svc.Mine(r.Context(), operatorID(r), parsePage(r, 999),
		r.Get("workName").String(), r.Get("workType").String(),
		r.Get("workStatus").String(), r.Get("workScene").String(),
		r.Get("workTag").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetStudentWorkInfo 作品详情（anon，浏览数 +1）。
// 契约：GET /teaching/teachingWork/studentWorkInfo?workId=。
func (c *Controller) GetStudentWorkInfo(r *ghttp.Request) {
	res, err := c.svc.StudentWorkInfo(r.Context(), r.Get("workId").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GreatWorkList 精选作品分页（anon，固定 work_status=2）。
// 契约：GET /teaching/teachingWork/greatWork。
func (c *Controller) GreatWorkList(r *ghttp.Request) {
	res, err := c.svc.GreatWork(r.Context(), parsePage(r, 10))
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// Leaderboard 作品排行分页（anon，固定 work_status>=3）。
// 契约：GET /teaching/teachingWork/leaderboard?orderBy=view|time|star。
func (c *Controller) Leaderboard(r *ghttp.Request) {
	res, err := c.svc.Leaderboard(r.Context(), parsePage(r, 10),
		r.Get("orderBy", "view").String(), r.Get("workStatus").String(), r.Get("userId").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// StarWork 点赞作品（anon，Redis 24h 幂等）。
// 契约：GET /teaching/teachingWork/starWork?workId=。
func (c *Controller) StarWork(r *ghttp.Request) {
	msg, err := c.svc.StarWork(r.Context(), r.Get("workId").String(), r.GetClientIp())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, msg, nil)
}

// GetWorkComments 作品评论分页（anon）。
// 契约：GET /teaching/teachingWork/getWorkComments?workId=&page=&pageSize=。
func (c *Controller) GetWorkComments(r *ghttp.Request) {
	res, err := c.svc.GetWorkComments(r.Context(),
		r.Get("workId").String(), r.Get("page", 1).Int(), r.Get("pageSize", 10).Int())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// SaveComment 保存作品评论。
// 契约：POST /teaching/teachingWork/saveComment，body {workId, comment}；
// 成功话术「评论成功」。
func (c *Controller) SaveComment(r *ghttp.Request) {
	var in workCommentReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	if err := c.svc.SaveComment(r.Context(), operatorID(r), in.WorkId, derefComment(in.Comment)); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "评论成功", nil)
}

// derefComment *string → string。
func derefComment(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Submit 提交作业（四路匹配）。
// 契约：POST /teaching/teachingWork/submit；话术「更新成功！」/「添加成功！」。
func (c *Controller) Submit(r *ghttp.Request) {
	var in workReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	msg, err := c.svc.Submit(r.Context(), operatorID(r), toWorkInput(in))
	if err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, msg, nil)
}

// SendWork 发送作业给用户列表。
// 契约：POST /teaching/teachingWork/sendWork，body {sendWorkId, userIdList}；
// 话术「发送成功，共%d个用户」。
func (c *Controller) SendWork(r *ghttp.Request) {
	var in sendWorkReq
	if err := r.Parse(&in); err != nil {
		common.Fail(r, "请求参数不合法")
		return
	}
	count, err := c.svc.SendWork(r.Context(), in.SendWorkId, in.UserIdList)
	if err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, fmt.Sprintf("发送成功，共%d个用户", count), nil)
}

// MineAdditionalWork 我的附加作业。
// 契约：GET /teaching/teachingWork/mineAdditionalWork?departId=&submit=&status=。
func (c *Controller) MineAdditionalWork(r *ghttp.Request) {
	res, err := c.svc.MineAdditionalWork(r.Context(), operatorID(r),
		r.Get("submit").String(), r.Get("status").Int(), nil)
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// WorkList 作业管理分页（教师端）。
// 契约：GET /teaching/teachingWork/list?pageNo=&pageSize=&username=...
// 时间过滤参数名为 teaching_work.updateTime_begin/_end（实际比较 create_time）。
func (c *Controller) WorkList(r *ghttp.Request) {
	res, err := c.svc.List(r.Context(), operatorID(r), parsePage(r, 10),
		r.Get("username").String(), r.Get("workName").String(),
		r.Get("realname").String(), r.Get("workScene").String(),
		r.Get("workTag").String(),
		r.Get("teaching_work.updateTime_begin").String(),
		r.Get("teaching_work.updateTime_end").String(),
		r.Get("workStatus").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// WorkAdd 新增作业（主表+子表）。
// 契约：POST /teaching/teachingWork/add → Result.ok(teachingWork)。
func (c *Controller) WorkAdd(r *ghttp.Request) {
	in := c.parseWorkPage(r)
	res, err := c.svc.AddMain(r.Context(), in)
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// WorkEdit 编辑作业（子表先删后插 + 批改计数）。
// 契约：PUT /teaching/teachingWork/edit；不存在 →「未找到对应数据」；
// 成功话术「编辑成功!」。
func (c *Controller) WorkEdit(r *ghttp.Request) {
	in := c.parseWorkPage(r)
	if err := c.svc.EditMain(r.Context(), operatorID(r), in); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "编辑成功!", nil)
}

// WorkDelete 删除作业（子表+文件记录+物理删）。
// 契约：DELETE /teaching/teachingWork/delete?id=；话术「删除成功!」。
func (c *Controller) WorkDelete(r *ghttp.Request) {
	if err := c.svc.DeleteMain(r.Context(), r.Get("id").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "删除成功!", nil)
}

// WorkDeleteBatch 批量删除作业。
// 契约：DELETE /teaching/teachingWork/deleteBatch?ids=；话术「批量删除成功！」。
func (c *Controller) WorkDeleteBatch(r *ghttp.Request) {
	if err := c.svc.DeleteBatchMain(r.Context(), splitIDs(r.Get("ids").String())); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "批量删除成功！", nil)
}

// WorkQueryById 作业详情（单实体，不注入 dict）。
// 契约：GET /teaching/teachingWork/queryById?id=；不存在 →「未找到对应数据」。
func (c *Controller) WorkQueryById(r *ghttp.Request) {
	res, err := c.svc.QueryByID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// QueryCorrectsByMainID 批改列表。
// 契约：GET /teaching/teachingWork/queryTeachingWorkCorrectByMainId?id=。
func (c *Controller) QueryCorrectsByMainID(r *ghttp.Request) {
	res, err := c.svc.QueryCorrectsByMainID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// QueryCommentsByMainID 评论列表。
// 契约：GET /teaching/teachingWork/queryTeachingWorkCommentByMainId?id=。
func (c *Controller) QueryCommentsByMainID(r *ghttp.Request) {
	res, err := c.svc.QueryCommentsByMainID(r.Context(), r.Get("id").String())
	if err != nil {
		writeError(r, err)
		return
	}
	common.OK(r, res)
}

// GetWorkTags 当前用户标签集合。
// 契约：GET /teaching/teachingWork/getWorkTags。
func (c *Controller) GetWorkTags(r *ghttp.Request) {
	common.OK(r, c.svc.GetWorkTags(r.Context(), operatorID(r)))
}

// SetWorkTag 设置作品标签。
// 契约：GET /teaching/teachingWork/setWorkTag?workId=&workTag=；话术「标记成功」。
func (c *Controller) SetWorkTag(r *ghttp.Request) {
	if err := c.svc.SetWorkTag(r.Context(), operatorID(r),
		r.Get("workId").String(), r.Get("workTag").String()); err != nil {
		writeError(r, err)
		return
	}
	common.OKMsg(r, "标记成功", nil)
}

// DelWorkTag 删除标签。
// 契约：DELETE /teaching/teachingWork/delWorkTag?tag=&force=。
func (c *Controller) DelWorkTag(r *ghttp.Request) {
	if err := c.svc.DelWorkTag(r.Context(), operatorID(r),
		r.Get("tag").String(), r.Get("force", false).Bool()); err != nil {
		writeError(r, err)
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
