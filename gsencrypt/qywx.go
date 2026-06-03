package gsencrypt

import (
	"github.com/sbzhu/weworkapi_golang/wxbizmsgcrypt"
	"github.com/spf13/cast"
)

type Qywx struct {
	ReceiverId     string //服务商企业ID
	Token          string
	EncodingAESKey string
}

func (h *Qywx) GetEchoStr(params map[string]any) (string, *wxbizmsgcrypt.CryptError) {
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(h.Token, h.EncodingAESKey, h.ReceiverId, wxbizmsgcrypt.XmlType)
	verifyMsgSign := cast.ToString(params[`msg_signature`])
	verifyTimestamp := cast.ToString(params[`timestamp`])
	verifyNonce := cast.ToString(params[`nonce`])
	verifyEchoStr := cast.ToString(params[`echostr`])
	echoStr, cryptErr := wxcpt.VerifyURL(verifyMsgSign, verifyTimestamp, verifyNonce, verifyEchoStr)
	if cryptErr != nil {
		return ``, cryptErr
	}
	return cast.ToString(echoStr), nil
}

func (h *Qywx) GetMsg(params map[string]any, postBody string) (string, *wxbizmsgcrypt.CryptError) {
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(h.Token, h.EncodingAESKey, h.ReceiverId, wxbizmsgcrypt.XmlType)
	reqMsgSign := cast.ToString(params[`msg_signature`])
	reqTimestamp := cast.ToString(params[`timestamp`])
	reqNonce := cast.ToString(params[`nonce`])
	reqData := []byte(postBody)
	msg, cryptErr := wxcpt.DecryptMsg(reqMsgSign, reqTimestamp, reqNonce, reqData)
	if nil != cryptErr {
		return ``, cryptErr
	}
	return cast.ToString(msg), nil
}
