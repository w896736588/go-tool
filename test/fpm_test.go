package test

import (
	"dev_tool/internal/pkg/p_api"
	"fmt"
	"sync"
	"testing"

	"gitee.com/Sxiaobai/gs/gstool"
)

var wg sync.WaitGroup

// TestFpm 测试fpm无session的情况
func TestFpmNoSession(t *testing.T) {

	fmt.Println("Curl命令解析器")
	fmt.Println("=================")

	// 示例curl命令
	example :=
		`curl 'http://dev6.zhima_chat_ai.applnk.cn/manage/addFAQFile' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: zh-CN' \
  -H 'App-Type;' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: multipart/form-data; boundary=----WebKitFormBoundary3XEWYdskWArGkNEI' \
  -H 'Cookie: Hm_lvt_7b1addfd51b407c68cce8920af8faa0f=1762414505,1762414532,1764303926; HMACCOUNT=1CA4943A26AF0708; Hm_lpvt_7b1addfd51b407c68cce8920af8faa0f=1764303937' \
  -H 'Origin: http://dev6.zhima_chat_ai.applnk.cn' \
  -H 'Referer: http://dev6.zhima_chat_ai.applnk.cn/' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36' \
  -H 'X-Requested-With: XMLHttpRequest' \
  -H 'lang: zh-CN' \
  -H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjQzOTk4MzUsInBhcmVudF9pZCI6IjAiLCJ0dGwiOjg2NDAwLCJ1c2VyX2lkIjoiMSIsInVzZXJfbmFtZSI6ImFkbWluIn0.LaRpE8F3yO7JfHftecuBVHyCQY0O88GGZql4e0tF4pg' \
  --data-raw $'------WebKitFormBoundary3XEWYdskWArGkNEI\r\nContent-Disposition: form-data; name="faq_files"; filename="无标题-1.txt"\r\nContent-Type: text/plain\r\n\r\n\r\n------WebKitFormBoundary3XEWYdskWArGkNEI\r\nContent-Disposition: form-data; name="chunk_type"\r\n\r\n1\r\n------WebKitFormBoundary3XEWYdskWArGkNEI\r\nContent-Disposition: form-data; name="chunk_size"\r\n\r\n1000\r\n------WebKitFormBoundary3XEWYdskWArGkNEI\r\nContent-Disposition: form-data; name="chunk_model"\r\n\r\nqwen-max\r\n------WebKitFormBoundary3XEWYdskWArGkNEI\r\nContent-Disposition: form-data; name="chunk_model_config_id"\r\n\r\n3\r\n------WebKitFormBoundary3XEWYdskWArGkNEI\r\nContent-Disposition: form-data; name="chunk_prompt"\r\n\r\n根据user角色提供的文本，学习和分析它，并整理学习成果：\r\n- 提出问题并给出每个问题的答案。\r\n- 答案需详细完整，尽可能保留原文描述，可以适当扩展答案描述。\r\n- 答案可以包含普通文字、链接、代码、表格、公示、媒体链接等 Markdown 元素。\r\n- 最多提出 10个问题。\r\n- 生成的问题和答案和源文本语言相同。\r\n------WebKitFormBoundary3XEWYdskWArGkNEI--\r\n' \
  --insecure`
	parse := p_api.NewParseCurl(example)
	err := parse.ParseCurl()
	if err != nil {
		fmt.Println("解析错误:", err)
		return
	}
	gstool.FmtPrintlnLogTime(`%s`, gstool.JsonFormat(parse.CurlStruct))
}
