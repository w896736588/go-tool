package gsdb

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/w896736588/go-tool/gsdefine"
	"github.com/w896736588/go-tool/gstool"
)

// RedisHash 获取hash中的某一个值
// Deprecated:  请使用 RedisGetMapHash
func RedisHash(client *GsRedis, key, subKey string, call func() map[string]string, expire time.Duration) string {
	getRet, getErr := client.Client.HGet(context.Background(), key, subKey).Result()
	if errors.Is(getErr, redis.Nil) {
		callRet := call()
		hsetData := make([]interface{}, 0)
		for mapKey, mapValue := range callRet {
			hsetData = append(hsetData, mapKey, mapValue)
		}
		if len(hsetData) == 0 {
			hsetData = append(hsetData, `-1`, `-1`)
		}
		_, setErr := client.Client.HSet(context.Background(), key, hsetData...).Result()
		if setErr != nil {
			gstool.FmtPrintlnLog(`RedisHash HMSet error : %s`, setErr.Error())
			return ``
		}
		return callRet[subKey]
	}
	if getErr != nil {
		gstool.FmtPrintlnLog(`RedisHash HGet error : %s`, getErr.Error())
		return ``
	}
	return getRet
}

// RedisHashE 获取hash中的某一个值
// Deprecated:  请使用 RedisGetMapHash
func RedisHashE(client *GsRedis, key, subKey string, call func() map[string]string, expire time.Duration) (string, error) {
	getRet, getErr := client.Client.HGet(context.Background(), key, subKey).Result()
	if errors.Is(getErr, redis.Nil) {
		callRet := call()
		hsetData := make([]interface{}, 0)
		for mapKey, mapValue := range callRet {
			hsetData = append(hsetData, mapKey, mapValue)
		}
		if len(hsetData) == 0 {
			hsetData = append(hsetData, `-1`, `-1`)
		}
		_, setErr := client.Client.HSet(context.Background(), key, hsetData...).Result()
		if setErr != nil {
			return ``, setErr
		}
		return callRet[subKey], nil
	}
	if getErr != nil {
		return ``, getErr
	}
	return getRet, nil
}

// RedisHashAll 获取hash中的所有值
// Deprecated:  请使用 RedisGetMapHash
func RedisHashAll(client *GsRedis, key string, call func() map[string]string, expire time.Duration) map[string]string {
	boolExist, boolExistErr := RedisKeyExist(client, key)
	if boolExistErr != nil {
		gstool.FmtPrintlnLog(`判断是否存在出错 %s`, boolExistErr.Error())
		return make(map[string]string)
	}
	if !boolExist {
		callRet := call()
		hsetData := make([]interface{}, 0)
		for mapKey, mapValue := range callRet {
			hsetData = append(hsetData, mapKey, mapValue)
		}
		if len(hsetData) == 0 {
			hsetData = append(hsetData, `-1`, `-1`)
		}
		_, setErr := client.Client.HSet(context.Background(), key, hsetData...).Result()
		if setErr != nil {
			gstool.FmtPrintlnLog(`设置缓存出错了 %s`, setErr.Error())
			return make(map[string]string)
		}
		//设置有效期
		expireErr := RedisSetExpire(client, key, expire)
		if expireErr != nil {
			gstool.FmtPrintlnLog(`设置有效期出错 %s`, expireErr.Error())
			return make(map[string]string)
		}
		return callRet
	}
	getRet, getErr := client.Client.HGetAll(context.Background(), key).Result()
	if getErr != nil {
		gstool.FmtPrintlnLog(`RedisHashAll HGetAll出错 %s`, getErr.Error())
		return make(map[string]string)
	}
	return getRet
}

// RedisHashAllE 获取hash中的所有值
// Deprecated:  请使用 RedisGetMapHash
func RedisHashAllE(client *GsRedis, key string, call func() map[string]string, expire time.Duration) (map[string]string, error) {
	boolExist, boolExistErr := RedisKeyExist(client, key)
	if boolExistErr != nil {
		return make(map[string]string), boolExistErr
	}
	if !boolExist {
		callRet := call()
		hsetData := make([]interface{}, 0)
		for mapKey, mapValue := range callRet {
			hsetData = append(hsetData, mapKey, mapValue)
		}
		if len(hsetData) == 0 {
			hsetData = append(hsetData, `-1`, `-1`)
		}
		_, setErr := client.Client.HSet(context.Background(), key, hsetData...).Result()
		if setErr != nil {
			return nil, setErr
		}
		//设置有效期
		expireErr := RedisSetExpire(client, key, expire)
		if expireErr != nil {
			return make(map[string]string), expireErr
		}
		return callRet, nil
	}
	getRet, getErr := client.Client.HGetAll(context.Background(), key).Result()
	if getErr != nil {
		return nil, getErr
	}
	return getRet, nil
}

// RedisGetKey 获取key的值
// Deprecated:  不再使用
func RedisGetKey(client *GsRedis, key string) string {
	getRet, getErr := client.Client.Get(context.Background(), key).Result()
	if errors.Is(getErr, redis.Nil) {
		return ``
	}
	if getErr != nil {
		gstool.FmtPrintlnLog(`RedisKey Get error : %s`, getErr.Error())
		return ``
	}
	return getRet
}

// RedisGetKeyE 获取key的值
// Deprecated:  不再使用
func RedisGetKeyE(client *GsRedis, key string) (string, error) {
	return client.Client.Get(context.Background(), key).Result()
}

// RedisGetStringCall 获取redis中的数据
// Deprecated:  不再使用
func RedisGetStringCall(client *GsRedis, key string, call func() string, expire time.Duration) string {
	ret, retErr := RedisGetKeyE(client, key)
	if errors.Is(retErr, redis.Nil) {
		result := call()
		if result == `` {
			client.Client.Set(context.Background(), key, gsdefine.RedisNotExist, expire)
		} else {
			client.Client.Set(context.Background(), key, call(), expire)
		}
		return result
	}
	if retErr != nil {
		return ``
	}
	if ret == gsdefine.RedisNotExist {
		return ``
	}
	return ret
}

// RedisGetStringCallE 获取redis中的数据
// Deprecated:  不再使用
func RedisGetStringCallE(client *GsRedis, key string, call func() string, expire time.Duration) (string, error) {
	ret, retErr := RedisGetKeyE(client, key)
	if errors.Is(retErr, redis.Nil) {
		result := call()
		if result == `` {
			client.Client.Set(context.Background(), key, gsdefine.RedisNotExist, expire)
		} else {
			client.Client.Set(context.Background(), key, call(), expire)
		}
		return result, nil
	}
	if retErr != nil {
		return ``, retErr
	}
	if ret == gsdefine.RedisNotExist {
		return ``, nil
	}
	return ret, nil
}

// RedisSetNx 获取锁
// Deprecated:  不再使用
func RedisSetNx(client *GsRedis, key, value string, expire time.Duration) bool {
	setNxRet, setNxErr := client.Client.SetNX(context.Background(), key, value, expire).Result()
	if setNxErr != nil {
		gstool.FmtPrintlnLog(`RedisSetNx SetNX error : %s`, setNxErr.Error())
		return false
	}
	return setNxRet
}

// RedisSetNxE 获取锁
// Deprecated:  不再使用
func RedisSetNxE(client *GsRedis, key, value string, expire time.Duration) (bool, error) {
	setNxRet, setNxErr := client.Client.SetNX(context.Background(), key, value, expire).Result()
	if setNxErr != nil {
		return false, setNxErr
	}
	return setNxRet, nil
}

// RedisKeyExist 判断key是否存在
// Deprecated:  不再使用
func RedisKeyExist(client *GsRedis, key string) (bool, error) {
	exists, err := client.Client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	if exists == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

// RedisSetExpire 设置缓存有效期
// Deprecated:  不再使用
func RedisSetExpire(client *GsRedis, key string, expire time.Duration) error {
	return client.Client.Expire(context.Background(), key, expire).Err()
}

//新版本的4个方法

// BaseStringCache map[string]any -> string
// 将一个map[string]any 存为 string
// Deprecated:  请使用 RedisGetMapString
func BaseStringCache(cli *GsRedis, key string, queryFunc func() (map[string]any, error), expired time.Duration) (map[string]any, error) {
	cacheInfo, err := cli.Client.Get(context.Background(), key).Result()
	if !errors.Is(err, redis.Nil) && err != nil { //报错了
		return nil, err
	}
	if !errors.Is(err, redis.Nil) { //缓存存在
		if cacheInfo == gsdefine.EmptyValue { //空值 防止穿透
			return map[string]any{}, nil
		} else {
			data := make(map[string]any)
			dErr := gstool.JsonDecode(cacheInfo, &data)
			if dErr != nil {
				return nil, gstool.Error(`解析数据%s失败`, cacheInfo)
			}
			return data, nil
		}
	}
	//缓存没有 查库
	info, infoErr := queryFunc()
	if infoErr != nil {
		return nil, gstool.Error(`获取db数据异常 %s`, infoErr.Error())
	}
	var infoJson string
	if info == nil || len(info) == 0 { //数据库数据为空 给一个空值
		infoJson = gsdefine.EmptyValue
		info = make(map[string]any)
	} else {
		infoJson = gstool.JsonEncode(info)
	}
	setErr := cli.Client.Set(context.Background(), key, infoJson, expired).Err()
	if setErr != nil {
		return nil, gstool.Error(`设置缓存异常 %s`, setErr.Error())
	}
	return info, nil
}

// BaseString2Cache map[string]any -> string
// 将一个map[string]any 存为 string 最后返回string
// Deprecated:  请使用 RedisGetMapString
func BaseString2Cache(cli *GsRedis, key string, queryFunc func() (map[string]any, error), expired time.Duration) (string, error) {
	cacheInfo, err := cli.Client.Get(context.Background(), key).Result()
	if !errors.Is(err, redis.Nil) && err != nil { //报错了
		return ``, err
	}
	if !errors.Is(err, redis.Nil) { //缓存存在
		if cacheInfo == gsdefine.EmptyValue { //空值 防止穿透
			return ``, nil
		} else {
			return cacheInfo, nil
		}
	}
	//缓存没有 查库
	info, infoErr := queryFunc()
	if infoErr != nil {
		return ``, gstool.Error(`获取db数据异常 %s`, infoErr.Error())
	}
	var infoJson string
	var returnJson string
	if info == nil || len(info) == 0 { //数据库数据为空 给一个空值
		infoJson = gsdefine.EmptyValue
		returnJson = ``
	} else {
		infoJson = gstool.JsonEncode(info)
		returnJson = infoJson
	}
	setErr := cli.Client.Set(context.Background(), key, infoJson, expired).Err()
	if setErr != nil {
		return ``, gstool.Error(`设置缓存异常 %s`, setErr.Error())
	}
	return returnJson, nil
}

// BaseHashArrayCache map[string]map[string]any -> hash[string]string init
// 将一个双层map存为 hash
// Deprecated:  请使用 RedisGetMapHash
func BaseHashArrayCache(cli *GsRedis, key string, queryFunc func() (map[string]map[string]any, error), expired time.Duration) (map[string]map[string]any, error) {
	existInt, err := cli.Client.Exists(context.Background(), key).Result()
	if err != nil { //报错了
		return nil, err
	}
	if existInt > 0 {
		//获取所有数据
		resultList, resultErr := cli.Client.HGetAll(context.Background(), key).Result()
		if resultErr != nil { //保存了
			return nil, resultErr
		}
		//移除空值 反转结果返回
		delete(resultList, gsdefine.EmptyKey)
		returnResult := make(map[string]map[string]any)
		for resultKey, resultVal := range resultList {
			data := make(map[string]any)
			dErr := gstool.JsonDecode(resultVal, &data)
			if dErr != nil {
				return nil, gstool.Error(`解析数据%s失败`, resultVal)
			}
			returnResult[resultKey] = data
		}
		return returnResult, nil
	}
	infoList, infoErr := queryFunc()
	if infoErr != nil {
		return nil, gstool.Error(`获取db数据异常 %s`, infoErr.Error())
	}
	//没有数据
	if infoList == nil || len(infoList) == 0 {
		cli.Client.HSet(context.Background(), key, gsdefine.EmptyKey, gsdefine.EmptyValue)
		if expired > 0 {
			cli.Client.Expire(context.Background(), key, expired)
		}
		return map[string]map[string]any{}, nil
	}
	setList := make([]any, 0)
	for infoKey, infoVal := range infoList {
		infoJson := gstool.JsonEncode(infoVal)
		setList = append(setList, infoKey, infoJson)
	}
	cli.Client.HMSet(context.Background(), key, setList...)
	if expired > 0 {
		cli.Client.Expire(context.Background(), key, expired)
	}
	return infoList, nil

}

// BaseHashCache map[string]string -> hash[string]string
// 将单层map存入hash
// Deprecated: 请使用 RedisGetMapHash
func BaseHashCache(cli *GsRedis, key string, queryFunc func() (map[string]string, error), expired time.Duration) (map[string]string, error) {
	existInt, err := cli.Client.Exists(context.Background(), key).Result()
	if err != nil { //报错了
		return nil, err
	}
	if existInt > 0 {
		//获取所有数据
		result, resultErr := cli.Client.HGetAll(context.Background(), key).Result()
		if resultErr != nil {
			return nil, resultErr
		}
		//移除默认填充的数据
		delete(result, gsdefine.EmptyKey)
		return result, nil
	}
	info, infoErr := queryFunc()
	if infoErr != nil {
		return nil, gstool.Error(`获取db数据异常 %s`, infoErr.Error())
	}
	if info == nil || len(info) == 0 {
		cli.Client.HSet(context.Background(), key, gsdefine.EmptyKey, gsdefine.EmptyValue)
		if expired > 0 {
			cli.Client.Expire(context.Background(), key, expired)
		}
		return map[string]string{}, nil
	}
	setList := make([]any, 0)
	for infoKey, infoVal := range info {
		setList = append(setList, infoKey, infoVal)
	}
	cli.Client.HMSet(context.Background(), key, setList...)
	if expired > 0 {
		cli.Client.Expire(context.Background(), key, expired)
	}
	return info, nil
}
