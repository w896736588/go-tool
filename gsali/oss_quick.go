package gsali

import (
	"path/filepath"
	"strings"
)

type OssQuick struct {
	OssClient *OssConfig
	Business  string
	//临时文件目录
	TempDir string
}

// OssQuickGetUrl OssGetUrl 通过内容获取oss的url txt文件
func (h *OssQuick) OssQuickGetUrl(content, business, bucketName string) (string, error) {
	return h.OssQuickGetUrlExt(content, business, bucketName, `.txt`)
}

// OssQuickGetUrlExt OssGetUrl 通过内容获取oss的url 自定义后缀
func (h *OssQuick) OssQuickGetUrlExt(content, business, bucketName, ext string) (string, error) {
	business = h.getBusiness(business)
	if h.TempDir == `` {
		panic(`临时文件目录不能为空`)
	}
	if !strings.HasPrefix(ext, `.`) {
		ext = `.` + ext
	}
	fileName := gstool.Md5(content)
	localPath := filepath.Join(h.TempDir, fileName+ext)
	fileCreateErr := gstool.FileCreate(h.TempDir, fileName+ext, content)
	if fileCreateErr != nil {
		return ``, fileCreateErr
	}
	date := gstool.DateCurrentDate2()
	objectName := filepath.Join(business, date, fileName+ext)
	bucketName = h.getBucketName(bucketName)
	ossUrl, err := h.OssClient.OssClientUploadFile(localPath, objectName, bucketName)
	if err != nil {
		return ``, err
	}
	return ossUrl, nil
}

// OssGetUrlByLocalFile 通过内容获取oss的
func (h *OssQuick) OssGetUrlByLocalFile(localPath, objectName, bucketName string) (string, error) {
	bucketName = h.getBucketName(bucketName)
	ossUrl, err := h.OssClient.OssClientUploadFile(localPath, objectName, bucketName)
	if err != nil {
		return ``, err
	}
	return ossUrl, nil
}

// OssQuickGetContent OssGetContent 通过url获取oss的内容
func (h *OssQuick) OssQuickGetContent(ossUrl, bucketName string) []byte {
	bucketName = h.getBucketName(bucketName)
	content, err := h.OssClient.OssGetFileContent(ossUrl, bucketName)
	if err != nil {
		return nil
	}
	return content
}

func (h *OssQuick) getBusiness(business string) string {
	if business != `` {
		return business
	}
	return h.Business
}

func (h *OssQuick) getBucketName(bucketName string) string {
	if bucketName != `` {
		return bucketName
	}
	return h.OssClient.BucketName
}
