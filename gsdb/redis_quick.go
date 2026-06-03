package gsdb

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdefine"
	"github.com/w896736588/go-tool/gstool"
)

type tBase interface {
	string | int | int64 | float64
}

type tMap interface {
	map[string]any | map[string]string | map[string]int | map[string]int64 | map[string]float64 |
		map[int]string | map[int]int | map[int]int64 | map[int]float64 | map[int]any |
		map[int64]string | map[int64]int | map[int64]int64 | map[int64]float64 | map[int64]any |
		map[float64]string | map[float64]int | map[float64]int64 | map[float64]float64 | map[float64]any
}

type tExtend interface {
	tBase | tMap
}

func hashToMap[TKey tBase, TVal tExtend](dataMap map[string]string) (data map[TKey]TVal, err error) {
	data = make(map[TKey]TVal)
	defer func() {
		if r := recover(); r != nil {
			err = gstool.Error(`panic hash to map error : %v`, r)
			data = make(map[TKey]TVal)
		}
	}()
	if dataMap == nil {
		return
	}
	newMap := make(map[TKey]TVal)
	for mapKey, mapVal := range dataMap {
		newKey, err := convertTBase[TKey](mapKey)
		if err != nil {
			return newMap, err
		}
		newVal, err := convertTExtend[TVal](mapVal)
		if err != nil {
			return newMap, err
		}
		newMap[newKey] = newVal
	}
	data = newMap
	return
}

func convertTBase[T tBase](value any) (T, error) {
	var t T
	var val any
	typeStr := reflect.TypeOf(t).String()
	switch typeStr {
	case `string`:
		val = cast.ToString(value)
	case `int`:
		val = cast.ToInt(value)
	case `int64`:
		val = cast.ToInt64(value)
	case `float64`:
		val = cast.ToFloat64(value)
	default:
		return t, gstool.Error(`类型转换失败,不支持的转换类型 %s`, typeStr)
	}
	result, ok := val.(T)
	if !ok {
		return t, gstool.Error(`类型断言失败，预期类型 %s，实际类型 %T，原始值 %#v`, typeStr, val, value)
	}
	return result, nil
}

func getZeroMap[T tExtend](typeStr string) (T, error) {
	var zero any
	switch typeStr {
	// map[string]... 系列
	case `map[string]any`, `map[string]interface {}`:
		zero = make(map[string]any)
	case `map[string]string`:
		zero = make(map[string]string)
	case `map[string]int`:
		zero = make(map[string]int)
	case `map[string]int64`:
		zero = make(map[string]int64)
	case `map[string]float64`:
		zero = make(map[string]float64)

	// map[int]... 系列
	case `map[int]any`, `map[int]interface {}`:
		zero = make(map[int]any)
	case `map[int]string`:
		zero = make(map[int]string)
	case `map[int]int`:
		zero = make(map[int]int)
	case `map[int]int64`:
		zero = make(map[int]int64)
	case `map[int]float64`:
		zero = make(map[int]float64)

	// map[int64]... 系列
	case `map[int64]any`, `map[int64]interface {}`:
		zero = make(map[int64]any)
	case `map[int64]string`:
		zero = make(map[int64]string)
	case `map[int64]int`:
		zero = make(map[int64]int)
	case `map[int64]int64`:
		zero = make(map[int64]int64)
	case `map[int64]float64`:
		zero = make(map[int64]float64)

	// map[float64]... 系列
	case `map[float64]any`, `map[float64]interface {}`:
		zero = make(map[float64]any)
	case `map[float64]string`:
		zero = make(map[float64]string)
	case `map[float64]int`:
		zero = make(map[float64]int)
	case `map[float64]int64`:
		zero = make(map[float64]int64)
	case `map[float64]float64`:
		zero = make(map[float64]float64)

	default:
		var t T
		return t, gstool.Error(`不支持创建该类型map，不支持的类型 %s`, typeStr)
	}
	ret, ok := zero.(T)
	if !ok {
		var t T
		return t, gstool.Error(`类型断言失败，预期类型 %s`, typeStr)
	}
	return ret, nil
}

func convertTExtend[T tExtend](value string) (T, error) {
	var t T
	var val any
	typeStr := reflect.TypeOf(t).String()
	switch typeStr {
	case `string`:
		val = cast.ToString(value)
	case `int`:
		val = cast.ToInt(value)
	case `int64`:
		val = cast.ToInt64(value)
	case `float64`:
		val = cast.ToFloat64(value)
	}
	//基础类型
	if !strings.HasPrefix(typeStr, `map[`) {
		result, ok := val.(T)
		if !ok {
			return t, gstool.Error(`基础类型断言失败，预期类型 %s，实际类型 %T，原始值 %#v`, typeStr, val, value)
		}
		return result, nil
	}
	//扩展类型
	return stringJsonMap[T](typeStr, value)
}

func stringJsonMap[T tExtend](targetType string, value string) (T, error) {
	zero, err := getZeroMap[T](targetType)
	if err != nil {
		return zero, err
	}
	err = gstool.JsonDecode(value, &zero)
	if err != nil {
		return zero, gstool.Error(`error：json解码失败,%v -> %s`, value, targetType)
	}
	return zero, nil
}

// RedisGetMapHashAll 在hash中构建和获取map
func RedisGetMapHashAll[TKey tBase, TVal tExtend](
	cli *GsRedis, key string, queryFunc func() (map[TKey]TVal, error), expired time.Duration) (map[TKey]TVal, error) {
	if key == `` {
		return make(map[TKey]TVal), gstool.Error(`Redis操作key不能为空`)
	}
	if queryFunc == nil {
		return make(map[TKey]TVal), gstool.Error(`查询函数queryFunc为nil，无法获取数据`)
	}
	existInt, err := cli.Client.Exists(context.Background(), key).Result()
	if err != nil {
		return make(map[TKey]TVal), gstool.Error(`Redis判断key是否存在失败，key %s，错误 %s`, key, err.Error())
	}
	if existInt > 0 {
		result, resultErr := cli.Client.HGetAll(context.Background(), key).Result()
		if resultErr != nil {
			return make(map[TKey]TVal), gstool.Error(`Redis获取hash数据失败，key %s，错误 %s`, key, resultErr.Error())
		}
		if result != nil {
			delete(result, gsdefine.EmptyKey)
		}
		return hashToMap[TKey, TVal](result)
	}
	info, infoErr := queryFunc()
	if infoErr != nil {
		return make(map[TKey]TVal), gstool.Error(`获取数据库数据异常，错误 %s`, infoErr.Error())
	}
	if info == nil || len(info) == 0 {
		if hsetErr := cli.Client.HSet(context.Background(), key, gsdefine.EmptyKey, gsdefine.EmptyValue).Err(); hsetErr != nil {
			return make(map[TKey]TVal), gstool.Error(`Redis写入默认空值失败，key %s，错误 %s`, key, hsetErr.Error())
		}
		if expired > 0 {
			if expireErr := cli.Client.Expire(context.Background(), key, expired).Err(); expireErr != nil {
				return make(map[TKey]TVal), gstool.Error(`Redis设置key过期时间失败，key %s，错误 %s`, key, expireErr.Error())
			}
		}
		return make(map[TKey]TVal), nil
	}
	setList := make([]any, 0, len(info)*2)
	for infoKey, infoVal := range info {
		if strings.HasPrefix(gstool.ReflectTypeString(infoVal), `map[`) {
			setList = append(setList, infoKey, gstool.JsonEncode(infoVal))
		} else {
			setList = append(setList, infoKey, infoVal)
		}
	}
	if hmsetErr := cli.Client.HMSet(context.Background(), key, setList...).Err(); hmsetErr != nil {
		return info, gstool.Error(`Redis批量写入hash数据失败，key %s，错误 %s`, key, hmsetErr.Error())
	}
	if expired > 0 {
		if expireErr := cli.Client.Expire(context.Background(), key, expired).Err(); expireErr != nil {
			return info, gstool.Error(`Redis设置key过期时间失败，key %s，错误 %s`, key, expireErr.Error())
		}
	}
	return info, nil
}

// RedisGetMapString 在string中构建和获取map
func RedisGetMapString[TKey tBase, TVal tExtend](cli *GsRedis, key string, queryFunc func() (map[TKey]TVal, error), expired time.Duration) (map[TKey]TVal, error) {
	var zero = make(map[TKey]TVal)
	typeStr := reflect.TypeOf(zero).String()
	cacheInfo, err := cli.Client.Get(context.Background(), key).Result()
	if !errors.Is(err, redis.Nil) && err != nil { //报错了
		return zero, err
	}
	if !errors.Is(err, redis.Nil) { //缓存存在
		if cacheInfo == gsdefine.EmptyValue { //空值 防止穿透
			return zero, nil
		} else {
			err = gstool.JsonDecode(cacheInfo, &zero)
			if err != nil {
				gstool.FmtPrintlnLogTime(`error：json解码失败,%v -> %s`, cacheInfo, typeStr)
				return zero, nil
			}
			return zero, nil
		}
	}
	//缓存没有 查库
	info, infoErr := queryFunc()
	if infoErr != nil {
		return zero, gstool.Error(`获取db数据异常 %s`, infoErr.Error())
	}
	var infoJson string
	if info == nil || len(info) == 0 { //数据库数据为空 给一个空值
		infoJson = gsdefine.EmptyValue
	} else {
		infoJson = gstool.JsonEncode(info)
	}
	setErr := cli.Client.Set(context.Background(), key, infoJson, expired).Err()
	if setErr != nil {
		return zero, gstool.Error(`设置缓存异常 %s`, setErr.Error())
	}
	return info, nil
}
