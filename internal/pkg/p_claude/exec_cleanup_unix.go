//go:build !windows

package p_claude

import (
	"log"
	"os/exec"
)

// cleanupOrphanedMcpProcesses Unix（macOS/Linux）实现。
// 通过 pkill 杀死所有命令行中包含 chrome-devtools-mcp 的残留进程。
// 用于 Go 进程崩溃重启后清理残留孤儿进程。
func cleanupOrphanedMcpProcesses() {
	//nolint:gosec // 清理孤儿进程的命令，参数由程序内部构造
	cmd := exec.Command("pkill", "-f", "chrome-devtools-mcp")
	// pkill 在找不到匹配进程时返回 1，视为正常
	if err := cmd.Run(); err != nil {
		log.Printf("[claude-exec] 清理残留 MCP 进程失败（忽略继续）: %v", err)
	}
}
