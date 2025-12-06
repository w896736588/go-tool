package base

import (
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"strings"
	"time"
)

type TCode struct {
}

func (h *TCode) FindCode(sshConfig map[string]any, dirPath string) []string {
	codeDirList := make([]string, 0)
	command := Command{}
	command.Sudo()
	command.FindGitDir(dirPath, 2)
	uniqueKey := Component.TBase.GetCombineKey(sshConfig[`id`], `code`)
	//这里不需要输出sse 传空
	client, err := Component.TShell.GetClient(sshConfig, uniqueKey, ``, nil)
	if err != nil {
		gstool.FmtPrintlnLogTime(`连接ssh失败 %s`, err.Error())
		return codeDirList
	}
	defer Component.TShell.RmClient(uniqueKey)
	defer client.CloseTerminal()
	ret, retErr := client.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	if retErr != nil {
		gstool.FmtPrintlnLogTime(`执行命令失败 %s`, retErr.Error())
	} else {
		codeDirList = append(codeDirList, strings.Split(ret, "\n")...)
	}
	return codeDirList
}
