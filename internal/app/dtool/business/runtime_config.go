package business

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const (
	runtimeConfigDatabaseFileExt    = `.db`
	runtimeConfigLogDatabaseSuffix  = `.log`
)

// ReloadEditableRuntimeConfig 重新把当前 viper 中的可编辑配置同步到运行时环境。 // Reload editable config values from viper into the runtime environment.
func ReloadEditableRuntimeConfig() {
	if component.ConfigViper == nil || component.EnvClient == nil {
		return
	}
	if component.EnvClient.ConfigBase == nil {
		component.EnvClient.ConfigBase = &define.Base{}
	}

	component.EnvClient.ConfigBase.DbFileName = component.ConfigViper.GetString(`base.dbFileName`)
	component.EnvClient.ConfigBase.DbPath = component.ConfigViper.GetString(`base.dbPath`)
	component.EnvClient.ConfigBase.DbIsGitRepo = component.ConfigViper.GetBool(`base.dbIsGitRepo`)
	component.EnvClient.ConfigBase.MemoryDBPath = component.ConfigViper.GetString(`base.memoryDbPath`)
	component.EnvClient.ConfigBase.MemoryDBName = component.ConfigViper.GetString(`base.memoryDbFileName`)
	component.EnvClient.ConfigBase.MemoryDBIsGitRepo = component.ConfigViper.GetBool(`base.memoryDbIsGitRepo`)
	component.EnvClient.ConfigBase.WebPath = component.ConfigViper.GetString(`base.webPath`)

	if component.EnvClient.WebConfig == nil {
		component.EnvClient.WebConfig = &define.WebConfig{}
	}
	if component.EnvClient.ConfigBase.WebPath == `` {
		component.EnvClient.WebConfig.WebPath = filepath.Join(component.EnvClient.RootPath, `web`, `dist`)
	} else {
		component.EnvClient.WebConfig.WebPath = component.EnvClient.ConfigBase.WebPath
	}

	if component.EnvClient.DbConfig == nil {
		component.EnvClient.DbConfig = &define.DbConfig{}
	}
	component.EnvClient.DbConfig.DbPath = component.EnvClient.ConfigBase.DbPath
	component.EnvClient.DbConfig.DbIsGitRepo = component.EnvClient.ConfigBase.DbIsGitRepo
	component.EnvClient.DbConfig.DbName = component.EnvClient.AppName + `.db`
	if component.EnvClient.ConfigBase.DbFileName != `` {
		component.EnvClient.DbConfig.DbName = component.EnvClient.ConfigBase.DbFileName
	}
	if component.EnvClient.DbConfig.DbPath == `` {
		component.EnvClient.DbConfig.DbPath = filepath.Join(component.EnvClient.RootPath, `config`, component.EnvClient.AppName)
	}

	if component.EnvClient.LogDbConfig == nil {
		component.EnvClient.LogDbConfig = &define.DbConfig{}
	}
	component.EnvClient.LogDbConfig.DbName = buildRuntimeLogDBName(component.EnvClient.DbConfig.DbName)
	component.EnvClient.LogDbConfig.DbPath = component.EnvClient.DbConfig.DbPath

	drive := `C`
	if _, err := os.Stat(`D:\`); err == nil {
		drive = `D`
	}
	component.EnvClient.WebkitDriverPath = gstool.SReplaces(component.ConfigViper.GetString(`path.webkit_driver_path`), map[string]string{
		`{DRIVE}`: drive,
	})
	component.EnvClient.WebkitDataPath = gstool.SReplaces(component.ConfigViper.GetString(`path.webkit_data_path`), map[string]string{
		`{DRIVE}`: drive,
	})
	component.EnvClient.WebkitDownloadPath = gstool.SReplaces(component.ConfigViper.GetString(`path.webkit_download_path`), map[string]string{
		`{DRIVE}`: drive,
	})
}

// buildRuntimeLogDBName 基于主库文件名生成 log 库文件名。 // Build log db file name from the main database file name.
func buildRuntimeLogDBName(mainDBName string) string {
	if strings.HasSuffix(mainDBName, runtimeConfigDatabaseFileExt) {
		return strings.TrimSuffix(mainDBName, runtimeConfigDatabaseFileExt) + runtimeConfigLogDatabaseSuffix + runtimeConfigDatabaseFileExt
	}
	return mainDBName + runtimeConfigLogDatabaseSuffix + runtimeConfigDatabaseFileExt
}
