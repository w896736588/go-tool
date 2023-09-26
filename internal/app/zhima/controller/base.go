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
	"xkf_tool/base_module"
)

//Login 登录
func Login(c *gin.Context) {
	reqBody := &base_module.LoginStruct{}
	err := gsgin.GinPostBody(c, reqBody)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), nil)
		return
	}
	unikey := gstool.Md5(reqBody.UserName + reqBody.Password)
	base_module.CreateGlobal(unikey)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `获取成功`, map[string]string{
		`unikey`: unikey,
	})
}

//RegisterService 注册各类服务
func RegisterService(c *gin.Context) {
	reqBody := &base_module.RegisterStruct{}
	err := gsgin.GinPostBody(c, reqBody)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, err.Error(), nil)
		return
	}
	global := base_module.GetGlobal(reqBody.Unikey)
	if global == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `请登录`, nil)
		return
	}
	base_module.Register(global, reqBody)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `ok`, nil)
}

//GetGlobalReqParams 拿到全局参数
func GetGlobalReqParams(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, error) {
	reqMap := make(map[string]*gstool.GsCons)
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	global, err := GetGlobal(reqMap)
	if err != nil {
		return nil, nil, err
	}
	return global, reqMap, nil
}

//BaseRedisGetReqDataRedis 基础方法
func BaseRedisGetReqDataRedis(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gsdb.GsRedis, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, err
	}
	redisName := reqMap[`RedisName`]
	if redisName == nil {
		gsgin.GinResponse(c, gsgin.ResponseError, `缺少RedisName参数`, nil)
		return nil, nil, nil, errors.New(`缺少RedisName参数`)
	}
	client, err := global.RedisGetClient(cast.ToString(redisName))
	if err != nil {
		return nil, nil, nil, err
	}

	return global, reqMap, client, nil
}

//BaseRedisCheckKeyExist 基础方法
func BaseRedisCheckKeyExist(redisCli *redis.Client, key string) error {
	//判断是否存在
	if existInt := redisCli.Exists(key).Val(); existInt <= 0 {
		return errors.New(fmt.Sprintf(`%s 不存在`, key))
	}
	return nil
}

func BaseResponseByError(c *gin.Context, err error) {
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), ``)
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, ``, ``)
	}
}
