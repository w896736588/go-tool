package gstool

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

// UrlRemoveDomain 将一个域名中的域名部分去掉
func UrlRemoveDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ``, err
	}
	pathWithoutDomain := strings.TrimPrefix(u.Path, "/")
	if pathWithoutDomain == u.Path {
		pathWithoutDomain = ""
	}

	u.Host = ""
	u.Scheme = ""
	newURL := u.String()
	if strings.Contains(newURL, u.Hostname()) {
		newURL = strings.Replace(newURL, u.Hostname(), "", 1)
	}
	if pathWithoutDomain != "" {
		newURL += "/" + pathWithoutDomain
	}

	return newURL, nil
}

// UrlIsEncoded 判断一个url是否是urlencode编码过的
func UrlIsEncoded(url string) bool {
	re := regexp.MustCompile("%[0-9A-Fa-f][0-9A-Fa-f]")
	return re.MatchString(url)
}

// UrlEncode 编码一个url
func UrlEncode(urlStr string) string {
	return url.QueryEscape(urlStr)
}

// UrlDecode 解码一个url
func UrlDecode(urlStr string) (string, error) {
	return url.QueryUnescape(urlStr)
}

// UrlMapToQueryString 将map[string]any转换为url.Values
func UrlMapToQueryString(mapData *map[string]any) string {
	paramValues := url.Values{}
	for k, v := range *mapData {
		paramValues.Set(k, cast.ToString(v))
	}
	return paramValues.Encode()
}

// UrlAppendParams 将map[string]any加入到url中
func UrlAppendParams(baseURL string, appendParams map[string]interface{}) string {
	originalURL, originalURLErr := url.Parse(baseURL)
	if originalURLErr != nil {
		return baseURL
	}
	params := originalURL.Query()
	for key, param := range appendParams {
		params.Add(key, UrlEncode(cast.ToString(param)))
	}
	originalURL.RawQuery = params.Encode()
	return originalURL.String()
}

// UrlAppendVals 将Url.Values加入到url中
func UrlAppendVals(baseURL string, params url.Values) string {
	originalURL, originalURLErr := url.Parse(baseURL)
	if originalURLErr != nil {
		return baseURL
	}
	originalURL.RawQuery = params.Encode()
	return originalURL.String()
}

func UrlGetHost(urlFull string) string {
	parsedURL, err := url.Parse(urlFull)
	if err != nil {
		return ``
	}
	host := parsedURL.Host
	return host
}

func UrlValid(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

func UrlGetContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}
	content, contentErr := io.ReadAll(resp.Body)
	if contentErr != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", contentErr)
	}
	return content, nil
}

func UrlGetParam(urlStr, field string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ``
	}
	queryParams := parsedURL.Query()
	return queryParams.Get(field)
}

func UrlGetFileName(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ``
	}
	return path.Base(u.Path)
}

// UrlParseParams 从URL字符串中解析参数
func UrlParseParams(urlStr string) ([]map[string]string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %v", err)
	}
	queryParams, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, fmt.Errorf("解析查询参数失败: %v", err)
	}
	params := make([]map[string]string, 0)
	for key, values := range queryParams {
		if len(values) > 0 {
			for _, value := range values {
				params = append(params, map[string]string{
					"key":   key,
					"value": value,
				})
			}
		}
	}
	return params, nil
}

// URLGetBase 获取域名和路径，移除所有参数
// 返回: 协议, 域名+路径（不包含协议前缀）
func URLGetBase(rawURL string) (string, string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ""
	}
	// 清除查询参数和片段
	u.RawQuery = ""
	u.Fragment = ""
	// 构建域名+路径部分（不包含协议）
	base := u.Host
	if u.Path != "" {
		base += u.Path
	}
	return u.Scheme, base
}
