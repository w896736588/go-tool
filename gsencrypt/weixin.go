package gsencrypt

import (
	"github.com/sbzhu/weworkapi_golang/wxbizmsgcrypt" // 复用企业微信的解密库
	"github.com/spf13/cast"
)

type Weixin struct {
	Token          string // 开放平台配置的Token
	EncodingAESKey string // 开放平台配置的AES Key
	AppId          string // 第三方应用的AppId 或 SuiteId
}

func (h *Weixin) GetEchoStr(params map[string]any) (string, *wxbizmsgcrypt.CryptError) {
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(h.Token, h.EncodingAESKey, h.AppId, wxbizmsgcrypt.XmlType)
	verifyMsgSign := cast.ToString(params["msg_signature"])
	verifyTimestamp := cast.ToString(params["timestamp"])
	verifyNonce := cast.ToString(params["nonce"])
	verifyEchoStr := cast.ToString(params["echostr"])
	echoStr, cryptErr := wxcpt.VerifyURL(verifyMsgSign, verifyTimestamp, verifyNonce, verifyEchoStr)
	if cryptErr != nil {
		return "", cryptErr
	}
	return cast.ToString(echoStr), nil
}

func (h *Weixin) GetMsg(params map[string]any, postBody string) (string, *wxbizmsgcrypt.CryptError) {
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(h.Token, h.EncodingAESKey, h.AppId, wxbizmsgcrypt.XmlType)
	reqMsgSign := cast.ToString(params["msg_signature"])
	reqTimestamp := cast.ToString(params["timestamp"])
	reqNonce := cast.ToString(params["nonce"])
	reqData := []byte(postBody)
	msg, cryptErr := wxcpt.DecryptMsg(reqMsgSign, reqTimestamp, reqNonce, reqData)
	if cryptErr != nil {
		return "", cryptErr
	}
	return cast.ToString(msg), nil
}
