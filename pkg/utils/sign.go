package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
)

// ValidateSign 验证签名
// timestamp: 时间戳
// sign: 已经 URL 编码的签名
// secret: 密钥
func ValidateSign(timestamp, sign, secret string) bool {
	// 1. URL decode签名
	decodedSign, err := url.QueryUnescape(sign)
	if err != nil {
		return false
	}

	// 2. 计算预期的签名
	signStr := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signStr))

	// 3. Base64 编码
	expectedSign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return decodedSign == expectedSign
}

// GenerateSign 生成签名 (用于测试)
func GenerateSign(timestamp, secret string) string {
	// 1. 计算签名
	signStr := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signStr))

	// 2. Base64 编码
	b64Sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 3. URL 编码
	return url.QueryEscape(b64Sign)
}
