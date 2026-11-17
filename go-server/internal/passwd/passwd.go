// Package passwd 处理密码散列的校验与升级。
//
// 迁移背景（重要）：
//
//	Java 版 LoginController 调用的是
//	    PasswordUtil.encrypt(username, password, sysUser.getSalt())
//	其中三个形参分别是 (plaintext, password, salt)。也就是说：
//	  - 被加密的明文 = 用户名
//	  - 加密用的口令 = 用户输入的密码
//	  - 盐 = sys_user.salt（每人 8 字节）
//	算法为 PBEWithMD5AndDES（PKCS#5 v1.5 的 PBKDF1(MD5) 派生 + DES-CBC）。
//	因此存量"散列"实际是可逆密文，校验方式是「用输入密码解密后与用户名比对」。
//
//	这带来两个必须修的问题：
//	  1. DES(56 位) + MD5 派生，强度已不足以抵抗离线爆破；
//	  2. 密码可逆 => 一旦库泄露，攻击者可直接解出明文密码（用户往往多站复用）。
//
// 迁移策略（rehash-on-login，可在不停机的情况下完成）：
//  1. 用户提交明文密码后，先用 LegacyPBE.Verify 校验存量密文；
//  2. 校验通过即用 bcrypt 重新散列并写回（rehash），该用户后续走新算法；
//  3. 全部存量用户迁移完成后，删除本文件的 Legacy 部分，彻底移除弱算法。
package passwd

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ErrMismatch 密码不匹配。
var ErrMismatch = errors.New("密码不匹配")

// Scheme 标识散列/密文方案。
type Scheme int

const (
	// SchemeUnknown 无法识别的存储格式。
	SchemeUnknown Scheme = iota
	// SchemeBcrypt 新方案（单向散列，推荐）。
	SchemeBcrypt
	// SchemeLegacyPBE Java 版 PBEWithMD5AndDES 存量密文。
	SchemeLegacyPBE
)

// bcryptCost 散列强度。12 约 200~300ms，兼顾安全与登录体验。
const bcryptCost = 12

// legacyIterations 与 Java 版 PasswordUtil.ITERATIONCOUNT 保持一致。
const legacyIterations = 1000

// legacySalt Java 版 PasswordUtil.Salt 默认值；实际使用时优先取 sys_user.salt。
const legacySalt = "63293188"

// Hash 用 bcrypt 生成新散列。
func Hash(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("密码不能为空")
	}
	if len(plain) > 72 {
		// bcrypt 仅取前 72 字节，超长会被静默截断，显式拒绝以免产生歧义
		return "", errors.New("密码长度超过 72 字节上限")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("生成密码散列失败: %w", err)
	}
	return string(b), nil
}

// VerifyBcrypt 校验 bcrypt 散列。
func VerifyBcrypt(plain, hashed string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)); err != nil {
		return ErrMismatch
	}
	return nil
}

// Detect 根据存储值判断方案。
func Detect(stored string) Scheme {
	switch {
	case strings.HasPrefix(stored, "$2a$"), strings.HasPrefix(stored, "$2b$"), strings.HasPrefix(stored, "$2y$"):
		return SchemeBcrypt
	case stored == "":
		return SchemeUnknown
	default:
		return SchemeLegacyPBE
	}
}

// NeedsRehash 判断该存储值是否需要升级到 bcrypt。
func NeedsRehash(stored string) bool {
	s := Detect(stored)
	if s != SchemeBcrypt {
		return true
	}
	cost, err := bcrypt.Cost([]byte(stored))
	if err != nil {
		return true
	}
	return cost < bcryptCost
}

// ---------- Legacy: PBEWithMD5AndDES ----------

// LegacyPBE 复刻 Java 版 PBEWithMD5AndDES 的加解密行为。
type LegacyPBE struct {
	// Iterations 迭代次数，对应 Java 的 ITERATIONCOUNT(1000)。
	Iterations int
}

// NewLegacyPBE 构造默认参数的兼容实现。
func NewLegacyPBE() *LegacyPBE {
	return &LegacyPBE{Iterations: legacyIterations}
}

// keyAndIV 按 PKCS#5 v1.5 PBKDF1(MD5) 派生 8 字节密钥与 8 字节 IV。
//
// 算法：T1 = MD5(password || salt)；Ti = MD5(T(i-1))；
// 串联 T1..Tn 直到够 16 字节，前 8 字节为 DES key，后 8 字节为 IV。
// 注意：Java 的 PBEKeySpec 对 ASCII 口令等价于取字节；本实现按字节直接拼接，
// 非 ASCII 口令需与 Java 侧先做一次比对验证。
func (p *LegacyPBE) keyAndIV(password, salt string) ([]byte, []byte, error) {
	if len(salt) != 8 {
		return nil, nil, fmt.Errorf("PBE 盐必须为 8 字节，当前 %d", len(salt))
	}

	// PKCS#5 v1.5 PBKDF1(MD5)：T1 = MD5(password || salt)。
	//
	// 关键细节（容易踩坑）：PBKDF1 规定当所需密钥长度 dkLen <= hLen(MD5=16) 时，
	// 输出就是 T1，迭代次数 *不参与* 结果计算。Java 的 PBEWithMD5AndDES 派生的
	// 正是 key(8) + iv(8) = 16 字节，因此 iterations 参数对最终密钥没有影响。
	// 这既是本实现必须与 Java 完全对齐的地方，也说明该方案的实际强度
	// 仅相当于"一次 MD5"，远低于迭代 1000 次给人的印象——进一步印证迁移到
	// bcrypt/Argon2id 的必要性。
	//
	// 若未来 dkLen > 16，则需按 PBKDF1 定义串联 T1..Tn，并真正应用迭代次数。
	t := md5.Sum(append([]byte(password), []byte(salt)...))
	return t[0:8], t[8:16], nil
}

// Decrypt 用「用户输入密码」解密存量密文，返回原文（正常情况下应为用户名）。
func (p *LegacyPBE) Decrypt(ciphertextHex, password, salt string) (string, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(ciphertextHex))
	if err != nil {
		return "", fmt.Errorf("密文格式非法: %w", err)
	}
	if len(raw) == 0 || len(raw)%des.BlockSize != 0 {
		return "", errors.New("密文长度非法")
	}
	key, iv, err := p.keyAndIV(password, salt)
	if err != nil {
		return "", err
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("DES 初始化失败: %w", err)
	}
	plain := make([]byte, len(raw))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, raw)

	// 去 PKCS#5 填充
	plain, err = pkcs5Unpad(plain, des.BlockSize)
	if err != nil {
		return "", ErrMismatch
	}
	return string(plain), nil
}

// Verify 校验「用户名 + 输入密码 + 盐」是否与存量密文匹配。
//
// 对应 Java 的 PasswordUtil.encrypt(username, password, salt) 的逆运算。
func (p *LegacyPBE) Verify(username, plainPassword, storedCipher, salt string) error {
	if storedCipher == "" {
		return ErrMismatch
	}
	useSalt := salt
	if len(useSalt) != 8 {
		useSalt = legacySalt
	}
	decrypted, err := p.Decrypt(storedCipher, plainPassword, useSalt)
	if err != nil {
		return ErrMismatch
	}
	// 常量时间比较，避免通过响应时间推断前缀匹配长度
	if !ConstantTimeCompare(decrypted, username) {
		return ErrMismatch
	}
	return nil
}

func pkcs5Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("填充非法")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, errors.New("填充非法")
	}
	if !bytes.Equal(data[len(data)-pad:], bytes.Repeat([]byte{byte(pad)}, pad)) {
		return nil, errors.New("填充非法")
	}
	return data[:len(data)-pad], nil
}

// ---------- 通用工具 ----------

// ConstantTimeCompare 常量时间字符串比较。
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// RandomSalt 生成随机盐（供未来切换到 Argon2id/PBKDF2 时使用）。
func RandomSalt(n int) (string, error) {
	if n <= 0 {
		n = 16
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成随机盐失败: %w", err)
	}
	return base64.RawStdEncoding.EncodeToString(buf), nil
}

// SHA256Hex 供需要与外部系统对账时使用。
func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
