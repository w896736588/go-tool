package define

const (
	HomeTaskConfigDailyReportPrompt  = `home_task_daily_report_prompt`
	HomeTaskConfigDailyReportModelID = `home_task_daily_report_model_id`
	HomeTaskConfigFragmentPrompt     = `home_task_fragment_prompt`
	HomeTaskConfigTapdSmartLinkID    = `home_task_tapd_smart_link_id`
	HomeTaskConfigTapdLinkLabel      = `home_task_tapd_link_label`
	HomeTaskConfigTapdCssSelector    = `home_task_tapd_css_selector`
	HomeTaskConfigTapdWaitSeconds    = `home_task_tapd_wait_seconds`
	HomeTaskConfigZentaoSmartLinkID  = `home_task_zentao_smart_link_id`
	HomeTaskConfigZentaoLinkLabel    = `home_task_zentao_link_label`
	HomeTaskConfigZentaoCssSelector  = `home_task_zentao_css_selector`
	HomeTaskConfigZentaoWaitSeconds  = `home_task_zentao_wait_seconds`
	// HomeTaskConfigRequirementFetchConfigs 需求抓取自定义配置列表（JSON数组），替代旧独立 key。
	HomeTaskConfigRequirementFetchConfigs = `home_task_requirement_fetch_configs`
	// Deprecated: home_task_prompt_dev 已迁移到工作流模板系统，请使用模板管理。
	HomeTaskConfigPromptDev = `home_task_prompt_dev`
	// Deprecated: home_task_prompt_api_gen 已迁移到工作流模板系统，请使用模板管理。
	HomeTaskConfigPromptApiGen = `home_task_prompt_api_gen`
	// Deprecated: home_task_prompt_api_test 已迁移到工作流模板系统，请使用模板管理。
	HomeTaskConfigPromptApiTest = `home_task_prompt_api_test`
	// Deprecated: home_task_prompt_design 已迁移到工作流模板系统，请使用模板管理。
	HomeTaskConfigPromptDesign      = `home_task_prompt_design`
	HomeTaskConfigDevEnvironment    = `home_task_dev_environment`
	HomeTaskConfigBranchNamePrompt  = `home_task_branch_name_prompt`
	HomeTaskConfigBranchNameModelID = `home_task_branch_name_model_id`
	// Deprecated: home_task_prompt_plain_text_requirement 已迁移到工作流模板系统。
	HomeTaskConfigPromptPlainTextReq = `home_task_prompt_plain_text_requirement`
	// Deprecated: home_task_prompt_design_plan_requirement 已迁移到工作流模板系统。
	HomeTaskConfigPromptDesignPlanReq = `home_task_prompt_design_plan_requirement`
	// Deprecated: home_task_prompt_browser_test 已迁移到工作流模板系统。
	HomeTaskConfigPromptBrowserTest = `home_task_prompt_browser_test`
	// Deprecated: home_task_prompt_code_review 已迁移到工作流模板系统。
	HomeTaskConfigPromptCodeReview = `home_task_prompt_code_review`
	// Deprecated: home_task_prompt_issue_fix 已迁移到工作流模板系统。
	HomeTaskConfigPromptIssueFix = `home_task_prompt_issue_fix`

	DtoolAPIDefaultToken = `Test432` // 接口开发API的token默认值，避免占位符替换结果为空

	// 需求抓取配置类型——内置类型名称，新建任务默认 fetch_type 为 tapd。
	RequirementFetchTypeTapd   = `tapd`
	RequirementFetchTypeZentao = `zentao`
)

// RequirementFetchConfig 需求抓取配置条目（可自定义扩展）。
type RequirementFetchConfig struct {
	Name        string `json:"name"`          // 显示名称，如"TAPD"、"禅道"、"飞书需求"
	Type        string `json:"type"`          // 唯一标识，内置: tapd/zentao，自定义: 自动生成UUID
	SmartLinkID int    `json:"smart_link_id"` // 自定义网页ID
	LinkLabel   string `json:"link_label"`    // 网页链接label
	CssSelector string `json:"css_selector"`  // CSS选择器
	WaitSeconds int    `json:"wait_seconds"`  // 抓取前等待秒数
}

// IsBuiltin 判断是否为内置类型（tapd/zentao）。
func (c RequirementFetchConfig) IsBuiltin() bool {
	return c.Type == RequirementFetchTypeTapd || c.Type == RequirementFetchTypeZentao
}
