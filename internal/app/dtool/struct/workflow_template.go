package _struct

// WorkflowTemplateListRequest 获取所有模板列表请求。
type WorkflowTemplateListRequest struct {
	// 暂无参数，预留扩展
}

// WorkflowTemplateListResponse 模板列表项（含步骤）。
type WorkflowTemplateListResponse struct {
	ID          int                            `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	IsDefault   int                            `json:"is_default"`
	SortOrder   int                            `json:"sort_order"`
	CreateTime  int64                          `json:"create_time"`
	UpdateTime  int64                          `json:"update_time"`
	Steps       []WorkflowTemplateStepResponse `json:"steps"`
}

// WorkflowTemplateSaveRequest 创建/更新工作流程模板请求。
type WorkflowTemplateSaveRequest struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WorkflowTemplateDeleteRequest 删除工作流程模板请求。
type WorkflowTemplateDeleteRequest struct {
	ID int `json:"id"`
}

// WorkflowTemplateStepSaveRequest 创建/更新模板步骤请求。
type WorkflowTemplateStepSaveRequest struct {
	ID            int    `json:"id"`
	TemplateID    int    `json:"template_id"`
	Name          string `json:"name"`
	StepKey       string `json:"step_key"`
	PromptContent string `json:"prompt_content"`
	SortOrder     int    `json:"sort_order"`
}

// WorkflowTemplateStepDeleteRequest 删除模板步骤请求。
type WorkflowTemplateStepDeleteRequest struct {
	ID int `json:"id"`
}

// WorkflowTemplateStepSortRequest 步骤排序请求。
type WorkflowTemplateStepSortRequest struct {
	TemplateID int   `json:"template_id"`
	StepIDs    []int `json:"step_ids"`
}

// WorkflowTemplateStepResponse 模板步骤响应。
type WorkflowTemplateStepResponse struct {
	ID            int    `json:"id"`
	TemplateID    int    `json:"template_id"`
	Name          string `json:"name"`
	StepKey       string `json:"step_key"`
	PromptContent string `json:"prompt_content"`
	SortOrder     int    `json:"sort_order"`
	IsFixed       int    `json:"is_fixed"`
	CreateTime    int64  `json:"create_time"`
	UpdateTime    int64  `json:"update_time"`
}
