//go:build windows

package p_claude

import (
	"log"
	"os/exec"
)

// cleanupOrphanedMcpProcesses Windows 实现。
// 通过 PowerShell 的 Get-CimInstance 查找命令行中包含 chrome-devtools-mcp 的 node 进程并终止。
// 用于 Go 进程崩溃重启后清理残留孤儿进程。
func cleanupOrphanedMcpProcesses() {
	//nolint:gosec // 清理孤儿进程的命令，参数由程序内部构造
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process -Filter "name='node.exe'" | Where-Object { $_.CommandLine -like '*chrome-devtools-mcp*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force }`,
	)
	if err := cmd.Run(); err != nil {
		log.Printf("[claude-exec] 清理残留 MCP 进程失败（忽略继续）: %v", err)
	}
}
