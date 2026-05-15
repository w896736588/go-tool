//go:build !windows

package p_claude

import (
	"context"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

// startClaude Unix 实现，使用 PTY 伪终端恢复子进程行缓冲。
func startClaude(ctx context.Context, args []string, workDir string, env []string) (ptyResult, error) {
	cmd := exec.CommandContext(ctx, `claude`, args...)
	cmd.Dir = workDir
	cmd.Env = env

	ptyFile, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 200, Rows: 40})
	if err != nil {
		return ptyResult{}, err
	}

	return ptyResult{
		reader: ptyFile,
		pid:    cmd.Process.Pid,
		waitFn: func() (int, error) {
			err := cmd.Wait()
			if err == nil {
				return 0, nil
			}
			if ee, ok := err.(*exec.ExitError); ok {
				if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
					return ws.ExitStatus(), nil
				}
				return 1, err
			}
			return 1, err
		},
		closeFn: func() {
			ptyFile.Close()
		},
	}, nil
}
