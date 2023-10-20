package controller

import (
	"context"
	"dev_tool/base_module"
	"dev_tool/internal/app/zhima/service"
	"errors"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var VipMap = map[string]string{
	`0`: `免费版`,
	`1`: `专业版`,
	`2`: `企业版`,
	`3`: `标准版`,
	`4`: `平台版`,
}

//VipChange vip版本切换
func VipChange(c *gin.Context) {
	_, reqMap, redisCli, mysqlCli, err := getVipReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	account := reqMap[`Account`]
	if account.IsEmpty() {
		gsgin.GinResponse(c, gsgin.ResponseError, `账号不能为空`, nil)
		return
	}
	userInfo := service.GetAdminUserId(mysqlCli, account.ToStr())
	if userInfo.G(`_id`).IsEmpty() {
		gsgin.GinResponse(c, gsgin.ResponseError, `找不到该账号`, nil)
		return
	}
	_, upErr := service.UpdateVip(mysqlCli, userInfo.G(`_id`).ToStr(), reqMap[`ExpireDay`].ToStr(), reqMap[`SystemType`].ToStr(), reqMap[`VipLevel`].ToStr())
	if upErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, upErr.Error(), nil)
		return
	}
	//移除缓存
	adminUserId := userInfo.G(`_id`)
	number := cast.ToString(adminUserId.ToInt() % 10)
	redisCli.Client.HDel(context.Background(), `wechatapp.vip.info.v20220308..`+number, adminUserId.ToStr())
	redisCli.Client.HDel(context.Background(), `wechatapp.kefu.vip.info.v20220308..`+number, adminUserId.ToStr())
	result, resultErr := queryVipType(c)
	if resultErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, resultErr.Error(), nil)
		return
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, result, nil)
		return
	}
}

//VipQuery vip版本查询
func VipQuery(c *gin.Context) {
	result, resultErr := queryVipType(c)
	if resultErr != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, resultErr.Error(), nil)
		return
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, result, nil)
		return
	}
}

// QueryVipType 查询VIP版本
func queryVipType(c *gin.Context) (string, error) {
	_, reqMap, _, mysqlCli, err := getVipReqData(c)
	if err != nil {
		return ``, err
	}
	account := reqMap[`Account`]
	if account.IsEmpty() {
		return ``, errors.New(`账号不能为空`)
	}
	userInfo := service.GetAdminUserId(mysqlCli, account.ToStr())
	if userInfo.G(`_id`).IsEmpty() {
		return ``, errors.New(`找不到该账号`)
	}
	vipInfo, queryErr := service.QueryVip(mysqlCli, userInfo.G(`_id`).ToStr(), reqMap[`SystemType`].ToStr())
	if queryErr != nil {
		return ``, queryErr
	}
	if vipInfo.IsZeroLen() {
		return `管理员ID：` + userInfo.G(`_id`).ToStr() + `未查到vip信息`, nil
	}
	return `管理员ID：` + userInfo.G(`_id`).ToStr() + `，vip版本：` + VipMap[vipInfo.G(`vip_type`).ToStr()] + `，过期时间：` + vipInfo.G(`expired_time`).ToStr(), nil
}

func getVipReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gsdb.GsRedis, *gsdb.GsMysql, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	redisName := reqMap[`redisName`]
	redisCli, err := global.RedisGetClient(redisName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	mysqlName := reqMap[`mysqlName`]
	mysqlCli, err := global.MysqlGetClient(mysqlName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return global, reqMap, redisCli, mysqlCli, nil
}
