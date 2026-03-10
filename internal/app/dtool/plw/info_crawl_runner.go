package plw

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
)

// InfoCrawlRunner 信息抓取执行器。
type InfoCrawlRunner struct {
	RunID   int
	TaskID  int
	Log     *gstool.GsSlog
	Context *ContextPageList
}

// InfoCrawlPageResult 单网页执行结果。
type InfoCrawlPageResult struct {
	Status         string
	ErrorMessage   string
	PlannerAction  string
	ExecuteLog     string
	RawText        string
	RawHTML        string
	ScreenshotPath string
}

// NewInfoCrawlRunner 创建信息抓取执行器。
func NewInfoCrawlRunner(runID, taskID int, log *gstool.GsSlog) *InfoCrawlRunner {
	return &InfoCrawlRunner{
		RunID:   runID,
		TaskID:  taskID,
		Log:     log,
		Context: NewContextList(log),
	}
}

// OpenLoginPage 打开网页登录页。
func (h *InfoCrawlRunner) OpenLoginPage(pageInfo map[string]any) error {
	if PlaywrightClient == nil || PlaywrightClient.Pw == nil {
		return errors.New(`浏览器核心未启动`)
	}
	contextPage, err := h.getOrCreatePersistentContext(pageInfo, false)
	if err != nil {
		return err
	}
	page, err := (*contextPage.Context).NewPage()
	if err != nil {
		return err
	}
	if _, err = page.Goto(cast.ToString(pageInfo[`url`])); err != nil {
		return err
	}
	PlaywrightClient.WaitForLoadState(&page, 3000)
	return nil
}

// CheckLoginStatus 检查网页登录状态。
func (h *InfoCrawlRunner) CheckLoginStatus(pageInfo map[string]any) (bool, error) {
	contextPage, err := h.getOrCreatePersistentContext(pageInfo, true)
	if err != nil {
		return false, err
	}
	checkSelector := strings.TrimSpace(cast.ToString(pageInfo[`login_check_selector`]))
	if checkSelector == `` {
		return true, nil
	}
	targetPage, err := h.ensureContextPage(contextPage, cast.ToString(pageInfo[`url`]))
	if err != nil {
		return false, err
	}
	PlaywrightClient.WaitForLoadState(targetPage, 2000)
	waitErr := (*targetPage).Locator(checkSelector).First().WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(3000),
	})
	if waitErr != nil {
		return false, nil
	}
	return true, nil
}

// RunPage 执行单个网页抓取。
func (h *InfoCrawlRunner) RunPage(pageInfo map[string]any, planner map[string]any, logFunc func(string)) InfoCrawlPageResult {
	result := InfoCrawlPageResult{
		Status: define.InfoCrawlRunPageStatusSuccess,
	}
	if cast.ToInt(pageInfo[`login_status`]) != define.InfoCrawlPageLoginStatusOk {
		result.Status = define.InfoCrawlRunPageStatusLoginRequired
		result.ErrorMessage = `请先完成网页登录`
		return result
	}
	contextPage, err := h.getOrCreatePersistentContext(pageInfo, true)
	if err != nil {
		result.Status = define.InfoCrawlRunPageStatusFailed
		result.ErrorMessage = err.Error()
		return result
	}
	targetPage, err := h.ensureContextPage(contextPage, cast.ToString(pageInfo[`url`]))
	if err != nil {
		result.Status = define.InfoCrawlRunPageStatusFailed
		result.ErrorMessage = err.Error()
		return result
	}
	PlaywrightClient.WaitForLoadState(targetPage, 3000)
	goal := strings.TrimSpace(cast.ToString(planner[`goal`]))
	actionList := normalizeActionList(planner[`actions`])
	actionSummaryList := make([]string, 0, len(actionList))
	logList := make([]string, 0)
	textParts := make([]string, 0)
	for index, action := range actionList {
		actionType := strings.TrimSpace(cast.ToString(action[`type`]))
		actionSummary := fmt.Sprintf(`step%d %s %s`, index+1, actionType, strings.TrimSpace(cast.ToString(action[`locator`])))
		actionSummaryList = append(actionSummaryList, strings.TrimSpace(actionSummary))
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 步骤%d %s`, cast.ToString(pageInfo[`name`]), index+1, actionType))
		}
		logList = append(logList, actionSummary)
		if err = h.runAction(targetPage, pageInfo, action, &textParts, &logList, logFunc); err != nil {
			result.Status = define.InfoCrawlRunPageStatusFailed
			result.ErrorMessage = err.Error()
			break
		}
	}
	if goal == `` {
		goal = cast.ToString(pageInfo[`note`])
	}
	if len(textParts) == 0 && result.Status == define.InfoCrawlRunPageStatusSuccess {
		bodyText, bodyErr := (*targetPage).Locator(`body`).TextContent()
		if bodyErr == nil {
			textParts = append(textParts, strings.TrimSpace(bodyText))
		}
	}
	result.PlannerAction = strings.TrimSpace(goal + `; ` + strings.Join(actionSummaryList, `; `))
	result.ExecuteLog = strings.Join(logList, "\n")
	result.RawText = strings.TrimSpace(strings.Join(textParts, "\n\n"))
	rawHTML, htmlErr := (*targetPage).Content()
	if htmlErr == nil {
		result.RawHTML = rawHTML
	}
	result.ScreenshotPath = h.captureScreenshot(targetPage, cast.ToString(pageInfo[`name`]))
	return result
}

// runAction 执行单个白名单动作。
func (h *InfoCrawlRunner) runAction(page *playwright.Page, pageInfo map[string]any, action map[string]any, textParts *[]string, logList *[]string, logFunc func(string)) error {
	actionType := strings.TrimSpace(cast.ToString(action[`type`]))
	locator := strings.TrimSpace(cast.ToString(action[`locator`]))
	value := strings.TrimSpace(cast.ToString(action[`value`]))
	switch actionType {
	case define.InfoCrawlPlannerActionWait:
		waitMillis := cast.ToInt(value)
		if waitMillis <= 0 {
			waitMillis = 1500
		}
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 等待 %dms`, cast.ToString(pageInfo[`name`]), waitMillis))
		}
		time.Sleep(time.Duration(waitMillis) * time.Millisecond)
		*logList = append(*logList, fmt.Sprintf(`wait %dms`, waitMillis))
		return nil
	case define.InfoCrawlPlannerActionClick:
		if locator == `` {
			return errors.New(`click 动作缺少 locator`)
		}
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 点击 %s`, cast.ToString(pageInfo[`name`]), locator))
		}
		if err := (*page).Locator(locator).First().Click(); err != nil {
			return err
		}
		time.Sleep(1200 * time.Millisecond)
		return nil
	case define.InfoCrawlPlannerActionExistWait:
		if locator == `` {
			return errors.New(`exist_wait 动作缺少 locator`)
		}
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 等待元素出现 %s`, cast.ToString(pageInfo[`name`]), locator))
		}
		err := (*page).Locator(locator).First().WaitFor(playwright.LocatorWaitForOptions{
			Timeout: playwright.Float(5000),
		})
		return err
	case define.InfoCrawlPlannerActionNoExistWait:
		if locator == `` {
			return errors.New(`no_exist_wait 动作缺少 locator`)
		}
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 等待元素消失 %s`, cast.ToString(pageInfo[`name`]), locator))
		}
		endTime := time.Now().Add(5 * time.Second)
		for time.Now().Before(endTime) {
			count, err := (*page).Locator(locator).Count()
			if err == nil && count == 0 {
				return nil
			}
			time.Sleep(300 * time.Millisecond)
		}
		return errors.New(`等待元素消失超时`)
	case define.InfoCrawlPlannerActionTextContent:
		if locator == `` {
			locator = `body`
		}
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 提取文本 %s`, cast.ToString(pageInfo[`name`]), locator))
		}
		content, err := (*page).Locator(locator).First().TextContent()
		if err != nil {
			return err
		}
		content = strings.TrimSpace(content)
		if content != `` {
			*textParts = append(*textParts, fmt.Sprintf("[%s]\n%s", cast.ToString(pageInfo[`name`]), content))
		}
		return nil
	case define.InfoCrawlPlannerActionBoolResult:
		if locator == `` {
			return errors.New(`bool_result 动作缺少 locator`)
		}
		if logFunc != nil {
			logFunc(fmt.Sprintf(`[网页] %s 检查元素是否存在 %s`, cast.ToString(pageInfo[`name`]), locator))
		}
		count, err := (*page).Locator(locator).Count()
		if err != nil {
			return err
		}
		*logList = append(*logList, fmt.Sprintf(`bool_result %s=%t`, locator, count > 0))
		return nil
	default:
		return errors.New(`存在未授权动作`)
	}
}

// getOrCreatePersistentContext 获取或创建持久化 context。
func (h *InfoCrawlRunner) getOrCreatePersistentContext(pageInfo map[string]any, reusePage bool) (*ContextPage, error) {
	if PlaywrightClient == nil || PlaywrightClient.Pw == nil {
		return nil, errors.New(`浏览器核心未启动`)
	}
	pageID := cast.ToInt(pageInfo[`id`])
	userDataDir := strings.TrimSpace(cast.ToString(pageInfo[`user_data_dir`]))
	if userDataDir == `` {
		userDataDir = common.DbMain.InfoCrawlBuildUserDataDir(cast.ToInt(pageInfo[`task_id`]), cast.ToString(pageInfo[`name`]))
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_info_crawl_task_page`, map[string]any{`id`: pageID}, map[string]any{
			`user_data_dir`: userDataDir,
			`update_time`:   time.Now().Unix(),
		}).Exec()
	}
	linkID := fmt.Sprintf(`info_crawl_page_%d`, pageID)
	existContext := h.Context.FindContextList(func(context *ContextPage) *ContextPage {
		if context.LinkId == linkID || context.UserDataPath == userDataDir {
			return context
		}
		return nil
	})
	if existContext != nil {
		return existContext, nil
	}
	_ = gstool.DirCreatePath(userDataDir)
	runParams := &PlaywrightRunParams{
		Id:              pageID,
		Link:            cast.ToString(pageInfo[`url`]),
		LinkIdLabel:     fmt.Sprintf(`info_crawl_page_%d`, pageID),
		LinkId:          linkID,
		OpenType:        define.OpenTypeWebkitChrome,
		CombineType:     define.CombineTypeFix,
		AutoCloseSecond: 0,
		LocatorTimeout:  3000,
		GetPageTimeout:  8000,
		Channel:         `chromium`,
		Domain:          gstool.UrlGetHost(cast.ToString(pageInfo[`url`])),
		StreamFunc: func(string, string) {
		},
	}
	context, _, err := h.Context.GetContextSaveUserData(runParams)
	if err != nil {
		return nil, err
	}
	if !reusePage {
		return context, nil
	}
	return context, nil
}

// ensureContextPage 确保 context 中存在目标页面。
func (h *InfoCrawlRunner) ensureContextPage(contextPage *ContextPage, targetURL string) (*playwright.Page, error) {
	pageList := (*contextPage.Context).Pages()
	for _, currentPage := range pageList {
		if currentPage.URL() == targetURL {
			return &currentPage, nil
		}
	}
	page, err := (*contextPage.Context).NewPage()
	if err != nil {
		return nil, err
	}
	if _, err = page.Goto(targetURL); err != nil {
		return nil, err
	}
	return &page, nil
}

// captureScreenshot 保存网页截图。
func (h *InfoCrawlRunner) captureScreenshot(page *playwright.Page, pageName string) string {
	safeName := strings.NewReplacer(`\\`, `_`, `/`, `_`, `:`, `_`, `*`, `_`, `?`, `_`, `"`, `_`, `<`, `_`, `>`, `_`, `|`, `_`).Replace(pageName)
	if safeName == `` {
		safeName = `page`
	}
	dir := filepath.Join(common.DbMain.Env.WebkitDownloadPath, `info_crawl_shots`)
	_ = gstool.DirCreatePath(dir)
	filePath := filepath.Join(dir, fmt.Sprintf(`run_%d_%s.png`, h.RunID, safeName))
	_, err := (*page).Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(filePath),
		FullPage: playwright.Bool(true),
	})
	if err != nil {
		return ``
	}
	return filePath
}

func normalizeActionList(rawActions any) []map[string]any {
	result := make([]map[string]any, 0)
	switch value := rawActions.(type) {
	case []map[string]any:
		return value
	case []any:
		for _, item := range value {
			if actionMap, ok := item.(map[string]any); ok {
				result = append(result, actionMap)
			}
		}
	}
	return result
}
