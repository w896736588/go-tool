package controller

import (
	"dev_tool/base_module"
	"errors"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"strings"
)

var (
	cdCommand = `cd /var/www/`
)

// GitCurrentBranch 查询目录的git分支
func GitCurrentBranch(c *gin.Context) {
	_, reqMap, shell, command, err := getGitReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	codePath := reqMap[`CodePath`].ToStr()
	command.Sudo()
	command.Cd(cdCommand + codePath)
	command.GitShowBranch()
	result, err := shell.RunShell3(command.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
	} else {
		gsgin.GinResponse(c, gsgin.ResponseSuccess, result, ``)
	}
}

// GitChangeBranch 切换分支
func GitChangeBranch(c *gin.Context) {
	_, reqMap, shell, command, err := getGitReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	command1 := base_module.Command{}
	command1.Init()
	command1.Sudo()
	command1.Cd(cdCommand + reqMap[`CodePath`].ToStr())
	command1.GitShowBranch()
	currentBranch, err := shell.RunShell3(command1.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
		return
	}

	command.Sudo()
	command.Cd(cdCommand + reqMap[`CodePath`].ToStr())
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	if !strings.Contains(currentBranch, reqMap[`BranchName`].ToStr()) {
		command.GitCheckout(reqMap[`BranchName`].ToStr())
	}
	command.GitPullOrigin(reqMap[`BranchName`].ToStr())
	command.GitShowBranch()
	result, err := shell.RunShell3(command.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, result, ``)
}

// GitPullBranchOrigin 拉取当前分支最新代码
func GitPullBranchOrigin(c *gin.Context) {
	_, reqMap, shell, command, err := getGitReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	command1 := base_module.Command{}
	command1.Init()
	command1.Sudo()
	command1.Cd(cdCommand + reqMap[`CodePath`].ToStr())
	command1.GitShowBranch()
	currentBranch, err := shell.RunShell3(command1.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
		return
	}

	currentBranch = strings.Replace(currentBranch, "\n", "", -1)
	command.Sudo()
	command.Cd(cdCommand + reqMap[`CodePath`].ToStr())
	command.GitIgnoreAll()
	command.GitCleanAll()
	command.GitFetch()
	command.GitPull()
	command.GitPullOrigin(strings.Replace(currentBranch, "\n", "", -1))
	command.GitShowBranch()

	result, err := shell.RunShell3(command.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, result, ``)
}

// QueryStatus 查询分支状态
func QueryStatus(c *gin.Context) {
	_, reqMap, shell, command, err := getGitReqData(c)
	if err != nil {
		gsgin.GinResponse(c, gsgin.ResponseError, err.Error(), nil)
		return
	}
	command.Sudo()
	command.Cd(cdCommand + reqMap[`CodePath`].ToStr())
	command.GitStatus()

	result, err := shell.RunShell3(command.GetCommand().ToByte())
	if err != nil {
		BaseResponseByError(c, err)
		return
	}
	gsgin.GinResponse(c, gsgin.ResponseSuccess, result, ``)
}

//getGitReqData 基础方法
func getGitReqData(c *gin.Context) (*base_module.Global, map[string]*gstool.GsCons, *gstool.GsShell, *base_module.Command, error) {
	global, reqMap, err := GetGlobalReqParams(c)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	shellName := reqMap[`ShellName`]
	if shellName == nil {
		return nil, nil, nil, nil, errors.New(`缺少ShellName参数`)
	}
	client, err := global.ShellGetClient(shellName.ToStr())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	command := &base_module.Command{}
	command.Init()
	return global, reqMap, client, command, nil
}
