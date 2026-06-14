import base from '../base'

// WorkflowTemplateList 获取所有模板列表（含步骤）。
function WorkflowTemplateList(callBack) {
  base.BasePost('/api/workflow/template/list', {}, callBack)
}

// WorkflowTemplateSave 创建/更新模板。
function WorkflowTemplateSave(data, callBack) {
  base.BasePost('/api/workflow/template/save', data, callBack)
}

// WorkflowTemplateDelete 删除模板。
function WorkflowTemplateDelete(id, callBack) {
  base.BasePost('/api/workflow/template/delete', {
    id: id,
  }, callBack)
}

// WorkflowTemplateStepSave 创建/更新模板步骤。
function WorkflowTemplateStepSave(data, callBack) {
  base.BasePost('/api/workflow/template/step/save', data, callBack)
}

// WorkflowTemplateStepDelete 删除模板步骤。
function WorkflowTemplateStepDelete(id, callBack) {
  base.BasePost('/api/workflow/template/step/delete', {
    id: id,
  }, callBack)
}

// WorkflowTemplateStepSort 步骤排序。
function WorkflowTemplateStepSort(templateId, stepIds, callBack) {
  base.BasePost('/api/workflow/template/step/sort', {
    template_id: templateId,
    step_ids: stepIds,
  }, callBack)
}

// WorkflowTemplateListBasic 获取简单模板列表（id+name，供下拉选择）。
function WorkflowTemplateListBasic(callBack) {
  base.BasePost('/api/workflow/template/list-basic', {}, callBack)
}

export default {
  WorkflowTemplateList,
  WorkflowTemplateSave,
  WorkflowTemplateDelete,
  WorkflowTemplateStepSave,
  WorkflowTemplateStepDelete,
  WorkflowTemplateStepSort,
  WorkflowTemplateListBasic,
}
