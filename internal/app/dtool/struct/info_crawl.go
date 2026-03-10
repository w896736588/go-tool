package _struct

// InfoCrawlTask 信息抓取任务。
type InfoCrawlTask struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Prompt     string `json:"prompt"`
	AiModelID  int    `json:"ai_model_id"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// InfoCrawlTaskPage 信息抓取网页配置。
type InfoCrawlTaskPage struct {
	ID                 int    `json:"id"`
	TaskID             int    `json:"task_id"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	Note               string `json:"note"`
	LoginCheckSelector string `json:"login_check_selector"`
	LoginStatus        int    `json:"login_status"`
	UserDataDir        string `json:"user_data_dir"`
	Sort               int    `json:"sort"`
	Status             int    `json:"status"`
	CreateTime         int64  `json:"create_time"`
	UpdateTime         int64  `json:"update_time"`
}

// InfoCrawlRun 信息抓取执行记录。
type InfoCrawlRun struct {
	ID               int    `json:"id"`
	TaskID           int    `json:"task_id"`
	Status           string `json:"status"`
	RunMessage       string `json:"run_message"`
	PromptSnapshot   string `json:"prompt_snapshot"`
	AiModelSnapshot  string `json:"ai_model_snapshot"`
	PlannerContent   string `json:"planner_content"`
	SummaryContent   string `json:"summary_content"`
	PageTotal        int    `json:"page_total"`
	PageSuccessTotal int    `json:"page_success_total"`
	PageFailedTotal  int    `json:"page_failed_total"`
	CreateTime       int64  `json:"create_time"`
	UpdateTime       int64  `json:"update_time"`
}

// InfoCrawlRunPage 信息抓取网页执行明细。
type InfoCrawlRunPage struct {
	ID             int    `json:"id"`
	RunID          int    `json:"run_id"`
	TaskPageID     int    `json:"task_page_id"`
	PageName       string `json:"page_name"`
	URL            string `json:"url"`
	Status         string `json:"status"`
	ErrorMessage   string `json:"error_message"`
	PlannerAction  string `json:"planner_action"`
	ExecuteLog     string `json:"execute_log"`
	RawText        string `json:"raw_text"`
	RawHTML        string `json:"raw_html"`
	ScreenshotPath string `json:"screenshot_path"`
	CreateTime     int64  `json:"create_time"`
	UpdateTime     int64  `json:"update_time"`
}

// InfoCrawlTaskSaveRequest 保存任务请求。
type InfoCrawlTaskSaveRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Prompt    string `json:"prompt"`
	AiModelID int    `json:"ai_model_id"`
}

// InfoCrawlTaskPageSaveRequest 保存网页请求。
type InfoCrawlTaskPageSaveRequest struct {
	ID                 int    `json:"id"`
	TaskID             int    `json:"task_id"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	Note               string `json:"note"`
	LoginCheckSelector string `json:"login_check_selector"`
	Sort               int    `json:"sort"`
}

// InfoCrawlTaskRunRequest 执行任务请求。
type InfoCrawlTaskRunRequest struct {
	TaskID          int    `json:"task_id"`
	SseDistributeID string `json:"sse_distribute_id"`
}

// InfoCrawlPlannerAction AI 规划的单个动作。
type InfoCrawlPlannerAction struct {
	Type    string `json:"type"`
	Locator string `json:"locator"`
	Value   string `json:"value"`
	OutKey  string `json:"out_key"`
	Tip     string `json:"tip"`
}

// InfoCrawlPlannerPage AI 规划的单个网页抓取方案。
type InfoCrawlPlannerPage struct {
	TaskPageID int                      `json:"task_page_id"`
	Goal       string                   `json:"goal"`
	Actions    []InfoCrawlPlannerAction `json:"actions"`
}

// InfoCrawlPlannerResult AI 规划结果。
type InfoCrawlPlannerResult struct {
	Pages []InfoCrawlPlannerPage `json:"pages"`
}
