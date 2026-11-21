// Package announcement 承载系统通告域业务用例
// （契约蓝本：SysAnnouncementController 10 个核心端点 + SysAnnouncementSendController
// 9 个端点；exportXls/importExcel/syncNotic websocket 广播有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - delete 为逻辑删（del_flag=1），list 恒过滤 del_flag=0；
//   - add：sendStatus=0 未发布；msgType=USER 时同步插入 sys_announcement_send
//     阅读标记行（userIds 尾逗号裁剪后拆分）；
//   - upDateAnnouncement：USER 类型补插缺失用户标记行 + 删除不在 userIds
//     清单内的多余行（NOT IN 语义）；
//   - doReleaseData：sendStatus=1 + sendTime=now + sender=当前用户；
//     doReovkeData：sendStatus=2 + cancelTime=now（websocket 广播未迁移）；
//   - listByUser：ALL 类型已发布通告补插用户标记行（endTime>=用户注册时间
//     过滤——新注册用户不看结束通知），再按 msgCategory 查未读（read_flag=0）
//     前 5 条 + 总数；
//   - getMyAnnouncementSend：sas LEFT JOIN sa，send_status='1' + del_flag='0'
//   - titile/sender LIKE + readFlag/busType 精确，LIMIT offset,pageSize。
package announcement

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// 通告状态/分类常量（对齐 CommonConstant/CommonSendStatus）。
const (
	msgTypeAll            = "ALL"
	msgTypeUser           = "USER"
	sendStatusUnpublished = "0" // 未发布
	sendStatusPublished   = "1" // 发布中（HAS_SEND）
	sendStatusRevoked     = "2" // 撤销发布
	noReadFlag            = "0"
	hasReadFlag           = "1"
)

// ErrNotFound 通告不存在（Java error500 原话术）。
var ErrNotFound = errors.New("未找到对应实体")

// AnnouncementInput add/edit 请求体（SysAnnouncement 可编辑字段）。
type AnnouncementInput struct {
	Id          string
	Titile      *string
	MsgContent  *string
	StartTime   *string
	EndTime     *string
	Priority    *string
	MsgCategory *string
	MsgType     *string
	SendStatus  *string
	BusType     *string
	BusId       *string
	OpenType    *string
	OpenPage    *string
	MsgAbstract *string
	UserIds     *string // 逗号分隔（USER 类型专用，尾逗号由服务端裁剪）
}

// SendInput send add/edit 请求体（SysAnnouncementSend 可编辑字段）。
type SendInput struct {
	Id       string
	AnntId   *string
	UserId   *string
	ReadFlag *string
}

// sendModel getMyAnnouncementSend 输出行（AnnouncementSendModel 消费面）。
type sendModel struct {
	Id          string `json:"id"`
	AnntId      string `json:"anntId"`
	UserId      string `json:"userId"`
	ReadFlag    string `json:"readFlag"`
	Titile      string `json:"titile"`
	MsgContent  string `json:"msgContent"`
	Sender      string `json:"sender"`
	Priority    string `json:"priority"`
	MsgCategory string `json:"msgCategory"`
	SendTime    string `json:"sendTime"`
	BusId       string `json:"busId"`
	OpenType    string `json:"openType"`
	OpenPage    string `json:"openPage"`
	MsgAbstract string `json:"msgAbstract"`
}

// Service 系统通告业务服务。
type Service struct{}

// New 构造服务。
func New() *Service {
	return &Service{}
}

// List 通告分页（del_flag=0 + 实体字段 eq 过滤 + column/order 排序）。
//
// 契约蓝本 SysAnnouncementController#list：QueryWrapper(entity) 为字段精确
// 等值（非 QueryGenerator 通配规则）；排序 camelToUnderline 后无白名单
// （Java 原样拼接，Go 收敛为列集合校验防注入）。
func (s *Service) List(ctx context.Context, param page.Param, titile, msgCategory, msgType, sender, orderColumn, orderType string) (*page.Result, error) {
	cols := dao.SysAnnouncement.Columns()
	m := dao.SysAnnouncement.Ctx(ctx).Where(cols.DelFlag, "0")
	if titile != "" {
		m = m.Where(cols.Titile, titile)
	}
	if msgCategory != "" {
		m = m.Where(cols.MsgCategory, msgCategory)
	}
	if msgType != "" {
		m = m.Where(cols.MsgType, msgType)
	}
	if sender != "" {
		m = m.Where(cols.Sender, sender)
	}
	m = applyOrder(m, orderColumn, orderType, cols)
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计通告数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysAnnouncement{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysAnnouncement
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询通告列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增通告。
//
// 契约蓝本 SysAnnouncementController#add + saveAnnouncement：delFlag=0、
// sendStatus=0 未发布；msgType=USER 时按 userIds（尾逗号裁剪）同步插入
// sys_announcement_send 标记行（read_flag=0）；单事务。
func (s *Service) Add(ctx context.Context, in AnnouncementInput) error {
	data := do.SysAnnouncement{Id: in.Id, DelFlag: "0", SendStatus: sendStatusUnpublished}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyAnnouncementInput(&data, in)
	return dao.SysAnnouncement.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.SysAnnouncement.Ctx(ctx).Data(data).Insert(); err != nil {
			return fmt.Errorf("操作失败: %w", err)
		}
		if derefStr(in.MsgType) == msgTypeUser {
			insertSendRows(ctx, data.Id.(string), derefStr(in.UserIds))
		}
		return nil
	})
}

// Edit 编辑通告（存在性校验；USER 类型同步标记行：补插缺失 + 删除多余）。
func (s *Service) Edit(ctx context.Context, in AnnouncementInput) error {
	cols := dao.SysAnnouncement.Columns()
	var old entity.SysAnnouncement
	if err := dao.SysAnnouncement.Ctx(ctx).Where(cols.Id, in.Id).Scan(&old); err != nil {
		return fmt.Errorf("查询通告失败: %w", err)
	}
	if old.Id == "" {
		return ErrNotFound
	}
	data := do.SysAnnouncement{}
	applyAnnouncementInput(&data, in)
	return dao.SysAnnouncement.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.SysAnnouncement.Ctx(ctx).
			Where(cols.Id, in.Id).Data(data).Update(); err != nil {
			return fmt.Errorf("操作失败: %w", err)
		}
		userIds := derefStr(in.UserIds)
		msgType := derefStr(in.MsgType)
		if msgType == "" {
			msgType = old.MsgType
		}
		if userIds == "" || msgType != msgTypeUser {
			return nil
		}
		trimmed := strings.TrimSuffix(userIds, ",")
		ids := strings.Split(trimmed, ",")
		sc := dao.SysAnnouncementSend.Columns()
		for _, uid := range ids {
			if uid == "" {
				continue
			}
			cnt, err := dao.SysAnnouncementSend.Ctx(ctx).
				Where(sc.AnntId, in.Id).Where(sc.UserId, uid).Count()
			if err != nil {
				return fmt.Errorf("查询标记行失败: %w", err)
			}
			if cnt > 0 {
				continue
			}
			if _, err := dao.SysAnnouncementSend.Ctx(ctx).Data(do.SysAnnouncementSend{
				Id: guid.S(), AnntId: in.Id, UserId: uid,
				ReadFlag: noReadFlag, ReadTime: gtime.Now(),
			}).Insert(); err != nil {
				return fmt.Errorf("插入标记行失败: %w", err)
			}
		}
		// 删除不在新清单内的多余标记行（对齐 Java NOT IN 分支）
		if len(ids) > 0 {
			_, err := dao.SysAnnouncementSend.Ctx(ctx).
				Where(sc.AnntId, in.Id).
				WhereNotIn(sc.UserId, ids).Delete()
			if err != nil {
				return fmt.Errorf("清理多余标记行失败: %w", err)
			}
		}
		return nil
	})
}

// Delete 逻辑删（del_flag=1）。
func (s *Service) Delete(ctx context.Context, id string) error {
	cols := dao.SysAnnouncement.Columns()
	var row entity.SysAnnouncement
	if err := dao.SysAnnouncement.Ctx(ctx).Where(cols.Id, id).Scan(&row); err != nil {
		return fmt.Errorf("查询通告失败: %w", err)
	}
	if row.Id == "" {
		return ErrNotFound
	}
	if _, err := dao.SysAnnouncement.Ctx(ctx).Where(cols.Id, id).
		Data(do.SysAnnouncement{DelFlag: "1"}).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量逻辑删。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// QueryByID 通告详情。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.SysAnnouncement, error) {
	var row entity.SysAnnouncement
	if err := dao.SysAnnouncement.Ctx(ctx).
		Where(dao.SysAnnouncement.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询通告失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// DoReleaseData 发布通告（sendStatus=1 + sendTime + sender=当前用户）。
func (s *Service) DoReleaseData(ctx context.Context, id, operatorName string) error {
	var row entity.SysAnnouncement
	if err := dao.SysAnnouncement.Ctx(ctx).
		Where(dao.SysAnnouncement.Columns().Id, id).Scan(&row); err != nil {
		return fmt.Errorf("查询通告失败: %w", err)
	}
	if row.Id == "" {
		return ErrNotFound
	}
	if _, err := dao.SysAnnouncement.Ctx(ctx).Where(dao.SysAnnouncement.Columns().Id, id).
		Data(do.SysAnnouncement{SendStatus: sendStatusPublished, SendTime: gtime.Now(),
			Sender: operatorName}).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	// msgType=ALL 时全员广播（对齐 Java webSocket.sendAllMessage；USER 类型
	// 广播对象为标记行用户——Go 简化为仅 ALL 广播，文档化）
	if broadcastTopic != nil && row.MsgType == msgTypeAll {
		broadcastTopic(id, row.Titile)
	}
	return nil
}

// DoReovkeData 撤销发布（sendStatus=2 + cancelTime）。
func (s *Service) DoReovkeData(ctx context.Context, id string) error {
	var row entity.SysAnnouncement
	if err := dao.SysAnnouncement.Ctx(ctx).
		Where(dao.SysAnnouncement.Columns().Id, id).Scan(&row); err != nil {
		return fmt.Errorf("查询通告失败: %w", err)
	}
	if row.Id == "" {
		return ErrNotFound
	}
	if _, err := dao.SysAnnouncement.Ctx(ctx).Where(dao.SysAnnouncement.Columns().Id, id).
		Data(do.SysAnnouncement{SendStatus: sendStatusRevoked, CancelTime: gtime.Now()}).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// ListByUser 我的消息面板。
//
// 契约蓝本 SysAnnouncementController#listByUser：
//  1. ALL 类型已发布通告（endTime>=用户注册时间过滤——新注册用户不看结束
//     通知）且用户无标记行的，补插 read_flag=0 标记行；
//  2. 按未读标记（read_flag='0'）查 msgCategory=1（通知公告）/2（系统消息）
//     前 5 条 + 总数。
func (s *Service) ListByUser(ctx context.Context, operatorID string) (map[string]interface{}, error) {
	user, err := getUser(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	// 1. 已有标记行的通告 id
	anntIDs, err := g.DB().Model("sys_announcement_send").Fields("annt_id").
		Where("user_id", operatorID).Array()
	if err != nil {
		return nil, fmt.Errorf("查询标记行失败: %w", err)
	}
	existing := make([]string, 0, len(anntIDs))
	for _, id := range anntIDs {
		existing = append(existing, id.String())
	}
	// 2. 补插 ALL 类型已发布且未标记的通告行
	ac := dao.SysAnnouncement.Columns()
	m := dao.SysAnnouncement.Ctx(ctx).
		Where(ac.MsgType, msgTypeAll).
		Where(ac.DelFlag, "0").
		Where(ac.SendStatus, sendStatusPublished).
		Where(ac.EndTime+" >= ?", user.CreateTime)
	if len(existing) > 0 {
		m = m.WhereNotIn(ac.Id, existing)
	}
	var missing []entity.SysAnnouncement
	if err := m.Scan(&missing); err != nil {
		return nil, fmt.Errorf("查询未标记通告失败: %w", err)
	}
	for _, a := range missing {
		if _, err := dao.SysAnnouncementSend.Ctx(ctx).Data(do.SysAnnouncementSend{
			Id: guid.S(), AnntId: a.Id, UserId: operatorID, ReadFlag: noReadFlag,
		}).Insert(); err != nil {
			return nil, fmt.Errorf("插入标记行失败: %w", err)
		}
	}
	// 3. 未读消息按分类查前 5 条 + 总数（querySysCementListByUserId SQL）
	const cementQuery = `SELECT * FROM sys_announcement
WHERE send_status = '1' AND del_flag = '0' AND msg_category = ?
 AND id IN (SELECT annt_id FROM sys_announcement_send WHERE user_id = ? AND read_flag = '0')
ORDER BY create_time DESC`
	anntList, anntTotal, sysList, sysTotal, err := queryCement(ctx, cementQuery, operatorID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"sysMsgList":   sysList,
		"sysMsgTotal":  sysTotal,
		"anntMsgList":  anntList,
		"anntMsgTotal": anntTotal,
	}, nil
}

// queryCement 按分类查未读通告前 5 条与总数。
func queryCement(ctx context.Context, query, operatorID string) ([]entity.SysAnnouncement, int, []entity.SysAnnouncement, int, error) {
	var anntList, sysList []entity.SysAnnouncement
	if err := g.DB().GetScan(ctx, &anntList, query+" LIMIT 5", "1", operatorID); err != nil {
		return nil, 0, nil, 0, fmt.Errorf("查询通知公告失败: %w", err)
	}
	if err := g.DB().GetScan(ctx, &sysList, query+" LIMIT 5", "2", operatorID); err != nil {
		return nil, 0, nil, 0, fmt.Errorf("查询系统消息失败: %w", err)
	}
	anntTotal, err := g.DB().GetOne(ctx,
		"SELECT COUNT(*) AS n FROM sys_announcement WHERE send_status = '1' AND del_flag = '0' "+
			"AND msg_category = ? AND id IN (SELECT annt_id FROM sys_announcement_send "+
			"WHERE user_id = ? AND read_flag = '0')", "1", operatorID)
	if err != nil {
		return nil, 0, nil, 0, fmt.Errorf("统计通知公告失败: %w", err)
	}
	sysTotal, err := g.DB().GetOne(ctx,
		"SELECT COUNT(*) AS n FROM sys_announcement WHERE send_status = '1' AND del_flag = '0' "+
			"AND msg_category = ? AND id IN (SELECT annt_id FROM sys_announcement_send "+
			"WHERE user_id = ? AND read_flag = '0')", "2", operatorID)
	if err != nil {
		return nil, 0, nil, 0, fmt.Errorf("统计系统消息失败: %w", err)
	}
	return anntList, anntTotal["n"].Int(), sysList, sysTotal["n"].Int(), nil
}

// GetMyAnnouncementSend 我的消息分页（JOIN 查询 + 过滤面）。
func (s *Service) GetMyAnnouncementSend(ctx context.Context, operatorID string, param page.Param,
	titile, sender, readFlag, busType string) (*page.Result, error) {
	where := "sa.send_status = '1' AND sa.del_flag = '0' AND sas.user_id = ?"
	args := []interface{}{operatorID}
	if titile != "" {
		where += " AND sa.titile LIKE ?"
		args = append(args, "%"+titile+"%")
	}
	if sender != "" {
		where += " AND sa.sender LIKE ?"
		args = append(args, "%"+sender+"%")
	}
	if readFlag != "" {
		where += " AND sas.read_flag = ?"
		args = append(args, readFlag)
	}
	if busType != "" {
		where += " AND sa.bus_type = ?"
		args = append(args, busType)
	}
	const base = `FROM sys_announcement_send sas
LEFT JOIN sys_announcement sa ON sas.annt_id = sa.id`
	total, err := g.DB().GetOne(ctx, "SELECT COUNT(*) AS n "+base+" WHERE "+where, args...)
	if err != nil {
		return nil, fmt.Errorf("统计我的消息失败: %w", err)
	}
	if total.IsEmpty() || total["n"].Int() == 0 {
		return page.New([]sendModel{}, 0, param.PageNo, param.PageSize), nil
	}
	totalN := total["n"].Int()
	offset := (param.PageNo - 1) * param.PageSize
	if offset < 0 {
		offset = 0
	}
	query := `SELECT sas.id, sas.annt_id, sas.user_id, sas.read_flag,
 sa.titile, sa.msg_content, sa.sender, sa.priority, sa.msg_category,
 sa.send_time, sa.bus_id, sa.open_type, sa.open_page, sa.msg_abstract ` +
		base + " WHERE " + where + " LIMIT ? OFFSET ?"
	args = append(args, param.PageSize, offset)
	var rows []sendModel
	if err := g.DB().GetScan(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("查询我的消息失败: %w", err)
	}
	return page.New(rows, totalN, param.PageNo, param.PageSize), nil
}

// insertSendRows 按 userIds（尾逗号裁剪）插入标记行。
func insertSendRows(ctx context.Context, anntID, userIds string) {
	trimmed := strings.TrimSuffix(userIds, ",")
	for _, uid := range strings.Split(trimmed, ",") {
		if uid == "" {
			continue
		}
		_, _ = dao.SysAnnouncementSend.Ctx(ctx).Data(do.SysAnnouncementSend{
			Id: guid.S(), AnntId: anntID, UserId: uid,
			ReadFlag: noReadFlag, ReadTime: gtime.Now(),
		}).Insert()
	}
}

// applyAnnouncementInput 请求体字段拷贝进 DO（UserIdes 为非表列不入 DO；
// Id 由调用方管理）。
func applyAnnouncementInput(data *do.SysAnnouncement, in AnnouncementInput) {
	if in.Titile != nil {
		data.Titile = *in.Titile
	}
	if in.MsgContent != nil {
		data.MsgContent = *in.MsgContent
	}
	if in.StartTime != nil && *in.StartTime != "" {
		data.StartTime = *in.StartTime
	}
	if in.EndTime != nil && *in.EndTime != "" {
		data.EndTime = *in.EndTime
	}
	if in.Priority != nil {
		data.Priority = *in.Priority
	}
	if in.MsgCategory != nil {
		data.MsgCategory = *in.MsgCategory
	}
	if in.MsgType != nil {
		data.MsgType = *in.MsgType
	}
	if in.SendStatus != nil {
		data.SendStatus = *in.SendStatus
	}
	if in.BusType != nil {
		data.BusType = *in.BusType
	}
	if in.BusId != nil {
		data.BusId = *in.BusId
	}
	if in.OpenType != nil {
		data.OpenType = *in.OpenType
	}
	if in.OpenPage != nil {
		data.OpenPage = *in.OpenPage
	}
	if in.MsgAbstract != nil {
		data.MsgAbstract = *in.MsgAbstract
	}
}

// applyOrder 排序参数应用（camelToUnderline + 列集合校验防注入）。
func applyOrder(m *gdb.Model, column, orderType string, cols interface{}) *gdb.Model {
	column = strings.TrimSpace(column)
	if column == "" || strings.TrimSpace(orderType) == "" {
		return m
	}
	col := camelToUnderline(column)
	allowed := map[string]struct{}{
		"titile": {}, "send_status": {}, "create_time": {}, "priority": {},
		"msg_category": {}, "msg_type": {}, "sender": {}, "id": {},
	}
	if _, ok := allowed[col]; !ok {
		return m
	}
	if strings.Contains(strings.ToUpper(orderType), "ASC") {
		return m.OrderAsc(col)
	}
	return m.OrderDesc(col)
}

// camelToUnderline 驼峰转下划线。
func camelToUnderline(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// derefStr *string → string。
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// getUser 加载用户实体（listByUser 需要 createTime）。
func getUser(ctx context.Context, operatorID string) (*entity.SysUser, error) {
	var u entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Where(dao.SysUser.Columns().Id, operatorID).Scan(&u); err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &u, nil
}

// SendList 通告发送记录分页（QueryWrapper(entity) 字段等值 + 排序）。
func (s *Service) SendList(ctx context.Context, param page.Param, anntID, userID, readFlag string) (*page.Result, error) {
	sc := dao.SysAnnouncementSend.Columns()
	m := dao.SysAnnouncementSend.Ctx(ctx)
	if anntID != "" {
		m = m.Where(sc.AnntId, anntID)
	}
	if userID != "" {
		m = m.Where(sc.UserId, userID)
	}
	if readFlag != "" {
		m = m.Where(sc.ReadFlag, readFlag)
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计发送记录数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.SysAnnouncementSend{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.SysAnnouncementSend
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询发送记录失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// SendAdd 新增发送记录（纯 save，id 为空时服务端生成）。
func (s *Service) SendAdd(ctx context.Context, in SendInput) error {
	data := do.SysAnnouncementSend{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	if in.AnntId != nil {
		data.AnntId = *in.AnntId
	}
	if in.UserId != nil {
		data.UserId = *in.UserId
	}
	if in.ReadFlag != nil {
		data.ReadFlag = *in.ReadFlag
	}
	if _, err := dao.SysAnnouncementSend.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// SendEdit 编辑发送记录（纯 updateById）。
func (s *Service) SendEdit(ctx context.Context, in SendInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.SysAnnouncementSend{}
	if in.AnntId != nil {
		data.AnntId = *in.AnntId
	}
	if in.UserId != nil {
		data.UserId = *in.UserId
	}
	if in.ReadFlag != nil {
		data.ReadFlag = *in.ReadFlag
	}
	sc := dao.SysAnnouncementSend.Columns()
	if _, err := dao.SysAnnouncementSend.Ctx(ctx).
		Where(sc.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// SendDelete 删除发送记录。
func (s *Service) SendDelete(ctx context.Context, id string) error {
	sc := dao.SysAnnouncementSend.Columns()
	if _, err := dao.SysAnnouncementSend.Ctx(ctx).Where(sc.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除发送记录失败: %w", err)
	}
	return nil
}

// SendDeleteBatch 批量删除发送记录。
func (s *Service) SendDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	sc := dao.SysAnnouncementSend.Columns()
	if _, err := dao.SysAnnouncementSend.Ctx(ctx).WhereIn(sc.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除发送记录失败: %w", err)
	}
	return nil
}

// SendQueryByID 发送记录详情。
func (s *Service) SendQueryByID(ctx context.Context, id string) (*entity.SysAnnouncementSend, error) {
	var row entity.SysAnnouncementSend
	if err := dao.SysAnnouncementSend.Ctx(ctx).
		Where(dao.SysAnnouncementSend.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询发送记录失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// EditByAnntIdAndUserId 标记单条通告已读（anntId+当前用户）。
func (s *Service) EditByAnntIdAndUserID(ctx context.Context, operatorID, anntID string) error {
	sc := dao.SysAnnouncementSend.Columns()
	_, err := dao.SysAnnouncementSend.Ctx(ctx).
		Where(sc.AnntId, anntID).Where(sc.UserId, operatorID).
		Data(do.SysAnnouncementSend{ReadFlag: hasReadFlag, ReadTime: gtime.Now()}).Update()
	if err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}
	return nil
}

// ReadAll 一键已读（user_id 全部标记）。
func (s *Service) ReadAll(ctx context.Context, operatorID string) error {
	sc := dao.SysAnnouncementSend.Columns()
	_, err := dao.SysAnnouncementSend.Ctx(ctx).
		Where(sc.UserId, operatorID).
		Data(do.SysAnnouncementSend{ReadFlag: hasReadFlag, ReadTime: gtime.Now()}).Update()
	if err != nil {
		return fmt.Errorf("全部已读失败: %w", err)
	}
	return nil
}
