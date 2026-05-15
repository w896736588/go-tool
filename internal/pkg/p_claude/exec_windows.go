//go:build windows

package p_claude

import (
	"bufio"
	"context"
	"log"
	"os/exec"
)

// startClaude Windows 实现。
// 与 chat_test.go 一致：stdout/stderr 均通过 goroutine 实时消费，
// cmd.Wait() 在后台 goroutine 执行，确保管道数据被实时读取。
func startClaude(ctx context.Context, args []string, workDir string, env []string) (ptyResult, error) {
	cmd := exec.Command(`claude`, args...)
	cmd.Dir = workDir
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ptyResult{}, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ptyResult{}, err
	}

	if err := cmd.Start(); err != nil {
		return ptyResult{}, err
	}

	lineCh := make(chan string, 256)
	stderrCh := make(chan string, 64)

	// 实时读取 stdout
	go func() {
		defer close(lineCh)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, maxScanTokenSize), maxScanTokenSize)
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}
	}()

	// 实时读取 stderr，收集内容用于错误定位
	go func() {
		defer close(stderrCh)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()
			log.Printf("[claude-exec] stderr: %s", text)
			stderrCh <- text
		}
	}()

	// 后台等待进程退出
	waitDone := make(chan struct{})
	var exitCode int
	var waitErr error
	go func() {
		defer close(waitDone)
		err := cmd.Wait()
		if err == nil {
			exitCode = 0
			return
		}
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
			return
		}
		exitCode = 1
		waitErr = err
	}()

	return ptyResult{
		lineCh:   lineCh,
		stderrCh: stderrCh,
		pid:      cmd.Process.Pid,
		waitFn: func() (int, error) {
			<-waitDone
			return exitCode, waitErr
		},
		closeFn: func() {
			cmd.Process.Kill()
		},
	}, nil
}
