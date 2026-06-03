package gstool

import "reflect"

// ReflectGetType 获取类型
func ReflectGetType(item any) reflect.Kind {
	return reflect.TypeOf(item).Kind()
}

func ReflectIsString(item any) bool {
	_, ok := item.(string)
	return ok
}

func ReflectTypeString(item any) string {
	reval := reflect.TypeOf(item)
	if reval == nil {
		return ``
	}
	return reval.String()
}
