// buffer.go 轻量字节缓冲（供 PNG 编码）与 base64 助手。
package captcha

import (
	"bytes"
	"encoding/base64"
)

// bytesBuffer bytes.Buffer 别名包装。
type bytesBuffer struct{ buf bytes.Buffer }

func (b *bytesBuffer) Write(p []byte) (int, error) { return b.buf.Write(p) }
func (b *bytesBuffer) bytes() []byte               { return b.buf.Bytes() }

// base64Std 标准 base64 编码。
func base64Std(data []byte) string { return base64.StdEncoding.EncodeToString(data) }
