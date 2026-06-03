package gsencrypt

import (
	"bytes"
	"errors"
)

// Pkcs5Padding PKCS#5填充
func Pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// Pkcs5Unpadding PKCS#5去填充
func Pkcs5Unpadding(ciphertext []byte) []byte {
	padding := int(ciphertext[len(ciphertext)-1])
	return ciphertext[:len(ciphertext)-padding]
}

// Pkcs7UnPadding 填充的反向操作
func Pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	// 获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

// Pkcs7Padding 填充
func Pkcs7Padding(data []byte, blockSize int) []byte {
	// 判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	// 补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func padKv(key string, length int) []byte {
	rightKey := key
	if len(rightKey) > length {
		rightKey = rightKey[:length]
	} else if len(rightKey) < length {
		for i := 0; i < length; i++ {
			rightKey = rightKey + key
			if len(rightKey) > length {
				rightKey = rightKey[:length]
				break
			}
		}
	}
	return []byte(rightKey)
}
