package gsgin

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// GinPostBody 取出body数据转换为结构体
func GinPostBody(c *gin.Context, bodyStruct any) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return gstool.JsonDecode(cast.ToString(body), bodyStruct)
}

// GinGetParams 取出get参数
func GinGetParams(c *gin.Context) url.Values {
	return c.Request.URL.Query()
}

// GinResponse 返回
func GinResponse(c *gin.Context, errcode int, errmsg string, body interface{}) {
	returnJson := gstool.JsonEncode(&Response{
		Errcode: errcode,
		Errmsg:  errmsg,
		Data:    body,
	})
	c.String(http.StatusOK, returnJson)
}

// GinResponseError 返回
func GinResponseError(c *gin.Context, errmsg string, body interface{}) {
	GinResponse(c, CodeError, errmsg, body)
}

// GinResponseSuccess 返回
func GinResponseSuccess(c *gin.Context, errmsg string, body interface{}) {
	GinResponse(c, CodeSuccess, errmsg, body)
}

// GinPostBodyToMap post的json字符串转为map
func GinPostBodyToMap(c *gin.Context, postMap *map[string]interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	errMap := gstool.JsonDecode(cast.ToString(body), postMap)
	if errMap != nil {
		return errMap
	}
	return nil
}

// GinPostFormMulti 获取表单数据 包括文件 Content-Type: multipart/form-data
func GinPostFormMulti(c *gin.Context) (map[string][]string, map[string][]*multipart.FileHeader, error) {
	formData := make(map[string][]string)
	formFileData := make(map[string][]*multipart.FileHeader, 0)
	form, formErr := c.MultipartForm()
	if formErr != nil {
		return formData, formFileData, formErr
	}
	return form.Value, form.File, nil
}

// GinPostFormMultiOne 所有key都只拿第一个参数 不支持传递数组
func GinPostFormMultiOne(c *gin.Context) (map[string]string, map[string]*multipart.FileHeader, error) {
	formData := make(map[string]string)
	formFileData := make(map[string]*multipart.FileHeader)
	formList, formFileList, formErr := GinPostFormMulti(c)
	if formErr != nil {
		return formData, formFileData, formErr
	}
	for key, valueList := range formList {
		formData[key] = valueList[0]
	}
	for keyFile, fileList := range formFileList {
		formFileData[keyFile] = fileList[0]
	}
	return formData, formFileData, nil
}
