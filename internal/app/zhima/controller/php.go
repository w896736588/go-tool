package controller

import (
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

//PhpPhpUnSerialize 反序列化（php）
func PhpPhpUnSerialize(c *gin.Context) {
	var err error
	_, reqMap, err := GetGlobalReqParamsM(c)
	var out interface{}
	out, err = gstool.PhpUnSerialize(cast.ToString(reqMap[`SerializeStr`]))
	if err != nil {
		gstool.FmtPrintlnLog(`反序列化失败 %s`, err.Error())
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), reqMap[`SerializeStr`])
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `成功`, gstool.JsonEncode(out))
}
