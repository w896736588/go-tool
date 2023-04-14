package lib_tool

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

//加密
func JsonEncode(params interface{}) string {
	str, err := json.Marshal(params)
	if err != nil {
		log.Errorf(`格式化失败 %#v %s`, params, err.Error())
	}
	return cast.ToString(str)
}

//反解json
func JsonDecode(str string) interface{} {
	var returnParams interface{}
	err := json.Unmarshal([]byte(str), returnParams)
	if err != nil {
		log.Errorf(`格式化失败 %#v %s`, str, err.Error())
	}
	return returnParams
}
