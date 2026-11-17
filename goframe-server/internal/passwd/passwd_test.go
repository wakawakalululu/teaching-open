package passwd

import (
	"testing"
)

// TestLegacyPBEVerify 用 4 组与 Java 版对齐的向量校验兼容层。
// 这些向量来自 docs/GO-MIGRATION-PLAN.md §4.3 与 _crosscheck/pbkdf1_des_ref.py，
// 与 Java 版 PasswordUtil.encrypt(username, password, salt) 的输出一致。
func TestLegacyPBEVerify(t *testing.T) {
	cases := []struct {
		username string
		password string
		salt     string
		cipher   string
	}{
		{"admin", "123456", "63293188", "c1461cf4492050e9"},
		{"teacher", "123456", "63293188", "44a7ec3167f68089"},
		{"student", "Abc@12345", "63293188", "ae7d77e35833d801"},
		{"张三", "passw0rd", "63293188", "2e235213908e99bf"},
	}

	legacy := NewLegacyPBE()
	for _, c := range cases {
		if err := legacy.Verify(c.username, c.password, c.cipher, c.salt); err != nil {
			t.Errorf("Verify(%q, %q) 应通过，得到错误: %v", c.username, c.password, err)
		}
	}
}

// TestLegacyPBEVerifyRejectsWrongPassword 错误密码应拒绝。
func TestLegacyPBEVerifyRejectsWrongPassword(t *testing.T) {
	legacy := NewLegacyPBE()
	if err := legacy.Verify("admin", "wrongpass", "c1461cf4492050e9", "63293188"); err != ErrMismatch {
		t.Errorf("错误密码应返回 ErrMismatch，得到: %v", err)
	}
}

// TestLegacyPBEVerifyRejectsWrongUsername 错误用户名应拒绝。
func TestLegacyPBEVerifyRejectsWrongUsername(t *testing.T) {
	legacy := NewLegacyPBE()
	if err := legacy.Verify("root", "123456", "c1461cf4492050e9", "63293188"); err != ErrMismatch {
		t.Errorf("错误用户名应返回 ErrMismatch，得到: %v", err)
	}
}

// TestLegacyPBEVerifyMalformed 畸形密文不应 panic。
func TestLegacyPBEVerifyMalformed(t *testing.T) {
	legacy := NewLegacyPBE()
	for _, bad := range []string{"", "zzzz", "abcd", "0123456789abcdef01"} {
		if err := legacy.Verify("admin", "123456", bad, "63293188"); err == nil {
			t.Errorf("畸形密文 %q 不应通过", bad)
		}
	}
}

// TestDetect 方案识别。
func TestDetect(t *testing.T) {
	if Detect("$2a$10$abcdefghijklmnopqrstuv") != SchemeBcrypt {
		t.Error("bcrypt 前缀应识别为 SchemeBcrypt")
	}
	if Detect("c1461cf4492050e9") != SchemeLegacyPBE {
		t.Error("非 bcrypt 应识别为 SchemeLegacyPBE")
	}
	if Detect("") != SchemeUnknown {
		t.Error("空值应识别为 SchemeUnknown")
	}
}

// TestBcryptRoundtrip bcrypt 散列往返。
func TestBcryptRoundtrip(t *testing.T) {
	hash, err := Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash 失败: %v", err)
	}
	if err := VerifyBcrypt("S3curePass!", hash); err != nil {
		t.Errorf("VerifyBcrypt 应通过，得到: %v", err)
	}
	if err := VerifyBcrypt("wrong", hash); err != ErrMismatch {
		t.Errorf("错误密码应返回 ErrMismatch，得到: %v", err)
	}
}
