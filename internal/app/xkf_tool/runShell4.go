package xkf_tool

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"time"
)

var RunShellCli4Map *gstool.HighMap

// GetRunShellCli4 shell 4
func GetRunShellCli4(sshConfig *SshConfig) *gstool.GsShell {
	if sshConfig.Host == `` {
		return nil
	}
	uniKey := GetSshUnikey(sshConfig)
	if GetRunShellCli4FromMap(uniKey) != nil {
		return GetRunShellCli4FromMap(uniKey)
	}
	gsShellTerConfig := gstool.ShellConfig{
		Host:          sshConfig.Host,
		Port:          cast.ToInt64(sshConfig.Port),
		Username:      sshConfig.Username,
		Password:      sshConfig.Password,
		TimeoutSecond: 100,
	}
	cliTerConf := gstool.GsShell{
		Config:              &gsShellTerConfig,
		IsOpenLog:           true,
		Logger:              Logger,
		TerminalRefreshTime: 100 * time.Millisecond,
		TerminalMaxTime:     10 * time.Second,
	}
	cliTerConfErr := cliTerConf.CreateClient()
	if cliTerConfErr != nil {
		panic(`创建交互式链接失败 ` + cliTerConfErr.Error())
	}
	RunShellCli4Map.Set(uniKey, &cliTerConf)
	return GetRunShellCli4FromMap(uniKey)
}

func GetRunShellCli4FromMap(uniKey string) *gstool.GsShell {
	shellRet, boolRet := RunShellCli4Map.Get(uniKey)
	if !boolRet {
		gstool.FmtPrintlnLog(`获取失败 不存在 %s`, uniKey)
		return nil
	}
	return shellRet.(*gstool.GsShell)
}

func InitRunShell4(reqBody *SshExec) (*gstool.GsShell, *gstool.GsShell) {
	xkfRunShell4 := GetRunShellCli4(&reqBody.SshConfig)
	wkRunShell4 := GetRunShellCli4(&reqBody.WkSshConfig)
	return xkfRunShell4, wkRunShell4
}
