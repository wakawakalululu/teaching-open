// captcha_test.go 验证码域零 DB 依赖单测：realKey 构造、内存 KV 校验链路。
package captcha

import (
	"context"
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
}
