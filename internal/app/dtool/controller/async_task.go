package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	_struct "dev_tool/internal/app/dtool/struct"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const (
	// asyncTaskTypeHomeTaskDailyReport 标识首页工作日报异步任务。 // asyncTaskTypeHomeTaskDailyReport identifies the home-task daily report async task.
	asyncTaskTypeHomeTaskDailyReport = `home_task_daily_report`
	// asyncTaskTypeMemoryFragmentArrange 标识知识片段整理异步任务。 // asyncTaskTypeMemoryFragmentArrange identifies the memory fragment arrange async task.
	asyncTaskTypeMemoryFragmentArrange = `memory_fragment_arrange`

	// asyncTaskActionSaveDailyReport 表示保存日报到知识片段。 // asyncTaskActionSaveDailyReport means saving the report as a memory fragment.
	asyncTaskActionSaveDailyReport = `save_daily_report`
	// asyncTaskActionOverwriteMemoryFragment 表示用整理结果覆盖知识片段。 // asyncTaskActionOverwriteMemoryFragment means overwriting the memory fragment with arranged content.
	asyncTaskActionOverwriteMemoryFragment = `overwrite_memory_fragment`
	// asyncTaskActionDiscard 表示丢弃结果。 // asyncTaskActionDiscard means discarding the async result.
	asyncTaskActionDiscard = `discard`
)

// asyncTaskBackgroundRunner 允许测试把后台执行切换成同步运行。 // asyncTaskBackgroundRunner allows tests to replace the goroutine runner with a synchronous runner.
var asyncTaskBackgroundRunner = func(_ int, run func()) {
	go run()
}

// buildAsyncHomeTaskDailyReportResult 允许测试替换日报结果构建过程。 // buildAsyncHomeTaskDailyReportResult allows tests to replace the daily report result builder.
var buildAsyncHomeTaskDailyReportResult = defaultBuildAsyncHomeTaskDailyReportResult

// buildAsyncMemoryArrangeResult 允许测试替换知识片段整理结果构建过程。 // buildAsyncMemoryArrangeResult allows tests to replace the memory arrange result builder.
var buildAsyncMemoryArrangeResult = defaultBuildAsyncMemoryArrangeResult

// AsyncTaskList 查询异步任务列表与汇总。 // AsyncTaskList returns async task summary plus recent items.
func AsyncTaskList(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskListRequest{}
	_ = gsgin.GinPostBody(c, &request)
	summary, err := db.AsyncTaskSummary(request.Limit)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, summary)
}

// AsyncTaskInfo 查询异步任务详情。 // AsyncTaskInfo returns a single async task detail.
func AsyncTaskInfo(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := db.AsyncTaskInfo(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// AsyncTaskDelete 删除异步任务记录。 // AsyncTaskDelete removes the async task record.
func AsyncTaskDelete(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if err := db.AsyncTaskDelete(request.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// AsyncTaskAction 对异步任务执行确认或丢弃操作。 // AsyncTaskAction applies confirm/discard actions to an async task result.
func AsyncTaskAction(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskActionRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := db.AsyncTaskInfo(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	action := strings.TrimSpace(request.Action)
	switch action {
	case asyncTaskActionDiscard:
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusRejected); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	case asyncTaskActionSaveDailyReport:
		resultMap, parseErr := asyncTaskDecodePayload(cast.ToString(info[`result_payload`]))
		if parseErr != nil {
			gsgin.GinResponseError(c, parseErr.Error(), nil)
			return
		}
		if strings.TrimSpace(cast.ToString(info[`task_type`])) != asyncTaskTypeHomeTaskDailyReport {
			gsgin.GinResponseError(c, `异步任务类型不支持保存日报`, nil)
			return
		}
		memoryDB, memoryOk := memoryDBOrResponse(c)
		if !memoryOk {
			return
		}
		memoryInfo, saveErr := memoryDB.MemoryFragmentSave(
			0,
			cast.ToString(resultMap[`report_title`]),
			cast.ToString(resultMap[`markdown`]),
			cast.ToStringSlice(resultMap[`suggested_tags`]),
		)
		if saveErr != nil {
			gsgin.GinResponseError(c, saveErr.Error(), nil)
			return
		}
		component.MemoryRuntime.ScheduleSync()
		broadcastMemoryFragmentUpsert(memoryInfo)
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusConfirmed); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		updatedInfo[`memory_fragment`] = memoryInfo
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	case asyncTaskActionOverwriteMemoryFragment:
		resultMap, parseErr := asyncTaskDecodePayload(cast.ToString(info[`result_payload`]))
		if parseErr != nil {
			gsgin.GinResponseError(c, parseErr.Error(), nil)
			return
		}
		if strings.TrimSpace(cast.ToString(info[`task_type`])) != asyncTaskTypeMemoryFragmentArrange {
			gsgin.GinResponseError(c, `异步任务类型不支持覆盖知识片段`, nil)
			return
		}
		memoryDB, memoryOk := memoryDBOrResponse(c)
		if !memoryOk {
			return
		}
		fragmentID := cast.ToString(resultMap[`fragment_id`])
		existingInfo, existingErr := memoryDB.MemoryFragmentInfo(fragmentID)
		if existingErr != nil {
			gsgin.GinResponseError(c, existingErr.Error(), nil)
			return
		}
		memoryInfo, saveErr := memoryDB.MemoryFragmentSave(
			fragmentID,
			cast.ToString(resultMap[`title`]),
			cast.ToString(resultMap[`arranged_content`]),
			cast.ToStringSlice(existingInfo[`tags`]),
		)
		if saveErr != nil {
			gsgin.GinResponseError(c, saveErr.Error(), nil)
			return
		}
		component.MemoryRuntime.ScheduleSync()
		broadcastMemoryFragmentUpsert(memoryInfo)
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusConfirmed); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		updatedInfo[`memory_fragment`] = memoryInfo
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	default:
		gsgin.GinResponseError(c, `异步任务操作不支持`, nil)
		return
	}
}

// createAsyncTask 创建任务并触发后台执行。 // createAsyncTask creates an async task record and starts background execution.
func createAsyncTask(taskType, title, sourceID string, requestPayload map[string]any, execute func(taskID int)) (map[string]any, error) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return nil, errors.New(`日志库未初始化`)
	}
	payloadText := gstool.JsonEncode(requestPayload)
	taskInfo, err := common.DbLog.AsyncTaskCreate(taskType, title, sourceID, payloadText)
	if err != nil {
		return nil, err
	}
	taskID := cast.ToInt(taskInfo[`id`])
	asyncTaskBackgroundRunner(taskID, func() {
		execute(taskID)
	})
	return taskInfo, nil
}

// runAsyncTaskAndPersistResult 统一处理后台执行状态流转。 // runAsyncTaskAndPersistResult centralizes async background status transitions and result persistence.
func runAsyncTaskAndPersistResult(taskID int, builder func() (map[string]any, error)) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return
	}
	if err := common.DbLog.AsyncTaskMarkRunning(taskID); err != nil {
		return
	}
	resultMap, err := builder()
	if err != nil {
		_ = common.DbLog.AsyncTaskMarkFailed(taskID, err.Error())
		return
	}
	resultPayload := gstool.JsonEncode(resultMap)
	if markErr := common.DbLog.AsyncTaskMarkAwaitConfirm(taskID, resultPayload); markErr != nil {
		return
	}
}

// asyncTaskDBOrResponse 返回日志库实例。 // asyncTaskDBOrResponse returns the log database instance or writes an error response.
func asyncTaskDBOrResponse(c *gin.Context) (*common.CSqlite, bool) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		gsgin.GinResponseError(c, `日志库未初始化`, nil)
		return nil, false
	}
	return common.DbLog, true
}

// asyncTaskDecodePayload 解析任务 JSON 结果。 // asyncTaskDecodePayload decodes the task JSON payload.
func asyncTaskDecodePayload(payload string) (map[string]any, error) {
	result := make(map[string]any)
	if strings.TrimSpace(payload) == `` {
		return result, nil
	}
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return nil, fmt.Errorf(`异步任务结果解析失败 %w`, err)
	}
	return result, nil
}

// defaultBuildAsyncHomeTaskDailyReportResult 构建日报异步任务结果。 // defaultBuildAsyncHomeTaskDailyReportResult builds the daily report async result.
func defaultBuildAsyncHomeTaskDailyReportResult(taskList []map[string]any, reportTime int64) (map[string]any, error) {
	reportAt := time.Unix(reportTime, 0)
	modelID, prompt, err := homeTaskDailyReportConfig()
	if err != nil {
		return nil, err
	}
	userPrompt, err := buildHomeTaskDailyReportUserPrompt(prompt, taskList, reportAt)
	if err != nil {
		return nil, err
	}
	result, modelInfo, err := common.DbMain.AIChatByModel(modelID, homeTaskDailyReportSystemPrompt(), userPrompt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`markdown`:       stripMarkdownCodeFence(result),
		`report_title`:   buildHomeTaskDailyReportTitle(reportAt),
		`prompt`:         prompt,
		`model_id`:       modelID,
		`model`:          modelInfo[`model`],
		`suggested_tags`: []string{homeTaskDailyReportMemoryTag},
	}, nil
}

// defaultBuildAsyncMemoryArrangeResult 构建知识片段整理异步任务结果。 // defaultBuildAsyncMemoryArrangeResult builds the memory fragment arrange async result.
func defaultBuildAsyncMemoryArrangeResult(title, content string) (map[string]any, error) {
	modelID, prompt, err := memoryArrangeConfig()
	if err != nil {
		return nil, err
	}
	userPrompt := buildMemoryArrangeUserPrompt(prompt, title, content)
	result, modelInfo, err := common.DbMain.AIChatByModel(modelID, memoryArrangeSystemPrompt(), userPrompt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`title`:            title,
		`original_content`: content,
		`arranged_content`: stripMarkdownCodeFence(result),
		`prompt`:           prompt,
		`model_id`:         modelID,
		`model`:            cast.ToString(modelInfo[`model`]),
	}, nil
}
