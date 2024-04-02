package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// hash plaintext with md5
func EncryptMd5(source string) (cryptext string) {
	if len(source) <= 0 {
		return ""
	}
	// 使用MD5加密
	signBytes := md5.Sum([]byte(source))
	// 把二进制转化为大写的十六进制
	cryptext = hex.EncodeToString(signBytes[:])
	return
}

// hash plaintext with md5
func EncryptMd5Byte(source []byte) (cryptext string) {
	// 使用MD5加密
	signBytes := md5.Sum(source)
	// 把二进制转化为大写的十六进制
	cryptext = hex.EncodeToString(signBytes[:])
	return
}
