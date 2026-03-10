package plw

import (
	"dev_tool/internal/app/dtool/common"
	"errors"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// InfoCrawlPlanner 信息抓取规划器。
type InfoCrawlPlanner struct {
	Log *gstool.GsSlog
}

// NewInfoCrawlPlanner 创建信息抓取规划器。
func NewInfoCrawlPlanner(log *gstool.GsSlog) *InfoCrawlPlanner {
	return &InfoCrawlPlanner{Log: log}
}

// Plan 生成网页抓取规划。
func (h *InfoCrawlPlanner) Plan(taskInfo map[string]any, pageList []map[string]any) (map[int]map[string]any, string, map[string]any, error) {
	if len(pageList) == 0 {
		return nil, ``, nil, errors.New(`至少需要一个网页配置`)
	}
	systemPrompt := common.DbMain.InfoCrawlPlannerSystemPrompt()
	userPrompt := common.DbMain.InfoCrawlBuildPlannerUserPrompt(taskInfo, pageList)
	content, modelInfo, err := common.DbMain.InfoCrawlChatByModel(castTaskAiModelID(taskInfo), systemPrompt, userPrompt)
	if err != nil {
		return nil, ``, nil, err
	}
	plannerResult, err := common.DbMain.InfoCrawlParsePlannerResult(content)
	if err != nil {
		return nil, content, modelInfo, err
	}
	rawPages := make([]map[string]any, 0, len(plannerResult.Pages))
	for _, page := range plannerResult.Pages {
		actionList := make([]map[string]any, 0, len(page.Actions))
		for _, action := range page.Actions {
			actionList = append(actionList, map[string]any{
				`type`:    action.Type,
				`locator`: action.Locator,
				`value`:   action.Value,
				`out_key`: action.OutKey,
				`tip`:     action.Tip,
			})
		}
		rawPages = append(rawPages, map[string]any{
			`task_page_id`: page.TaskPageID,
			`goal`:         page.Goal,
			`actions`:      actionList,
		})
	}
	plannerMap := common.DbMain.InfoCrawlNormalizePlannerMap(rawPages)
	if err = common.DbMain.InfoCrawlValidatePlanner(castTaskID(taskInfo), pageList, plannerMap); err != nil {
		return nil, content, modelInfo, err
	}
	return plannerMap, content, modelInfo, nil
}

func castTaskAiModelID(taskInfo map[string]any) int {
	return max(0, cast.ToInt(taskInfo[`ai_model_id`]))
}

func castTaskID(taskInfo map[string]any) int {
	return max(0, cast.ToInt(taskInfo[`id`]))
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}

// InfoCrawlSummaryGenerator 生成汇总结果。
type InfoCrawlSummaryGenerator struct{}

// Build 构建汇总结果。
func (h *InfoCrawlSummaryGenerator) Build(taskInfo map[string]any, runTime string, runPageList []map[string]any) (string, error) {
	systemPrompt := common.DbMain.InfoCrawlSummarySystemPrompt()
	userPrompt := common.DbMain.InfoCrawlBuildSummaryUserPrompt(taskInfo, runTime, runPageList)
	content, _, err := common.DbMain.InfoCrawlChatByModel(cast.ToInt(taskInfo[`ai_model_id`]), systemPrompt, userPrompt)
	if err != nil {
		return ``, err
	}
	return content, nil
}
