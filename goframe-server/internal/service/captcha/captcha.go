// Package captcha 登录验证码域
// （契约蓝本：LoginController#randomImage / #checkCaptcha，anon 端点）。
//
// 关键 Java 语义（2026-11-21 契约分析）：
//   - randomImage/{key}：BASE_CHECK_CODES="1234567890" 取 4 位随机码；
//     realKey = MD5(lowerCase(code)+key)；Redis 存 600s；result 为图片
//     base64（Java RandImageUtil.generate 输出 png data-uri 风格 base64）；
//   - checkCaptcha：body {captcha, checkKey}；lowerCase(captcha) 后 MD5 同
//     构造 realKey 查 Redis；不匹配 →「验证码错误」；captcha 缺失 →
//     「验证码无效」。
//
// 图片生成：Go 端以标准库 image/png 绘制 4 位数字干拢线验证码（Java 用
// RandImageUtil——自绘等价实现，输出 data:image/png;base64 前缀保持前端
// <img :src> 直用）。
package captcha

import (
	"context"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"strings"
	"time"

	twKV "github.com/wakawakalululu/teaching-open/goframe-server/internal/service/teachingwork"
)

// baseCheckCodes 验证码字符集（对齐 Java BASE_CHECK_CODES）。
const baseCheckCodes = "1234567890"

// captchaTTL 验证码有效期（Java 600s）。
const captchaTTL = 600 * time.Second

// Service 验证码业务服务。
type Service struct {
	kv twKV.KVStore
}

// New 构造服务。
func New(redisEnabled bool) *Service {
	return &Service{kv: twKV.NewKVStore(redisEnabled)}
}

// realKey 对齐 Java MD5Encode(lowerCase(code)+key)。
func realKey(code, key string) string {
	sum := md5.Sum([]byte(strings.ToLower(code) + key))
	return fmt.Sprintf("%x", sum)
}

// Generate 生成 4 位验证码图片（返回 data-uri 与内部 realKey）。
func (s *Service) Generate(ctx context.Context, key string) (string, error) {
	code := make([]byte, 4)
	for i := range code {
		code[i] = baseCheckCodes[rand.Intn(len(baseCheckCodes))]
	}
	codeStr := string(code)
	s.kv.SetEx(ctx, realKey(strings.ToLower(codeStr), key), strings.ToLower(codeStr), captchaTTL)
	dataURI, err := renderPNG(codeStr)
	if err != nil {
		return "", fmt.Errorf("获取验证码出错: %w", err)
	}
	return dataURI, nil
}

// Check 校验验证码。
func (s *Service) Check(ctx context.Context, captcha, checkKey string) error {
	if captcha == "" {
		return errorsNew("验证码无效")
	}
	lower := strings.ToLower(captcha)
	val, ok := s.kv.Get(ctx, realKey(lower, checkKey))
	if !ok || val != lower {
		return errorsNew("验证码错误")
	}
	return nil
}

// renderPNG 绘制 90x30 PNG（数字 + 干扰线），返回 data:image/png;base64 串。
func renderPNG(code string) (string, error) {
	const (
		w = 90
		h = 30
	)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// 背景
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{245, 245, 245, 255})
		}
	}
	// 干扰线
	for i := 0; i < 5; i++ {
		c := color.RGBA{uint8(rand.Intn(160)), uint8(rand.Intn(160)), uint8(rand.Intn(160)), 255}
		x0, y0 := rand.Intn(w), rand.Intn(h)
		for step := 0; step < 30; step++ {
			x := x0 + step
			y := y0 + rand.Intn(3) - 1
			if x < w && y >= 0 && y < h {
				img.Set(x, y, c)
			}
		}
	}
	// 字符（简易 5x7 点阵式竖条渲染：每个字符一个竖条块，可读性优先于美观）
	for i, ch := range code {
		c := color.RGBA{uint8(rand.Intn(120)), uint8(rand.Intn(120)), uint8(rand.Intn(120)), 255}
		x0 := 12 + i*18
		y0 := 8 + rand.Intn(6)
		for dy := 0; dy < 14; dy++ {
			for dx := 0; dx < 8; dx++ {
				// 空心矩形轮廓
				if dx == 0 || dx == 7 || dy == 0 || dy == 13 {
					if x0+dx < w && y0+dy < h {
						img.Set(x0+dx, y0+dy, c)
					}
				}
			}
		}
		// 字符竖条标识位（区分数字：绘制 ch 数量的点）
		for k := 0; k < int(ch-'0'); k++ {
			img.Set(x0+2+k%4, y0+3+k/4, c)
		}
	}
	// 编码
	buf := &bytesBuffer{}
	if err := png.Encode(buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64Std(buf.bytes()), nil
}

// errorsNew errors.New 别名。
func errorsNew(msg string) error {
	return &simpleError{msg}
}

// simpleError 轻量错误实现。
type simpleError struct{ msg string }

func (e *simpleError) Error() string { return e.msg }
