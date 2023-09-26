package controller

import (
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/cast"
	"time"
	"xkf_tool/internal/app/xkf_tool"
)

//RedisAvailableList 拿到注册的可用的redis列表
func RedisAvailableList(c *gin.Context) {
	global, _, err := GetGlobalReqParams(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	runList := make([]string, 0)
	global.RedisEachConfigList(func(s string, cons *gstool.GsCons) {
		_, err := global.RedisGetClient(s)
		if err != nil {
			return
		}
		runList = append(runList, s)
	})
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, runList)
}

//RedisKeys 搜索key
func RedisKeys(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	search := reqMap[`Search`]
	if search == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少搜索内容参数`, nil)
		return
	}
	var resultMap []string
	resultMap, err = redisCli.Client.Keys(cast.ToString(search)).Result()
	if err == redis.Nil {
		resultMap = make([]string, 0)
	} else if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	//拿到key类型
	returnList := make([]map[string]any, 0)
	for _, cacheKey := range resultMap {
		returnList = append(returnList, map[string]any{
			`CacheKey`: cacheKey,
			`Type`:     ` `,
			`Loading`:  true,
		})
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, returnList)
}

//RedisSearch 获取一个key的明细
func RedisSearch(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	cacheKey := gstool.ConsNew(reqMap[`CacheKey`])
	if cacheKey.Value() == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少搜索的key`, nil)
		return
	}
	//找到key是什么类型
	keyType, err := redisCli.Client.Type(cacheKey.ToStr()).Result()
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), ``)
		return
	}
	if keyType == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `缓存已不存在`, ``)
		return
	}
	//通用的返回结果
	var gsCons *gstool.GsCons
	if keyType == gsdb.RedisKeyString {
		var result string
		result, err = redisCli.Client.Get(cacheKey.ToStr()).Result()
		gsCons = gstool.ConsNew(result)
	} else if keyType == gsdb.RedisKeyHash {
		var resultMap map[string]string
		resultMap, err = redisCli.Client.HGetAll(cacheKey.ToStr()).Result()
		gsCons = gstool.ConsNew(resultMap)
	} else if keyType == gsdb.RedisKeyList {
		var resultArray []string
		resultArray, err = redisCli.Client.LRange(cacheKey.ToStr(), 0, 100000).Result()
		gsCons = gstool.ConsNew(resultArray)
	} else if keyType == gsdb.RedisKeySet {
		var resultArray []string
		resultArray, err = redisCli.Client.SMembers(cacheKey.ToStr()).Result()
		gsCons = gstool.ConsNew(resultArray)
	} else if keyType == gsdb.RedisKeyZSet {
		var resultArray []redis.Z
		resultArray, err = redisCli.Client.ZRangeWithScores(cacheKey.ToStr(), 0, 100000).Result()
		gsCons = gstool.ConsNew(resultArray)
	} else {
		gsgin.GinResponse(c, gsgin.ResponseError, `暂不支持的缓存类型 `+keyType, ``)
		return
	}
	if err == redis.Nil {
		gsgin.GinResponse(c, gsgin.ResponseError, fmt.Sprintf(`%s 已经不存在`, cacheKey), ``)
		return
	} else if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), ``)
		return
	}
	if gsCons == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, fmt.Sprintf(`不支持的类型 %s`, keyType), ``)
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取成功`, gsCons.Value())
	}

}

//RedisKeysType 获取redis的key类型
func RedisKeysType(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	keyList := gstool.ArrayGetFromAny(reqMap[`KeyList`])
	if len(*keyList) == 0 {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少keyList参数`, nil)
		return
	}
	//拿到key类型
	returnList := make([]map[string]interface{}, 0)
	for _, cacheKey := range *keyList {
		keyType, err := redisCli.Client.Type(cacheKey.ToStr()).Result()
		if err == nil && keyType != `` {
			returnList = append(returnList, map[string]interface{}{
				`CacheKey`: cacheKey.ToStr(),
				`Type`:     keyType,
			})
		}
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取成功`, returnList)
}

// RedisKeyType 获取单个key类型
func RedisKeyType(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		BaseResponseByError(c, err)
		return
	}

	existErr := BaseRedisCheckKeyExist(redisCli.Client, cast.ToString(reqMap[`CacheKey`]))
	if existErr != nil {
		BaseResponseByError(c, err)
		return
	}
	cacheKey := cast.ToString(reqMap[`CacheKey`])
	//找到key是什么类型
	keyType, err := redisCli.Client.Type(cacheKey).Result()
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	} else if keyType == `` {
		gsgin.GinResponse(c, gsgin.ResponseError, `获取元素类型失败`, nil)
		return
	}
	//找到过期时间
	var ttl time.Duration
	ttl, err = redisCli.Client.TTL(cacheKey).Result()
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), ``)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, map[string]interface{}{
		`Type`: keyType,
		`TTL`:  cast.ToInt(ttl.Seconds()),
	})
}

//RedisSaveString 保存字符串值
func RedisSaveString(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		BaseResponseByError(c, err)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		BaseResponseByError(c, err)
		return
	}

	ttlTime, err := redisCli.Client.TTL(cacheKey).Result()
	//永久
	err = redisCli.Client.Set(cacheKey, reqMap[`Value`], ttlTime).Err()
	BaseResponseByError(c, err)
}

//RedisDelKey 删除key
func RedisDelKey(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	//永久
	err = redisCli.Client.Del(cacheKey).Err()
	BaseResponseByError(c, err)
}

//RedisDelSub 删除子元素
func RedisDelSub(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	cacheKey := reqMap[`CacheKey`].ToStr()
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	cacheType := reqMap[`cacheType`].ToStr()
	if cacheType == `` {
		BaseResponseByError(c, errors.New(`cacheType类型不能为空`))
		return
	}
	subKey := reqMap[`Sub`].ToStr()
	if subKey == `` {
		BaseResponseByError(c, errors.New(`Sub 类型不能为空`))
		return
	}
	if cacheType == xkf_tool.CacheString {
		gsgin.GinResponse(c, gsgin.ResponseError, `不支持字符串`, ``)
		return
	} else if cacheType == xkf_tool.CacheHash {
		err = redisCli.Client.HDel(cacheKey, subKey).Err()
	} else if cacheType == xkf_tool.CacheList {
		err = redisCli.Client.LRem(cacheKey, 0, subKey).Err()
	} else if cacheType == xkf_tool.CacheSet {
		err = redisCli.Client.SRem(cacheKey, subKey).Err()
	} else if cacheType == xkf_tool.CacheZSet {
		err = redisCli.Client.ZRem(cacheKey, subKey).Err()
	}
	BaseResponseByError(c, err)
}

//RedisEditTtl 修改过期时间
func RedisEditTtl(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	ttl := cast.ToInt(reqMap[`TTL`])
	dru := time.Duration(ttl) * time.Second
	err = redisCli.Client.Expire(cacheKey, dru).Err()
	BaseResponseByError(c, err)
}

//RedisDelAllKey 批量删除key
func RedisDelAllKey(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	cacheKeyList := reqMap[`CacheKeys`].Value().([]string)
	err = redisCli.Client.Del(cacheKeyList...).Err()
	BaseResponseByError(c, err)
}

//RedisCreateCache 新增key
func RedisCreateCache(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	//判断是否存在
	if reqMap[`BoolCreate`].ToInt() == 1 {
		if existInt := redisCli.Client.Exists(cacheKey).Val(); existInt > 0 {
			gsgin.GinResponse(c, gsgin.ResponseError, fmt.Sprintf(`%s 已经存在`, cacheKey), ``)
			return
		}
	} else {
		err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
		if err == nil {
			gsgin.GinResponse(c, gsgin.ResponseError, fmt.Sprintf(`%s key不存在`, cacheKey), ``)
			return
		}
	}

	cacheType := cast.ToString(reqMap[`CacheType`])
	if cacheType == gsdb.RedisKeyString {
		err = redisCli.Client.Set(cacheKey, reqMap[`CacheValue`], time.Duration(cast.ToInt64(reqMap[`TTL`]))*time.Second).Err()
	} else if cacheType == gsdb.RedisKeyHash {
		err = redisCli.Client.HSet(cacheKey, reqMap[`CacheField`].ToStr(), reqMap[`CacheValue`].ToStr()).Err()
	} else if cacheType == gsdb.RedisKeyList {
		if reqMap[`LPushValue`].ToStr() != `` {
			err = redisCli.Client.LPush(cacheKey, reqMap[`LPushValue`].ToStr()).Err()
		} else if reqMap[`RPushValue`].ToStr() != `` {
			err = redisCli.Client.RPush(cacheKey, reqMap[`RPushValue`].ToStr()).Err()
		} else {
			err = redisCli.Client.RPush(cacheKey, reqMap[`CacheValue`].ToStr()).Err()
		}
	} else if cacheType == gsdb.RedisKeySet {
		err = redisCli.Client.SAdd(cacheKey, reqMap[`CacheMember`].ToStr()).Err()
	} else if cacheType == gsdb.RedisKeyZSet {
		err = redisCli.Client.ZAdd(cacheKey, redis.Z{
			Score:  reqMap[`CacheScore`].ToFloat64(),
			Member: reqMap[`CacheMember`].ToStr(),
		}).Err()
	}
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), ``)
		return
	}
	//处理过期时间
	if reqMap[`BoolCreate`].ToInt() == 1 && reqMap[`TTL`].ToInt() != 0 {
		err = redisCli.Client.Expire(cacheKey, time.Duration(reqMap[`TTL`].ToInt())*time.Second).Err()
	}
	BaseResponseByError(c, err)
}

//RedisEditSub 编辑子元素
func RedisEditSub(c *gin.Context) {
	_, reqMap, redisCli, err := BaseRedisGetReqDataRedis(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	cacheKey := cast.ToString(reqMap[`CacheKey`])
	err = BaseRedisCheckKeyExist(redisCli.Client, cacheKey)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	cacheType := reqMap[`CacheType`].ToStr()
	cacheField := reqMap[`CacheField`].ToStr()
	cacheValue := reqMap[`CacheValue`].ToStr()
	cacheIndex := reqMap[`CacheIndex`].ToInt64()
	if cacheType == gsdb.RedisKeyHash {
		err = redisCli.Client.HSet(cacheKey, cacheField, cacheValue).Err()
	} else if cacheType == gsdb.RedisKeyList {
		err = redisCli.Client.LSet(cacheKey, cacheIndex, cacheValue).Err()
	} else if cacheType == gsdb.RedisKeyZSet {
		err = redisCli.Client.ZAdd(cacheKey, redis.Z{
			Score:  reqMap[`CacheScore`].ToFloat64(),
			Member: reqMap[`CacheMember`].ToStr(),
		}).Err()
	}
	BaseResponseByError(c, err)
}
