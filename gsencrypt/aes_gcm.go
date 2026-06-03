package gsencrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/spf13/cast"
)

type AesGcm struct {
	key []byte
}

func NewAesGcm(key string) *AesGcm {
	aesGcm := &AesGcm{
		key: padKv(key, 32),
	}
	return aesGcm
}

// Encrypt 加密
func (h *AesGcm) Encrypt(plaintext []byte) (string, error) {
	block, blockErr := aes.NewCipher(h.key)
	if blockErr != nil {
		return "", blockErr
	}
	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return "", gcmErr
	}
	blockSize := block.BlockSize()
	plaintext = Pkcs7Padding(plaintext, blockSize)
	nonce := make([]byte, gcm.NonceSize())
	if _, randErr := rand.Read(nonce); randErr != nil {
		return "", randErr
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (h *AesGcm) Decrypt(decryptData string) (string, error) {
	ciphertextBytes, ciphertextBytesErr := base64.URLEncoding.DecodeString(decryptData)
	if ciphertextBytesErr != nil {
		return ``, ciphertextBytesErr
	}
	block, blockErr := aes.NewCipher(h.key)
	if blockErr != nil {
		return ``, blockErr
	}
	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return ``, gcmErr
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return ``, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plaintext, plaintextErr := gcm.Open(nil, nonce, ciphertext, nil)
	if plaintextErr != nil {
		return ``, plaintextErr
	}
	plaintext, unpaddErr := Pkcs7UnPadding(plaintext)
	if unpaddErr != nil {
		return ``, unpaddErr
	}
	return cast.ToString(plaintext), nil
}
