package gstool

import (
	"encoding/base64"
)

func Base64Encode(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func Base64Decode(content string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(content)
}
