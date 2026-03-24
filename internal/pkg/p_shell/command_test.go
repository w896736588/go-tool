package p_shell

import (
	"strings"
	"testing"
)

const (
	// dockerLogTruncateCommandSnippet 约束日志清理命令必须直接作用于 Docker 容器日志目录。
	dockerLogTruncateCommandSnippet = `truncate -s 0 /var/lib/docker/containers/*/*-json.log`
	// dockerSpaceAnalysisCommandSnippet 约束空间分析命令必须使用 docker inspect --size 获取容器占用。
	dockerSpaceAnalysisCommandSnippet = `docker inspect --size --format`
)

func TestDockerSpaceAnalysisCommand(t *testing.T) {
	commandText := NewCommand().Sudo().DockerSpaceAnalysis().GetCommand().ToStr()

	if !strings.Contains(commandText, dockerSpaceAnalysisCommandSnippet) {
		t.Fatalf("DockerSpaceAnalysis command = %q, want contains %q", commandText, dockerSpaceAnalysisCommandSnippet)
	}
	if strings.Contains(commandText, `sh -lc`) {
		t.Fatalf("DockerSpaceAnalysis command = %q, should avoid login shell for compatibility", commandText)
	}
	if !strings.Contains(commandText, `sh -c`) {
		t.Fatalf("DockerSpaceAnalysis command = %q, want contains %q", commandText, `sh -c`)
	}
}

func TestDockerContainerLogTruncateCommand(t *testing.T) {
	commandText := NewCommand().Sudo().DockerContainerLogTruncate().GetCommand().ToStr()

	if !strings.Contains(commandText, dockerLogTruncateCommandSnippet) {
		t.Fatalf("DockerContainerLogTruncate command = %q, want contains %q", commandText, dockerLogTruncateCommandSnippet)
	}
}
