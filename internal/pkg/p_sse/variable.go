package p_sse

import (
	"dev_tool/internal/pkg/p_define"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

type SseVariable struct {
	Sse             *gsgin.Sse          //sse
	SseDistributeId string              //具体业务接收的id
	RunUniqueId     string              //当前执行任务的唯一ID 用来控制任务停止输出sse
	StopCall        func(string) string //判断是否停止
}

func (h *SseVariable) Send(msg any, enter bool, typs ...string) {
	//如果本次任务已经停止 那么不再输出
	if h.StopCall(h.RunUniqueId) == `stop` {
		h.Sse.CleanMsg()
		return
	}
	typ := p_define.SseContentTypeMsg
	if len(typs) > 0 && typs[0] != `` {
		typ = typs[0]
	}
	if gstool.AnyTypeString(msg) == `string` {
		msg = cast.ToString(msg) + "\n"
	}
	sendData := gstool.JsonEncode(p_define.SseData{
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
