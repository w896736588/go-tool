package gsssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/w896736588/go-tool/gstool"
	"golang.org/x/crypto/ssh"
)

const (
	StatusWait = iota
	StatusRunning
	StatusError
	StatusStop
)

const SshBroken = `notice : ssh connection is broken`
const RunTimeOut = `notice : ssh command run time out`

type Hook struct {
	funcReceiveMsg func(msg string) string                                                //接收ssh输出的消息回调 将会以返回值替换传入的值输出
	funcBroken     func(msg string)                                                       //连接已中断
	funcAuthPrompt func(prompt string, stdin io.WriteCloser, session *ssh.Session) string //账号密码提示输入回调
}

type SshTerminal struct {
	sshHandle *Ssh
	//通道
	chanCommand    chan string //发送命令管道
	chanReceiveMsg chan string //接收命令管道
	//session
	session            *ssh.Session
	command            string             //当前执行的命令
	lockCommand        sync.Mutex         //命令锁
	waitPty            sync.WaitGroup     //等待pty启动完成
	waitCtxCancel      context.CancelFunc //等待命令返回结果
	runStatus          int                //0 待运行 1 运行中 2 运行失败
	runErr             error              //失败内容
	runResult          string             //命令结束之前的输出
	hook               Hook               //回调
	exceptionList      []string           //异常结束标记
	maxBuffer          int                //按行读取时，最大允许的容量 默认64 * 1024 也就是64kb
	ptyConfig          PtyConfig          //终端配置
	authPromptKeywords []string           //账号密码提示关键词 如果遇到将会ctrl+c结束并给与提示
}

type PtyConfig struct {
	Term   string `json:"term"   yaml:"term"`   // 终端类型
	Height int    `json:"height" yaml:"height"` // 行数
	Width  int    `json:"width"  yaml:"width"`  // 列数 会影响输出的内容是否换行
	// 把常用的 TerminalModes 单独拆字段，零值即“不设置”
	Echo        uint32 `json:"echo"           yaml:"echo"`       // 把用户敲的命令回显 0 关闭，1 开启
	RawMode     bool   `json:"raw_mode"       yaml:"raw_mode"`   // true 时把 ECHO、ICANON 等全部关掉
	InputSpeed  uint32 `json:"input_speed"  yaml:"input_speed"`  // 波特率
	OutputSpeed uint32 `json:"output_speed" yaml:"output_speed"` // 波特率
}

func NewSshTerminal(sshHandle *Ssh) *SshTerminal {
	return &SshTerminal{
		sshHandle: sshHandle,
	}
}

func (h *SshTerminal) SetPtyConfig(ptyConfig PtyConfig) {
	h.ptyConfig = ptyConfig
}

func (h *SshTerminal) SetFuncBroken(broken func(msg string)) {
	h.hook.funcBroken = broken
}

func (h *SshTerminal) SetFuncReceiveMsg(receive func(string) string) {
	h.hook.funcReceiveMsg = receive
}

func (h *SshTerminal) SetFuncAuthPrompt(prompt func(prompt string, stdin io.WriteCloser, session *ssh.Session) string) {
	h.hook.funcAuthPrompt = prompt
}

func (h *SshTerminal) SetMaxBufferSize(maxBuffer int) {
	if maxBuffer < 64*1024 {
		maxBuffer = 64 * 1024
	}
	h.maxBuffer = maxBuffer
}

func (h *SshTerminal) SetAuthPromptKeywords(authPromptKeywords []string) {
	h.authPromptKeywords = authPromptKeywords
}

// RunCommandWait 通过终端一次性执行命令 等待完成
func (h *SshTerminal) RunCommandWait(command string, maxRun time.Duration) (string, error) {
	defer func() {
		h.lockCommand.Unlock()
	}()
	h.lockCommand.Lock()
	if h.runStatus == StatusStop {
		return ``, errors.New(`连接已停止`)
	}
	if h.runStatus == StatusError {
		return ``, errors.New(`连接已中断，等待重连`)
	}
	checkErr := h.checkAndRunTerminal()
	if checkErr != nil {
		return ``, checkErr
	}
	//初始化
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), maxRun)
	defer cancel()
	h.command = command
	//先清空残留、设置收集器，再发送命令，避免命令发出后到准备完成之间的竞态窗口导致输出丢失
	h.runResult = ``
	h.waitCtxCancel = cancel
	h.chanCommand <- command
	waitErr := h.runWaitTimeout(ctx)
	result := h.runResult
	//命令结束后置空waitCtxCancel，阻止combineMsg在两次调用之间追加残留PTY输出
	h.waitCtxCancel = nil
	if waitErr != nil {
		if h.runStatus == StatusError {
			h.handleRunError(result, h.runErr)
		}
		return result, waitErr
	}
	if h.runStatus == StatusError {
		runErr := h.handleRunError(result, h.runErr)
		return result, runErr
	}
	return result, nil
}

// RunCommand 通过终端一次性执行命令 不等待完成
func (h *SshTerminal) RunCommand(command string) error {
	defer h.lockCommand.Unlock()
	h.lockCommand.Lock()

	if h.runStatus == StatusStop {
		return errors.New(`连接已停止`)
	}
	if h.runStatus == StatusError {
		return errors.New(`连接已中断，等待重连`)
	}
	checkErr := h.checkAndRunTerminal()
	if checkErr != nil {
		return checkErr
	}
	h.runResult = ``
	h.command = command
	h.chanCommand <- command
	if h.runStatus == StatusError {
		h.handleRunError(h.runResult, h.runErr)
	}
	return nil
}

func (h *SshTerminal) handleRunError(result string, runErr error) error {
	if runErr == nil {
		runErr = errors.New(`连接异常中断`)
	}
	h.toFunctionBroken(fmt.Sprintf(`result：%s,error：%s`, result, runErr.Error()))
	h.sshHandle.Close()
	return runErr
}

func (h *SshTerminal) runWaitTimeout(ctx context.Context) error {
	<-ctx.Done()
	switch ctxErr := ctx.Err(); {
	case errors.Is(ctxErr, context.DeadlineExceeded):
		timeoutErr := fmt.Errorf(`%s: %w`, RunTimeOut, context.DeadlineExceeded)
		h.setError(timeoutErr)
		h.toChanReceiveMsg(RunTimeOut)
		// 命令超时也按断开连接处理
		h.handleRunError(h.runResult, timeoutErr)
		return timeoutErr
	case errors.Is(ctxErr, context.Canceled):
		return nil
	default:
		unknownErr := fmt.Errorf("异常终止 %v", ctxErr)
		h.setError(unknownErr)
		return unknownErr
	}
}

func (h *SshTerminal) checkAndRunTerminal() error {
	if h.runStatus == StatusWait {
		h.waitPty.Add(1)
		go h.startTerminal()
		h.waitPty.Wait()
		if h.runErr != nil {
			return h.runErr
		}
	}
	return nil
}

func (h *SshTerminal) startTerminal() {
	var waitPtyDone sync.Once
	donePtyReady := func() {
		waitPtyDone.Do(h.waitPty.Done)
	}
	defer donePtyReady()
	if h.runStatus == StatusRunning {
		h.setError(errors.New(`正在运行中`))
		return
	}
	if h.runStatus == StatusError {
		h.runStatus = StatusWait
		h.runErr = nil
	}
	if h.sshHandle.client == nil {
		clientErr := h.sshHandle.ConnectAuthPassword()
		if clientErr != nil {
			h.setError(fmt.Errorf(`初始化client失败: %w`, clientErr))
			return
		}
	}
	var sessionErr error
	h.session, sessionErr = h.sshHandle.client.NewSession()
	if sessionErr != nil {
		h.setError(sessionErr)
		return
	}
	defer func() {
		if h.session != nil {
			sessionCloseErr := h.session.Close()
			if sessionCloseErr != nil {
				h.setError(sessionCloseErr)
			}
		}

	}()
	//启动pty
	if h.ptyConfig.InputSpeed == 0 {
		h.ptyConfig.InputSpeed = 14400
	}
	if h.ptyConfig.OutputSpeed == 0 {
		h.ptyConfig.OutputSpeed = 14400
	}
	if h.ptyConfig.Term == `` {
		h.ptyConfig.Term = "xterm"
	}
	if h.ptyConfig.Width == 0 {
		h.ptyConfig.Width = 300
	}
	if h.ptyConfig.Height == 0 {
		h.ptyConfig.Height = 100
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          h.ptyConfig.Echo,
		ssh.TTY_OP_ISPEED: h.ptyConfig.InputSpeed,
		ssh.TTY_OP_OSPEED: h.ptyConfig.OutputSpeed,
	}
	if ptyErr := h.session.RequestPty(h.ptyConfig.Term, h.ptyConfig.Height, h.ptyConfig.Width, modes); ptyErr != nil {
		h.setError(ptyErr)
		return
	}

	// 将会话的stdout和stderr设置为非阻塞的管道
	stdout, stdoutErr := h.session.StdoutPipe()
	if stdoutErr != nil {
		h.setError(stdoutErr)
		return
	}
	stderr, stderrErr := h.session.StderrPipe()
	if stderrErr != nil {
		h.setError(stderrErr)
		return
	}
	stdin, stdinErr := h.session.StdinPipe()
	if stdinErr != nil {
		h.setError(stdinErr)
		return
	}
	//启动
	if shellErr := h.session.Shell(); shellErr != nil {
		h.setError(shellErr)
		return
	}
	//初始化
	h.initParams()
	//接收终端输出
	go h.receiveMsg(stdin, stdout, stderr)
	// 发送命令到会话
	go h.receiveCommand(stdin)
	// 接收消息
	go h.combineMsg(stdin)
	h.runStatus = StatusRunning
	// PTY 启动完成后即可放行命令执行等待，不必等到 session 结束。
	donePtyReady()
	waitErr := h.session.Wait()
	if waitErr != nil {
		h.setError(waitErr)
		h.toChanReceiveMsg(SshBroken)
		time.Sleep(time.Second)
		h.sshHandle.Close()
	}
	return
}

func (h *SshTerminal) toChanReceiveMsg(msg string) {
	defer func() {
		if r := recover(); r != nil {
			gstool.FmtPrintlnLogTime(`尝试写入msg：%s失败 %v`, msg, r)
		}
	}()
	h.chanReceiveMsg <- msg
	return
}

func (h *SshTerminal) initParams() {
	h.chanCommand = make(chan string, 1) //为了保证同步执行 这里只允许一次执行一条命令
	h.chanReceiveMsg = make(chan string)
	h.exceptionList = []string{`-bash: syntax error`, SshBroken, RunTimeOut} //这种异常标记的是不会输出结束标记的
	h.authPromptKeywords = []string{}
}

// 等待输入命令 执行
func (h *SshTerminal) receiveCommand(stdin io.WriteCloser) {
	for {
		select {
		case command, ok := <-h.chanCommand:
			if !ok {
				return
			}
			cm := command + " \n"
			_, writeErr := stdin.Write([]byte(cm))
			if writeErr != nil {
				h.toChanReceiveMsg(SshBroken)
				h.setError(writeErr)
				return
			}
		}
	}
}

func (h *SshTerminal) receiveMsg(stdin io.WriteCloser, std, stderr io.Reader) {
	go func() {
		reader := io.MultiReader(std, stderr)
		buf := make([]byte, 512) // 小缓冲区，保证实时性
		for {
			if h.runStatus == StatusWait || h.runStatus == StatusStop {
				return
			}
			n, err := reader.Read(buf)
			if n > 0 {
				receiveMsg := string(buf[:n])
				h.toChanReceiveMsg(receiveMsg)
			}
			if err != nil {
				if err != io.EOF {
					h.toChanReceiveMsg(fmt.Sprintf("读取错误: %v", err))
				}
				return
			}
		}
	}()
}

func (h *SshTerminal) detectAuthPrompt(msg string) bool {
	if h.runStatus != StatusRunning {
		return false
	}
	if len(h.authPromptKeywords) == 0 {
		return false
	}
	for _, keyword := range h.authPromptKeywords {
		if strings.Contains(msg, keyword) {
			return true
		}
	}
	return false
}
func (h *SshTerminal) detectCommandEnd() bool {
	//常规错误，不会输出结束标记
	for _, exception := range h.exceptionList {
		if strings.Contains(h.runResult, exception) {
			return true
		}
	}
	reg, err := regexp.Compile(h.GetPatternRegex())
	if err != nil {
		gstool.FmtPrintlnLogTime("正则表达式编译失败：%s", err.Error())
		return true
	}
	return reg.MatchString(h.runResult)
}

func (h *SshTerminal) GetPatternRegex() string {
	return h.sshHandle.sshConfig.UserName + `@[a-zA-Z0-9\-\.]+:[^\s]+?[$#%]`
}

func (h *SshTerminal) FilterEndTip(msg string) string {
	reg, err := regexp.Compile(h.GetPatternRegex())
	if err != nil {
		gstool.FmtPrintlnLogTime("正则表达式编译失败：%s", err.Error())
		return msg
	}
	return reg.ReplaceAllString(msg, "")
}

func (h *SshTerminal) FilterCommand(msg string) string {
	return strings.Replace(msg, h.command, ``, 1)
}

func (h *SshTerminal) combineMsg(stdin io.WriteCloser) {
	for {
		select {
		case msg, ok := <-h.chanReceiveMsg:
			if !ok {
				return
			}
			//消息回调
			h.toFunctionReceiveMsg(msg)
			//执行输出的结果
			if h.waitCtxCancel == nil {
				continue
			}
			h.runResult += msg
			//检测是否包含账号/密码提示关键词
			if h.detectAuthPrompt(msg) {
				h.runResult = h.toFunctionAuthPrompt(h.runResult, stdin, h.session)
				continue
			}

			//检测命令结束或异常结束
			if h.detectCommandEnd() {
				h.setCommandEnd()
				continue
			}
		}
	}
}

func (h *SshTerminal) toFunctionBroken(msg string) {
	if h.hook.funcBroken != nil {
		h.hook.funcBroken(msg)
	}
}

func (h *SshTerminal) toFunctionReceiveMsg(msg string) {
	if h.hook.funcReceiveMsg != nil {
		msg = h.hook.funcReceiveMsg(msg)
	}
}

func (h *SshTerminal) toFunctionAuthPrompt(prompt string, stdin io.WriteCloser, session *ssh.Session) string {
	if h.hook.funcAuthPrompt != nil {
		return h.hook.funcAuthPrompt(prompt, stdin, session)
	}
	return prompt
}

func (h *SshTerminal) setCommandEnd(msg ...string) {
	if h.waitCtxCancel != nil {
		h.waitCtxCancel()
	}
	if len(msg) > 0 {
		if h.hook.funcReceiveMsg != nil {
			h.hook.funcReceiveMsg(msg[0])
		}
	}
}

func (h *SshTerminal) setError(err error) {
	h.runStatus = StatusError
	h.runErr = err
}

// CloseTerminal 主动关闭
func (h *SshTerminal) CloseTerminal() {
	if h.session == nil {
		return
	}
	if h.runStatus == StatusWait {
		return
	}
	h.runStatus = StatusStop
	if h.session != nil {
		closeSessionErr := h.session.Close()
		if closeSessionErr != nil {
			//这里可能会报错 不管
		}
	}
	if h.chanCommand != nil {
		close(h.chanCommand)
	}
	if h.chanReceiveMsg != nil {
		close(h.chanReceiveMsg)
	}
	h.session = nil
}
