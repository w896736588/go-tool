package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// InfoCrawlTaskList 查询信息抓取任务列表。
func InfoCrawlTaskList(c *gin.Context) {
	list, err := common.DbMain.InfoCrawlTaskList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_list`: list,
	})
}

// InfoCrawlTaskInfo 查询信息抓取任务详情。
func InfoCrawlTaskInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `任务id不能为空`, nil)
		return
	}
	info, err := common.DbMain.InfoCrawlTaskInfo(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// InfoCrawlTaskSave 保存信息抓取任务。
func InfoCrawlTaskSave(c *gin.Context) {
	request := _struct.InfoCrawlTaskSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.InfoCrawlTaskSave(request.ID, request.Name, request.Prompt, request.AiModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// InfoCrawlTaskDelete 删除信息抓取任务。
func InfoCrawlTaskDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if err := common.DbMain.InfoCrawlTaskDelete(cast.ToInt(dataMap[`id`])); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// InfoCrawlTaskPageSave 保存网页配置。
func InfoCrawlTaskPageSave(c *gin.Context) {
	request := _struct.InfoCrawlTaskPageSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.InfoCrawlTaskPageSave(
		request.ID,
		request.TaskID,
		request.Name,
		request.URL,
		request.Note,
		request.LoginCheckSelector,
		request.Sort,
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// InfoCrawlTaskPageDelete 删除网页配置。
func InfoCrawlTaskPageDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if err := common.DbMain.InfoCrawlTaskPageDelete(cast.ToInt(dataMap[`id`])); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// InfoCrawlTaskPageOpenLogin 打开网页登录页。
func InfoCrawlTaskPageOpenLogin(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	pageID := cast.ToInt(dataMap[`task_page_id`])
	if pageID <= 0 {
		gsgin.GinResponseError(c, `网页id不能为空`, nil)
		return
	}
	if !ensureSmartLinkNodeInstalled(c, nil) {
		return
	}
	pageInfo, err := common.DbMain.InfoCrawlTaskPageRow(pageID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	runner := plw.NewInfoCrawlRunner(0, cast.ToInt(pageInfo[`task_id`]), plw.PlaywrightClient.Log)
	if err = runner.OpenLoginPage(pageInfo); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `请在浏览器中完成登录`, nil)
}

// InfoCrawlTaskPageCheckLogin 检查网页登录状态。
func InfoCrawlTaskPageCheckLogin(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	pageID := cast.ToInt(dataMap[`task_page_id`])
	if pageID <= 0 {
		gsgin.GinResponseError(c, `网页id不能为空`, nil)
		return
	}
	if !ensureSmartLinkNodeInstalled(c, nil) {
		return
	}
	pageInfo, err := common.DbMain.InfoCrawlTaskPageRow(pageID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	runner := plw.NewInfoCrawlRunner(0, cast.ToInt(pageInfo[`task_id`]), plw.PlaywrightClient.Log)
	ok, err := runner.CheckLoginStatus(pageInfo)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if ok {
		_ = common.DbMain.InfoCrawlTaskPageSetLoginStatus(pageID, define.InfoCrawlPageLoginStatusOk)
		gsgin.GinResponseSuccess(c, `登录状态正常`, map[string]any{
			`login_status`:      define.InfoCrawlPageLoginStatusOk,
			`login_status_desc`: `已登录`,
		})
		return
	}
	_ = common.DbMain.InfoCrawlTaskPageSetLoginStatus(pageID, define.InfoCrawlPageLoginStatusExpired)
	gsgin.GinResponseError(c, `未检测到登录状态，请重新登录`, map[string]any{
		`login_status`:      define.InfoCrawlPageLoginStatusExpired,
		`login_status_desc`: `登录失效`,
	})
}

// InfoCrawlTaskRun 执行信息抓取任务。
func InfoCrawlTaskRun(c *gin.Context) {
	request := _struct.InfoCrawlTaskRunRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.TaskID <= 0 {
		gsgin.GinResponseError(c, `任务id不能为空`, nil)
		return
	}
	if !ensureSmartLinkNodeInstalled(c, nil) {
		return
	}
	taskInfo, err := common.DbMain.InfoCrawlTaskRow(request.TaskID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	modelInfo, err := common.DbMain.InfoCrawlAiModelInfo(cast.ToInt(taskInfo[`ai_model_id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	pageList, err := common.DbMain.InfoCrawlTaskPageList(request.TaskID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if len(pageList) == 0 {
		gsgin.GinResponseError(c, `至少需要一个网页配置`, nil)
		return
	}
	common.DbMain.InfoCrawlSortPages(pageList)
	runID, err := common.DbMain.InfoCrawlRunCreate(request.TaskID, taskInfo, modelInfo)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: request.SseDistributeID,
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`run_id`:             runID,
		`status`:             define.InfoCrawlRunStatusRunning,
		`page_total`:         len(pageList),
		`page_success_total`: 0,
		`page_failed_total`:  0,
		`run_message`:        `任务已提交，正在后台执行`,
	})
	go runInfoCrawlTaskAsync(runID, request.TaskID, taskInfo, modelInfo, pageList, sse)
}

// runInfoCrawlTaskAsync 异步执行信息抓取任务。
func runInfoCrawlTaskAsync(runID, taskID int, taskInfo, modelInfo map[string]any, pageList []map[string]any, sse *p_sse.SseShell) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			stackText := string(debug.Stack())
			errorMessage := fmt.Sprintf(`后台执行异常：%v`, recoverErr)
			_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
				`status`:          define.InfoCrawlRunStatusFailed,
				`run_message`:     errorMessage,
				`page_total`:      len(pageList),
				`summary_content`: ``,
			})
			sse.Send(`[任务] `+errorMessage+"\n", `error`)
			gstool.FmtPrintlnLogTime(`info crawl async panic run_id=%d err=%v stack=%s`, runID, recoverErr, stackText)
		}
	}()
	_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
		`status`:      define.InfoCrawlRunStatusRunning,
		`run_message`: `任务已提交，正在后台执行`,
		`page_total`:  len(pageList),
	})
	planner := plw.NewInfoCrawlPlanner(plw.PlaywrightClient.Log)
	sse.Send(`[任务] 已进入后台执行，run_id=` + cast.ToString(runID) + "\n")
	sse.Send(`[任务] 开始执行 ` + cast.ToString(taskInfo[`name`]) + "\n")
	sse.Send(`[任务] 网页数量 ` + cast.ToString(len(pageList)) + `，AI模型 ` + cast.ToString(modelInfo[`name`]) + "\n")
	sse.Send(`[规划] 正在生成抓取计划` + "\n")
	plannerMap, plannerContent, _, err := planner.Plan(taskInfo, pageList)
	if err != nil {
		sse.Send(`[规划] 生成失败：`+err.Error()+"\n", `error`)
		_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
			`status`:          define.InfoCrawlRunStatusFailed,
			`run_message`:     `生成抓取计划失败：` + err.Error(),
			`planner_content`: plannerContent,
			`page_total`:      len(pageList),
		})
		return
	}
	_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
		`planner_content`: plannerContent,
		`page_total`:      len(pageList),
	})
	sse.Send(`[规划] 计划生成完成` + "\n")
	sse.Send(`[执行] 开始按网页顺序抓取` + "\n")
	runner := plw.NewInfoCrawlRunner(runID, taskID, plw.PlaywrightClient.Log)
	runPageList := make([]map[string]any, 0, len(pageList))
	successCount := 0
	failedCount := 0
	for pageIndex, pageInfo := range pageList {
		pageID := cast.ToInt(pageInfo[`id`])
		pagePlanner, ok := plannerMap[pageID]
		if !ok {
			pagePlanner = map[string]any{
				`goal`:    cast.ToString(pageInfo[`note`]),
				`actions`: []map[string]any{{`type`: define.InfoCrawlPlannerActionTextContent, `locator`: `body`, `value`: ``, `out_key`: `page_text`, `tip`: `抓取正文`}},
			}
		}
		sse.Send(`[网页] ` + cast.ToString(pageIndex+1) + `/` + cast.ToString(len(pageList)) + ` ` + cast.ToString(pageInfo[`name`]) + ` 开始执行` + "\n")
		pageResult := runner.RunPage(pageInfo, pagePlanner, func(msg string) {
			sse.Send(msg + "\n")
		})
		runPageRow := map[string]any{
			`run_id`:          runID,
			`task_page_id`:    pageID,
			`page_name`:       cast.ToString(pageInfo[`name`]),
			`url`:             cast.ToString(pageInfo[`url`]),
			`status`:          pageResult.Status,
			`error_message`:   pageResult.ErrorMessage,
			`planner_action`:  pageResult.PlannerAction,
			`execute_log`:     pageResult.ExecuteLog,
			`raw_text`:        pageResult.RawText,
			`raw_html`:        pageResult.RawHTML,
			`screenshot_path`: pageResult.ScreenshotPath,
		}
		_, _ = common.DbMain.InfoCrawlRunPageCreate(runPageRow)
		runPageList = append(runPageList, runPageRow)
		switch pageResult.Status {
		case define.InfoCrawlRunPageStatusSuccess:
			successCount++
			sse.Send(`[网页] ` + cast.ToString(pageInfo[`name`]) + ` 执行完成，已抓取文本 ` + cast.ToString(len(pageResult.RawText)) + ` 字` + "\n")
		case define.InfoCrawlRunPageStatusLoginRequired:
			failedCount++
			_ = common.DbMain.InfoCrawlTaskPageSetLoginStatus(pageID, define.InfoCrawlPageLoginStatusExpired)
			sse.Send(`[网页] `+cast.ToString(pageInfo[`name`])+` 需要重新登录`+"\n", `error`)
		default:
			failedCount++
			sse.Send(`[网页] `+cast.ToString(pageInfo[`name`])+` 执行失败：`+pageResult.ErrorMessage+"\n", `error`)
		}
	}
	summaryContent := ``
	runStatus := buildInfoCrawlRunStatus(successCount, failedCount)
	runMessage := buildInfoCrawlRunMessage(successCount, failedCount)
	if successCount > 0 {
		sse.Send(`[汇总] 正在生成总结` + "\n")
		runTime := gstool.TimeUnixToString(time.Now(), `Y-m-d H:i:s`)
		summaryGenerator := &plw.InfoCrawlSummaryGenerator{}
		summaryContent, err = summaryGenerator.Build(taskInfo, runTime, runPageList)
		if err != nil {
			sse.Send(`[汇总] 生成失败：`+err.Error()+"\n", `error`)
			if failedCount == 0 {
				runStatus = define.InfoCrawlRunStatusPartialFailed
			}
			runMessage = strings.TrimSpace(runMessage + `；AI汇总失败：` + err.Error())
		} else {
			sse.Send(`[汇总] 汇总完成` + "\n")
		}
	}
	_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
		`status`:             runStatus,
		`run_message`:        strings.Trim(runMessage, `；`),
		`summary_content`:    summaryContent,
		`page_total`:         len(pageList),
		`page_success_total`: successCount,
		`page_failed_total`:  failedCount,
	})
	sse.Send(`[任务] 执行完成，成功 ` + cast.ToString(successCount) + ` 个，失败 ` + cast.ToString(failedCount) + ` 个` + "\n")
}

// InfoCrawlRunList 查询执行历史。
func InfoCrawlRunList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	taskID := cast.ToInt(dataMap[`task_id`])
	if taskID <= 0 {
		gsgin.GinResponseError(c, `任务id不能为空`, nil)
		return
	}
	list, err := common.DbMain.InfoCrawlRunList(taskID, cast.ToInt(dataMap[`limit`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`run_list`: list,
	})
}

// InfoCrawlRunInfo 查询执行详情。
func InfoCrawlRunInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := cast.ToInt(dataMap[`id`])
	if id <= 0 {
		gsgin.GinResponseError(c, `执行记录id不能为空`, nil)
		return
	}
	info, err := common.DbMain.InfoCrawlRunInfo(id)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

func buildInfoCrawlRunStatus(successCount, failedCount int) string {
	if successCount > 0 && failedCount == 0 {
		return define.InfoCrawlRunStatusSuccess
	}
	if successCount > 0 && failedCount > 0 {
		return define.InfoCrawlRunStatusPartialFailed
	}
	return define.InfoCrawlRunStatusFailed
}

func buildInfoCrawlRunMessage(successCount, failedCount int) string {
	messageList := make([]string, 0)
	if successCount > 0 {
		messageList = append(messageList, `成功页面 `+cast.ToString(successCount)+` 个`)
	}
	if failedCount > 0 {
		messageList = append(messageList, `失败页面 `+cast.ToString(failedCount)+` 个`)
	}
	if len(messageList) == 0 {
		return `未执行任何页面`
	}
	return strings.Join(messageList, `；`)
}
