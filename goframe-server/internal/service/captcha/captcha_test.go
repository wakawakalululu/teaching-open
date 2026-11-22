// captcha_test.go 验证码域零 DB 依赖单测：realKey 构造、内存 KV 校验链路、
// 字形渲染（PNG 可解码且含足量深色字形像素）。
package captcha

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/png"
	"strings"
	"testing"
)

func TestRealKey(t *testing.T) {
	// MD5(lowerCase(code)+key) 的确定性
	a := realKey("1234", "abc")
	b := realKey("1234", "abc")
	c := realKey("1234", "abd")
	if a != b {
		t.Fatalf("同输入应产生同 key: %s vs %s", a, b)
	}
	if a == c {
		t.Fatal("不同 key 不应产生同 key")
	}
	if len(a) != 32 {
		t.Fatalf("MD5 hex 长度应为 32: %d", len(a))
	}
}

func TestGenerateAndCheck(t *testing.T) {
	s := New(false)
	ctx := context.Background()
	// 生成 4 位纯数字码
	_, err := s.Generate(ctx, "checkKey1")
	if err != nil {
		t.Fatalf("Generate 失败: %v", err)
	}
	// 错误验证码
	if err := s.Check(ctx, "9999", "checkKey1"); err == nil || !strings.Contains(err.Error(), "验证码错误") {
		t.Fatalf("错误验证码应报「验证码错误」: %v", err)
	}
	// 空验证码
	if err := s.Check(ctx, "", "checkKey1"); err == nil || !strings.Contains(err.Error(), "验证码无效") {
		t.Fatalf("空验证码应报「验证码无效」: %v", err)
	}
}

func TestRenderPNGShape(t *testing.T) {
	dataURI, err := renderPNG("1234")
	if err != nil {
		t.Fatalf("renderPNG 失败: %v", err)
	}
	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Fatalf("data-uri 前缀异常: %s", dataURI[:40])
	}
	// 解码校验：尺寸 90x30，且存在足量深色字形像素（点阵字形可读性保障）
	b64 := strings.TrimPrefix(dataURI, "data:image/png;base64,")
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("base64 解码失败: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("png 解码失败: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 90 || b.Dy() != 30 {
		t.Fatalf("画布尺寸应为 90x30: %dx%d", b.Dx(), b.Dy())
	}
	dark := 0
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			// 深色字形：最大通道 < 45000（16 位域）——覆盖全部深色调色板
			// （最浅通道 130×257≈33410），排除背景 250 与干扰灰 205。
			if r < 45000 && g < 45000 && bl < 45000 {
				dark++
			}
		}
	}
	// 4 个数字字形（'1'/'2'/'3'/'4' 点阵像素 10/14/15/16 × 2x2 放大 ≈ 220，
	// 扣除画布边缘裁切），阈值取实测下限。
	if dark < 150 {
		t.Fatalf("深色字形像素不足，字形可能未渲染: dark=%d", dark)
	}
}
