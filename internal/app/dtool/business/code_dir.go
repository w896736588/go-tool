package business

import (
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"strings"
	"time"
)

type TCode struct {
}

func FindCode(sshConfig map[string]any, dirPath string) []string {
	codeDirList := make([]string, 0)
	command := p_shell.Command{}
	command.Sudo()
	command.FindGitDir(dirPath, 2)
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshConfig[`id`], `code`)
	//这里不需要输出sse 传空
	client, err := p_shell.ShellClient.GetClient(sshConfig, uniqueKey, &p_sse.SseShell{}, nil)
	if err != nil {
		gstool.FmtPrintlnLogTime(`连接ssh失败 %s`, err.Error())
		return codeDirList
	}
	defer p_shell.ShellClient.RmClient(uniqueKey)
	defer client.CloseTerminal()
	ret, retErr := client.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	if retErr != nil {
		gstool.FmtPrintlnLogTime(`执行命令失败 %s`, retErr.Error())
	} else {
		codeDirList = append(codeDirList, strings.Split(ret, "\n")...)
	}
	return codeDirList
}
