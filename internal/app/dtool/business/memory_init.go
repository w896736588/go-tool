package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// ReadMemoryConfigFromINI 从 ini 读取记忆库配置 / read memory db config from ini.
func ReadMemoryConfigFromINI() common.MemoryConfig {
	if component.ConfigViper == nil {
		return common.MemoryConfig{}
	}
	memoryDir := strings.TrimSpace(component.ConfigViper.GetString(`base.memoryDbPath`))
	memoryDBName := strings.TrimSpace(component.ConfigViper.GetString(`base.memoryDbFileName`))
	config := common.MemoryConfig{
		Dir:    memoryDir,
		DBName: memoryDBName,
	}
	// 仅在目录和文件名都齐全时拼接完整路径 / build full path only when both path and file name exist.
	if config.Dir != `` && config.DBName != `` {
		config.DBPath = filepath.Join(config.Dir, config.DBName)
	}
	return config
}

func LoadMemoryStore() error {
	common.MemoryRuntime.Reset()

	config := ReadMemoryConfigFromINI()
	if config.Dir == `` || config.DBName == `` {
		gstool.FmtPrintlnLogTime(`记忆库未在配置文件中配置，跳过初始化`)
		return nil
	}
	if err := gstool.DirCreatePath(config.Dir); err != nil {
		return fmt.Errorf(`创建记忆目录失败 %w`, err)
	}

	memoryGit := NewMemoryGit()
	isGitRepo, err := memoryGit.IsGitRepo(config.Dir)
	if err != nil {
		return fmt.Errorf(`检测记忆目录 git 仓库失败 %w`, err)
	}
	if isGitRepo {
		if err = memoryGit.Pull(config.Dir); err != nil {
			return fmt.Errorf(`拉取记忆目录失败 %w`, err)
		}
	}

	memoryClient, err := p_db.InitSqlite(config.Dir, config.DBName)
	if err != nil {
		return fmt.Errorf(`连接记忆 sqlite 失败 %w`, err)
	}
	memoryDB := &common.CSqlite{Client: memoryClient, Env: component.EnvClient}
	NewMemoryDataBaseUp(memoryDB, component.EnvClient.MemoryDatabaseUpPath).Run()
	common.MemoryRuntime.SetGitSyncer(memoryGit)
	config.IsGitRepo = isGitRepo
	config.DBPath = filepath.Join(config.Dir, config.DBName)
	common.MemoryRuntime.Configure(config, memoryDB)
	return nil
}
