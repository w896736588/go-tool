package controller

import (
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"time"
	"xkf_tool/base_module"
	"xkf_tool/internal/app/zhima/service"
)

//LoginLink 登录地址
func LoginLink(c *gin.Context) {
	_, reqMap, encrypt, mysqlCli, err := getLoginReqData(c)
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
	loginHost := reqMap[`LoginHost`].ToStr()

	//拿到一个应用ID和一个渠道ID
	wechatAppId, channelId, errQuery := service.QueryOneWechatAppIdChannelId(mysqlCli, cast.ToInt(userInfo.G(`_id`).ToInt()))
	if errQuery != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, errQuery.Error(), nil)
		return
	}
	redirectUrl := reqMap[`LoginUrl`].ToStr()
	redirectUrl = strings.Replace(redirectUrl, `{wechatapp_id}`, wechatAppId, -1)
	redirectUrl = strings.Replace(redirectUrl, `{channel_id}`, channelId, -1)
	token := gstool.JsonEncode(map[string]string{
		`login_type`: `1`,
		`user_id`:    cast.ToString(userInfo.G(`_id`)),
		`param`: gstool.JsonEncode(map[string]string{
			`uri`: redirectUrl,
		}),
		`time`: cast.ToString(time.Now().Unix()), //仅10秒内有效
	})
	data, err := encrypt.EncryptDataDesCBC(token)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	token = url.QueryEscape(data)
	gsgin.GinResponse(c, gsgin.ResponseSuccess, loginHost+`index/LoginRedirect?token=`+token, nil)
	return
}

func getLoginReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gstool.Encrypt, *gsdb.GsMysql, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	encrypt := global.GetEncrypt()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	mysqlName := reqMap[`mysqlName`]
	mysqlCli, err := global.MysqlGetClient(mysqlName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return global, reqMap, encrypt, mysqlCli, nil
}
