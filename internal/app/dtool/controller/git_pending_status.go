package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"path/filepath"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
)

// gitPendingStatusChecker 抽象 git 状态查询能力，方便复用和单元测试。
// gitPendingStatusChecker abstracts git status checks for reuse and unit testing.
type gitPendingStatusChecker interface {
	IsGitRepo(dir string) (bool, error)
	HasFileChanges(dir, fileName string) (bool, error)
}

// detectGitPendingStatus 统一计算主库和记忆库红点状态，确保提示口径与实际同步行为一致。
// detectGitPendingStatus computes badge flags and keeps them aligned with real sync behavior.
func detectGitPendingStatus(gitSyncer gitPendingStatusChecker, mainConfig business.MainDBConfig, memoryConfig common.MemoryConfig, hasMainDBEnv bool) (bool, bool) {
	mainDBPending := false
	memoryPending := false

	// 中文注释：主库同步实际只处理 sqlite 主文件，因此红点也只跟随主文件变更。
	// English comment: Main-db sync only targets the primary sqlite file, so the badge follows that file only.
	if hasMainDBEnv && mainConfig.GitRepoEnabled && mainConfig.Dir != `` && mainConfig.DBPath != `` {
		if isGit, err := gitSyncer.IsGitRepo(mainConfig.Dir); err == nil && isGit {
			fileName := filepath.Base(mainConfig.DBPath)
			if hasChanges, checkErr := gitSyncer.HasFileChanges(mainConfig.Dir, fileName); checkErr == nil && hasChanges {
				mainDBPending = true
			}
		}
	}

	// 中文注释：记忆库同步以整个目录为目标，所以这里继续使用目录级变更判断。
	// English comment: Memory sync operates on the whole directory, so the badge remains directory-based.
	if memoryConfig.GitRepoEnabled && memoryConfig.Dir != `` {
		if isGit, err := gitSyncer.IsGitRepo(memoryConfig.Dir); err == nil && isGit {
			if hasChanges, checkErr := gitSyncer.HasFileChanges(memoryConfig.Dir, `.`); checkErr == nil && hasChanges {
				memoryPending = true
			}
		}
	}

	return mainDBPending, memoryPending
}

// GitPendingStatus 检测主库和记忆库是否存在待提交的 git 变更。
// GitPendingStatus reports whether the main db or memory db has pending git changes.
func GitPendingStatus(c *gin.Context) {
	gitSyncer := business.NewMemoryGit()
	mainConfig := business.ReadMainDBConfig()
	memoryConfig := business.ReadMemoryConfigFromINI()
	mainDBPending, memoryPending := detectGitPendingStatus(gitSyncer, mainConfig, memoryConfig, component.EnvClient != nil && component.EnvClient.DbConfig != nil)

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`main_db_pending`: mainDBPending,
		`memory_pending`:  memoryPending,
	})
}
