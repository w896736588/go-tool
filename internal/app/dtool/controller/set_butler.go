package controller

import (
	"context"
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// butlerRunningStatuses 管家任务"运行中"的状态集合
var butlerRunningStatuses = []string{
	define.ButlerTaskStatusPending,
	define.ButlerTaskStatusExecuting,
	define.ButlerTaskStatusVerifying,
}

// checkButlerRunningTask 检查是否有正在运行的管家任务，有则返回 false 并响应错误信息。
func checkButlerRunningTask(c *gin.Context) bool {
	statusList := make([]string, len(butlerRunningStatuses))
	for i, s := range butlerRunningStatuses {
		statusList[i] = fmt.Sprintf(`'%s'`, s)
	}
	row, _ := common.DbMain.Client.QueryBySql(
		fmt.Sprintf(`SELECT COUNT(*) as cnt FROM tbl_butler_task WHERE status IN (%s)`, strings.Join(statusList, `,`)),
	).One()
	if cast.ToInt(row[`cnt`]) > 0 {
		gsgin.GinResponseError(c, `有管家任务正在运行，请等待任务完成后再编辑配置`, nil)
		return false
	}
	return true
}

// maskSecret 脱敏密钥字符串：保留前 6 和后 4 位，中间用星号替换。
func maskSecret(secret string) string {
	if len(secret) <= 10 {
		return `******`
	}
	return secret[:6] + `******` + secret[len(secret)-4:]
}

// SetButlerBotConfigList 查询管家机器人配置列表
func SetButlerBotConfigList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_butler_bot_config`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	// 脱敏处理：app_secret 字段用星号替换
	for i, item := range all {
		if item[`app_secret`] != `` {
			all[i][`app_secret`] = maskSecret(cast.ToString(item[`app_secret`]))
		}
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

// SetButlerBotConfigAdd 新增或更新管家机器人配置（id==0 新增，id>0 更新）。
// 当 status 变为 1（启用）时自动连接该机器人；当 status 变为非 1（禁用）时自动断开。
func SetButlerBotConfigAdd(c *gin.Context) {
	if !checkButlerRunningTask(c) {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`platform`, `name`, `app_key`, `app_secret`, `robot_code`, `status`})
	var err error
	now := time.Now().Unix()
	botId := cast.ToInt(dataMap[`id`])
	newStatus := cast.ToInt(updateData[`status`])

	if botId == 0 {
		updateData[`created_at`] = now
		updateData[`updated_at`] = now
		_, err = common.DbMain.Client.QuickCreate(`tbl_butler_bot_config`, updateData).Exec()
		// 新增后获取自增 id
		if err == nil && newStatus == 1 {
			lastInsertIdRow, _ := common.DbMain.Client.QueryBySql(`SELECT last_insert_rowid() as id`).One()
			botId = cast.ToInt(lastInsertIdRow[`id`])
		}
	} else {
		// 查询更新前的完整记录（status、app_secret 等），用于保留前端未传的敏感字段
		oldRow, oldErr := common.DbMain.Client.QuickQuery(`tbl_butler_bot_config`, `status, app_secret`, map[string]any{`id`: botId}).One()
		oldStatus := 0
		if oldErr == nil && len(oldRow) > 0 {
			oldStatus = cast.ToInt(oldRow[`status`])
		}
		// 编辑时 app_secret 已脱敏不回填，前端可能传空字符串，此时保留旧值避免被清空
		if cast.ToString(updateData[`app_secret`]) == `` && oldErr == nil && len(oldRow) > 0 {
			updateData[`app_secret`] = cast.ToString(oldRow[`app_secret`])
		}
		updateData[`updated_at`] = now
		_, err = common.DbMain.Client.QuickUpdate(`tbl_butler_bot_config`,
			map[string]any{`id`: dataMap[`id`]}, updateData).Exec()
		// 启用/禁用状态变更时自动连接/断开
		if err == nil {
			runtime := business.GetButlerRuntime()
			if runtime != nil {
				if newStatus == 1 && oldStatus != 1 {
					// 启用 → 自动连接
					go handleBotEnable(botId)
				} else if newStatus != 1 && oldStatus == 1 {
					// 禁用 → 自动断开
					runtime.DisconnectBot(botId)
				}
			}
		}
	}
	if err != nil {
		gsgin.GinResponseError(c, `保存失败: `+err.Error(), nil)
		return
	}
	// 新增且启用 → 自动连接
	if botId > 0 && newStatus == 1 {
		runtime := business.GetButlerRuntime()
		if runtime != nil {
			go handleBotEnable(botId)
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// handleBotEnable 机器人启用后的连接处理：从数据库读取完整配置并触发连接。
func handleBotEnable(botId int) {
	runtime := business.GetButlerRuntime()
	if runtime == nil {
		return
	}
	row, err := common.DbMain.Client.QuickQuery(`tbl_butler_bot_config`, `*`, map[string]any{`id`: botId}).One()
	if err != nil || len(row) == 0 {
		gstool.FmtPrintlnLogTime(`[butler] 启用机器人连接失败：配置不存在 id=%d`, botId)
		return
	}
	botConfig := &define.BotConfigItem{
		Id:        cast.ToInt(row[`id`]),
		Platform:  cast.ToString(row[`platform`]),
		Name:      cast.ToString(row[`name`]),
		AppKey:    cast.ToString(row[`app_key`]),
		AppSecret: cast.ToString(row[`app_secret`]),
		RobotCode: cast.ToString(row[`robot_code`]),
		Status:    cast.ToInt(row[`status`]),
	}
	if err := runtime.ConnectBot(botConfig); err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 启用机器人连接失败 %s`, err.Error())
	}
}

// SetButlerBotConfigDelete 删除管家机器人配置，同时断开该机器人的连接。
func SetButlerBotConfigDelete(c *gin.Context) {
	if !checkButlerRunningTask(c) {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	botId := cast.ToInt(dataMap[`id`])
	if botId == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	// 删除前先断开连接
	runtime := business.GetButlerRuntime()
	if runtime != nil {
		runtime.DisconnectBot(botId)
	}
	_, _ = common.DbMain.Client.QuickDelete(`tbl_butler_bot_config`, map[string]any{
		`id`: botId,
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetButlerBotConfigTest 测试机器人配置：通过钉钉 Stream SDK 验证 AppKey/AppSecret 能否成功建立连接。
// Stream 模式机器人本身不需要 RobotCode，此处只验证 SDK 层面的连通性。
func SetButlerBotConfigTest(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	botId := cast.ToInt(dataMap[`id`])
	if botId == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	// 从主库读取原始配置（未脱敏）
	row, rowErr := common.DbMain.Client.QuickQuery(`tbl_butler_bot_config`, `*`, map[string]any{`id`: botId}).One()
	if rowErr != nil || len(row) == 0 {
		gsgin.GinResponseError(c, `机器人配置不存在`, nil)
		return
	}
	appKey := cast.ToString(row[`app_key`])
	appSecret := cast.ToString(row[`app_secret`])
	name := cast.ToString(row[`name`])
	if appKey == `` || appSecret == `` {
		gsgin.GinResponseError(c, `AppKey/AppSecret 未配置，无法测试`, nil)
		return
	}
	// 创建临时 StreamClient，通过 CheckConfigValid 验证凭证格式
	cli := client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(appKey, appSecret)),
	)
	if err := cli.CheckConfigValid(); err != nil {
		gsgin.GinResponseError(c, `配置校验失败: `+err.Error(), nil)
		return
	}
	// 调用钉钉 SDK 获取 WebSocket 连接端点，验证 AppKey/AppSecret 是否真实可用
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	endpoint, err := cli.GetConnectionEndpoint(ctx)
	if err != nil {
		gsgin.GinResponseError(c, `Stream 连接失败: `+err.Error(), nil)
		return
	}
	// Stream 连接验证成功，无需额外的手动 Open API 调用
	testResult := `Stream 连接端点获取成功`
	if endpoint != nil && endpoint.Endpoint != `` {
		testResult = `Stream 连接端点获取成功 (` + endpoint.Endpoint + `)`
	}

	gsgin.GinResponseSuccess(c, `机器人 "`+name+`" 连接测试成功`, map[string]any{
		`endpoint`:    endpoint,
		`test_result`: testResult,
	})
}

// SetButlerMessageList 查询管家机器人消息日志（分页），按 bot_config_id 过滤。
// 同时兼容 bot_config_id=0 的旧数据（消息属于该机器人但旧版本未记录 bot_config_id）。
func SetButlerMessageList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	botConfigId := cast.ToInt(dataMap[`bot_config_id`])
	if botConfigId == 0 {
		gsgin.GinResponseError(c, `bot_config_id不能为空`, nil)
		return
	}
	page := cast.ToInt(dataMap[`page`])
	if page <= 0 {
		page = 1
	}
	pageSize := cast.ToInt(dataMap[`page_size`])
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	// 查询总数：同时匹配 bot_config_id 为指定值或 0（兼容旧数据）
	countRow, _ := common.DbMain.Client.QueryBySql(
		`SELECT COUNT(*) as total FROM tbl_butler_message WHERE bot_config_id = ? OR bot_config_id = 0`, botConfigId,
	).One()
	total := cast.ToInt(countRow[`total`])
	// 查询分页数据（按 id 倒序，最新在前），同时匹配 bot_config_id 为指定值或 0
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_butler_message WHERE bot_config_id = ? OR bot_config_id = 0 ORDER BY id DESC LIMIT ? OFFSET ?`,
		botConfigId, pageSize, offset,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, gin.H{
		`list`:      rows,
		`total`:     total,
		`page`:      page,
		`page_size`: pageSize,
	})
}

// SetButlerRoleList 查询管家角色列表
func SetButlerRoleList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_butler_role`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

// SetButlerRoleAdd 新增或更新管家角色（id==0 新增，id>0 更新）
func SetButlerRoleAdd(c *gin.Context) {
	if !checkButlerRunningTask(c) {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `persona`, `tone`, `system_prompt`, `init_greeting`, `status`})
	var err error
	now := time.Now().Unix()
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`created_at`] = now
		updateData[`updated_at`] = now
		_, err = common.DbMain.Client.QuickCreate(`tbl_butler_role`, updateData).Exec()
	} else {
		updateData[`updated_at`] = now
		_, err = common.DbMain.Client.QuickUpdate(`tbl_butler_role`,
			map[string]any{`id`: dataMap[`id`]}, updateData).Exec()
	}
	if err != nil {
		gsgin.GinResponseError(c, `保存失败: `+err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetButlerRoleDelete 删除管家角色
func SetButlerRoleDelete(c *gin.Context) {
	if !checkButlerRunningTask(c) {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	_, _ = common.DbMain.Client.QuickDelete(`tbl_butler_role`, map[string]any{
		`id`: cast.ToInt(dataMap[`id`]),
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetButlerConfigList 查询管家运行参数列表
func SetButlerConfigList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_butler_config`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	// 关联角色和机器人配置名称（主库），模型和 agent cli 名称（主库）
	for i, item := range all {
		roleId := cast.ToInt(item[`role_id`])
		botConfigId := cast.ToInt(item[`bot_config_id`])
		modelId := cast.ToInt(item[`model_id`])
		fcModelId := cast.ToInt(item[`fc_model_id`])
		agentCliId := cast.ToInt(item[`agent_cli_id`])
		if roleId > 0 {
			roleOne, roleErr := common.DbMain.Client.QuickQuery(`tbl_butler_role`, `name`, map[string]any{`id`: roleId}).One()
			if roleErr == nil && roleOne[`name`] != `` {
				all[i][`role_name`] = roleOne[`name`]
			}
		}
		if botConfigId > 0 {
			botOne, botErr := common.DbMain.Client.QuickQuery(`tbl_butler_bot_config`, `name`, map[string]any{`id`: botConfigId}).One()
			if botErr == nil && botOne[`name`] != `` {
				all[i][`bot_config_name`] = botOne[`name`]
			}
		}
		if modelId > 0 {
			modelOne, modelErr := common.DbMain.Client.QuickQuery(`tbl_ai_model`, `name`, map[string]any{`id`: modelId}).One()
			if modelErr == nil && modelOne[`name`] != `` {
				all[i][`model_name`] = modelOne[`name`]
			}
		}
		if fcModelId > 0 {
			fcModelOne, fcModelErr := common.DbMain.Client.QuickQuery(`tbl_ai_model`, `name`, map[string]any{`id`: fcModelId}).One()
			if fcModelErr == nil && fcModelOne[`name`] != `` {
				all[i][`fc_model_name`] = fcModelOne[`name`]
			}
		}
		if agentCliId > 0 {
			agentOne, agentErr := common.DbMain.Client.QuickQuery(`tbl_agent_cli`, `name`, map[string]any{`id`: agentCliId}).One()
			if agentErr == nil && agentOne[`name`] != `` {
				all[i][`agent_cli_name`] = agentOne[`name`]
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

// SetButlerConfigAdd 新增或更新管家运行参数（id==0 新增，id>0 更新）
func SetButlerConfigAdd(c *gin.Context) {
	if !checkButlerRunningTask(c) {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `role_id`, `model_id`, `fc_model_id`, `agent_cli_id`, `bot_config_id`, `active_timeout_minutes`, `max_history`, `auto_clean_on_new_topic`, `index_doc_path`, `auto_init_on_start`, `status`})
	var err error
	now := time.Now().Unix()
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`created_at`] = now
		updateData[`updated_at`] = now
		_, err = common.DbMain.Client.QuickCreate(`tbl_butler_config`, updateData).Exec()
	} else {
		updateData[`updated_at`] = now
		_, err = common.DbMain.Client.QuickUpdate(`tbl_butler_config`,
			map[string]any{`id`: dataMap[`id`]}, updateData).Exec()
	}
	if err != nil {
		gsgin.GinResponseError(c, `保存失败: `+err.Error(), nil)
		return
	}
	// 状态为启用时，尝试重启管家核心（首次创建或无核心时自动启动）
	if cast.ToInt(updateData[`status`]) == 1 {
		runtime := business.GetButlerRuntime()
		if runtime != nil {
			go runtime.RestartCore()
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetButlerConfigDelete 删除管家运行参数
func SetButlerConfigDelete(c *gin.Context) {
	if !checkButlerRunningTask(c) {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	_, _ = common.DbMain.Client.QuickDelete(`tbl_butler_config`, map[string]any{
		`id`: cast.ToInt(dataMap[`id`]),
	}).Exec()
	// 删除后重启核心（如无启用配置则自动停止）
	runtime := business.GetButlerRuntime()
	if runtime != nil {
		go runtime.RestartCore()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
