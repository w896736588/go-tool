package common

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// InfoCrawlTaskList 查询任务列表。
func (h *CSqlite) InfoCrawlTaskList() ([]map[string]any, error) {
	list, err := h.Client.QueryBySql(`
select id,name,prompt,ai_model_id,status,create_time,update_time
from tbl_info_crawl_task
where status = ?
order by update_time desc, id desc`, define.InfoCrawlTaskStatusNormal).All()
	if err != nil {
		return nil, err
	}
	h.infoCrawlFillTimeDesc(list)
	return list, nil
}

// InfoCrawlTaskInfo 查询任务详情。
func (h *CSqlite) InfoCrawlTaskInfo(id int) (map[string]any, error) {
	task, err := h.InfoCrawlTaskRow(id)
	if err != nil {
		return nil, err
	}
	pageList, err := h.InfoCrawlTaskPageList(id)
	if err != nil {
		return nil, err
	}
	runList, err := h.InfoCrawlRunList(id, 20)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`task`:      task,
		`page_list`: pageList,
		`run_list`:  runList,
	}, nil
}

// InfoCrawlTaskRow 查询单个任务。
func (h *CSqlite) InfoCrawlTaskRow(id int) (map[string]any, error) {
	task, err := h.Client.QuickQuery(`tbl_info_crawl_task`, `*`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(task) == 0 {
		return nil, errors.New(`任务不存在`)
	}
	h.infoCrawlFillRowTimeDesc(task)
	return task, nil
}

// InfoCrawlTaskSave 保存任务。
func (h *CSqlite) InfoCrawlTaskSave(id int, name, prompt string, aiModelID int) (map[string]any, error) {
	now := time.Now().Unix()
	name = strings.TrimSpace(name)
	prompt = strings.TrimSpace(prompt)
	if name == `` {
		return nil, errors.New(`任务名称不能为空`)
	}
	if prompt == `` {
		return nil, errors.New(`任务提示词不能为空`)
	}
	if aiModelID <= 0 {
		return nil, errors.New(`请选择AI模型`)
	}
	if _, err := h.InfoCrawlAiModelInfo(aiModelID); err != nil {
		return nil, err
	}
	if id <= 0 {
		newID, err := h.Client.QuickCreate(`tbl_info_crawl_task`, map[string]any{
			`name`:        name,
			`prompt`:      prompt,
			`ai_model_id`: aiModelID,
			`status`:      define.InfoCrawlTaskStatusNormal,
			`create_time`: now,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		id = cast.ToInt(newID)
	} else {
		_, err := h.Client.QuickUpdate(`tbl_info_crawl_task`, map[string]any{
			`id`:     id,
			`status`: define.InfoCrawlTaskStatusNormal,
		}, map[string]any{
			`name`:        name,
			`prompt`:      prompt,
			`ai_model_id`: aiModelID,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return nil, err
		}
	}
	return h.InfoCrawlTaskRow(id)
}

// InfoCrawlTaskDelete 软删除任务。
func (h *CSqlite) InfoCrawlTaskDelete(id int) error {
	if id <= 0 {
		return errors.New(`任务id不能为空`)
	}
	now := time.Now().Unix()
	if _, err := h.Client.QuickUpdate(`tbl_info_crawl_task`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}, map[string]any{
		`status`:      define.InfoCrawlTaskStatusDelete,
		`update_time`: now,
	}).Exec(); err != nil {
		return err
	}
	_, _ = h.Client.QuickUpdate(`tbl_info_crawl_task_page`, map[string]any{
		`task_id`: id,
		`status`:  define.InfoCrawlTaskStatusNormal,
	}, map[string]any{
		`status`:      define.InfoCrawlTaskStatusDelete,
		`update_time`: now,
	}).Exec()
	return nil
}

// InfoCrawlTaskPageList 查询任务网页列表。
func (h *CSqlite) InfoCrawlTaskPageList(taskID int) ([]map[string]any, error) {
	list, err := h.Client.QueryBySql(`
select id,task_id,name,url,note,login_check_selector,login_status,user_data_dir,sort,status,create_time,update_time
from tbl_info_crawl_task_page
where task_id = ? and status = ?
order by sort asc, id asc`, taskID, define.InfoCrawlTaskStatusNormal).All()
	if err != nil {
		return nil, err
	}
	h.infoCrawlFillPageStatusDesc(list)
	h.infoCrawlFillTimeDesc(list)
	return list, nil
}

// InfoCrawlTaskPageRow 查询单个网页配置。
func (h *CSqlite) InfoCrawlTaskPageRow(id int) (map[string]any, error) {
	info, err := h.Client.QuickQuery(`tbl_info_crawl_task_page`, `*`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`网页配置不存在`)
	}
	h.infoCrawlFillPageStatusDesc([]map[string]any{info})
	h.infoCrawlFillRowTimeDesc(info)
	return info, nil
}

// InfoCrawlTaskPageSave 保存网页配置。
func (h *CSqlite) InfoCrawlTaskPageSave(id, taskID int, name, pageURL, note, loginCheckSelector string, sortValue int) (map[string]any, error) {
	now := time.Now().Unix()
	name = strings.TrimSpace(name)
	pageURL = strings.TrimSpace(pageURL)
	note = strings.TrimSpace(note)
	loginCheckSelector = strings.TrimSpace(loginCheckSelector)
	if taskID <= 0 {
		return nil, errors.New(`任务id不能为空`)
	}
	if _, err := h.InfoCrawlTaskRow(taskID); err != nil {
		return nil, err
	}
	if name == `` {
		return nil, errors.New(`网页名称不能为空`)
	}
	if pageURL == `` {
		return nil, errors.New(`网页URL不能为空`)
	}
	if !strings.HasPrefix(strings.ToLower(pageURL), `http://`) && !strings.HasPrefix(strings.ToLower(pageURL), `https://`) {
		return nil, errors.New(`网页URL必须以 http:// 或 https:// 开头`)
	}
	if id <= 0 {
		userDataDir := h.InfoCrawlBuildUserDataDir(taskID, name)
		newID, err := h.Client.QuickCreate(`tbl_info_crawl_task_page`, map[string]any{
			`task_id`:              taskID,
			`name`:                 name,
			`url`:                  pageURL,
			`note`:                 note,
			`login_check_selector`: loginCheckSelector,
			`login_status`:         define.InfoCrawlPageLoginStatusNo,
			`user_data_dir`:        userDataDir,
			`sort`:                 sortValue,
			`status`:               define.InfoCrawlTaskStatusNormal,
			`create_time`:          now,
			`update_time`:          now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		id = cast.ToInt(newID)
	} else {
		oldInfo, err := h.InfoCrawlTaskPageRow(id)
		if err != nil {
			return nil, err
		}
		userDataDir := cast.ToString(oldInfo[`user_data_dir`])
		if userDataDir == `` {
			userDataDir = h.InfoCrawlBuildUserDataDir(taskID, name)
		}
		_, err = h.Client.QuickUpdate(`tbl_info_crawl_task_page`, map[string]any{
			`id`:     id,
			`status`: define.InfoCrawlTaskStatusNormal,
		}, map[string]any{
			`task_id`:              taskID,
			`name`:                 name,
			`url`:                  pageURL,
			`note`:                 note,
			`login_check_selector`: loginCheckSelector,
			`user_data_dir`:        userDataDir,
			`sort`:                 sortValue,
			`update_time`:          now,
		}).Exec()
		if err != nil {
			return nil, err
		}
	}
	return h.InfoCrawlTaskPageRow(id)
}

// InfoCrawlTaskPageDelete 软删除网页配置。
func (h *CSqlite) InfoCrawlTaskPageDelete(id int) error {
	if id <= 0 {
		return errors.New(`网页id不能为空`)
	}
	_, err := h.Client.QuickUpdate(`tbl_info_crawl_task_page`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}, map[string]any{
		`status`:      define.InfoCrawlTaskStatusDelete,
		`update_time`: time.Now().Unix(),
	}).Exec()
	return err
}

// InfoCrawlTaskPageSetLoginStatus 更新网页登录状态。
func (h *CSqlite) InfoCrawlTaskPageSetLoginStatus(id, loginStatus int) error {
	_, err := h.Client.QuickUpdate(`tbl_info_crawl_task_page`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}, map[string]any{
		`login_status`: loginStatus,
		`update_time`:  time.Now().Unix(),
	}).Exec()
	return err
}

// InfoCrawlRunCreate 创建执行记录。
func (h *CSqlite) InfoCrawlRunCreate(taskID int, taskInfo map[string]any, aiModelInfo map[string]any) (int, error) {
	now := time.Now().Unix()
	newID, err := h.Client.QuickCreate(`tbl_info_crawl_run`, map[string]any{
		`task_id`:            taskID,
		`status`:             define.InfoCrawlRunStatusRunning,
		`run_message`:        ``,
		`prompt_snapshot`:    cast.ToString(taskInfo[`prompt`]),
		`ai_model_snapshot`:  gstool.JsonEncode(aiModelInfo),
		`planner_content`:    ``,
		`summary_content`:    ``,
		`page_total`:         0,
		`page_success_total`: 0,
		`page_failed_total`:  0,
		`create_time`:        now,
		`update_time`:        now,
	}).Exec()
	return cast.ToInt(newID), err
}

// InfoCrawlRunUpdate 更新执行记录。
func (h *CSqlite) InfoCrawlRunUpdate(id int, updateData map[string]any) error {
	updateData[`update_time`] = time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_info_crawl_run`, map[string]any{
		`id`: id,
	}, updateData).Exec()
	return err
}

// InfoCrawlRunPageCreate 创建网页执行明细。
func (h *CSqlite) InfoCrawlRunPageCreate(data map[string]any) (int, error) {
	data[`create_time`] = time.Now().Unix()
	data[`update_time`] = time.Now().Unix()
	newID, err := h.Client.QuickCreate(`tbl_info_crawl_run_page`, data).Exec()
	return cast.ToInt(newID), err
}

// InfoCrawlRunList 查询执行历史。
func (h *CSqlite) InfoCrawlRunList(taskID, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	list, err := h.Client.QueryBySql(fmt.Sprintf(`
select id,task_id,status,run_message,prompt_snapshot,ai_model_snapshot,planner_content,summary_content,page_total,page_success_total,page_failed_total,create_time,update_time
from tbl_info_crawl_run
where task_id = ?
order by id desc
limit %d`, limit), taskID).All()
	if err != nil {
		return nil, err
	}
	h.infoCrawlFillTimeDesc(list)
	return list, nil
}

// InfoCrawlRunInfo 查询执行详情。
func (h *CSqlite) InfoCrawlRunInfo(id int) (map[string]any, error) {
	runInfo, err := h.Client.QuickQuery(`tbl_info_crawl_run`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(runInfo) == 0 {
		return nil, errors.New(`执行记录不存在`)
	}
	runPageList, err := h.Client.QueryBySql(`
select *
from tbl_info_crawl_run_page
where run_id = ?
order by id asc`, id).All()
	if err != nil {
		return nil, err
	}
	h.infoCrawlFillTimeDesc([]map[string]any{runInfo})
	h.infoCrawlFillTimeDesc(runPageList)
	return map[string]any{
		`run_info`:      runInfo,
		`run_page_list`: runPageList,
	}, nil
}

// InfoCrawlAiModelInfo 查询 AI 模型配置。
func (h *CSqlite) InfoCrawlAiModelInfo(id int) (map[string]any, error) {
	info, err := h.Client.QueryBySql(`
select m.*,p.name as provider_name,p.provider_type,p.base_url,p.api_key
from tbl_ai_model m
left join tbl_ai_provider p on p.id = m.provider_id
where m.id = ? and m.status = 1 and p.status = 1`, id).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`AI模型不存在或已停用`)
	}
	return info, nil
}

// InfoCrawlBuildUserDataDir 生成网页登录目录。
func (h *CSqlite) InfoCrawlBuildUserDataDir(taskID int, pageName string) string {
	safeName := strings.TrimSpace(pageName)
	if safeName == `` {
		safeName = `page`
	}
	safeName = p_common.Replace(safeName, map[string]string{
		`\\`: `_`,
		`/`:  `_`,
		`:`:  `_`,
		`*`:  `_`,
		`?`:  `_`,
		`"`:  `_`,
		`<`:  `_`,
		`>`:  `_`,
		`|`:  `_`,
		` `:  `_`,
	})
	return filepath.Join(h.Env.WebkitDataPath, `info_crawl`, fmt.Sprintf(`task_%d_%s`, taskID, safeName))
}

// infoCrawlFillTimeDesc 填充时间描述字段。
func (h *CSqlite) infoCrawlFillTimeDesc(list []map[string]any) {
	for _, item := range list {
		h.infoCrawlFillRowTimeDesc(item)
	}
}

// infoCrawlFillRowTimeDesc 填充单行时间描述字段。
func (h *CSqlite) infoCrawlFillRowTimeDesc(row map[string]any) {
	row[`create_time_desc`] = h.infoCrawlFormatTime(cast.ToInt64(row[`create_time`]))
	row[`update_time_desc`] = h.infoCrawlFormatTime(cast.ToInt64(row[`update_time`]))
}

// infoCrawlFillPageStatusDesc 填充网页登录状态描述。
func (h *CSqlite) infoCrawlFillPageStatusDesc(list []map[string]any) {
	for _, item := range list {
		switch cast.ToInt(item[`login_status`]) {
		case define.InfoCrawlPageLoginStatusOk:
			item[`login_status_desc`] = `已登录`
		case define.InfoCrawlPageLoginStatusExpired:
			item[`login_status_desc`] = `登录失效`
		default:
			item[`login_status_desc`] = `未登录`
		}
	}
}

// infoCrawlFormatTime 格式化时间。
func (h *CSqlite) infoCrawlFormatTime(unixTime int64) string {
	if unixTime <= 0 {
		return ``
	}
	return gstool.TimeUnixToString(time.Unix(unixTime, 0), `Y-m-d H:i:s`)
}

// InfoCrawlValidatePlanner 校验 AI 规划结果。
func (h *CSqlite) InfoCrawlValidatePlanner(taskID int, pageList []map[string]any, planner map[int]map[string]any) error {
	pageIDMap := make(map[int]bool)
	for _, page := range pageList {
		pageIDMap[cast.ToInt(page[`id`])] = true
	}
	for pageID, pagePlanner := range planner {
		if !pageIDMap[pageID] {
			return errors.New(`规划结果包含非法网页ID`)
		}
		actionList, ok := pagePlanner[`actions`].([]map[string]any)
		if !ok {
			return errors.New(`规划动作格式错误`)
		}
		if len(actionList) > define.InfoCrawlPlannerActionMaxCount {
			return errors.New(`规划动作数量超过上限`)
		}
		for _, action := range actionList {
			actionType := cast.ToString(action[`type`])
			if !h.infoCrawlAllowAction(actionType) {
				return errors.New(`规划动作包含未授权类型`)
			}
			if len(cast.ToString(action[`locator`])) > 500 {
				return errors.New(`规划 locator 过长`)
			}
			if actionType == define.InfoCrawlPlannerActionWait && cast.ToInt(action[`value`]) > define.InfoCrawlPlannerWaitMaxMillis {
				return errors.New(`等待时间超过上限`)
			}
		}
	}
	return nil
}

// infoCrawlAllowAction 判断动作是否允许。
func (h *CSqlite) infoCrawlAllowAction(actionType string) bool {
	allowList := []string{
		define.InfoCrawlPlannerActionWait,
		define.InfoCrawlPlannerActionClick,
		define.InfoCrawlPlannerActionExistWait,
		define.InfoCrawlPlannerActionNoExistWait,
		define.InfoCrawlPlannerActionTextContent,
		define.InfoCrawlPlannerActionBoolResult,
	}
	return gstool.ArrayExistValue(&allowList, actionType)
}

// InfoCrawlNormalizePlannerMap 规范化规划结果映射。
func (h *CSqlite) InfoCrawlNormalizePlannerMap(rawPages []map[string]any) map[int]map[string]any {
	result := make(map[int]map[string]any)
	for _, rawPage := range rawPages {
		pageID := cast.ToInt(rawPage[`task_page_id`])
		actionList := make([]map[string]any, 0)
		switch actions := rawPage[`actions`].(type) {
		case []map[string]any:
			actionList = actions
		case []any:
			for _, item := range actions {
				if actionMap, ok := item.(map[string]any); ok {
					actionList = append(actionList, actionMap)
				}
			}
		}
		result[pageID] = map[string]any{
			`task_page_id`: pageID,
			`goal`:         cast.ToString(rawPage[`goal`]),
			`actions`:      actionList,
		}
	}
	return result
}

// InfoCrawlSortPages 按 sort 和 id 排序页面。
func (h *CSqlite) InfoCrawlSortPages(pageList []map[string]any) {
	sort.SliceStable(pageList, func(i, j int) bool {
		if cast.ToInt(pageList[i][`sort`]) != cast.ToInt(pageList[j][`sort`]) {
			return cast.ToInt(pageList[i][`sort`]) < cast.ToInt(pageList[j][`sort`])
		}
		return cast.ToInt(pageList[i][`id`]) < cast.ToInt(pageList[j][`id`])
	})
}
