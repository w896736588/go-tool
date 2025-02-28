package controller

import (
	"dev_tool/base"
	ai2 "dev_tool/internal/pkg/ai"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
)

func AiRun(c *gin.Context) {
	data, err := getAiComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	base.Component.GsLog.Debugf(`对话开始 参数：%s %s`, "\n", gstool.JsonEncode(data))
	go func() {
		aiRet, aiProcess, aiErr := ai2.Ai(data)
		if aiErr != nil {
			base.Component.TSocket.SendMsgReal(`0#code`, `执行失败 `+aiErr.Error())
		} else {
			base.Component.GsLog.Debugf(`%s`, gstool.JsonEncode(map[string]any{
				`ret`:     aiRet,
				`process`: aiProcess,
			}))
		}
	}()
	gsgin.GinResponseSuccess(c, ``, nil)
	return

}

func getAiComponent(c *gin.Context) (map[string]interface{}, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, err
	}
	return reqMap, nil
}
