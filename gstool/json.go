package gstool

import (
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// JsonEncode 加密
func JsonEncode(params interface{}) string {
	str, err := json.Marshal(params)
	if err != nil {
		log.Errorf(`编码json失败 %#v %s`, params, err.Error())
	}
	return cast.ToString(str)
}

func JsonEncodeErr(params interface{}) (string, error) {
	str, err := json.Marshal(params)
	if err != nil {
		return ``, err
	}
	return cast.ToString(str), nil
}

// JsonDecode 反解json
func JsonDecode(str string, target any) error {
	err := json.Unmarshal([]byte(str), target)
	if err != nil {
		return Error(`json decode error: %s , %s`, str, err.Error())
	}
	return nil
}

// JsonToMap json转为map
func JsonToMap(b []byte) map[string]interface{} {
	m := make(map[string]interface{})
	err := json.Unmarshal(b, &m)
	if err != nil {
		return m
	}
	return m
}

// JsonDecodeUseNum 反解json，支持超长数字不被截断
func JsonDecodeUseNum(str string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(str))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return Error(`json decode use num error: %s , %s`, str, err.Error())
	}
	return nil
}

// JsonMarshalTrim json转义字符串中的特殊符号 并移除左侧和右侧的双引号
func JsonMarshalTrim(str string) string {
	//进行转义特殊符号
	b, err := json.Marshal(str)
	if err != nil {
		return str
	} else {
		s := cast.ToString(b)
		s = strings.TrimLeft(s, `"`)
		s = strings.TrimRight(s, `"`)
		return s
	}
}

func JsonFormat(data any) string {
	jsonData, _ := json.MarshalIndent(data, "", "    ")
	return cast.ToString(jsonData)
}

// JsonParseFromStr 从字符串中解析出每一个json，组成数组返回 不符合的字符会被过滤掉包括控制符
// 注意：json的key是不允许出现数字的，所以如果传入的是数字的key，最终显示的也是字符串的key
func JsonParseFromStr(data []byte) []string {
	var (
		outJsons    []string
		start       = -1
		depth       int  //根据括号嵌套深度
		isInJsonStr bool //当前字符是否属于json
		isEsc       bool //下一个字符是转义序列，不要把它当作普通引号或括号
		i           int
	)
	for i = 0; i < len(data); i++ {
		currentByte := data[i]
		// 处理字符串转义
		if isInJsonStr {
			if isEsc {
				isEsc = false
				continue
			}
			if currentByte == '\\' {
				isEsc = true
				continue
			}
			if currentByte == '"' {
				isInJsonStr = false
			}
			continue
		}
		// 不在字符串里
		switch currentByte {
		case '"':
			isInJsonStr = true
			isEsc = false
		case '{':
			if depth == 0 {
				start = i // 记录顶层 { 起点
			}
			depth++
		case '}':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					outJsons = append(outJsons, string(data[start:i+1]))
					start = -1
				}
			}
		}
	}
	return outJsons
}
