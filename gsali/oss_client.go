package gsali

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/w896736588/go-tool/gstool"
)

type OssConfig struct {
	Endpoint        string //例如http://oss-cn-hangzhou.aliyuncs.com 根据实际情况
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
}

// OssClientUploadFile 上传文件
// localFilePath 完整本地文件名
// filePath 存储的文件名。例如/abc/sdg/aaa.txt
func (h *OssConfig) OssClientUploadFile(localFilePath, filePath, bucketName string) (string, error) {
	defer func() {
		if gstool.FileIsExisted(localFilePath) {
			_ = gstool.FileDelete(localFilePath)
		}
	}()
	_, bucket, getError := h.OssGetClientBucket(bucketName)
	if getError != nil {
		return ``, getError
	}
	uploadErr := bucket.PutObjectFromFile(filePath, localFilePath)
	if uploadErr != nil {
		return ``, uploadErr
	}
	return bucketName + `.` + h.Endpoint + `/` + filePath, nil
}

// OssGetClientBucket 获取client和bucket
func (h *OssConfig) OssGetClientBucket(bucketName string) (*oss.Client, *oss.Bucket, error) {
	checkErr := h.check(bucketName)
	if checkErr != nil {
		return nil, nil, checkErr
	}
	client, clientErr := oss.New(h.Endpoint, h.AccessKeyId, h.AccessKeySecret)
	if clientErr != nil {
		return nil, nil, clientErr
	}
	bucket, bucketErr := client.Bucket(h.BucketName)
	if bucketErr != nil {
		return nil, nil, bucketErr
	}
	return client, bucket, nil
}

// OssDownloadFile 下载文件
// ossUrl oss文件地址
// downloadFilePath 需要下载到笨的文件地址
func (h *OssConfig) OssDownloadFile(ossUrl, downloadFilePath, bucketName string) error {

	objectName, objectNameErr := h.GetObjectName(ossUrl, bucketName)
	if objectNameErr != nil {
		return objectNameErr
	}
	_, bucket, getError := h.OssGetClientBucket(bucketName)
	if getError != nil {
		return getError
	}
	// 下载文件。
	downloadErr := bucket.GetObjectToFile(objectName, downloadFilePath)
	if downloadErr != nil {
		return downloadErr
	}
	return nil
}

// OssGetFileContent 获取文件内容
// ossUrl oss文件地址
// downloadFilePath 需要下载到笨的文件地址
func (h *OssConfig) OssGetFileContent(ossUrl, bucketName string) ([]byte, error) {
	objectName, objectNameErr := h.GetObjectName(ossUrl, bucketName)
	if objectNameErr != nil {
		return nil, objectNameErr
	}
	_, bucket, getError := h.OssGetClientBucket(bucketName)
	if getError != nil {
		return nil, getError
	}
	object, objectErr := bucket.GetObject(objectName)
	if objectErr != nil {
		return nil, objectErr
	}

	defer func(object io.ReadCloser) {
		closeErr := object.Close()
		if closeErr != nil {
			gstool.FmtPrintlnLog(`关闭读取器失败 %s`, closeErr.Error())
		}
	}(object)
	var buf bytes.Buffer
	_, bufErr := buf.ReadFrom(object)
	if bufErr != nil {
		return nil, bufErr
	}
	return buf.Bytes(), nil
}

// OssDeleteFile 删除文件
// ossUrl oss文件地址
// downloadFilePath 需要下载到笨的文件地址
func (h *OssConfig) OssDeleteFile(ossUrl, bucketName string) error {
	objectName, objectNameErr := h.GetObjectName(ossUrl, bucketName)
	if objectNameErr != nil {
		return objectNameErr
	}
	_, bucket, getError := h.OssGetClientBucket(bucketName)
	if getError != nil {
		return getError
	}
	deleteErr := bucket.DeleteObject(objectName)
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}

// GetObjectName 通过oss拿到objectName
func (h *OssConfig) GetObjectName(ossUrl, bucketName string) (string, error) {
	ossUrl = strings.Replace(ossUrl, `https://`, ``, 1)
	ossUrl = strings.Replace(ossUrl, `http://`, ``, 1)
	ossUrl = strings.Replace(ossUrl, bucketName+`.`, ``, 1)
	ossUrl = strings.Replace(ossUrl, h.Endpoint+`/`, ``, 1)
	return ossUrl, nil
}

func (h *OssConfig) check(bucketName string) error {
	if h.AccessKeySecret == `` || h.AccessKeyId == `` || h.Endpoint == `` {
		return errors.New(`配置不能为空`)
	}
	if bucketName == `` && h.BucketName == `` {
		return errors.New(`bucket不能为空`)
	}
	if bucketName != `` {
		h.BucketName = bucketName
	}
	return nil
}
