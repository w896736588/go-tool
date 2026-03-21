package controller

import (
	"dev_tool/internal/app/dtool/common"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// MemoryFragmentStatus 返回记忆库配置状态。
func MemoryFragmentStatus(c *gin.Context) {
	config := common.MemoryRuntime.Config()
	lastPushTime := common.MemoryRuntime.LastPushTime()
	lastPushError := common.MemoryRuntime.LastPushError()
	lastPushTimeDesc := `-`
	if lastPushTime > 0 {
		lastPushTimeDesc = gstool.TimeUnixToString(time.Unix(lastPushTime, 0), `Y-m-d H:i:s`)
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`configured`:          common.MemoryRuntime.IsConfigured(),
		`memory_dir`:          config.Dir,
		`memory_db_name`:      config.DBName,
		`is_git_repo`:         config.IsGitRepo,
		`last_push_time`:      lastPushTime,
		`last_push_time_desc`: lastPushTimeDesc,
		`last_push_error`:     lastPushError,
	})
}

// MemoryFragmentList 查询记忆片段列表。
func MemoryFragmentList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := memoryDB.MemoryFragmentList(cast.ToInt(dataMap[`limit`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentInfo 查询单个记忆片段详情。
func MemoryFragmentInfo(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	info, err := memoryDB.MemoryFragmentInfo(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// MemoryFragmentSave 保存记忆片段。
func MemoryFragmentSave(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	info, err := memoryDB.MemoryFragmentSave(
		cast.ToInt(dataMap[`id`]),
		cast.ToString(dataMap[`title`]),
		cast.ToString(dataMap[`content`]),
		memoryFragmentParseTags(dataMap[`tags`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	common.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, info)
}

// MemoryFragmentDelete 软删除记忆片段。
func MemoryFragmentDelete(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	_, err := memoryDB.MemoryFragmentSoftDelete(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	common.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// MemoryFragmentHistoryList 查询记忆片段历史记录。
func MemoryFragmentHistoryList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	list, err := memoryDB.MemoryFragmentHistoryList(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentTagList 查询记忆片段标签列表。
func MemoryFragmentTagList(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	list, err := memoryDB.MemoryFragmentTagList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// MemoryFragmentSearch 搜索记忆片段。
func MemoryFragmentSearch(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	list, err := memoryDB.MemoryFragmentSearch(
		cast.ToString(dataMap[`mode`]),
		cast.ToString(dataMap[`query`]),
		memoryFragmentParseTags(dataMap[`selected_tags`]),
		cast.ToInt(dataMap[`limit`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

func memoryDBOrResponse(c *gin.Context) (*common.CSqlite, bool) {
	if err := common.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`configured`: false,
		})
		return nil, false
	}
	return common.MemoryRuntime.DB(), true
}

// memoryFragmentParseTags 解析请求中的标签数组。
func memoryFragmentParseTags(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			result = append(result, cast.ToString(item))
		}
		return result
	case string:
		if value == `` {
			return []string{}
		}
		return []string{value}
	default:
		return []string{}
	}
}
