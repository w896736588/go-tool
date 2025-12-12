package base

import (
	"dev_tool/base/define"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// SseShell 通过ssh转发的消息
type SseShell struct {
	Sse             *gsgin.Sse //sse
	SseDistributeId string     //具体业务接收的id
}

func (h *SseShell) Send(msg any, typs ...string) {
	typ := define.SseContentTypeMsg
	if len(typs) > 0 && typs[0] != `` {
		typ = typs[0]
	}
	sendData := gstool.JsonEncode(define.SseData{
		SseDistributeId: h.SseDistributeId,
		Data:            msg,
		Type:            typ,
	})
	err := h.Sse.SendToChan(sendData)
	if err != nil {
		gstool.FmtPrintlnLogTime(`发送sse错误 %s`, err.Error())
	}
}

type SseVariable struct {
	Sse             *gsgin.Sse //sse
	SseDistributeId string     //具体业务接收的id
	RunUniqueId     string     //当前执行任务的唯一ID 用来控制任务停止输出sse
}

func (h *SseVariable) Send(msg any, enter bool, typs ...string) {
	//如果本次任务已经停止 那么不再输出
	if Component.TVariable.Get(h.RunUniqueId) == `stop` {
		h.Sse.CleanMsg()
		return
	}
	typ := define.SseContentTypeMsg
	if len(typs) > 0 && typs[0] != `` {
		typ = typs[0]
	}
	if gstool.AnyTypeString(msg) == `string` {
		msg = cast.ToString(msg) + "\n"
	}
	sendData := gstool.JsonEncode(define.SseData{
		SseDistributeId: h.SseDistributeId,
		Data:            msg,
		Type:            typ,
	})
	if enter {
		sendData += "\n"
	}
	err := h.Sse.SendToChan(sendData)
	if err != nil {
		gstool.FmtPrintlnLogTime(`发送sse错误 %s`, err.Error())
	}
}

func (h *SseVariable) SetRunUniqueId(runUniqueId string) {
	h.RunUniqueId = runUniqueId
}

func (h *SseVariable) CleanMsg() {
	h.Sse.CleanMsg()
}
