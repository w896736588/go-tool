package controller

import (
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
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
func GetGlobalReqParams(c *gin.Context) (*base_module.Global, map[string]interface{}, error) {
	reqMap := make(map[string]interface{})
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
