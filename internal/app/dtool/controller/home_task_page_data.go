package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"

	"dev_tool/internal/pkg/p_define"
)

// HomeTaskPageDataLoad 接收前端请求后，通过 SSE 推送 HomeTask 页面全部附加数据。
// 主列表由 HomeTaskList 直接返回，此接口只负责附加数据的 SSE 推送。
func HomeTaskPageDataLoad(c *gin.Context) {
	request := _struct.HomeTaskPageDataLoadRequest{}
	_ = gsgin.GinPostBody(c, &request)

	if strings.TrimSpace(request.ClientID) == "" {
		gsgin.GinResponseError(c, "client_id不能为空", nil)
		return
	}

	sseClient := gsgin.SseGetByClientId(strings.TrimSpace(request.ClientID))
	if sseClient == nil {
		gsgin.GinResponseError(c, "SSE连接未建立，请刷新页面", nil)
		return
	}

	// 先返回成功，然后在 goroutine 中收集数据并通过 SSE 推送
	gsgin.GinResponseSuccess(c, "数据加载已触发", nil)

	go pushHomeTaskPageData(sseClient, request.TaskIDs)
}

// pushHomeTaskPageData 并发收集所有附加数据并通过 SSE 推送给客户端。
func pushHomeTaskPageData(sse *gsgin.Sse, taskIDs []int) {
	result := make(map[string]any)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 1. Git 仓库列表 + 分组
	wg.Add(1)
	go func() {
		defer wg.Done()
		gitGroupList, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
			`type`: define.GroupTypeGit,
		}).All()
		for k := range gitGroupList {
			gitGroupList[k][`id`] = cast.ToString(gitGroupList[k][`id`])
		}
		gitList, _ := common.DbMain.Client.QuickQuery(`tbl_git`, `*`, nil).All()
		for k := range gitList {
			gitList[k][`id`] = cast.ToString(gitList[k][`id`])
			gitList[k][`git_group_id`] = cast.ToString(gitList[k][`git_group_id`])
		}
		mu.Lock()
		result[`git_group_list`] = gitGroupList
		result[`git_list`] = gitList
		mu.Unlock()
	}()

	// 2. API 集合基础列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		list, _ := common.DbMain.Client.QueryBySql(`
select c.id,
       c.name,
       c.create_time,
       c.update_time,
       count(d.id) as child_count
from tbl_api_collection c
left join tbl_api_dir d on d.collection_id = c.id and d.archived = 0
group by c.id, c.name, c.create_time, c.update_time
order by c.id asc`).All()
		result2 := make([]map[string]any, 0, len(list))
		for _, row := range list {
			row[`id`] = cast.ToString(row[`id`])
			result2 = append(result2, row)
		}
		mu.Lock()
		result[`api_collection_list`] = result2
		mu.Unlock()
	}()

	// 3. Docker Compose 列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		list, _ := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
			`status`: 1,
		}).All()
		for k := range list {
			list[k][`id`] = cast.ToString(list[k][`id`])
		}
		mu.Lock()
		result[`docker_list`] = list
		mu.Unlock()
	}()

	// 4. MySQL 列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		list, _ := common.DbMain.Client.QuickQuery(`tbl_mysql`, `*`, nil).All()
		for k := range list {
			list[k][`id`] = cast.ToString(list[k][`id`])
		}
		mu.Lock()
		result[`mysql_list`] = list
		mu.Unlock()
	}()

	// 5. SmartLink 列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		list, _ := common.DbMain.Client.QueryBySql(`select * from tbl_smart_link where status = ? order by weight asc`, define.SmartLinkStatusNormal).All()
		for k := range list {
			list[k][`id`] = cast.ToString(list[k][`id`])
		}
		mu.Lock()
		result[`smart_link_list`] = list
		mu.Unlock()
	}()

	// 6. 记忆库文件夹列表
	wg.Add(1)
	go func() {
		defer wg.Done()
		if component.MemoryRuntime != nil && component.MemoryRuntime.EnsureConfigured() == nil {
			memoryDB := component.MemoryRuntime.DB()
			if memoryDB != nil {
				folderList, err := memoryDB.MemoryFragmentFolderList()
				if err == nil {
					mu.Lock()
					result[`memory_folder_list`] = folderList
					mu.Unlock()
				}
			}
		}
	}()

	// 7. 工作流节点状态（需要 taskIDs）
	if len(taskIDs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeStatusesMap, unreadCountMap, err := common.DbMain.TaskWorkflowBatchWorkflowSummaryByHomeTaskIDs(taskIDs)
			if err == nil {
				mu.Lock()
				result[`workflow_node_statuses_map`] = nodeStatusesMap
				result[`workflow_unread_count_map`] = unreadCountMap
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// 不在 goroutine 中做本地目录检查（这些操作本身很快，但涉及文件系统调用）
	// 从 task list 中提取的 local_dirs 和 branch 信息由前端传入时携带

	// 推送所有聚合数据
	sendSseData(sse, define.SseHomeTaskPageData, result)
}

// CheckAndPushLocalDirs 检查本地目录存在性并 SSE 推送（由前端 POST 触发）。
func CheckAndPushLocalDirs(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	clientID := strings.TrimSpace(cast.ToString(dataMap[`client_id`]))
	if clientID == "" {
		gsgin.GinResponseError(c, "client_id不能为空", nil)
		return
	}

	sseClient := gsgin.SseGetByClientId(clientID)
	if sseClient == nil {
		gsgin.GinResponseError(c, "SSE连接未建立", nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)

	pathsRaw, _ := dataMap[`paths`].([]any)
	if len(pathsRaw) == 0 {
		sendSseData(sseClient, define.SseHomeTaskPageData+`_dir_status`, map[string]any{})
		return
	}

	go func() {
		result := make(map[string]bool, len(pathsRaw))
		for _, p := range pathsRaw {
			dirPath := strings.TrimSpace(cast.ToString(p))
			if dirPath == `` {
				continue
			}
			if _, ok := result[dirPath]; ok {
				continue
			}
			info, statErr := os.Stat(dirPath)
			result[dirPath] = statErr == nil && info.IsDir()
		}
		sendSseData(sseClient, define.SseHomeTaskPageData+`_dir_status`, map[string]any{
			`dir_status_map`: result,
		})
	}()
}

// CheckAndPushBranchStatus 检查本地分支匹配状态并 SSE 推送（由前端 POST 触发）。
func CheckAndPushBranchStatus(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	clientID := strings.TrimSpace(cast.ToString(dataMap[`client_id`]))
	if clientID == "" {
		gsgin.GinResponseError(c, "client_id不能为空", nil)
		return
	}

	sseClient := gsgin.SseGetByClientId(clientID)
	if sseClient == nil {
		gsgin.GinResponseError(c, "SSE连接未建立", nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)

	itemsRaw, _ := dataMap[`items`].([]any)
	if len(itemsRaw) == 0 {
		sendSseData(sseClient, define.SseHomeTaskPageData+`_branch_status`, map[string]any{})
		return
	}

	go func() {
		result := make(map[string]map[string]any, len(itemsRaw))
		for _, raw := range itemsRaw {
			item, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			localDir := strings.TrimSpace(cast.ToString(item[`local_dir`]))
			branchName := strings.TrimSpace(cast.ToString(item[`branch_name`]))
			if localDir == `` || branchName == `` {
				continue
			}
			key := localDir + localBranchBatchCheckKeySep + branchName
			if _, exists := result[key]; exists {
				continue
			}
			result[key] = buildLocalBranchCheckResult(localDir, branchName)
		}
		sendSseData(sseClient, define.SseHomeTaskPageData+`_branch_status`, map[string]any{
			`branch_status_map`: result,
		})
	}()
}

// sendSseData 向指定 SSE 客户端推送一个带分发 ID 的数据包。
func sendSseData(sse *gsgin.Sse, distributeID string, data map[string]any) {
	if sse == nil {
		return
	}
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: distributeID,
		Data:            data,
		Type:            p_define.SseContentTypeMsg,
	})
	if err := sse.SendToChan(msg); err != nil {
		gstool.FmtPrintlnLogTime(`home_task_page_data 推送错误 %s`, err.Error())
	}
}
