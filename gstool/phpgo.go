package gstool

import (
	"github.com/leeqvip/gophp/serialize"
	"github.com/spf13/cast"
)

// PhpSerialize 序列化
func PhpSerialize(str string) (string, error) {
	out, err := serialize.Marshal([]byte(str))
	if err != nil {
		return ``, err
	}
	return cast.ToString(out), nil
}

// PhpUnSerialize 反序列化
func PhpUnSerialize(str string) (interface{}, error) {
	out, err := serialize.UnMarshal([]byte(str))
	if err != nil {
		return ``, err
	}
	return out, nil
}
