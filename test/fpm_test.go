package test

import (
	"dev_tool/internal/pkg/p_api"
	"fmt"
	"sync"
	"testing"
)

var wg sync.WaitGroup

// TestFpm 测试fpm无session的情况
func TestFpmNoSession(t *testing.T) {

	fmt.Println("Curl命令解析器")
	fmt.Println("=================")

	// 示例curl命令
	example :=
		`curl 'http://dev8.zhima_chat_ai.applnk.cn/manage/editParagraph' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: zh-CN' \
  -H 'App-Type;' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryQYPZgcq4OSGeui6V' \
  -H 'Cookie: Hm_lvt_7b1addfd51b407c68cce8920af8faa0f=1764205341,1764206100,1764209353,1764225064; HMACCOUNT=1CA4943A26AF0708; Hm_lpvt_7b1addfd51b407c68cce8920af8faa0f=1764235287' \
  -H 'Origin: http://dev8.zhima_chat_ai.applnk.cn' \
  -H 'Referer: http://dev8.zhima_chat_ai.applnk.cn/' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36' \
  -H 'X-Requested-With: XMLHttpRequest' \
  -H 'lang: zh-CN' \
  -H 'token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjQzMTkzNjAsInBhcmVudF9pZCI6IjAiLCJ0dGwiOjg2NDAwLCJ1c2VyX2lkIjoiMSIsInVzZXJfbmFtZSI6ImFkbWluIn0.VI6_o-UQUp7KGXKP6kWL4_JvYvxwVlZKzL13vfsr4W8' \
  --data-raw $'------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="title"\r\n\r\n\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="content"\r\n\r\n\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="question"\r\n\r\naaa的下一个词是什么\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="answer"\r\n\r\nbbb\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="category_id"\r\n\r\n0\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="group_id"\r\n\r\n0\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="similar_questions"\r\n\r\n["dsfds","fdsfds"]\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="id"\r\n\r\n30\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V\r\nContent-Disposition: form-data; name="library_id"\r\n\r\n7\r\n------WebKitFormBoundaryQYPZgcq4OSGeui6V--\r\n' \
  --insecure`

	fmt.Printf("原始命令: %s\n", example)
	parsed, err := p_api.ParseCurlCommand(example)
	if err != nil {
		fmt.Printf("解析错误: %v\n", err)
		return
	}

	fmt.Printf("解析结果:\n%s", parsed.String())
	fmt.Printf("等效命令: %s\n", parsed.ToCurlCommand())

	fmt.Println("程序结束")
}
