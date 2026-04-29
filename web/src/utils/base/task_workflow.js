import base from '../base'

// TaskWorkflowCreateOrGet 查询或创建任务工作流。
function TaskWorkflowCreateOrGet(homeTaskId, callBack) {
  base.BasePost('/api/task/workflow/create_or_get', {
    home_task_id: homeTaskId,
  }, callBack)
}

// TaskWorkflowInfo 查询任务工作流详情。
function TaskWorkflowInfo(workflowId, callBack) {
  base.BasePost('/api/task/workflow/info', {
    workflow_id: workflowId,
  }, callBack)
}

export default {
  TaskWorkflowCreateOrGet,
  TaskWorkflowInfo,
}
