// Package upload 提供本地文件上传与静态预览的 HTTP 处理器
// （契约蓝本：CommonController 的 /sys/common/upload 与 /sys/common/static/**）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - upload（uploadType=local）：落盘 {STORAGE_LOCAL_ROOT}/{biz}/{uuid.后缀}，
//     dbpath = biz/uuid.ext（\ 归一为 /），放 Result.message；jeditor 参数
//     存在时只回 message="local"；非 local 模式 Java 走 minio/alioss（无 qiniu
//     分支），Go 当前仅实现 local（其余模式返回明确错误，文档化）；
//   - static/**：本地预览/下载，路径剥离 ".."，anon（图片预览不限 token）；
//     支持 Range 断点（V2.8 视频拖动进度依赖）。
package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// Service 本地上传/静态服务。
type Service struct {
	uploadType string
	localRoot  string
}

// New 构造服务。
func New(uploadType, localRoot string) *Service {
	return &Service{uploadType: uploadType, localRoot: localRoot}
}

// Upload 通用上传（local 模式）。
// 契约：POST /sys/common/upload，form: file + biz + jeditor。
func (s *Service) Upload(r *ghttp.Request) {
	if s.uploadType != "local" {
		// Java 走 minio/alioss 后端（无 qiniu 分支）；Go 当前仅 local，
		// 其余模式明确报错（文档化差距）
		common0(r, false, "当前上传方式未实现: "+s.uploadType)
		return
	}
	file := r.GetUploadFile("file")
	if file == nil {
		common0(r, false, "上传失败")
		return
	}
	biz := r.Get("biz", "temp").String()
	ext := filepath.Ext(file.Filename)
	name := guid.S() + ext
	relPath := strings.ReplaceAll(filepath.Join(biz, name), "\\", "/")
	absPath := filepath.Join(s.localRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		g.Log().Errorf(r.Context(), "创建上传目录失败 err=%s", err.Error())
		common0(r, false, "上传失败")
		return
	}
	if _, err := file.Save(absPath, true); err != nil {
		g.Log().Errorf(r.Context(), "保存上传文件失败 err=%s", err.Error())
		common0(r, false, "上传失败")
		return
	}
	if r.Get("jeditor").String() != "" {
		// jeditor 场景只回 "local"（对齐 Java）
		common0(r, true, "local")
		return
	}
	common0(r, true, relPath)
}

// Static 本地静态文件预览/下载（anon；支持 Range 断点续传）。
// 契约：GET /sys/common/static/{path}。
func (s *Service) Static(r *ghttp.Request) {
	rel := r.Get("path").String()
	// 剥离路径穿越（对齐 Java 去掉 ".."）
	rel = strings.ReplaceAll(rel, "..", "")
	absPath := filepath.Join(s.localRoot, filepath.FromSlash(rel))
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		r.Response.WriteStatus(404)
		return
	}
	f, err := os.Open(absPath)
	if err != nil {
		r.Response.WriteStatus(404)
		return
	}
	defer func() { _ = f.Close() }()
	// Content-Type 按扩展名映射：图片/视频内联渲染（门户封面 <img> 依赖），
	// 其余保持 octet-stream 下载语义（V2.8 静态预览行为）。
	ct := contentTypeByExt(filepath.Ext(absPath))
	r.Response.Header().Set("Content-Type", ct)
	if ct == "application/octet-stream" {
		r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%q", info.Name()))
	}
	// Range 断点支持（V2.8 本地文件视频进度拖动依赖）
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		r.Response.WriteHeader(200)
		writeFullFile(r, f)
		return
	}
	start, end, ok := parseRange(rangeHeader, info.Size())
	if !ok {
		r.Response.WriteHeader(200)
		writeFullFile(r, f)
		return
	}
	r.Response.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
	r.Response.Header().Set("Accept-Ranges", "bytes")
	r.Response.WriteHeader(206)
	buf := make([]byte, 32*1024)
	pos := start
	for pos <= end {
		if _, err := f.Seek(pos, 0); err != nil {
			return
		}
		n, err := f.Read(buf)
		if n > 0 {
			writeEnd := pos + int64(n) - 1
			if writeEnd > end {
				n -= int(writeEnd - end)
			}
			if werr := writeChunk(r, buf[:n]); werr != nil {
				return
			}
			pos += int64(n)
		}
		if err != nil {
			break
		}
	}
}

// parseRange 解析 Range: bytes=start-end（end 可省略）。
// writeFullFile 将整个文件内容写入响应（必须读出字节，不可直接传 *os.File——
// ghttp Response.Write 对非字节类型只写格式化表示，曾导致静态文件 2 字节）。
func writeFullFile(r *ghttp.Request, f *os.File) {
	if _, err := f.Seek(0, 0); err != nil {
		return
	}
	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			if werr := writeChunk(r, buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// contentTypeByExt 扩展名 → Content-Type（覆盖门户/课程常见静态类型）。
func contentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

func parseRange(header string, size int64) (int64, int64, bool) {
	if !strings.HasPrefix(header, "bytes=") {
		return 0, 0, false
	}
	spec := strings.TrimPrefix(header, "bytes=")
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	var start, end int64
	if parts[0] == "" {
		// suffix range：末尾 N 字节
		var n int64
		if _, err := fmt.Sscanf(parts[1], "%d", &n); err != nil || n <= 0 {
			return 0, 0, false
		}
		start = size - n
		if start < 0 {
			start = 0
		}
		end = size - 1
	} else {
		if _, err := fmt.Sscanf(parts[0], "%d", &start); err != nil || start < 0 || start >= size {
			return 0, 0, false
		}
		if parts[1] == "" {
			end = size - 1
		} else {
			if _, err := fmt.Sscanf(parts[1], "%d", &end); err != nil || end < start {
				return 0, 0, false
			}
			if end >= size {
				end = size - 1
			}
		}
	}
	return start, end, true
}

// common0 简化响应壳（message 承载 dbpath/local，与 Java Result 形状一致）。
func common0(r *ghttp.Request, success bool, message string) {
	r.Response.WriteJson(g.Map{
		"success":   success,
		"message":   message,
		"code":      map[bool]int{true: 200, false: 500}[success],
		"result":    nil,
		"timestamp": time.Now().UnixMilli(),
	})
}

// writeChunk 写一段响应体（封装以统一错误处理）。
func writeChunk(r *ghttp.Request, data []byte) error {
	r.Response.Write(data)
	return nil
}
