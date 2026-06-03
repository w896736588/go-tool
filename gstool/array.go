package gstool

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdefine"
)

// ArrayFilterEmpty 过滤掉空值 false 0值 空字符串 空切片 都属于空
func ArrayFilterEmpty[T any](arrayList *[]T) []T {
	newList := make([]T, 0)
	for _, value := range *arrayList {
		if !AnyIsEmpty(value, ``) {
			newList = append(newList, value)
		}
	}
	return newList
}

// ArrayDeleteValue 删除切片中某个元素
func ArrayDeleteValue[T comparable](arrayList *[]T, deleteValue T) {
	index := ArrayFindValueIndex(arrayList, deleteValue)
	if index == -1 {
		return
	}
	*arrayList = append((*arrayList)[:index], (*arrayList)[index+1:]...)
}

// ArrayAppendNotExist 当元素不存在时加入到切片中
func ArrayAppendNotExist[T comparable](arrayList *[]T, appendValue T) {
	if !ArrayExistValue(arrayList, appendValue) {
		*arrayList = append(*arrayList, appendValue)
	}
}

// ArrayFindValueIndex 查找元素在切片中的位置
func ArrayFindValueIndex[T comparable](arrayList *[]T, appendValue T) int {
	for key, val := range *arrayList {
		if val == appendValue {
			return key
		}
	}
	return -1
}

// ArrayGetFromAny 将一个any类型数据转为切片
func ArrayGetFromAny(arrayList any) *[]GsCons {
	var returnList = make([]GsCons, 0)
	if arrayList == nil {
		return &returnList
	}
	if arrTrans, ok := arrayList.([]interface{}); ok {
		for _, val := range arrTrans {
			returnList = append(returnList, *ConsNew(val))
		}
		return &returnList
	} else {
		return &returnList
	}
}

// ArrayExistValue 判断是否在数组中
func ArrayExistValue[T comparable](array *[]T, value T) bool {
	if ArrayFindValueIndex(array, value) == -1 {
		return false
	}
	return true
}

// ArrayChunkRun 分组运行
func ArrayChunkRun[T any](source []T, chunkNum int, backFunc func([]T)) {
	sourceLen := len(source)
	runData := make([]T, 0)
	for i := 0; i < sourceLen; i++ {
		runData = append(runData, source[i])
		if len(runData) == chunkNum {
			backFunc(runData)
			runData = make([]T, 0)
		}
	}
	if len(runData) != 0 {
		backFunc(runData)
	}
}

// ArrayColumn 获取数组map中的某个key
func ArrayColumn(source *[]map[string]any, column string) []any {
	targetList := make([]any, 0)
	for _, value := range *source {
		if value[column] != nil {
			targetList = append(targetList, value[column])
		}
	}
	return targetList
}

// Array2Str interface转整数
func Array2Str[T any](source *[]T) []string {
	newList := make([]string, 0)
	for _, value := range *source {
		newList = append(newList, cast.ToString(value))
	}
	return newList
}

// Array2Int interface转整数
func Array2Int[T any](source *[]T) []int {
	newList := make([]int, 0)
	for _, value := range *source {
		newList = append(newList, cast.ToInt(value))
	}
	return newList
}

// Array2Int64 interface转整数
func Array2Int64[T any](source *[]T) []int64 {
	newList := make([]int64, 0)
	for _, value := range *source {
		newList = append(newList, cast.ToInt64(value))
	}
	return newList
}

// ArrayAIncludeAllBStr A数据中是否包含了B数组中的所有
func ArrayAIncludeAllBStr(sourceA *[]string, sourceB *[]string) bool {
	for _, valueB := range *sourceB {
		boolFind := false
		for _, valueA := range *sourceA {
			if valueB == valueA {
				boolFind = true
				break
			}
		}
		//未找到 那么说明A不包含B的所有
		if !boolFind {
			return false
		}
	}
	return true
}

// ArrayAIncludeAllBInt A数据中是否包含了B数组中的所有
func ArrayAIncludeAllBInt(sourceA *[]int, sourceB *[]int) bool {
	for _, valueB := range *sourceB {
		boolFind := false
		for _, valueA := range *sourceA {
			if valueB == valueA {
				boolFind = true
				break
			}
		}
		//未找到 那么说明A不包含B的所有
		if !boolFind {
			return false
		}
	}
	return true
}

// ArrayANotExistAnyBStr A数据中不包含B数组中的任意一个
func ArrayANotExistAnyBStr(sourceA *[]string, sourceB *[]string) bool {
	for _, valueB := range *sourceB {
		boolFind := false
		for _, valueA := range *sourceA {
			if valueB == valueA {
				boolFind = true
				break
			}
		}
		if boolFind {
			return false
		}
	}
	return true
}

// ArrayANotExistAnyBInt A数据中不存在B数组中的任意一个
func ArrayANotExistAnyBInt(sourceA *[]int, sourceB *[]int) bool {
	for _, valueB := range *sourceB {
		boolFind := false
		for _, valueA := range *sourceA {
			if valueB == valueA {
				boolFind = true
				break
			}
		}
		if boolFind {
			return false
		}
	}
	return true
}

// ArrayContainString 检测字符串是否在数组中 包含
func ArrayContainString(search string, sourceList *[]string) bool {
	for _, source := range *sourceList {
		if strings.Contains(source, search) {
			return true
		}
	}
	return false
}

// ArrayWalk 遍历处理数组
func ArrayWalk[T any](sourceList *[]T, walk func(int, T) T) {
	for key, value := range *sourceList {
		(*sourceList)[key] = walk(key, value)
	}
}

// ArraySort 根据值排序
func ArraySort[T comparable](m []T, sortType string) []T {
	// 对切片进行降序排序
	sort.Slice(m, func(i, j int) bool {
		if sortType == gsdefine.SortDesc {
			return cast.ToString(m[i]) > cast.ToString(m[j])
		} else {
			return cast.ToString(m[i]) < cast.ToString(m[j])
		}
	})
	return m
}

// ArrayMapSort 根据key排序
func ArrayMapSort[T comparable](arrayMap *[]map[string]T, field, sortType string) {
	sort.Slice(*arrayMap, func(i, j int) bool {
		if sortType == gsdefine.SortDesc {
			return cast.ToString((*arrayMap)[i][field]) > cast.ToString((*arrayMap)[j][field])
		} else {
			return cast.ToString((*arrayMap)[i][field]) < cast.ToString((*arrayMap)[j][field])
		}
	})
}

// ArrayMapFilterContainField 数组中的map包含某个值
func ArrayMapFilterContainField[T any](arrayList *[]map[string]T, field string, value T) {
	newList := make([]map[string]T, 0)
	for _, mapValue := range *arrayList {
		if strings.Contains(cast.ToString(mapValue[field]), cast.ToString(value)) {
			newList = append(newList, mapValue)
		}
	}
	*arrayList = newList
}

// ArrayMapFilterField 数组中的map等于某个值
func ArrayMapFilterField[T any](arrayList *[]map[string]any, field string, value T) {
	newList := make([]map[string]any, 0)
	for _, mapValue := range *arrayList {
		if cast.ToString(mapValue[field]) == cast.ToString(value) {
			newList = append(newList, mapValue)
		}
	}
	*arrayList = newList
}

// ArrayRandValue 随机返回一个值
func ArrayRandValue[T any](arrayList []T) T {
	if len(arrayList) == 0 {
		var zero T
		return zero
	}
	// 使用当前时间作为随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return arrayList[r.Intn(len(arrayList))]
}

func ArrayToSliceAny(value any) ([]any, bool) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice {
		return nil, false
	}

	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	return result, true
}
func ArrayWalkDesc[T any](arrayList []T, call func(T) bool) {
	for i := len(arrayList) - 1; i >= 0; i-- {
		isContinue := call(arrayList[i])
		if !isContinue {
			break
		}
	}
}

// ArrayReverseSlice 反转切片
func ArrayReverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
