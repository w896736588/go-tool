package base

import (
	"dev_tool/base/define"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
)

type FullSse struct {
	Sse             *gsgin.Sse //sse
	SseDistributeId string     //具体业务接收的id
}

func (h *FullSse) SendDistribute(msg any, typ string) {
	sendData := gstool.JsonEncode(define.SseData{
		SseDistributeId: h.SseDistributeId,
		Data:            msg,
		Type:            typ,
	})
	_ = h.Sse.SendToChan(sendData)
}

func GetVariableSseSend(fullSse *FullSse, runUniqueId string) func(msg string, enter bool) {
	return func(msg string, enter bool) {
		//如果本次任务已经停止 那么不再输出
		if Component.TVariable.Get(runUniqueId) == `stop` {
			fullSse.Sse.CleanMsg()
			return
		}
		if enter {
			msg += "\n"
		}
		//发送结构化数据
		fullSse.SendDistribute(msg, define.SseContentTypeMsg)
	}
}

func GetGitSseSend(fullSse *FullSse) func(msg string) {
	return func(msg string) {
		//发送结构化数据
		fullSse.SendDistribute(msg, define.SseContentTypeMsg)
	}
}

func GetShellOutSseSend(fullSse *FullSse) func(msg any, typ string) {
	return func(msg any, typ string) {
		if typ == `` {
			typ = define.SseContentTypeMsg
		}
		//发送结构化数据
		fullSse.SendDistribute(msg, typ)
	}
}
