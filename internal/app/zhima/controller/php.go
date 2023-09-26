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
	_, reqMap, err := GetGlobalReqParams(c)
	var out string
	out, err = gstool.PhpUnSerialize(cast.ToString(reqMap[`SerializeStr`]))
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), reqMap[`SerializeStr`])
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, `成功`, out)
}
