package gstool

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type Http struct {
	Client *http.Client
	Logger *GsSlog
}

// NewHttp 创建http对象
func NewHttp(timeout time.Duration, logger *GsSlog) *Http {
	return &Http{
		Client: &http.Client{
			Timeout: timeout, //default 10 seconds
			Transport: &http.Transport{
				DisableKeepAlives: true, //禁用保持长连接
			},
		},
		Logger: logger,
	}
}

// HttpPostJson 发起post json请求
func (h *Http) HttpPostJson(url, postBody string, header *map[string]string) ([]byte, error) {
	var jsonData = []byte(postBody)
	request, err := http.NewRequest(`POST`, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	if header == nil {
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	} else {
		for headerKey, headerVal := range *header {
			request.Header.Set(headerKey, headerVal)
		}
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	response, errDo := h.Client.Do(request)
	if errDo != nil {
		return nil, errDo
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			h.Logger.Errof(`close error： %s`, closeErr.Error())
		}
	}(response.Body)
	resData, errIo := io.ReadAll(response.Body)
	if errIo != nil {
		return nil, errIo
	}
	return resData, nil
}

// HttpGet 发起get请求
func (h *Http) HttpGet(url string) ([]byte, error) {
	request, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return nil, err
	}
	response, errDo := h.Client.Do(request)
	if errDo != nil {
		return nil, errDo
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			h.Logger.Errof(`close error： %s`, closeErr.Error())
		}
	}(response.Body)
	resData, errIo := io.ReadAll(response.Body)
	if errIo != nil {
		return nil, errIo
	}
	return resData, nil
}

// HttpPostMultiForm 发起post 表单上传文件请求
func (h *Http) HttpPostMultiForm(url string, body *bytes.Buffer) ([]byte, error) {
	request, requestErr := http.NewRequest(`POST`, url, body)
	if requestErr != nil {
		return nil, requestErr
	}
	request.Header.Set(`Content-Type`, `multipart/form-data`)
	response, errDo := h.Client.Do(request)
	if errDo != nil {
		return nil, errDo
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			h.errorLog(`close error： %s`, closeErr.Error())
		}
	}(response.Body)
	resData, errIo := io.ReadAll(response.Body)
	if errIo != nil {
		return nil, errIo
	}
	return resData, nil
}

func (h *Http) errorLog(msg string, params ...any) {
	if h.Logger != nil {
		h.Logger.Errof(msg, params...)
	} else {
		FmtPrintlnLog(msg, params...)
	}
}
