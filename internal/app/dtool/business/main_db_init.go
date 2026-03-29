package business

import (
	"dev_tool/internal/app/dtool/component"
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const mainDBSyncCommitMessage = `chore: sync main db`

// mainDBGitSyncer 定义主库 git 同步所需能力。 // Defines the git sync capabilities required by the main database.
type mainDBGitSyncer interface {
	IsGitRepo(dir string) (bool, error)
	Pull(dir string) error
	HasFileChanges(dir, fileName string) (bool, error)
	AddFile(dir, fileName string) error
	Commit(dir, fileName, message string) error
	Push(dir string) error
}

// MainDBConfig 描述主库 sqlite 与 git 同步配置。 // Describes the main sqlite database and its git sync settings.
type MainDBConfig struct {
	Dir            string
	DBName         string
	DBPath         string
	IsGitRepo      bool
	GitRepoEnabled bool
}

type preparedMainDBBootstrap struct {
	Config MainDBConfig
	Git    mainDBGitSyncer
}

var preparedMainDBStore *preparedMainDBBootstrap

// newMainDBGitSyncer 允许测试替换 git 同步实现。 // Allows tests to replace the git sync implementation.
var newMainDBGitSyncer = func() mainDBGitSyncer {
	return NewMemoryGit()
}

// ReadMainDBConfig 读取已经解析过默认值的主库配置。 // Reads the resolved main database config after defaults are applied.
func ReadMainDBConfig() MainDBConfig {
	if component.EnvClient == nil || component.EnvClient.DbConfig == nil {
		return MainDBConfig{}
	}
	config := MainDBConfig{
		Dir:    strings.TrimSpace(component.EnvClient.DbConfig.DbPath),
		DBName: strings.TrimSpace(component.EnvClient.DbConfig.DbName),
	}
	if component.EnvClient.ConfigBase != nil {
		config.GitRepoEnabled = component.EnvClient.ConfigBase.DbIsGitRepo
	}
	if config.Dir != `` && config.DBName != `` {
		config.DBPath = filepath.Join(config.Dir, config.DBName)
	}
	return config
}

// PrepareMainDBStore 在主库初始化前按配置执行 git pull。 // Performs git pull before opening the main database when enabled.
func PrepareMainDBStore() error {
	config := ReadMainDBConfig()
	if config.Dir == `` || config.DBName == `` {
		preparedMainDBStore = nil
		return nil
	}
	if err := gstool.DirCreatePath(config.Dir); err != nil {
		return fmt.Errorf(`创建主库目录失败 %w`, err)
	}

	gitSyncer := newMainDBGitSyncer()
	// 只有显式开启 git 模式时才执行仓库检测与拉取。 // Only detect and pull when git mode is explicitly enabled.
	if config.GitRepoEnabled {
		isGitRepo, err := gitSyncer.IsGitRepo(config.Dir)
		if err != nil {
			return fmt.Errorf(`检测主库目录 git 仓库失败 %w`, err)
		}
		// 开启了 git 模式却不是仓库时直接失败，避免静默行为。 // Fail fast when git mode is enabled but the directory is not a repository.
		if !isGitRepo {
			return fmt.Errorf(`主库目录未检测到 git 仓库，请检查 base.dbIsGitRepo 和 dbPath 配置`)
		}
		if err = gitSyncer.Pull(config.Dir); err != nil {
			return fmt.Errorf(`拉取主库目录失败 %w`, err)
		}
		config.IsGitRepo = true
	}

	preparedMainDBStore = &preparedMainDBBootstrap{
		Config: config,
		Git:    gitSyncer,
	}
	if component.EnvClient != nil && component.EnvClient.DbConfig != nil {
		component.EnvClient.DbConfig.DbIsGitRepo = config.GitRepoEnabled
	}
	return nil
}

// SyncMainDBStoreOnShutdown 在程序关闭时检查主库文件并执行自动 push。 // Checks the main db file and performs auto-push during shutdown.
func SyncMainDBStoreOnShutdown() error {
	if preparedMainDBStore == nil {
		return nil
	}
	config := preparedMainDBStore.Config
	if !config.IsGitRepo {
		return nil
	}
	fileName := filepath.Base(config.DBPath)
	hasChanges, err := preparedMainDBStore.Git.HasFileChanges(config.Dir, fileName)
	if err != nil {
		return err
	}
	if !hasChanges {
		return nil
	}
	if err = preparedMainDBStore.Git.AddFile(config.Dir, fileName); err != nil {
		return err
	}
	if err = preparedMainDBStore.Git.Commit(config.Dir, fileName, mainDBSyncCommitMessage); err != nil {
		return err
	}
	if err = preparedMainDBStore.Git.Push(config.Dir); err != nil {
		return err
	}
	return nil
}
