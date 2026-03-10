import base from '../base'

// InfoCrawlTaskList 查询任务列表。
function InfoCrawlTaskList(callBack) {
  base.BasePost('/api/InfoCrawlTaskList', {}, callBack)
}

// InfoCrawlTaskInfo 查询任务详情。
function InfoCrawlTaskInfo(id, callBack) {
  base.BasePost('/api/InfoCrawlTaskInfo', { id: id }, callBack)
}

// InfoCrawlTaskSave 保存任务。
function InfoCrawlTaskSave(data, callBack) {
  base.BasePost('/api/InfoCrawlTaskSave', data, callBack)
}

// InfoCrawlTaskDelete 删除任务。
function InfoCrawlTaskDelete(id, callBack) {
  base.BasePost('/api/InfoCrawlTaskDelete', { id: id }, callBack)
}

// InfoCrawlTaskPageSave 保存网页配置。
function InfoCrawlTaskPageSave(data, callBack) {
  base.BasePost('/api/InfoCrawlTaskPageSave', data, callBack)
}

// InfoCrawlTaskPageDelete 删除网页配置。
function InfoCrawlTaskPageDelete(id, callBack) {
  base.BasePost('/api/InfoCrawlTaskPageDelete', { id: id }, callBack)
}

// InfoCrawlTaskPageOpenLogin 打开网页登录页。
function InfoCrawlTaskPageOpenLogin(taskPageId, callBack) {
  base.BasePost('/api/InfoCrawlTaskPageOpenLogin', { task_page_id: taskPageId }, callBack)
}

// InfoCrawlTaskPageCheckLogin 检查网页登录状态。
function InfoCrawlTaskPageCheckLogin(taskPageId, callBack) {
  base.BasePost('/api/InfoCrawlTaskPageCheckLogin', { task_page_id: taskPageId }, callBack)
}

// InfoCrawlTaskRun 执行任务。
function InfoCrawlTaskRun(data, callBack) {
  base.BasePost('/api/InfoCrawlTaskRun', data, callBack)
}

// InfoCrawlRunList 查询执行历史。
function InfoCrawlRunList(taskId, limit, callBack) {
  base.BasePost('/api/InfoCrawlRunList', { task_id: taskId, limit: limit }, callBack)
}

// InfoCrawlRunInfo 查询执行详情。
function InfoCrawlRunInfo(id, callBack) {
  base.BasePost('/api/InfoCrawlRunInfo', { id: id }, callBack)
}

export default {
  InfoCrawlTaskList,
  InfoCrawlTaskInfo,
  InfoCrawlTaskSave,
  InfoCrawlTaskDelete,
  InfoCrawlTaskPageSave,
  InfoCrawlTaskPageDelete,
  InfoCrawlTaskPageOpenLogin,
  InfoCrawlTaskPageCheckLogin,
  InfoCrawlTaskRun,
  InfoCrawlRunList,
  InfoCrawlRunInfo,
}
