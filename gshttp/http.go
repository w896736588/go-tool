package gshttp

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

const (
	ContentTypeGet = iota
	ContentTypePostMultiForm
	ContentTypePostForm
	ContentTypePostJson
)

type Client struct {
	sourceUrl        string
	urlValues        *url.Values
	contentType      int
	headerMap        map[string]string
	cookieList       []*http.Cookie
	allowHttpStatus  []int //允许成功的http状态码，默认为200
	disableKeepAlive bool  //true禁止keep-alive，默认开启
	body             *bytes.Buffer
	response         *http.Response
	streamFac        StreamInterface //流式输出处理器，仅支持一个
	responseByte     []byte
	err              error
	mw               *multipart.Writer
	httpClient       *http.Client //复用HTTP客户端，支持keep-alive
}

func Get(url string) *Client {
	return &Client{
		sourceUrl:        url,
		contentType:      ContentTypeGet,
		cookieList:       []*http.Cookie{},
		headerMap:        make(map[string]string),
		disableKeepAlive: true,
	}
}

func PostForm(sourceUrl string) *Client {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return &Client{
		sourceUrl:        sourceUrl,
		contentType:      ContentTypePostForm,
		headerMap:        headers,
		body:             &bytes.Buffer{},
		cookieList:       []*http.Cookie{},
		urlValues:        &url.Values{},
		disableKeepAlive: true,
	}
}

func PostJson(url string) *Client {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	return &Client{
		sourceUrl:        url,
		contentType:      ContentTypePostJson,
		headerMap:        headers,
		body:             &bytes.Buffer{},
		cookieList:       []*http.Cookie{},
		disableKeepAlive: true,
	}
}

func QuickGet(url string, timeout time.Duration) ([]byte, error) {
	return Get(url).Request(timeout).Result()
}

func QuickPostJson(url, body string, timeout time.Duration) ([]byte, error) {
	return PostJson(url).BodyStr(body).Request(timeout).Result()
}

func QuickPostBuffer(url string, buffer *bytes.Buffer, timeout time.Duration) ([]byte, error) {
	return PostMultiForm(url).BodyBuffer(buffer).Request(timeout).Result()
}

func PostMultiForm(url string) *Client {
	client := &Client{
		sourceUrl:        url,
		contentType:      ContentTypePostMultiForm,
		body:             &bytes.Buffer{},
		cookieList:       []*http.Cookie{},
		disableKeepAlive: true,
	}
	client.mw = multipart.NewWriter(client.body)
	headers := make(map[string]string)
	headers["Content-Type"] = client.mw.FormDataContentType()
	client.headerMap = headers
	return client
}

func (h *Client) SetStreamFac(fac StreamInterface) *Client {
	if h.streamFac != nil {
		h.err = errors.New("streamFac只能设置一个，重复设置无效")
		return h
	}
	h.streamFac = fac
	return h
}

func (h *Client) SetAllowHttpStatus(statusCode ...int) *Client {
	h.allowHttpStatus = statusCode
	return h
}

func (h *Client) isAllowHttpStatus() bool {
	if h.response == nil {
		return true
	}
	if len(h.allowHttpStatus) == 0 {
		return h.response.StatusCode == 200
	}
	return gstool.ArrayExistValue(&h.allowHttpStatus, h.response.StatusCode)
}

func (h *Client) OpenKeepAlive() *Client {
	h.disableKeepAlive = false
	return h
}

func (h *Client) BodyMap(pM map[string]any) *Client {
	if h.err != nil {
		return h
	}
	switch h.contentType {
	case ContentTypePostJson:
		_, copyErr := io.Copy(h.body, bytes.NewReader([]byte(gstool.JsonEncode(pM))))
		h.err = copyErr
	case ContentTypePostForm:
		for k, v := range pM {
			// 检查是否为数组/切片
			if arr, ok := v.([]interface{}); ok {
				for _, item := range arr {
					h.urlValues.Add(k, cast.ToString(item))
				}
			} else {
				h.urlValues.Add(k, cast.ToString(v))
			}
		}
		_, encodeErr := io.WriteString(h.body, h.urlValues.Encode())
		if encodeErr != nil {
			h.err = encodeErr
			return h
		}
	case ContentTypePostMultiForm:
		for k, v := range pM {
			// 检查是否为数组/切片
			if arr, ok := v.([]interface{}); ok {
				// 为数组中的每个元素创建一个字段
				for _, item := range arr {
					tw, twErr := h.mw.CreateFormField(k)
					if twErr != nil {
						h.err = twErr
						return h
					}
					_, writeErr := tw.Write([]byte(cast.ToString(item)))
					if writeErr != nil {
						h.err = writeErr
						return h
					}
				}
			} else {
				// 普通字段
				tw, twErr := h.mw.CreateFormField(k)
				if twErr != nil {
					h.err = twErr
					return h
				}
				_, writeErr := tw.Write([]byte(cast.ToString(v)))
				if writeErr != nil {
					h.err = writeErr
					return h
				}
			}
		}
	default:
		h.err = errors.New(`未知的类型`)
	}
	return h
}

func (h *Client) BodyStr(body string) *Client {
	if h.err != nil {
		return h
	}
	switch h.contentType {
	case ContentTypePostJson, ContentTypePostForm, ContentTypePostMultiForm:
		_, copyErr := io.Copy(h.body, bytes.NewReader([]byte(body)))
		h.err = copyErr
	default:
		h.err = errors.New(`不支持向get请求写入body`)
	}
	return h
}

func (h *Client) BodyBuffer(body *bytes.Buffer) *Client {
	if h.err != nil {
		return h
	}
	switch h.contentType {
	case ContentTypePostJson, ContentTypePostForm, ContentTypePostMultiForm:
		h.body = body
	default:
		h.err = errors.New(`不支持向get请求写入body`)
	}
	return h
}

func (h *Client) Headers(m map[string]string) *Client {
	if h.headerMap == nil {
		h.headerMap = m
	} else {
		for k, v := range m {
			h.headerMap[k] = v
		}
	}
	return h
}

func (h *Client) GetHeaders() map[string]string {
	return h.headerMap
}

func (h *Client) Cookies(m []*http.Cookie) *Client {
	h.cookieList = m
	return h
}

func (h *Client) BodyFile(formKey, filePath, fileName string) *Client {
	if h.contentType != ContentTypePostMultiForm {
		h.err = errors.New(`上传文件仅支持PostMultiForm`)
		return h
	}
	file, openErr := os.Open(filePath)
	if openErr != nil {
		h.err = openErr
		return h
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			//TODO
		}
	}(file)
	if fileName == `` {
		fileName = file.Name()
	}
	fw, fwErr := h.mw.CreateFormFile(formKey, fileName)
	if fwErr != nil {
		h.err = fwErr
		return h
	}
	_, copyErr := io.Copy(fw, file)
	if copyErr != nil {
		h.err = copyErr
		return h
	}
	return h
}

func (h *Client) Request(timeout time.Duration) *Client {
	if h.err != nil {
		return h
	}
	var req *http.Request
	var reqErr error
	if h.contentType == ContentTypePostMultiForm && h.mw != nil {
		if err := h.mw.Close(); err != nil {
			h.err = err
			return h
		}
	}
	switch h.contentType {
	case ContentTypePostJson, ContentTypePostForm, ContentTypePostMultiForm:
		req, reqErr = http.NewRequest("POST", h.sourceUrl, h.body)
	case ContentTypeGet:
		req, reqErr = http.NewRequest("GET", h.sourceUrl, nil)
	default:
		h.err = errors.New(`未知的类型`)
		return h
	}
	if reqErr != nil {
		h.err = reqErr
		return h
	}
	for k, v := range h.headerMap {
		req.Header.Add(k, v)
	}
	if h.contentType == ContentTypePostMultiForm && h.mw != nil {
		contentType := h.mw.FormDataContentType()
		req.Header.Set("Content-Type", contentType)
		h.headerMap["Content-Type"] = contentType
	}
	for _, v := range h.cookieList {
		req.AddCookie(v)
	}
	if h.httpClient == nil {
		h.httpClient = &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: h.disableKeepAlive,
			},
		}
	}
	h.httpClient.Timeout = timeout
	var responseErr error
	h.response, responseErr = h.httpClient.Do(req)
	if responseErr != nil {
		h.err = responseErr
		return h
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			//TODO
		}
	}(h.response.Body)
	// 读取响应体内容
	if h.streamFac != nil {
		h.streamFac.ReceiveSplit(h.response, &h.responseByte)
	} else {
		responseData, responseDataErr := io.ReadAll(h.response.Body)
		if responseDataErr != nil {
			h.err = responseDataErr
			return h
		}
		h.responseByte = responseData
	}
	// 校验http状态码
	if !h.isAllowHttpStatus() {
		h.err = errors.New(cast.ToString(h.response.StatusCode) + " " + string(h.responseByte))
		return h
	}

	return h
}

func (h *Client) ResponseHeader() http.Header {
	if h.response == nil {
		return http.Header{}
	}
	return h.response.Header
}

func (h *Client) Response() *http.Response {
	return h.response
}

func (h *Client) Result() ([]byte, error) {
	return h.responseByte, h.err
}

func (h *Client) ResultStr() (string, error) {
	return cast.ToString(h.responseByte), h.err
}

func (h *Client) HttpStatus() (int, error) {
	return h.Response().StatusCode, h.err
}

func (h *Client) ErrInfo() error {
	return h.err
}

func (h *Client) JsonDecode(a any) error {
	if h.err != nil {
		return h.err
	}
	decodeErr := gstool.JsonDecode(cast.ToString(h.responseByte), a)
	if decodeErr != nil {
		return decodeErr
	}
	return nil
}
