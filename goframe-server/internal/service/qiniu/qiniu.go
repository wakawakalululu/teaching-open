// Package qiniu 七牛上传令牌生成（移植自 Java QiniuController + qiniu SDK
// Auth.uploadToken 的令牌格式）。
//
// 令牌格式：AccessKey + ':' + sign + ':' + encodedPutPolicy，其中
// encodedPutPolicy = urlsafe-base64(JSON putPolicy)、sign = HMAC-SHA1
// （secretKey，对 encodedPutPolicy）的十六进制小写。过期时间固定 3600s
// （对齐 Java expires 常量）。
package qiniu

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Config 七牛凭证与桶配置（未配置 AK 时生成空 AccessKey 令牌，与 Java
// 空配置行为一致）。
type Config struct {
	AccessKey string
	SecretKey string
	Bucket    string
}

// putPolicy 上传策略（对齐 SDK 默认字段：scope/deadline；key 限定覆盖上传）。
type putPolicy struct {
	Scope      string `json:"scope"`
	Deadline   int64  `json:"deadline"`
	InsertOnly int    `json:"insertOnly,omitempty"`
}

// UploadToken 生成上传令牌；key 非空时限定了覆盖上传的 scope。
func (c Config) UploadToken(key string) string {
	scope := c.Bucket
	if key != "" {
		scope = c.Bucket + ":" + key
	}
	policy := putPolicy{
		Scope:    scope,
		Deadline: time.Now().Add(3600 * time.Second).Unix(),
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return ""
	}
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(policyJSON)
	sign := hmacSHA1Hex(c.SecretKey, encoded)
	return strings.Join([]string{c.AccessKey, sign, encoded}, ":")
}

// hmacSHA1Hex HMAC-SHA1 十六进制小写签名（与 qiniu SDK Mac.Sign 一致）。
func hmacSHA1Hex(secret, data string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(data))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
