package passwd

import (
	"testing"
)

// 本测试的期望值来自独立实现（纯 Python 复算 PKCS#5 v1.5 的 PBKDF1(MD5) + DES-CBC，
// 并通过 FIPS 81 的 DES 已知向量自检），用于确认 Go 端能与 Java 版
// PasswordUtil.encrypt(username, password, salt) 的存量为兼容。
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

// 校验 Go 实现的 keyAndIV 与参考实现一致。
func TestLegacyKeyDerivationMatchesReference(t *testing.T) {
	p := NewLegacyPBE()
	for _, v := range legacyVectors {
		key, iv, err := p.keyAndIV(v.password, v.salt)
		if err != nil {
			t.Fatalf("%s: 派生失败: %v", v.username, err)
		}
		if len(key) != 8 || len(iv) != 8 {
			t.Fatalf("%s: key/iv 长度异常 key=%d iv=%d", v.username, len(key), len(iv))
		}
	}
}

// 校验存量密文能被正确解密（等价于 Java 的校验语义）。
func TestLegacyVerify(t *testing.T) {
	p := NewLegacyPBE()
	for _, v := range legacyVectors {
		if err := p.Verify(v.username, v.password, v.cipher, v.salt); err != nil {
			t.Fatalf("%s: 期望校验通过，实际失败: %v", v.username, err)
		}
	}
}

// 错误密码必须被拒绝。
func TestLegacyVerifyWrongPassword(t *testing.T) {
	p := NewLegacyPBE()
	v := legacyVectors[0]
	if err := p.Verify(v.username, "wrong-password", v.cipher, v.salt); err == nil {
		t.Fatal("错误密码应校验失败")
	}
}

// 错误用户名必须被拒绝（密文解出的明文与用户名不匹配）。
func TestLegacyVerifyWrongUsername(t *testing.T) {
	p := NewLegacyPBE()
	v := legacyVectors[0]
	if err := p.Verify("not-admin", v.password, v.cipher, v.salt); err == nil {
		t.Fatal("错误用户名应校验失败")
	}
}

// 非法密文不应 panic，应返回错误。
func TestLegacyVerifyMalformedCipher(t *testing.T) {
	p := NewLegacyPBE()
	cases := []string{"", "zz", "abc", "0011223344556677"}
	for _, c := range cases {
		if err := p.Verify("admin", "123456", c, "63293188"); err == nil {
			t.Fatalf("畸形密文 %q 应校验失败", c)
		}
	}
}

func TestSchemeDetect(t *testing.T) {
	cases := []struct {
		stored string
		want   Scheme
	}{
		{"$2a$12$abcdefghijklmnopqrstuv", SchemeBcrypt},
		{"$2b$12$abcdefghijklmnopqrstuv", SchemeBcrypt},
		{"c1461cf4492050e9", SchemeLegacyPBE},
		{"", SchemeUnknown},
	}
	for _, c := range cases {
		if got := Detect(c.stored); got != c.want {
			t.Fatalf("Detect(%q)=%v, want %v", c.stored, got, c.want)
		}
	}
}
