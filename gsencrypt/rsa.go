package gsencrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

type Rsa2048 struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// NewRsa2048 初始化 注意秘钥应该是----开头结尾的
func NewRsa2048(publicKey, privateKey string) (*Rsa2048, error) {
	rsa2048 := &Rsa2048{}
	//装载公钥
	if publicKey != `` {
		blockPub, _ := pem.Decode([]byte(publicKey))
		if blockPub == nil || blockPub.Type != "PUBLIC KEY" {
			return nil, gstool.Error("cs1 failed to decode PEM block containing public key %#v", blockPub)
		}
		PublicKeyObj, PublicKeyErr := x509.ParsePKIXPublicKey(blockPub.Bytes)
		if PublicKeyErr != nil {
			return nil, PublicKeyErr
		}
		rsa2048.PublicKey = PublicKeyObj.(*rsa.PublicKey)
	}

	//装载私钥
	if privateKey != `` {
		if rsa2048.isCs1(privateKey) {
			blockPri, _ := pem.Decode([]byte(privateKey))
			gstool.FmtPrintlnLogTime(`%#v`, blockPri)
			if blockPri == nil || blockPri.Type != "RSA PRIVATE KEY" {
				return nil, errors.New("failed to decode PEM block containing private key")
			}
			privateKeyObj, privateKeyErr := x509.ParsePKCS1PrivateKey(blockPri.Bytes)
			if privateKeyErr != nil {
				return nil, privateKeyErr
			}
			rsa2048.PrivateKey = privateKeyObj
		} else {
			blockPri, _ := pem.Decode([]byte(privateKey))
			if blockPri == nil || blockPri.Type != "PRIVATE KEY" {
				return nil, gstool.Error("cs8初始化失败，failed to decode PEM block containing private key %#v", blockPri)
			}
			privateKeyObjTemp, privateKeyErr := x509.ParsePKCS8PrivateKey(blockPri.Bytes)
			if privateKeyErr != nil {
				return nil, privateKeyErr
			}

			privateKeyObj, ok := privateKeyObjTemp.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("private key type is not rsa cs8")
			}
			rsa2048.PrivateKey = privateKeyObj
		}
	}
	return rsa2048, nil
}

// Rsa2048GeneralCs1Key 生成cs1格式的公私钥 推荐2048
func Rsa2048GeneralCs1Key(bits int) (string, string, error) {
	privateKey, privateKeyErr := rsa.GenerateKey(rand.Reader, bits)
	if privateKeyErr != nil {
		return "", "", privateKeyErr
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicKeyBytes, publicKeyBytesErr := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if publicKeyBytesErr != nil {
		return "", "", publicKeyBytesErr
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return cast.ToString(privateKeyPEM), cast.ToString(pubKeyPEM), nil
}

// Rsa2048GeneralCs8Key 生成cs8格式的公私钥 推荐2048
func Rsa2048GeneralCs8Key(bits int) (string, string, error) {
	privateKey, privateKeyErr := rsa.GenerateKey(rand.Reader, bits)
	if privateKeyErr != nil {
		return "", "", privateKeyErr
	}
	cs8Ret, cs8Err := x509.MarshalPKCS8PrivateKey(privateKey)
	if cs8Err != nil {
		return ``, ``, cs8Err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: cs8Ret,
	})
	publicKeyBytes, publicKeyBytesErr := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if publicKeyBytesErr != nil {
		return "", "", publicKeyBytesErr
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return cast.ToString(privateKeyPEM), cast.ToString(pubKeyPEM), nil
}

// EncryptCs1v15 加密 注意：如果是base64的需要进行base64反解析
func (h *Rsa2048) EncryptCs1v15(plaintext []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, h.PublicKey, plaintext)
}

// DecryptCs1v15 解密注意如果是base64需要反解后传入
func (h *Rsa2048) DecryptCs1v15(decryptData []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, h.PrivateKey, decryptData)
}

// EncryptOaep 加密 注意：如果是base64的需要进行base64反解析
func (h *Rsa2048) EncryptOaep(plaintext []byte, label []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, h.PublicKey, plaintext, label)
}

// DecryptOaep 解密注意如果是base64需要反解后传入
func (h *Rsa2048) DecryptOaep(decryptData []byte, label []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, h.PrivateKey, decryptData, label)
}

func (h *Rsa2048) isCs1(key string) bool {
	if strings.Contains(key, `BEGIN RSA`) {
		return true
	}
	return false
}
