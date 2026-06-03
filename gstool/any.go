package gstool

import (
	"go/types"

	"github.com/spf13/cast"
)

// AnyIsEmpty 判断是否为空 空字符串 数字0 false 空切片 空map
func AnyIsEmpty(value any, typeStr string) bool {
	if typeStr == `` {
		typeStr = AnyTypeString(value)
	}
	switch typeStr {
	case `bool`:
		if cast.ToBool(value) == false {
			return true
		}
	case `int`:
		if cast.ToInt(value) == 0 {
			return true
		}
	case `float`:
		if cast.ToFloat32(value) == 0 {
			return true
		}
	case `string`:
		if cast.ToString(value) == `` {
			return true
		}
	case `array`:
		if slice, ok := value.([]any); ok {
			if len(slice) == 0 {
				return true
			}
		}
	case `map`:
		if mapValue, ok := value.(map[string]any); ok {
			if len(mapValue) == 0 {
				return true
			}
		}
	}

	return false
}

// AnyTypeString 获取类型
func AnyTypeString(value any) string {
	switch value.(type) {
	case bool:
		return `bool`
	case int, int8, int16, int32, int64, uint, uint16, uint32, uint64:
		return `int`
	case float32, float64:
		return `float`
	case string:
		return `string`
	case types.Array:
		return `array`
	case types.Map:
		return `map`
	}
	return ``
}
