package controller

import (
	"net/http"

	"dev_tool/internal/app/dtool/common"
	_struct "dev_tool/internal/app/dtool/struct"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// WorkflowTemplateList 获取所有工作流程模板列表（含步骤）。
func WorkflowTemplateList(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	templates, err := common.DbMain.WorkflowTemplateList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: templates,
	})
}

// WorkflowTemplateSave 创建/更新工作流程模板。
func WorkflowTemplateSave(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	templateID, err := common.DbMain.WorkflowTemplateSave(request.ID, request.Name, request.Description)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	template, infoErr := common.DbMain.WorkflowTemplateInfo(cast.ToInt(templateID))
	if infoErr != nil {
		gsgin.GinResponseError(c, infoErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`template`: template,
	})
}

// WorkflowTemplateDelete 删除工作流程模板。
func WorkflowTemplateDelete(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.ID <= 0 {
		gsgin.GinResponseError(c, `模板id不能为空`, nil)
		return
	}
	if err := common.DbMain.WorkflowTemplateDelete(request.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// WorkflowTemplateStepSave 创建/更新模板步骤。
func WorkflowTemplateStepSave(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateStepSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	stepID, err := common.DbMain.WorkflowTemplateStepSave(
		request.ID, request.TemplateID, request.Name,
		request.StepKey, request.PromptContent, request.SortOrder,
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	step, infoErr := common.DbMain.WorkflowTemplateStepInfo(cast.ToInt(stepID))
	if infoErr != nil {
		gsgin.GinResponseError(c, infoErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`step`: step,
	})
}

// WorkflowTemplateStepDelete 删除模板步骤。
func WorkflowTemplateStepDelete(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateStepDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.ID <= 0 {
		gsgin.GinResponseError(c, `步骤id不能为空`, nil)
		return
	}
	if err := common.DbMain.WorkflowTemplateStepDelete(request.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// WorkflowTemplateStepSort 更新模板步骤排序。
func WorkflowTemplateStepSort(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateStepSortRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.TemplateID <= 0 {
		gsgin.GinResponseError(c, `模板id不能为空`, nil)
		return
	}
	if err := common.DbMain.WorkflowTemplateStepSort(request.TemplateID, request.StepIDs); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// ===================== 模板相关辅助接口 =====================

// WorkflowTemplateListBasic 获取简单的模板列表（仅 id+name，供下拉选择）。
func WorkflowTemplateListBasic(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "主库未初始化", "data": nil})
		return
	}
	templates, err := common.DbMain.WorkflowTemplateList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error(), "data": nil})
		return
	}
	type basicItem struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		IsDefault int    `json:"is_default"`
	}
	result := make([]basicItem, 0, len(templates))
	for _, t := range templates {
		result = append(result, basicItem{
			ID:        cast.ToInt(t[`id`]),
			Name:      cast.ToString(t[`name`]),
			IsDefault: cast.ToInt(t[`is_default`]),
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "", "data": result})
}
