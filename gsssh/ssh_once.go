package gsssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SshOnce struct {
	sshHandle *Ssh
}

func NewSshOnce(sshHandle *Ssh) *SshOnce {
	return &SshOnce{
		sshHandle: sshHandle,
	}
}

// RunCommandOnce 建立一个临时session执行一次
func (h *SshOnce) RunCommandOnce(shell string) (string, error) {
	var (
		session *ssh.Session
		err     error
	)
	if h.sshHandle.client == nil {
		connectionErr := h.sshHandle.ConnectAuthPassword()
		if connectionErr != nil {
			return ``, gstool.Error(`链接ssh失败 %s`, connectionErr.Error())
		}
	}
	if session, err = h.sshHandle.client.NewSession(); err != nil {
		return ``, err
	}
	defer func(session *ssh.Session) {
		closeErr := session.Close()
		if closeErr != nil && closeErr != io.EOF {
			gstool.FmtPrintlnLogTime(`关闭ssh失败 %s`, closeErr.Error())
		}
	}(session)
	output, outputErr := session.CombinedOutput(shell)
	return string(output), outputErr
}

// UploadFile 上传文件到远程
func (h *SshOnce) UploadFile(remoteFilePath, fileContent, filePath string) error {
	var (
		err error
	)
	if h.sshHandle.client == nil {
		connectionErr := h.sshHandle.ConnectAuthPassword()
		if connectionErr != nil {
			return connectionErr
		}
	}
	// 打开SFTP会话
	sftpSession, err := sftp.NewClient(h.sshHandle.client)
	if err != nil {
		gstool.FmtPrintlnLogTime(`打开sftp失败 %s`, err.Error())
		return err
	}
	// 打开本地文件
	// 如果是文件路径 localFile, err = os.Open(localFilePath) 还要加上close
	var localFile io.Reader
	if filePath != `` {
		// 如果是文件路径，打开文件
		file, opErr := os.Open(filePath)
		if opErr != nil {
			return opErr
		}
		defer func() {
			err := file.Close()
			if err != nil {
				gstool.FmtPrintlnLogTime(`关闭文件失败 %s`, err.Error())
				return
			}
		}()
		localFile = file
	} else {
		localFile = strings.NewReader(fileContent)
	}

	// 在远程服务器上创建文件
	remoteFile, err := sftpSession.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer func(remoteFile *sftp.File) {
		deferError := remoteFile.Close()
		if deferError != nil {
			gstool.FmtPrintlnLogTime(`%s`, deferError.Error())
		}
	}(remoteFile)

	// 复制文件内容
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return err
	}
	return nil
}

// UploadFileProcess 上传文件到远程，支持进度回调
func (h *SshOnce) UploadFileProcess(remoteFilePath, localFilePath string, progressCallback func(bytesWritten, totalBytes int64)) error {
	var (
		err        error
		totalBytes int64
	)
	if h.sshHandle.client == nil {
		connectionErr := h.sshHandle.ConnectAuthPassword()
		if connectionErr != nil {
			return gstool.Error(`链接ssh失败 %s`, connectionErr.Error())
		}
	}
	// 打开SFTP会话
	sftpSession, err := sftp.NewClient(h.sshHandle.client)
	if err != nil {
		gstool.FmtPrintlnLogTime("打开sftp失败 %s", err.Error())
		return err
	}
	defer func() {
		err := sftpSession.Close()
		if err != nil && err != io.EOF {
			gstool.FmtPrintlnLogTime(`关闭session失败 %v`, err)
		}
	}()

	// 准备本地文件内容
	var localFile io.Reader
	// 如果是文件路径，打开文件并获取大小
	file, opErr := os.Open(localFilePath)
	if opErr != nil {
		return opErr
	}
	defer func() {
		err := file.Close()
		if err != nil {
			gstool.FmtPrintlnLogTime(`关闭文件失败 %s`, err.Error())
			return
		}
	}()

	// 获取文件大小用于进度计算
	fileInfo, fileInfoErr := file.Stat()
	if fileInfoErr != nil {
		return fileInfoErr
	}
	totalBytes = fileInfo.Size()
	localFile = file

	// 在远程服务器上创建文件
	remoteFile, remoteFileErr := sftpSession.Create(remoteFilePath)
	if remoteFileErr != nil {
		return remoteFileErr
	}
	defer func() {
		if closeErr := remoteFile.Close(); closeErr != nil {
			gstool.FmtPrintlnLogTime("关闭远程文件失败: %s", closeErr.Error())
		}
	}()

	// 创建带进度跟踪的Reader
	progressRead := &progressReader{
		reader:           localFile,
		totalBytes:       totalBytes,
		progressCallback: progressCallback,
	}

	// 复制文件内容
	_, copyErr := io.Copy(remoteFile, progressRead)
	if copyErr != nil {
		return copyErr
	}

	// 确保最后一次进度回调是100%
	progressCallback(totalBytes, totalBytes)
	return nil
}

func (h *SshOnce) UploadFileSCP(remoteFilePath, localFilePath string, expired time.Duration) error {
	if h.sshHandle.client == nil {
		connectionErr := h.sshHandle.ConnectAuthPassword()
		if connectionErr != nil {
			return gstool.Error(`链接ssh失败 %s`, connectionErr.Error())
		}
	}
	// 1. 准备会话和文件
	session, err := h.sshHandle.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %v", err)
	}
	defer func() {
		err := session.Close()
		if err != nil {
			gstool.FmtPrintlnLogTime(`关闭session失败 %v`, err)
		}
	}()

	// 获取本地文件信息
	file, err := os.Open(localFilePath)
	if err != nil {
		return gstool.Error("Failed to open local file: %v", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			gstool.FmtPrintlnLogTime(`关闭文件失败 %s`, err.Error())
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return gstool.Error("Failed to get file info: %v", err)
	}

	// 准备SCP命令 - 确保远程路径是目录时以/结尾
	fileName := filepath.Base(localFilePath)
	scpCommand := fmt.Sprintf("scp -t %s", remoteFilePath)

	// 设置标准输入输出
	stdin, err := session.StdinPipe()
	if err != nil {
		return gstool.Error("Failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return gstool.Error("Failed to create stdout pipe: %v", err)
	}

	// 启动远程SCP进程
	if err := session.Start(scpCommand); err != nil {
		return gstool.Error("Failed to start scp command: %v", err)
	}

	// 发送文件元数据
	_, err = fmt.Fprintf(stdin, "C%04o %d %s\n", 0644, fileInfo.Size(), fileName)
	if err != nil {
		return gstool.Error("Failed to send file metadata: %v", err)
	}

	// 读取远程确认
	buf := make([]byte, 1)
	if _, err := stdout.Read(buf); err != nil {
		return gstool.Error("Failed to read remote confirmation: %v", err)
	}

	if buf[0] != 0 {
		return gstool.Error("Remote scp error: %v", buf)
	}

	// 发送文件内容
	if _, err := io.Copy(stdin, file); err != nil {
		return gstool.Error("Failed to send file content: %v", err)
	}

	// 发送结束标志并关闭stdin
	if _, err := fmt.Fprint(stdin, "\x00"); err != nil {
		return gstool.Error("Failed to send end marker: %v", err)
	}

	// 重要：关闭stdin管道，通知远程进程没有更多数据了
	if err := stdin.Close(); err != nil {
		return gstool.Error("Failed to close stdin: %v", err)
	}

	// 等待命令完成，但设置超时以防万一
	err = session.Wait()
	if err != nil {
		return gstool.Error("SCP command failed: %v", err)
	}
	return nil
}

func (h *SshOnce) UploadFileProcessScp(remoteFilePath, localFilePath string, progressCallback func(bytesWritten, totalBytes int64)) error {
	if h.sshHandle.client == nil {
		connectionErr := h.sshHandle.ConnectAuthPassword()
		if connectionErr != nil {
			return gstool.Error(`链接ssh失败 %s`, connectionErr.Error())
		}
	}
	// 1. 准备会话和文件
	session, err := h.sshHandle.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %v", err)
	}
	defer func() {
		err := session.Close()
		if err != nil && err != io.EOF {
			gstool.FmtPrintlnLogTime(`关闭session失败 %v`, err)
		}
	}()

	// 获取本地文件信息
	file, err := os.Open(localFilePath)
	if err != nil {
		return gstool.Error("Failed to open local file: %v", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			gstool.FmtPrintlnLogTime(`关闭文件失败 %s`, err.Error())
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return gstool.Error("Failed to get file info: %v", err)
	}

	// 准备SCP命令
	fileName := filepath.Base(localFilePath)
	scpCommand := fmt.Sprintf("scp -t %s", remoteFilePath)

	// 设置标准输入输出
	stdin, err := session.StdinPipe()
	if err != nil {
		return gstool.Error("Failed to create stdin pipe: %v", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return gstool.Error("Failed to create stdout pipe: %v", err)
	}

	// 启动远程SCP进程
	if err := session.Start(scpCommand); err != nil {
		return gstool.Error("Failed to start scp command: %v", err)
	}

	// 发送文件元数据
	_, err = fmt.Fprintf(stdin, "C%04o %d %s\n", 0644, fileInfo.Size(), fileName)
	if err != nil {
		return gstool.Error("Failed to send file metadata: %v", err)
	}

	// 读取远程确认
	buf := make([]byte, 1)
	if _, err := stdout.Read(buf); err != nil {
		return gstool.Error("Failed to read remote confirmation: %v", err)
	}

	if buf[0] != 0 {
		return gstool.Error("Remote scp error: %v", buf)
	}

	// 创建进度跟踪器
	totalBytes := fileInfo.Size()

	// 创建自定义reader来跟踪进度
	progressReader := &progressReader{
		reader:           file,
		totalBytes:       totalBytes,
		progressCallback: progressCallback,
	}

	// 发送文件内容
	if _, err := io.Copy(stdin, progressReader); err != nil {
		return gstool.Error("Failed to send file content: %v", err)
	}

	// 发送结束标志并关闭stdin
	if _, err := fmt.Fprint(stdin, "\x00"); err != nil {
		return gstool.Error("Failed to send end marker: %v", err)
	}

	if err := stdin.Close(); err != nil {
		return gstool.Error("Failed to close stdin: %v", err)
	}

	// 等待命令完成
	err = session.Wait()
	if err != nil {
		return gstool.Error("SCP command failed: %v", err)
	}
	progressCallback(totalBytes, totalBytes)
	return nil
}

type progressReader struct {
	reader           io.Reader
	bytesRead        int64
	totalBytes       int64
	progressCallback func(bytesWritten, totalBytes int64)
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.bytesRead += int64(n)

	if r.progressCallback != nil {
		r.progressCallback(r.bytesRead, r.totalBytes)
	}

	return n, err
}
