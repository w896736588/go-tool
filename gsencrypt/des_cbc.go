package gsencrypt

import (
	"encoding/base64"
	"github.com/forgoer/openssl"
	"github.com/spf13/cast"
)

type DesCbc struct {
	Key string
	Iv  string
}

// NewDesCbc create a des CBC object
func NewDesCbc(key, iv string) *DesCbc {
	return &DesCbc{
		Key: key,
		Iv:  iv,
	}
}

// Encrypt DesCBC加密
func (handle *DesCbc) Encrypt(src string) (string, error) {
	byteKey := []byte(handle.Key)
	for i := len(byteKey); i < 8; i++ {
		byteKey = append(byteKey, 0x00)
	}
	byteIv := []byte(handle.Iv)
	for i := len(byteIv); i < 8; i++ {
		byteIv = append(byteIv, 0x00)
	}
	byteRet, err := openssl.DesCBCEncrypt([]byte(src), byteKey, byteIv, openssl.PKCS7_PADDING)
	if err != nil {
		return ``, err
	}
	return cast.ToString(base64.StdEncoding.EncodeToString(byteRet)), nil
}

// Decrypt DesCBC解密
func (handle *DesCbc) Decrypt(src string) (string, error) {
	byteRet, err := openssl.DesCBCDecrypt([]byte(src), []byte(handle.Key), []byte(handle.Iv), openssl.PKCS7_PADDING)
	if err != nil {
		return ``, err
	}
	return cast.ToString(byteRet), nil
}
