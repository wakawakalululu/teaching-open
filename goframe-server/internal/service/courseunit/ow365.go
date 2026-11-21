// ow365.go 文件预览串生成（移植自 Java org.jeecg.modules.common.util.Ow365Util）。
//
// 规则（Ow365Util#getFileUrlStr）：入参为逗号分隔 key 串，逐个经 getFileUrl 拼
// 完整 URL；当 preview=ow365 且 owKey 有效（非空且非 "??"）且扩展名 ∈
// {ppt,pptx,xls,xlsx,doc,docx,pdf} 时，前缀 aess:(https)/aes:(其他) + 加密串；
// 加密为 DES/CBC/PKCS5Padding（key=owKey 前 8 字节，iv=owIv），密文 Base64 后
// URL 安全替换 +→_ /→* =→-。owIv 长度非 8 时 Java init 抛异常整体禁用，
// Go 等价 passthrough。
package courseunit

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"strings"
)

// Ow365 ow365 预览配置（preview/key/iv）。
type Ow365 struct {
	Preview string
	Key     string
	Iv      string
}

// owExtensions ow365 支持的扩展名白名单。
var owExtensions = []string{"ppt", "pptx", "xls", "xlsx", "doc", "docx", "pdf"}

// enabled 是否启用加密（对齐 Java init 成功一次的判定）。
func (o *Ow365) enabled() bool {
	return o != nil && o.Preview == "ow365" && o.Key != "" && o.Key != "??" && len(o.Iv) == 8
}

// FileURLStr 逗号分隔 key 串 → 预览串（每个元素先经 fileURL 拼完整地址）。
func (o *Ow365) FileURLStr(keys string, fileURLFn func(string) string) string {
	if keys == "" || keys == "null" {
		return ""
	}
	parts := strings.Split(keys, ",")
	out := make([]string, 0, len(parts))
	for _, key := range parts {
		if key == "" {
			continue
		}
		full := fileURLFn(key)
		if o.enabled() && owExtMatch(key) {
			prefix := "aes:"
			if strings.HasPrefix(full, "https://") {
				prefix = "aess:"
			}
			if enc, err := o.encrypt(full); err == nil {
				full = prefix + enc
			}
		}
		out = append(out, full)
	}
	return strings.Join(out, ",")
}

// owExtMatch 扩展名白名单判断（endsWith 语义）。
func owExtMatch(key string) bool {
	lower := strings.ToLower(key)
	for _, ext := range owExtensions {
		if strings.HasSuffix(lower, "."+ext) {
			return true
		}
	}
	return false
}

// encrypt DES-CBC/PKCS5Padding + URL 安全 Base64（+→_ /→* =→-）。
func (o *Ow365) encrypt(plain string) (string, error) {
	block, err := des.NewCipher([]byte(o.Key[:8]))
	if err != nil {
		return "", err
	}
	data := padPKCS5([]byte(plain), block.BlockSize())
	iv := []byte(o.Iv[:8])
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(data, data)
	return base64.StdEncoding.EncodeToString(data), nil
}

// padPKCS5 PKCS5 填充。
func padPKCS5(data []byte, blockSize int) []byte {
	padSize := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padSize)}, padSize)
	return append(data, pad...)
}
