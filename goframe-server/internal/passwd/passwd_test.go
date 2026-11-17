package passwd

import (
	"testing"
)

// legacyVectors 与 Java 版 PasswordUtil.encrypt 对齐的已知向量。
// 同时用于 Verify（解密方向）与 Encrypt（加密方向）双向验证。
var legacyVectors = []struct {
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

// TestLegacyPBEVerify 用 4 组与 Java 版对齐的向量校验兼容层。
// 这些向量来自 docs/GO-MIGRATION-PLAN.md §4.3 与 _crosscheck/pbkdf1_des_ref.py，
// 与 Java 版 PasswordUtil.encrypt(username, password, salt) 的输出一致。
func TestLegacyPBEVerify(t *testing.T) {
	legacy := NewLegacyPBE()
	for _, c := range legacyVectors {
		if err := legacy.Verify(c.username, c.password, c.cipher, c.salt); err != nil {
			t.Errorf("Verify(%q, %q) 应通过，得到错误: %v", c.username, c.password, err)
		}
	}
}

// TestLegacyPBEEncryptMatchesJava 加密方向必须与 Java 输出逐字节一致。
func TestLegacyPBEEncryptMatchesJava(t *testing.T) {
	legacy := NewLegacyPBE()
	for _, c := range legacyVectors {
		got, err := legacy.Encrypt(c.username, c.password, c.salt)
		if err != nil {
			t.Fatalf("Encrypt(%q) 失败: %v", c.username, err)
		}
		if got != c.cipher {
			t.Errorf("Encrypt(%q, %q) = %s，期望 %s", c.username, c.password, got, c.cipher)
		}
	}
}

// TestLegacyPBERoundtrip 加密后解密应还原。
func TestLegacyPBERoundtrip(t *testing.T) {
	legacy := NewLegacyPBE()
	const (
		plain = "测试用户名"
		pwd   = "p@ssw0rd!"
		salt  = "12345678"
	)
	cipherText, err := legacy.Encrypt(plain, pwd, salt)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}
	back, err := legacy.Decrypt(cipherText, pwd, salt)
	if err != nil {
		t.Fatalf("Decrypt 失败: %v", err)
	}
	if back != plain {
		t.Errorf("往返失败: 得到 %q，期望 %q", back, plain)
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
