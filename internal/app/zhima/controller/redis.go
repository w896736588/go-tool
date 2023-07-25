package controller

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/spf13/cast"
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
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	redisName := reqMap[`RedisName`]
	if redisName == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少redisName参数`, nil)
		return
	}
	search := reqMap[`Search`]
	if search == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少搜索内容参数`, nil)
		return
	}
	redisCli, err := global.RedisGetClient(cast.ToString(redisName))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
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
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	redisName := reqMap[`RedisName`]
	if redisName == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少redisName参数`, nil)
		return
	}
	cacheKey := gstool.GsNew(reqMap[`CacheKey`])
	if cacheKey.Value() == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少搜索的key`, nil)
		return
	}
	redisCli, err := global.RedisGetClient(cast.ToString(redisName))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}

	//找到key是什么类型
	keyType, err := redisCli.Client.Type(cacheKey.ToStr()).Result()
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), ``)
		return
	}
	if keyType == `` {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取key类型失败`, ``)
		return
	}
	//通用的返回结果
	var gsCons *gstool.GsCons
	if keyType == gsdb.RedisKeyString {
		var result string
		result, err = redisCli.Client.Get(cacheKey.ToStr()).Result()
		gsCons = gstool.GsNew(result)
	} else if keyType == gsdb.RedisKeyHash {
		var resultMap map[string]string
		resultMap, err = redisCli.Client.HGetAll(cacheKey.ToStr()).Result()
		gsCons = gstool.GsNew(resultMap)
	} else if keyType == gsdb.RedisKeyList {
		var resultArray []string
		resultArray, err = redisCli.Client.LRange(cacheKey.ToStr(), 0, 100000).Result()
		gsCons = gstool.GsNew(resultArray)
	} else if keyType == gsdb.RedisKeySet {
		var resultArray []string
		resultArray, err = redisCli.Client.SMembers(cacheKey.ToStr()).Result()
		gsCons = gstool.GsNew(resultArray)
	} else if keyType == gsdb.RedisKeyZSet {
		var resultArray []redis.Z
		resultArray, err = redisCli.Client.ZRangeWithScores(cacheKey.ToStr(), 0, 100000).Result()
		gsCons = gstool.GsNew(resultArray)
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

//RedisKeyType 获取redis的key类型
func RedisKeyType(c *gin.Context) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	redisName := reqMap[`RedisName`]
	if redisName == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少redisName参数`, nil)
		return
	}
	keyList, boolRet := gstool.ArrayIsString(reqMap[`keyList`])
	if !boolRet {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少keyList参数`, nil)
		return
	}
	redisCli, err := global.RedisGetClient(cast.ToString(redisName))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	//拿到key类型
	returnList := make([]map[string]interface{}, 0)
	for _, cacheKey := range keyList {
		keyType, err := redisCli.Client.Type(cacheKey).Result()
		if err == nil && keyType != `` {
			returnList = append(returnList, map[string]interface{}{
				`CacheKey`: cacheKey,
				`Type`:     keyType,
			})
		}
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取成功`, returnList)
}
