package controller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dev_tool/internal/app/dtool/component"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
)

// agentDownloadSpec 定义客户端下载时需要的目标平台和文件名。
// agentDownloadSpec describes the target platform and artifact filename for client downloads.
type agentDownloadSpec struct {
	Goos     string
	FileName string
}

// resolveAgentDownloadSpec 根据 os 查询参数解析目标平台，避免下载请求打到未知平台。
// resolveAgentDownloadSpec maps the os query parameter to a supported download target.
func resolveAgentDownloadSpec(rawOS string) (agentDownloadSpec, bool) {
	switch strings.ToLower(strings.TrimSpace(rawOS)) {
	case "windows":
		return agentDownloadSpec{
			Goos:     "windows",
			FileName: "dtool-agent.exe",
		}, true
	case "darwin":
		return agentDownloadSpec{
			Goos:     "darwin",
			FileName: "dtool-agent",
		}, true
	case "linux":
		return agentDownloadSpec{
			Goos:     "linux",
			FileName: "dtool-agent",
		}, true
	default:
		return agentDownloadSpec{}, false
	}
}

// buildAgentDownloadBinary 按目标平台现编 dtool-agent，确保下载入口至少能返回可执行文件。
// buildAgentDownloadBinary builds the dtool-agent binary on demand so the download endpoint can return a runnable artifact.
func buildAgentDownloadBinary(spec agentDownloadSpec) (string, error) {
	rootPath := component.EnvClient.RootPath
	outputDir := filepath.Join(rootPath, "tmp", "agent_download")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	outputPath := filepath.Join(outputDir, spec.FileName)
	buildCmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", outputPath, "./cmd/dtool-agent")
	buildCmd.Dir = rootPath
	buildCmd.Env = append(
		os.Environ(),
		"GOOS="+spec.Goos,
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	)

	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		return "", fmt.Errorf("build dtool-agent failed: %w, output: %s", buildErr, string(buildOutput))
	}

	return outputPath, nil
}

// AgentDownload 提供本地客户端安装包下载，当前实现为按目标平台即时构建二进制。
// AgentDownload serves the local client installer by building the agent binary for the requested platform on demand.
func AgentDownload(c *gin.Context) {
	spec, ok := resolveAgentDownloadSpec(c.Query("os"))
	if !ok {
		gsgin.GinResponseError(c, "unsupported os, expected windows, darwin or linux", nil)
		return
	}

	binaryPath, err := buildAgentDownloadBinary(spec)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.FileAttachment(binaryPath, spec.FileName)
}
