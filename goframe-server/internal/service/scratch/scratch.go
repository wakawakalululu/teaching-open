// Package scratch 承载 Scratch 素材库与背包域业务用例
// （契约蓝本：TeachingScratchAssetsController 7 个核心端点 +
// TeachingScratchBackpackController 4 个端点；exportXls/importExcel 有意未迁移）。
//
// 关键 Java 语义（2026-11-21 契约分析，出处见各方法注释）：
//   - getScratchAssets 返回**裸 JSONArray**（非 Result 包装）：逐条解析
//     asset_data JSON（素材元数据由服务端存储的 JSON 还原）；
//   - delete：assetType==4（角色）的物理文件清理为 TODO（Java 现状跳过），
//     其余类型清理 sys_file 记录（md5Ext 精确匹配）；
//   - 背包：Java 存七牛（Go 无七牛上传后端 → 落本地存储并按 uploadType 拼
//     URL，文档化偏差）；列表/删除走 Redis list（键 scratch:backpackList:{uid}
//     与 Java 一致；条目存 JSON 串，Java FastJson 序列化的旧条目解析失败
//     时跳过——灰度期已知偏差）；getFile 读取 scratch:backpackBody:{filename}
//     （save 不写该键，恒空响应——Java 现状保留）。
package scratch

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/wakawakalululu/teaching-open/goframe-server/internal/dao"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/do"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/entity"
	"github.com/wakawakalululu/teaching-open/goframe-server/internal/model/page"
)

// Redis 键（与 Java CacheConstant 逐字符一致）。
const (
	backpackListKeyFmt = "scratch:backpackList:" // + uid（list）
	backpackBodyKeyFmt = "scratch:backpackBody:" // + filename（string，恒空的遗留键）
)

// ErrNotFound 素材不存在（Java queryById error 原话术）。
var ErrNotFound = errors.New("未找到对应数据")

// AssetInput add/edit 请求体（TeachingScratchAssets 可编辑字段）。
type AssetInput struct {
	Id        string
	AssetType *int
	AssetName *string
	AssetData *string
	Md5Ext    *string
	Tags      *string
	DelFlag   *int
}

// BackpackModel 背包条目（对齐 Java ScratchBackpackModel JSON 字段）。
type BackpackModel struct {
	Body      string `json:"body"`
	Id        string `json:"id"`
	Mime      string `json:"mime"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
	Type      string `json:"type"`
}

// Service Scratch 素材库与背包业务服务。
type Service struct {
	kv           KVListStore
	staticDomain string
	qiniuDomain  string
	uploadType   string
	localRoot    string
}

// New 构造服务。
func New(redisEnabled bool, staticDomain, qiniuDomain, uploadType, localRoot string) *Service {
	return &Service{
		kv:           NewKVListStore(redisEnabled),
		staticDomain: staticDomain,
		qiniuDomain:  qiniuDomain,
		uploadType:   uploadType,
		localRoot:    localRoot,
	}
}

// fileURL 文件展示地址（与既有域同款语义）。
func (s *Service) fileURL(key string) string {
	if key == "" || key == "null" {
		return ""
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	domain := s.qiniuDomain
	if s.uploadType == "local" || domain == "" {
		domain = s.staticDomain
	}
	if domain == "" {
		return key
	}
	return strings.TrimSuffix(domain, "/") + "/" + key
}

// GetScratchAssets 素材库列表（裸 JSONArray，非 Result 包装）。
//
// 契约蓝本 TeachingScratchAssetsController#getScratchAssets：asset_type 精确
// 过滤，逐条 asset_data JSON 解析为对象返回。
func (s *Service) GetScratchAssets(ctx context.Context, assetType int) ([]*gvar.Var, error) {
	cols := dao.TeachingScratchAssets.Columns()
	var rows []entity.TeachingScratchAssets
	if err := dao.TeachingScratchAssets.Ctx(ctx).
		Where(cols.AssetType, assetType).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询素材库失败: %w", err)
	}
	out := make([]*gvar.Var, 0, len(rows))
	for _, r := range rows {
		item := parseJSONRaw(r.AssetData)
		out = append(out, item)
	}
	return out, nil
}

// List 素材分页（QueryGenerator 消费面：assetType 精确 + assetName 规则）。
func (s *Service) List(ctx context.Context, param page.Param, assetType string, assetName string) (*page.Result, error) {
	cols := dao.TeachingScratchAssets.Columns()
	m := dao.TeachingScratchAssets.Ctx(ctx)
	if assetType != "" {
		m = m.Where(cols.AssetType, assetType)
	}
	if assetName != "" {
		m = m.WhereLike(cols.AssetName, "%"+assetName+"%")
	}
	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("统计素材数量失败: %w", err)
	}
	if total == 0 {
		return page.New([]entity.TeachingScratchAssets{}, 0, param.PageNo, param.PageSize), nil
	}
	var rows []entity.TeachingScratchAssets
	if err := m.Page(param.PageNo, param.PageSize).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询素材列表失败: %w", err)
	}
	return page.New(rows, total, param.PageNo, param.PageSize), nil
}

// Add 新增素材（纯 save，id 为空时服务端生成）。
func (s *Service) Add(ctx context.Context, in AssetInput) error {
	data := do.TeachingScratchAssets{Id: in.Id}
	if in.Id == "" {
		data.Id = guid.S()
	}
	applyAssetInput(&data, in)
	if _, err := dao.TeachingScratchAssets.Ctx(ctx).Data(data).Insert(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Edit 编辑素材（纯 updateById）。
func (s *Service) Edit(ctx context.Context, in AssetInput) error {
	if in.Id == "" {
		return nil
	}
	data := do.TeachingScratchAssets{}
	applyAssetInput(&data, in)
	cols := dao.TeachingScratchAssets.Columns()
	if _, err := dao.TeachingScratchAssets.Ctx(ctx).
		Where(cols.Id, in.Id).Data(data).Update(); err != nil {
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

// Delete 删除素材（assetType==4 角色清理为 TODO——Java 现状跳过；
// 其余类型清理 sys_file 记录）。
func (s *Service) Delete(ctx context.Context, id string) error {
	var asset entity.TeachingScratchAssets
	if err := dao.TeachingScratchAssets.Ctx(ctx).
		Where(dao.TeachingScratchAssets.Columns().Id, id).Scan(&asset); err != nil {
		return fmt.Errorf("查询素材失败: %w", err)
	}
	if asset.Id == "" {
		return nil
	}
	if asset.AssetType != 4 {
		if err := s.cleanupSysFileByKey(ctx, asset.Md5Ext); err != nil {
			return err
		}
	}
	cols := dao.TeachingScratchAssets.Columns()
	if _, err := dao.TeachingScratchAssets.Ctx(ctx).Where(cols.Id, id).Delete(); err != nil {
		return fmt.Errorf("删除素材失败: %w", err)
	}
	return nil
}

// DeleteBatch 批量删除素材（同单删清理语义）。
func (s *Service) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	cols := dao.TeachingScratchAssets.Columns()
	var assets []entity.TeachingScratchAssets
	if err := dao.TeachingScratchAssets.Ctx(ctx).WhereIn(cols.Id, ids).Scan(&assets); err != nil {
		return fmt.Errorf("查询素材失败: %w", err)
	}
	for _, asset := range assets {
		if asset.AssetType != 4 {
			if err := s.cleanupSysFileByKey(ctx, asset.Md5Ext); err != nil {
				return err
			}
		}
	}
	if _, err := dao.TeachingScratchAssets.Ctx(ctx).WhereIn(cols.Id, ids).Delete(); err != nil {
		return fmt.Errorf("批量删除素材失败: %w", err)
	}
	return nil
}

// QueryByID 素材详情。
func (s *Service) QueryByID(ctx context.Context, id string) (*entity.TeachingScratchAssets, error) {
	var row entity.TeachingScratchAssets
	if err := dao.TeachingScratchAssets.Ctx(ctx).
		Where(dao.TeachingScratchAssets.Columns().Id, id).Scan(&row); err != nil {
		return nil, fmt.Errorf("查询素材失败: %w", err)
	}
	if row.Id == "" {
		return nil, ErrNotFound
	}
	return &row, nil
}

// cleanupSysFileByKey 按 file_path 精确匹配清理 sys_file 记录（与既有域同款）。
func (s *Service) cleanupSysFileByKey(ctx context.Context, key string) error {
	if key == "" || key == "null" {
		return nil
	}
	cols := dao.SysFile.Columns()
	var file entity.SysFile
	if err := dao.SysFile.Ctx(ctx).Where(cols.FilePath, key).Scan(&file); err != nil {
		return fmt.Errorf("查询文件记录失败: %w", err)
	}
	if file.Id == "" {
		return nil
	}
	if file.FileLocation == 1 || file.FileLocation == 2 {
		if _, err := dao.SysFile.Ctx(ctx).Where(cols.Id, file.Id).Delete(); err != nil {
			return fmt.Errorf("删除文件记录失败: %w", err)
		}
	}
	return nil
}

// applyAssetInput 请求体字段拷贝进 DO。
func applyAssetInput(data *do.TeachingScratchAssets, in AssetInput) {
	if in.AssetType != nil {
		data.AssetType = *in.AssetType
	}
	if in.AssetName != nil {
		data.AssetName = *in.AssetName
	}
	if in.AssetData != nil {
		data.AssetData = *in.AssetData
	}
	if in.Md5Ext != nil {
		data.Md5Ext = *in.Md5Ext
	}
	if in.Tags != nil {
		data.Tags = *in.Tags
	}
	if in.DelFlag != nil {
		data.DelFlag = *in.DelFlag
	}
}

// BackpackContent 背包列表（Redis list 分段读取，对齐 Java lGet(offset,
// offset+limit)；LRANGE 含端点故 end = offset+limit-1）。
func (s *Service) BackpackContent(ctx context.Context, operatorID string, limit, offset int) ([]*BackpackModel, error) {
	entries := s.kv.LRange(ctx, backpackListKeyFmt+operatorID, int64(offset), int64(offset+limit-1))
	out := make([]*BackpackModel, 0, len(entries))
	for _, e := range entries {
		if m := parseBackpack(e); m != nil {
			out = append(out, m)
		}
	}
	return out, nil
}

// BackpackSave 保存到背包。
//
// 契约蓝本 TeachingScratchBackpackController#save：body/thumbnail 为
// base64 编码内容；Java 上传七牛（Go 无七牛上传后端 → 落本地存储
// scratch/backpackBody/{id}[.jpg] 并按 uploadType 拼 URL——文档化偏差，
// 功能在 local 模式等价）；条目 RPUSH 进 scratch:backpackList:{uid}。
func (s *Service) BackpackSave(r *ghttp.Request, ctx context.Context, operatorID string, body, thumbnail, typ, mime, name string) (*BackpackModel, error) {
	id := strings.ReplaceAll(guid.S(), "-", "")
	bodyKey := "scratch/backpackBody/" + id
	thumbKey := "scratch/backpackBody/" + id + ".jpg"
	if err := s.saveLocal(ctx, bodyKey, body); err != nil {
		return nil, err
	}
	if err := s.saveLocal(ctx, thumbKey, thumbnail); err != nil {
		return nil, err
	}
	m := &BackpackModel{
		Id:        id,
		Name:      name,
		Mime:      mime,
		Type:      typ,
		Thumbnail: s.fileURL(thumbKey),
		Body:      s.fileURL(bodyKey),
	}
	entry, err := jsonMarshal(m)
	if err != nil {
		return nil, fmt.Errorf("序列化背包条目失败: %w", err)
	}
	s.kv.RPush(ctx, backpackListKeyFmt+operatorID, entry)
	return m, nil
}

// BackpackDelete 删除背包条目（列表遍历匹配 id；本地模式同步删物理文件；
// 七牛远端删除为文档化待补齐项）。
func (s *Service) BackpackDelete(ctx context.Context, operatorID, id string) error {
	key := backpackListKeyFmt + operatorID
	entries := s.kv.LRange(ctx, key, 0, -1)
	for _, e := range entries {
		m := parseBackpack(e)
		if m == nil || m.Id != id {
			continue
		}
		// Java 删七牛对象；Go 本地模式删物理文件（文档化：七牛删除待补齐）
		if s.uploadType == "local" && s.localRoot != "" {
			_ = os.Remove(filepath.Join(s.localRoot, "scratch/backpackBody/"+id))
			_ = os.Remove(filepath.Join(s.localRoot, "scratch/backpackBody/"+id+".jpg"))
		}
		s.kv.LRem(ctx, key, e)
	}
	return nil
}

// saveLocal 保存 base64 内容到本地存储。
func (s *Service) saveLocal(ctx context.Context, key, contentB64 string) error {
	if s.localRoot == "" {
		return fmt.Errorf("本地存储根目录未配置")
	}
	raw, err := base64.StdEncoding.DecodeString(contentB64)
	if err != nil {
		// 兼容 URL 安全变体（前端 Scratch 导出可能使用）
		raw, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(contentB64)
		if err != nil {
			return fmt.Errorf("背包内容 base64 解码失败: %w", err)
		}
	}
	abs := filepath.Join(s.localRoot, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(abs, raw, 0o644); err != nil {
		return fmt.Errorf("写入背包文件失败: %w", err)
	}
	return nil
}

// BackpackGetFile 读取背包 body 遗留键（Java 现状：save 不写该键，恒空）。
func (s *Service) BackpackGetFile(ctx context.Context, filename string) []byte {
	val, ok := s.kv.Get(ctx, backpackBodyKeyFmt+filename)
	if !ok || val == "" {
		return []byte{}
	}
	raw, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return []byte{}
	}
	return raw
}

// parseBackpack 解析背包条目 JSON（Java FastJson 旧条目解析失败跳过——
// 灰度期已知偏差）。
func parseBackpack(entry string) *BackpackModel {
	m := &BackpackModel{}
	if err := jsonUnmarshal(entry, m); err != nil {
		return nil
	}
	return m
}

// KVListStore 背包 list 存储（Redis 主 + 内存兜底）。
type KVListStore interface {
	RPush(ctx context.Context, key, val string)
	LRange(ctx context.Context, key string, start, stop int64) []string
	LRem(ctx context.Context, key, val string)
	Get(ctx context.Context, key string) (string, bool)
}

// NewKVListStore 按开关选择实现。
func NewKVListStore(redisEnabled bool) KVListStore {
	if redisEnabled {
		return &redisListStore{}
	}
	return &memListStore{lists: map[string][]string{}, vals: map[string]string{}}
}

// redisListStore Redis 实现（RPUSH/LRANGE/LREM 对齐 jeecg RedisUtil）。
type redisListStore struct{}

func (s *redisListStore) RPush(ctx context.Context, key, val string) {
	_, _ = g.Redis().Do(ctx, "RPUSH", key, val)
}

func (s *redisListStore) LRange(ctx context.Context, key string, start, stop int64) []string {
	v, err := g.Redis().Do(ctx, "LRANGE", key, start, stop)
	if err != nil || v == nil {
		return nil
	}
	return v.Strings()
}

func (s *redisListStore) LRem(ctx context.Context, key, val string) {
	_, _ = g.Redis().Do(ctx, "LREM", key, 0, val)
}

func (s *redisListStore) Get(ctx context.Context, key string) (string, bool) {
	v, err := g.Redis().Do(ctx, "GET", key)
	if err != nil || v == nil {
		return "", false
	}
	return v.String(), true
}

// memListStore 进程内存实现。
type memListStore struct {
	lists map[string][]string
	vals  map[string]string
}

func (s *memListStore) RPush(_ context.Context, key, val string) {
	s.lists[key] = append(s.lists[key], val)
}

func (s *memListStore) LRange(_ context.Context, key string, start, stop int64) []string {
	l := s.lists[key]
	n := int64(len(l))
	if start < 0 {
		start = 0
	}
	if stop >= n {
		stop = n - 1
	}
	if start > stop || start >= n {
		return nil
	}
	out := make([]string, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		out = append(out, l[i])
	}
	return out
}

func (s *memListStore) LRem(_ context.Context, key, val string) {
	l := s.lists[key]
	out := l[:0]
	for _, item := range l {
		if item != val {
			out = append(out, item)
		}
	}
	s.lists[key] = out
}

func (s *memListStore) Get(_ context.Context, key string) (string, bool) {
	v, ok := s.vals[key]
	return v, ok
}

// gjsonRaw 裸 JSON 对象占位（getScratchAssets 返回元素）。
type gjsonRaw = *gvar.Var

// jsonMarshal JSON 序列化。
func jsonMarshal(v interface{}) (string, error) {
	b, err := gjson.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// jsonUnmarshal JSON 反序列化。
func jsonUnmarshal(data string, v interface{}) error {
	return gjson.Unmarshal([]byte(data), v)
}

// parseJSONRaw 解析素材元数据 JSON（非法 JSON 返回 null 对象）。
func parseJSONRaw(data string) *gvar.Var {
	v := gjson.New(data)
	if v == nil {
		return nil
	}
	raw := v.Var()
	return raw
}
