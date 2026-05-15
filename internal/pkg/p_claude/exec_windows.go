//go:build windows

package p_claude

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/UserExistsError/conpty"
)

// startClaude Windows 实现，使用 ConPTY API 创建伪终端，恢复子进程行缓冲。
func startClaude(ctx context.Context, args []string, workDir string, env []string) (ptyResult, error) {
	claudePath, err := exec.LookPath(`claude`)
	if err != nil {
		return ptyResult{}, fmt.Errorf("claude not found in PATH: %w", err)
	}
	cmdLine := buildWindowsCmdLine(append([]string{claudePath}, args...))

	cpty, err := conpty.Start(cmdLine,
		conpty.ConPtyWorkDir(workDir),
		conpty.ConPtyEnv(env),
		conpty.ConPtyDimensions(200, 40),
	)
	if err != nil {
		return ptyResult{}, err
	}

	closed := false
	closeOnce := func() {
		if !closed {
			closed = true
			cpty.Close()
		}
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			closeOnce()
		case <-done:
		}
	}()

	return ptyResult{
		reader: cpty,
		pid:    cpty.Pid(),
		waitFn: func() (int, error) {
			defer close(done)
			code, err := cpty.Wait(context.Background())
			return int(code), err
		},
		closeFn: func() { closeOnce() },
	}, nil
}
