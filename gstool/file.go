package gstool

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/xuri/excelize/v2"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileCreate 文件操作
func FileCreate(dirPath, fileName, fileContent string) error {
	exists, existErr := DirPathExists(dirPath)
	if existErr != nil {
		return existErr
	}
	if !exists {
		createDirErr := DirCreatePath(dirPath)
		if createDirErr != nil {
			return createDirErr
		}
	}
	filePath := filepath.Join(dirPath, fileName)
	return FilePutContent(filePath, fileContent)
}

// FilePutContent 向文件中添加内容
func FilePutContent(filePath, fileContent string) error {
	filePath = strings.Replace(filePath, `//`, `/`, 1)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			FmtPrintlnLog(`关闭文件失败 %s`, closeErr.Error())
		}
	}(file)
	_, err = file.Write([]byte(fileContent))
	if err != nil {
		return err
	}
	return nil
}

// FilePutContentCover 向文件中添加内容 覆盖
func FilePutContentCover(filePath, fileContent string) error {
	filePath = strings.Replace(filePath, `//`, `/`, 1)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			FmtPrintlnLog(`关闭文件失败 %s`, closeErr.Error())
		}
	}(file)
	_, err = file.Write([]byte(fileContent))
	if err != nil {
		return err
	}
	return nil
}

// FileDelete 删除文件
func FileDelete(filePath string) error {
	if FileIsExisted(filePath) {
		removeErr := os.Remove(filePath)
		if removeErr != nil {
			return removeErr
		}
	}
	return nil
}

// FileIsExisted 检查文件是否存在
func FileIsExisted(filePath string) bool {
	if strings.TrimSpace(filePath) == "" {
		return false
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// FileGetContent 读取文件内容
func FileGetContent(filePath string) (string, error) {
	if !FileIsExisted(filePath) {
		return ``, errors.New(`文件不存在 ` + filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return ``, err
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			FmtPrintlnLog(`关闭文件失败 %s `, closeErr.Error())
		}
	}(file) // 确保文件在使用完毕后关闭

	// 读取文件内容
	content, readErr := io.ReadAll(file)
	if readErr != nil {
		return ``, errors.New(`读取内容失败 ` + readErr.Error())
	}
	return string(content), nil
}

// FileSearchStrLineNumDesc 查找字符串在文件中的行数 从文件末尾开始
func FileSearchStrLineNumDesc(filePath, search string) (int, error) {
	if !FileIsExisted(filePath) {
		return 0, errors.New(`文件不存在 ` + filePath)
	}
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		return 0, fileErr
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			FmtPrintlnLog(`关闭文件失败 %s %s`, filePath, closeErr.Error())
		}
	}(file)

	reader := bufio.NewReader(file)
	var lineList []string
	var lineNum int

	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return 0, readErr
		}
		lineList = append(lineList, strings.TrimSuffix(line, "\n")) // 去除行尾的换行符
		lineNum++
	}
	for i := len(lineList) - 1; i >= 0; i-- {
		if strings.Contains(lineList[i], search) {
			return i, nil
		}
	}
	return 0, nil
}

// FileInsertToLine 在指定行插入内容
func FileInsertToLine(filePath, insertContent string, lineNum int) error {
	if !FileIsExisted(filePath) {
		return errors.New(`文件不存在 ` + filePath)
	}
	content, contentErr := FileGetContent(filePath)
	if contentErr != nil {
		return contentErr
	}

	lineList := strings.Split(string(content), "\n")
	if lineNum < 1 || lineNum > len(lineList) {
		return fmt.Errorf(`错误的行数: %d`, lineNum)
	}
	resultLineList := make([]string, 0)
	for index, line := range lineList {
		if index+1 == lineNum {
			resultLineList = append(resultLineList, insertContent)
		}
		resultLineList = append(resultLineList, line)
	}
	newContent := strings.Join(resultLineList, "\n") + "\n"
	return FilePutContent(filePath, newContent)
}

// FileInsertToStrPosition 在指定字符串的下一行插入字符串
func FileInsertToStrPosition(filePath, insertContent, searchStr string) error {
	searchNum, searchErr := FileSearchStrLineNumDesc(filePath, searchStr)
	if searchErr != nil {
		return searchErr
	}
	insertErr := FileInsertToLine(filePath, insertContent, searchNum)
	if insertErr != nil {
		return insertErr
	}
	return nil
}

func FileGetNameByPath(filePath string) string {
	return filepath.Base(filePath)
}

// FileExtType 获取文件类型(读取js生成的csv时不对，会被识别为zip)
func FileExtType(localFilePath string) (*types.Type, error) {
	file, err := os.Open(localFilePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			FmtPrintlnLogTime(`close file error :%s`, closeErr.Error())
		}
	}(file)

	// 限制读取前261字节（即filetype库默认的大小限制）
	head := make([]byte, 261)
	_, err = file.Read(head)
	if err != nil {
		return nil, err
	}

	// 检测文件类型
	kind, kindErr := filetype.Match(head)
	if kindErr != nil {
		return nil, kindErr
	}
	if kind.Extension == `unknown` {
		ext := filepath.Ext(FileGetNameByPath(localFilePath))
		switch ext {
		case `txt`:
			kind.Extension = `text`
			kind.MIME.Type = `text`
			kind.MIME.Value = `text/plain`
			kind.MIME.Subtype = `plain`
		}
	}
	return &kind, nil
}

func FileIsCsv(filePath string) bool {
	// 打开 CSV 文件
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	reader := csv.NewReader(file)
	_, headerErr := reader.Read()
	if headerErr != nil {
		return false
	}
	for {
		_, recordErr := reader.Read()
		if recordErr != nil {
			if recordErr.Error() == "EOF" {
				break // 文件读取完毕
			}
			return false
		}
	}
	return true
}

func FileIsXlsx(filePath string) bool {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Close()
	}()
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return false
	}
	return true
}

func FileIsTxt(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	mimeType := http.DetectContentType(data)
	if strings.Contains(mimeType, `text/plain`) {
		return true
	}
	return false
}

func FileSize(filePath, format string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	size := float64(fileInfo.Size())
	var result float64
	var unit string

	switch strings.ToLower(format) {
	case "bit":
		result = size * 8
		unit = "bit"
	case "b":
		result = size
		unit = "B"
	case "kb":
		result = size / 1024
		unit = "KB"
	case "mb":
		result = size / (1024 * 1024)
		unit = "MB"
	case "gb":
		result = size / (1024 * 1024 * 1024)
		unit = "GB"
	default:
		return "", fmt.Errorf("不支持的格式: %s (支持: bit, b, kb, mb, gb)", format)
	}
	if result == math.Trunc(result) {
		return fmt.Sprintf("%.0f %s", result, unit), nil
	}
	return fmt.Sprintf("%.2f %s", result, unit), nil
}
