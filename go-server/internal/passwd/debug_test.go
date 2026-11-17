package passwd

import (
	"encoding/hex"
	"fmt"
	"testing"
)

// 调试：打印 Go 端派生出的 key/iv 与解密出的原始字节，与参考实现比对。
func TestDebugLegacyDecrypt(t *testing.T) {
	p := NewLegacyPBE()
	key, iv, err := p.keyAndIV("123456", "63293188")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("key=%s iv=%s\n", hex.EncodeToString(key), hex.EncodeToString(iv))

	cipherHex := "c1461cf4492050e9"
	plain, err := p.Decrypt(cipherHex, "123456", "63293188")
	fmt.Printf("decrypt result=%q err=%v\n", plain, err)
}
