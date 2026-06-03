package gstool

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdefine"
)

// MapKeyExist key是否存在
func MapKeyExist[T any](mapValue *map[string]T, key string) bool {
	if _, ok := (*mapValue)[key]; ok {
		return true
	}
	return false
}

// MapCopy 复制map
func MapCopy[T any](sourceMap map[string]T) map[string]T {
	copyMap := make(map[string]T)
	for k, v := range sourceMap {
		copyMap[k] = v
	}
	return copyMap
}

// MapToHttpUrl map转为url请求请求参数
func MapToHttpUrl(requestMap *map[string]string, boolUrlEncode bool) string {
	returnArr := make([]string, 0)
	for key, value := range *requestMap {
		if boolUrlEncode {
			value = url.QueryEscape(value)
		}
		returnArr = append(returnArr, fmt.Sprintf(`%s=%s`, key, value))
	}
	return strings.Join(returnArr, `&`)
}

// MapGetKeys 获取map的key列表
func MapGetKeys[T any](requestMap *map[string]T) []string {
	returnList := make([]string, 0)
	for key, _ := range *requestMap {
		returnList = append(returnList, key)
	}
	return returnList
}

// MapTakeKeys 获取map的key列表
func MapTakeKeys[T comparable](m *map[T]any, keys []T) map[T]any {
	returnMap := make(map[T]any, 0)
	for key, value := range *m {
		if ArrayExistValue(&keys, key) {
			returnMap[key] = value
		}
	}
	return returnMap
}

// MapSortByKey 根据key排序
func MapSortByKey[T comparable](m map[T]any, sortType string) map[T]any {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	nM := make(map[T]any)
	// 对切片进行降序排序
	sort.Slice(keys, func(i, j int) bool {
		if sortType == gsdefine.SortDesc {
			return cast.ToString(keys[i]) > cast.ToString(keys[j])
		} else {
			return cast.ToString(keys[i]) < cast.ToString(keys[j])
		}
	})

	// 按排序后的键遍历 map
	for _, k := range keys {
		nM[k] = m[k]
	}
	return nM
}
