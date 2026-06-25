package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
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

// WorkflowTemplateSetDefault 设置默认模板。
func WorkflowTemplateSetDefault(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateSetDefaultRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.ID <= 0 {
		gsgin.GinResponseError(c, `模板id不能为空`, nil)
		return
	}
	if err := common.DbMain.WorkflowTemplateSetDefault(request.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
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

	// 校验文档占位符在同一模板内是否重复
	if err := validateStepDocumentPlaceholders(request.TemplateID, request.ID, request.StepDocuments); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	stepID, err := common.DbMain.WorkflowTemplateStepSave(
		request.ID, request.TemplateID, request.Name,
		request.StepKey, request.PromptContent, request.StepDocuments, request.Remark, request.SortOrder,
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

// validateStepDocumentPlaceholders 校验文档占位符在同一模板内是否重复（跨步骤检测）。
func validateStepDocumentPlaceholders(templateID, excludeStepID int, stepDocumentsJSON string) error {
	if templateID <= 0 {
		return nil
	}
	// 解析当前步骤的文档占位符
	currentDocs := common.WorkflowTemplateStepDocumentsParse(stepDocumentsJSON)
	currentPlaceholders := make(map[string]string) // placeholder -> docName
	for _, doc := range currentDocs {
		ph := strings.TrimSpace(doc.Placeholder)
		if ph == `` {
			continue
		}
		if prev, exists := currentPlaceholders[ph]; exists {
			return fmt.Errorf(`文档占位符 %s 重复（%s 与 %s）`, ph, prev, doc.Name)
		}
		currentPlaceholders[ph] = doc.Name
	}
	if len(currentPlaceholders) == 0 {
		return nil
	}

	// 读取模板的所有步骤
	steps, err := common.DbMain.WorkflowTemplateStepsByTemplateID(templateID)
	if err != nil {
		return nil
	}
	for _, step := range steps {
		stepID := cast.ToInt(step[`id`])
		if stepID == excludeStepID {
			continue
		}
		otherDocs := common.WorkflowTemplateStepDocumentsParse(cast.ToString(step[`step_documents`]))
		for _, doc := range otherDocs {
			ph := strings.TrimSpace(doc.Placeholder)
			if ph == `` {
				continue
			}
			if _, exists := currentPlaceholders[ph]; exists {
				return fmt.Errorf(`文档占位符 %s 与步骤"%s"中的文档"%s"重复`, ph, cast.ToString(step[`name`]), doc.Name)
			}
		}
	}
	return nil
}

// WorkflowSkillList 动态获取 skills 目录下的所有 skill 名称列表。
func WorkflowSkillList(c *gin.Context) {
	if component.EnvClient == nil {
		gsgin.GinResponseError(c, `环境未初始化`, nil)
		return
	}
	skillsDir := filepath.Join(component.EnvClient.RootPath, `skills`)
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		gsgin.GinResponseError(c, `读取skills目录失败`, nil)
		return
	}
	type skillItem struct {
		Name string `json:"name"`
	}
	result := make([]skillItem, 0)
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), `.`) {
			result = append(result, skillItem{Name: entry.Name()})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: result,
	})
}

// WorkflowTemplateImport 导入工作流程模板（含步骤）。
func WorkflowTemplateImport(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.WorkflowTemplateImportRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.Name == `` {
		gsgin.GinResponseError(c, `模板名称不能为空`, nil)
		return
	}
	// 转换步骤数据
	steps := make([]common.WorkflowTemplateImportStepData, 0, len(request.Steps))
	for _, s := range request.Steps {
		steps = append(steps, common.WorkflowTemplateImportStepData{
			Name:          s.Name,
			StepKey:       s.StepKey,
			PromptContent: s.PromptContent,
			StepDocuments: s.StepDocuments,
			Remark:        s.Remark,
			IsFixed:       s.IsFixed,
			SortOrder:     s.SortOrder,
		})
	}
	templateID, err := common.DbMain.WorkflowTemplateImport(request.Name, request.Description, steps)
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

// WorkflowTemplateListBasic 获取简单的模板列表（仅 id+name，供下拉选择）。
func WorkflowTemplateListBasic(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, "主库未初始化", nil)
		return
	}
	templates, err := common.DbMain.WorkflowTemplateList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
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
	gsgin.GinResponseSuccess(c, "", map[string]any{
		"list": result,
	})
}
