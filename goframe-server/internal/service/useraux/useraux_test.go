package useraux

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

// TestGenerateUserId 校验生成 ID 与 Java 契约完全同构：
// UUID.randomUUID().toString().replace("-", "") → 32 位小写十六进制。
func TestGenerateUserId(t *testing.T) {
	svc := New()
	first, err := svc.GenerateUserId()
	if err != nil {
		t.Fatalf("GenerateUserId 返回错误: %v", err)
	}
	if len(first) != 32 {
		t.Fatalf("生成 ID 长度应为 32，实际 %d: %q", len(first), first)
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("生成 ID 应为纯十六进制: %q, err=%v", first, err)
	}
	if first != strings.ToLower(first) {
		t.Fatalf("生成 ID 应为小写: %q", first)
	}
	second, err := svc.GenerateUserId()
	if err != nil {
		t.Fatalf("GenerateUserId 第二次调用返回错误: %v", err)
	}
	if first == second {
		t.Fatalf("两次生成的 ID 不应相同: %q", first)
	}
}

// TestMaskPhone 对齐 Java desensitizePhone（SysUserController.java L1310-1315）：
// 长度 ≥7 保留前 3 位与后 4 位、中间 ****；不足 7 位原样返回。
func TestMaskPhone(t *testing.T) {
	cases := []struct {
		name  string
		phone string
		want  string
	}{
		{"11位手机号", "13812345678", "138****5678"},
		{"恰好7位", "1234567", "123****4567"},
		{"6位不脱敏", "123456", "123456"},
		{"空串", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := maskPhone(tc.phone); got != tc.want {
				t.Fatalf("maskPhone(%q) = %q, want %q", tc.phone, got, tc.want)
			}
		})
	}
}

// TestSysUserBriefJSONWhitelist 锁定 querySysUser 输出白名单：
// 只允许 username/phone 两个键，防止敏感字段（password/salt 等）混入。
func TestSysUserBriefJSONWhitelist(t *testing.T) {
	raw, err := json.Marshal(SysUserBrief{Username: "u1", Phone: "138****5678"})
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}
	if len(m) != 2 {
		t.Fatalf("输出键数量应为 2，实际 %d: %v", len(m), m)
	}
	if _, ok := m["username"]; !ok {
		t.Fatalf("缺少 username 键: %v", m)
	}
	if _, ok := m["phone"]; !ok {
		t.Fatalf("缺少 phone 键: %v", m)
	}
}

// TestBizErrorMessages 锁定业务错误话术与 Java 契约逐字一致。
func TestBizErrorMessages(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"未找到对应实体", ErrEntityNotFound, "未找到对应实体"},
		{"用户账号已存在", ErrUserExists, "用户账号已存在"},
		{"验证失败", ErrVerifyFailed, "验证失败"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.want {
				t.Fatalf("话术漂移: got %q, want %q", tc.err.Error(), tc.want)
			}
		})
	}
}
