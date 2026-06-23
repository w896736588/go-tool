package business

import (
	"dev_tool/internal/app/dtool/butler"
	"dev_tool/internal/app/dtool/butler/bot"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// ButlerRuntime 管家运行时，管理所有已启用的机器人网关和管家核心实例。
// 支持多机器人同时连接：每个 status=1 的机器人配置对应一个 Gateway，
// 所有 Gateway 的消息合并投递到管家核心的统一消息通道。
// 同时实现 bot.GatewayProvider 接口，供管家核心获取 Gateway 实例。
type ButlerRuntime struct {
	db         *common.CSqlite
	butlerEnv  *define.ButlerEnv
	core       *butler.Core
	msgChan    chan bot.IncomingMessage
	gateways   map[int]bot.Gateway           // botConfigId → Gateway
	botConfigs map[int]*define.BotConfigItem // botConfigId → BotConfigItem
	mu         sync.Mutex
}

// NewButlerRuntime 创建管家运行时实例，db 为管家数据库，butlerEnv 为管家运行时环境。
func NewButlerRuntime(db *common.CSqlite, butlerEnv *define.ButlerEnv) *ButlerRuntime {
	return &ButlerRuntime{
		db:         db,
		butlerEnv:  butlerEnv,
		msgChan:    make(chan bot.IncomingMessage, 128),
		gateways:   make(map[int]bot.Gateway),
		botConfigs: make(map[int]*define.BotConfigItem),
	}
}

// Start 启动管家运行时：先连接所有已启用的机器人，再尝试加载管家配置并启动核心。
// 管家配置不存在时仅跳过核心启动，机器人仍可正常连接/断开。
func (r *ButlerRuntime) Start() error {
	// 1. 优先连接所有已启用的机器人（不依赖管家配置）
	r.connectAllEnabledBots()

	// 2. 加载管家配置（取 status=1 的第一条）
	butlerConfig, err := r.loadButlerConfig()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 管家配置未就绪 %s，跳过AI核心启动（机器人仍可用）`, err.Error())
		return nil
	}
	gstool.FmtPrintlnLogTime(`[butler] 管家配置: name=%s role_id=%d bot_config_id=%d`,
		butlerConfig.Name, butlerConfig.RoleId, butlerConfig.BotConfigId)

	// 3. 加载角色
	role, err := r.loadRole(butlerConfig.RoleId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 加载角色失败 %s，使用默认空角色`, err.Error())
		role = &define.RoleItem{}
	}
	gstool.FmtPrintlnLogTime(`[butler] 角色: name=%s persona=%s`, role.Name, role.Persona)

	// 4. 创建管家核心（传入 ButlerRuntime 作为 GatewayProvider）
	r.core = butler.NewCore(r.db, butlerConfig, r.butlerEnv, role, r, r.msgChan)
	// 主管家注入归档提交回调
	r.injectArchiveCallback(butlerConfig)
	r.core.Start()

	// 5. 启动归档管家（若存在启用的归档管家配置）
	r.startArchiveButler()

	gstool.FmtPrintlnLogTime(`[butler] 管家运行时已启动，已连接 %d 个机器人`, len(r.gateways))
	return nil
}

// RestartCore 根据最新管家配置重新启动核心。
// 先停止已有核心（如有），再重新加载配置并启动新核心。
// 管家配置不存在时仅停止旧核心，不报错。
func (r *ButlerRuntime) RestartCore() {
	r.mu.Lock()
	if r.core != nil {
		r.core.Stop()
		r.core = nil
	}
	r.mu.Unlock()

	butlerConfig, err := r.loadButlerConfig()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] RestartCore: 管家配置未就绪 %s，核心已停止`, err.Error())
		return
	}
	gstool.FmtPrintlnLogTime(`[butler] RestartCore: 管家配置 name=%s role_id=%d bot_config_id=%d`,
		butlerConfig.Name, butlerConfig.RoleId, butlerConfig.BotConfigId)

	role, err := r.loadRole(butlerConfig.RoleId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] RestartCore: 加载角色失败 %s，使用默认空角色`, err.Error())
		role = &define.RoleItem{}
	}
	gstool.FmtPrintlnLogTime(`[butler] RestartCore: 角色 name=%s persona=%s`, role.Name, role.Persona)

	r.mu.Lock()
	r.core = butler.NewCore(r.db, butlerConfig, r.butlerEnv, role, r, r.msgChan)
	r.mu.Unlock()
	r.injectArchiveCallback(butlerConfig)
	r.core.Start()
	gstool.FmtPrintlnLogTime(`[butler] RestartCore: 核心已重启，已连接 %d 个机器人`, len(r.gateways))
}

// Stop 停止管家运行时：断开所有机器人连接 → 停止管家核心。
func (r *ButlerRuntime) Stop() {
	gstool.FmtPrintlnLogTime(`[butler] 开始停止管家运行时`)
	if r.core != nil {
		r.core.Stop()
	}
	r.mu.Lock()
	for id, gw := range r.gateways {
		gw.Close()
		r.updateBotConnStatus(id, define.ConnStatusDisconnected, ``)
	}
	r.gateways = nil
	r.botConfigs = nil
	r.mu.Unlock()
	gstool.FmtPrintlnLogTime(`[butler] 管家运行时已停止`)
}

// ConnectBot 连接指定机器人配置。如果该机器人已连接则跳过。
// 成功连接后更新数据库连接状态为 ConnStatusConnected。
func (r *ButlerRuntime) ConnectBot(botConfig *define.BotConfigItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.gateways[botConfig.Id]; exists {
		gstool.FmtPrintlnLogTime(`[butler] 机器人 %s (id=%d) 已连接，跳过`, botConfig.Name, botConfig.Id)
		return nil
	}

	gateway := bot.NewDingTalkGateway(botConfig, r.msgChan)
	if err := gateway.Start(); err != nil {
		r.updateBotConnStatus(botConfig.Id, define.ConnStatusFailed, err.Error())
		return err
	}
	r.gateways[botConfig.Id] = gateway
	r.botConfigs[botConfig.Id] = botConfig
	r.updateBotConnStatus(botConfig.Id, define.ConnStatusConnected, ``)
	gstool.FmtPrintlnLogTime(`[butler] 机器人 %s (id=%d) 连接成功`, botConfig.Name, botConfig.Id)
	return nil
}

// DisconnectBot 断开指定机器人的连接，并更新数据库连接状态。
func (r *ButlerRuntime) DisconnectBot(botConfigId int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	gw, exists := r.gateways[botConfigId]
	if !exists {
		return
	}
	gw.Close()
	delete(r.gateways, botConfigId)
	delete(r.botConfigs, botConfigId)
	r.updateBotConnStatus(botConfigId, define.ConnStatusDisconnected, ``)
	gstool.FmtPrintlnLogTime(`[butler] 机器人 (id=%d) 已断开连接`, botConfigId)
}

// IsBotConnected 查询指定机器人是否已连接。
func (r *ButlerRuntime) IsBotConnected(botConfigId int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.gateways[botConfigId]
	return exists
}

// GetGateway 获取指定机器人的网关实例（bot.GatewayProvider 接口实现）。
func (r *ButlerRuntime) GetGateway(botConfigId int) bot.Gateway {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.gateways[botConfigId]
}

// GetAllGateways 返回所有已连接的机器人网关（bot.GatewayProvider 接口实现）。
func (r *ButlerRuntime) GetAllGateways() map[int]bot.Gateway {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 返回副本避免外部并发修改
	result := make(map[int]bot.Gateway, len(r.gateways))
	for id, gw := range r.gateways {
		result[id] = gw
	}
	return result
}

// connectAllEnabledBots 扫描 tbl_butler_bot_config 中所有 status=1 的记录并逐一连接。
func (r *ButlerRuntime) connectAllEnabledBots() {
	rows, err := r.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_bot_config WHERE status = 1 ORDER BY id ASC`,
	).All()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 查询已启用的机器人配置失败 %s`, err.Error())
		return
	}
	for _, row := range rows {
		botConfig := r.rowToBotConfigItem(row)
		if err := r.ConnectBot(botConfig); err != nil {
			gstool.FmtPrintlnLogTime(`[butler] 机器人 %s (id=%d) 连接失败 %s`, botConfig.Name, botConfig.Id, err.Error())
			// 连接失败不 panic，继续连接其他机器人
		}
	}
}

// loadButlerConfig 从共用库读取启用的管家配置（status=1 的第一条）。
func (r *ButlerRuntime) loadButlerConfig() (*define.ButlerConfigItem, error) {
	row, err := r.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_config WHERE status = 1 ORDER BY id ASC`,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf(`未找到启用的管家配置，请在 dtool 中配置 tbl_butler_config`)
	}
	return r.rowToButlerConfigItem(row), nil
}

// loadRole 根据 roleId 读取角色配置。
func (r *ButlerRuntime) loadRole(roleId int) (*define.RoleItem, error) {
	row, err := r.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_role WHERE id = ? AND status = 1`, roleId,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf(`未找到角色 id=%d`, roleId)
	}
	return &define.RoleItem{
		Id:           cast.ToInt(row[`id`]),
		Name:         cast.ToString(row[`name`]),
		Persona:      cast.ToString(row[`persona`]),
		Tone:         cast.ToString(row[`tone`]),
		SystemPrompt: cast.ToString(row[`system_prompt`]),
		InitGreeting: cast.ToString(row[`init_greeting`]),
		Status:       cast.ToInt(row[`status`]),
	}, nil
}

// rowToBotConfigItem 将数据库行映射为 BotConfigItem 结构体。
func (r *ButlerRuntime) rowToBotConfigItem(row map[string]any) *define.BotConfigItem {
	return &define.BotConfigItem{
		Id:        cast.ToInt(row[`id`]),
		Platform:  cast.ToString(row[`platform`]),
		Name:      cast.ToString(row[`name`]),
		AppKey:    cast.ToString(row[`app_key`]),
		AppSecret: cast.ToString(row[`app_secret`]),
		RobotCode: cast.ToString(row[`robot_code`]),
		Status:    cast.ToInt(row[`status`]),
	}
}

// rowToButlerConfigItem 将数据库行映射为 ButlerConfigItem 结构体。
func (r *ButlerRuntime) rowToButlerConfigItem(row map[string]any) *define.ButlerConfigItem {
	return &define.ButlerConfigItem{
		Id:                   cast.ToInt(row[`id`]),
		Name:                 cast.ToString(row[`name`]),
		RoleId:               cast.ToInt(row[`role_id`]),
		ModelId:              cast.ToInt(row[`model_id`]),
		FcModelId:            cast.ToInt(row[`fc_model_id`]),
		AgentCliId:           cast.ToInt(row[`agent_cli_id`]),
		BotConfigId:          cast.ToInt(row[`bot_config_id`]),
		ActiveTimeoutMinutes: cast.ToInt(row[`active_timeout_minutes`]),
		MaxHistoryStore:      cast.ToInt(row[`max_history_store`]),
		IndexDocPath:         cast.ToString(row[`index_doc_path`]),
		AutoInitOnStart:      cast.ToInt(row[`auto_init_on_start`]),
		MaxLoop:              cast.ToInt(row[`max_loop`]),
		ToolCallPushEnabled:  cast.ToInt(row[`tool_call_push_enabled`]),
		ButlerType:           cast.ToInt(row[`butler_type`]),
		Status:               cast.ToInt(row[`status`]),
	}
}

// updateBotConnStatus 更新机器人连接状态到 tbl_butler_bot_config。
func (r *ButlerRuntime) updateBotConnStatus(botConfigId, status int, errMsg string) {
	_, err := r.db.Client.QuickUpdate(`tbl_butler_bot_config`, map[string]any{`id`: botConfigId}, map[string]any{
		`conn_status`:    status,
		`conn_status_at`: time.Now().Unix(),
		`conn_error`:     errMsg,
	}).Exec()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 更新连接状态失败 %s`, err.Error())
	}
}

// InitButlerRuntime 构建管家运行时环境并启动管家，连接所有已启用的机器人。
// 供 dtool config.go 的 initButlerRuntime 调用。
func InitButlerRuntime() {
	butlerEnv := buildButlerEnv()
	runtime := NewButlerRuntime(component.DbMain, butlerEnv)
	component.ButlerRuntime = runtime
	if err := runtime.Start(); err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 管家运行时启动失败 %s，管家功能不可用`, err.Error())
		// 不 panic，允许 dtool 在无管家的情况下继续运行
		return
	}
}

// GetButlerRuntime 获取管家运行时实例的完整类型，供 controller 调用 ConnectBot 等方法。
// 返回 nil 表示管家运行时未初始化。
func GetButlerRuntime() *ButlerRuntime {
	if component.ButlerRuntime == nil {
		return nil
	}
	runtime, ok := component.ButlerRuntime.(*ButlerRuntime)
	if !ok {
		return nil
	}
	return runtime
}

// buildButlerEnv 从 dtool Env 构建 butler 的 define.ButlerEnv，供管家内部包使用。
// 管家表已合并到主库，不再使用独立数据库文件。
func buildButlerEnv() *define.ButlerEnv {
	env := component.EnvClient
	memoryDbPath := common.ResolveDefaultDToolDir(env.ConfigBase.MemoryDBPath)
	// 构建 dtool API 基地址：默认取第一个 API 端口
	baseURL := `http://localhost:17170`
	if len(env.ApiPorts) > 0 {
		baseURL = fmt.Sprintf(`http://localhost:%s`, env.ApiPorts[0])
	}
	return &define.ButlerEnv{
		RootPath:      env.RootPath,
		ConfigPath:    env.ConfigPath,
		ConfigFile:    env.ConfigFile,
		DbPath:        env.DbConfig.DbPath,
		DbName:        env.DbConfig.DbName,
		LogDbPath:     env.LogDbConfig.DbPath,
		MemoryDbPath:  memoryDbPath,
		DatabaseUpDir: env.DatabaseUpPath,
		LogPath:       env.LogPath,
		DtoolBaseURL:  baseURL,
	}
}

// DingtalkSendSingleChatMsg 通过钉钉 Open API 发送单聊文本消息。
// appKey/appSecret 用于获取 access_token，robotCode 为机器人编码，userId 为接收者内部 ID。
// 供 controller 层测试连通性时使用。
func DingtalkSendSingleChatMsg(appKey, appSecret, robotCode, userId, text string) error {
	return bot.SendDingtalkSingleChatMsg(appKey, appSecret, robotCode, userId, text)
}

// DingtalkGetAccessToken 获取钉钉 access_token（供测试接口获取管理员列表时使用）。
func DingtalkGetAccessToken(appKey, appSecret string) (string, error) {
	return bot.GetDingtalkAccessToken(appKey, appSecret)
}

// DingtalkGetAppAdmins 获取钉钉应用管理员列表（供测试接口获取接收者时使用）。
func DingtalkGetAppAdmins(accessToken, appKey string) ([]string, error) {
	return bot.GetDingtalkAppAdmins(accessToken, appKey)
}

// ==================== 归档管线 ====================

// injectArchiveCallback 向管家核心注入归档提交回调（仅主管家生效）。
func (r *ButlerRuntime) injectArchiveCallback(config *define.ButlerConfigItem) {
	if config.ButlerType != define.ButlerTypeMain {
		return
	}
	r.core.SetArchiveSubmit(func(configId, taskId int, sessionId string, files []string, conversation string) {
		id, err := r.db.CreateArchiveRecord(configId, taskId, sessionId, files, conversation)
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler-archive] 创建归档记录失败 config_id=%d session=%s err=%s`, configId, sessionId, err.Error())
		} else if id <= 0 {
			gstool.FmtPrintlnLogTime(`[butler-archive] 创建归档记录异常 id=%d config_id=%d session=%s`, id, configId, sessionId)
		} else {
			gstool.FmtPrintlnLogTime(`[butler-archive] 已创建归档记录 id=%d config_id=%d files=%d`, id, configId, len(files))
		}
	})
}

// startArchiveButler 启动归档管家后台协程（若存在 butler_type=2 的启用配置）。
func (r *ButlerRuntime) startArchiveButler() {
	archiveConfig, err := r.loadArchiveConfig()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-archive] 归档管家配置未就绪 %s`, err.Error())
		return
	}
	gstool.FmtPrintlnLogTime(`[butler-archive] 归档管家已启动 config_id=%d name=%s`, archiveConfig.Id, archiveConfig.Name)
	go r.archiveLoop(archiveConfig)
}

// loadArchiveConfig 加载启用状态的归档管家配置（butler_type=2, status=1，取第一条）。
func (r *ButlerRuntime) loadArchiveConfig() (*define.ButlerConfigItem, error) {
	row, err := r.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_config WHERE butler_type = ? AND status = ? ORDER BY id ASC LIMIT 1`,
		define.ButlerTypeArchive, 1,
	).One()
	if err != nil || len(row) == 0 {
		return nil, fmt.Errorf(`未找到启用的归档管家配置`)
	}
	return r.rowToButlerConfigItem(row), nil
}

// archiveLoop 归档管家轮询主循环，每 30 秒检查一次待处理的归档记录。
func (r *ButlerRuntime) archiveLoop(config *define.ButlerConfigItem) {
	gstool.FmtPrintlnLogTime(`[butler-archive] 轮询协程已启动 config_id=%d 等待首次轮询(30s后)`, config.Id)

	defer func() {
		if rec := recover(); rec != nil {
			gstool.FmtPrintlnLogTime(`[butler-archive] 轮询协程 panic 退出 config_id=%d panic=%v`, config.Id, rec)
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		items, err := r.db.ListPendingArchives(5)
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler-archive] 查询待处理归档失败 %s`, err.Error())
			continue
		}
		if len(items) > 0 {
			gstool.FmtPrintlnLogTime(`[butler-archive] 发现 %d 条待处理归档记录`, len(items))
		}
		for _, item := range items {
			r.processArchiveItem(config, item)
		}
	}
}

// processArchiveItem 处理单条归档记录：AI 评估 → 生成通用脚本 → 写入文件。
func (r *ButlerRuntime) processArchiveItem(config *define.ButlerConfigItem, item map[string]any) {
	archiveId := cast.ToInt(item[`id`])

	defer func() {
		if rec := recover(); rec != nil {
			gstool.FmtPrintlnLogTime(`[butler-archive] 处理归档记录 panic id=%d panic=%v`, archiveId, rec)
			_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored,
				fmt.Sprintf(`归档管家 panic: %v`, rec),
				fmt.Sprintf(`panic: %v`, rec), ``, ``)
		}
	}()

	logBuilder := &strings.Builder{}
	logBuilder.WriteString(fmt.Sprintf("开始处理归档记录 id=%d\n", archiveId))

	// 标记为处理中
	_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusProcessing, logBuilder.String(), ``, ``, ``)

	conversation := cast.ToString(item[`conversation`])

	// 步骤1：AI 评估是否值得自进化
	evalModelId := config.FcModelId
	if evalModelId <= 0 {
		evalModelId = config.ModelId
	}
	if evalModelId <= 0 {
		logBuilder.WriteString("未配置 AI 模型，标记为已忽略\n")
		_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored, logBuilder.String(), `未配置 AI 模型`, ``, ``)
		return
	}

	evalPrompt := fmt.Sprintf(`你是归档管家（自进化评估器），负责判断主管家产生的文件+对话是否有复用价值。

## 对话与文件内容
%s

请判断：
1. 这次任务的操作模式是否具有通用复用价值？（例如：查询配置、操作数据库、调用API等）
2. 是否适合抽象为通用 Python 脚本？

输出格式（严格按此格式）：
评估结论：[YES/NO]
理由：[简要说明]
脚本名建议：[如果YES，建议的脚本文件名，如 query_git_branches.py]
脚本描述：[如果YES，一行描述]`,
		conversation)

	logBuilder.WriteString(fmt.Sprintf("AI评估中 模型=%d prompt长度=%d\n", evalModelId, len(evalPrompt)))
	result, _, evalErr := r.db.AIChatByModel(evalModelId,
		`你是归档管家（自进化评估器），负责判断主管家产生的文件+对话是否有复用价值。`,
		evalPrompt)
	if evalErr != nil {
		logBuilder.WriteString(fmt.Sprintf("AI评估失败 %s\n", evalErr.Error()))
		_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored, logBuilder.String(), fmt.Sprintf(`AI评估失败: %s`, evalErr.Error()), ``, ``)
		return
	}

	logBuilder.WriteString(fmt.Sprintf("AI评估结果: %s\n", strings.TrimSpace(result)))

	// 判断是否值得自进化
	if !strings.Contains(result, `评估结论：YES`) && !strings.Contains(result, `评估结论: YES`) {
		logBuilder.WriteString("评估结论：无需自进化，标记为已忽略\n")
		_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored, logBuilder.String(), result, ``, ``)
		return
	}

	// 步骤2：生成通用脚本
	logBuilder.WriteString("评估结论：需要自进化，开始生成脚本\n")
	genPrompt := fmt.Sprintf(`基于以下主管家执行记录，编写通用 Python 脚本。

## 评估结论
%s

## 对话与执行详情
%s

请输出两部分（用 ===SCRIPT=== 和 ===INDEX=== 分隔）：
===SCRIPT===
[完整的 Python 脚本代码]
===INDEX===
[scripts.md 索引条目格式：## [skill名称] 描述\n- 脚本: script_name.py\n- 来源: 归档管家自进化]`,
		result, conversation)

	genResult, _, genErr := r.db.AIChatByModel(evalModelId,
		`你是 dtool 智能管家，负责编写通用工具 Python 脚本。只输出脚本代码和索引，不输出其他内容。`,
		genPrompt)
	if genErr != nil {
		logBuilder.WriteString(fmt.Sprintf("脚本生成失败 %s\n", genErr.Error()))
		_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored, logBuilder.String(), fmt.Sprintf(`脚本生成失败: %s`, genErr.Error()), ``, ``)
		return
	}

	// 解析脚本和索引
	scriptContent, indexEntry, scriptName := r.parseArchiveGenResult(genResult)
	if scriptContent == `` {
		logBuilder.WriteString("AI 未生成有效脚本内容\n")
		_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored, logBuilder.String(), genResult, ``, ``)
		return
	}

	// 写入脚本文件
	scriptFile, writeErr := common.WriteArchiveScript(r.butlerEnv.RootPath, scriptName, scriptContent)
	if writeErr != nil {
		logBuilder.WriteString(fmt.Sprintf("写脚本文件失败 %s\n", writeErr.Error()))
		_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusIgnored, logBuilder.String(), fmt.Sprintf(`写文件失败: %s`, writeErr.Error()), ``, ``)
		return
	}
	logBuilder.WriteString(fmt.Sprintf("脚本已写入 %s\n", scriptFile))

	// 追加索引
	if indexEntry != `` {
		_ = common.AppendArchiveIndex(r.butlerEnv.RootPath, `dtool-butler`, scriptName, indexEntry)
		logBuilder.WriteString("索引已追加 scripts.md\n")
	}

	logBuilder.WriteString("自进化完成\n")
	_ = r.db.UpdateArchiveStatus(archiveId, define.ArchiveStatusDone, logBuilder.String(), result, scriptFile, indexEntry)
	r.notifyArchiveEvent(config, archiveId, cast.ToString(item[`session_id`]), scriptFile, define.ArchiveStatusDone, scriptName)
}

// parseArchiveGenResult 解析 AI 生成的脚本内容和索引条目。
// 返回: 脚本内容, 索引描述, 脚本文件名
func (r *ButlerRuntime) parseArchiveGenResult(genResult string) (scriptContent, indexEntry, scriptName string) {
	parts := strings.SplitN(genResult, `===SCRIPT===`, 2)
	if len(parts) < 2 {
		return ``, ``, ``
	}
	remaining := parts[1]
	scriptParts := strings.SplitN(remaining, `===INDEX===`, 2)
	scriptContent = strings.TrimSpace(scriptParts[0])
	if len(scriptParts) > 1 {
		indexEntry = strings.TrimSpace(scriptParts[1])
	}

	if idx := strings.Index(indexEntry, `.py`); idx > 0 {
		start := strings.LastIndex(indexEntry[:idx], ` `)
		if start < 0 {
			start = strings.LastIndex(indexEntry[:idx], `-`)
		}
		if start >= 0 {
			scriptName = strings.TrimSpace(indexEntry[start:idx] + `.py`)
		}
	}
	if scriptName == `` {
		scriptName = fmt.Sprintf(`archive_%d.py`, time.Now().Unix())
	}
	return
}

// notifyArchiveEvent 推送归档处理通知到 Webhook。无 webhook 配置时静默跳过。
func (r *ButlerRuntime) notifyArchiveEvent(config *define.ButlerConfigItem, archiveId int, sessionId, detail, status, reason string) {
	webhookCfg := GetWebhookConfigByAgentCliId(config.AgentCliId)
	if webhookCfg == nil {
		return
	}

	var title, text string
	switch status {
	case define.ArchiveStatusProcessing:
		title = fmt.Sprintf("[归档开始] #%d", archiveId)
		text = fmt.Sprintf("会话: %s\n文件: %s", sessionId, truncateForNotify(detail, 200))
	case define.ArchiveStatusDone:
		title = fmt.Sprintf("[归档完成] #%d", archiveId)
		text = fmt.Sprintf("会话: %s\n产出脚本: %s\n描述: %s", sessionId, detail, truncateForNotify(reason, 100))
	case define.ArchiveStatusIgnored:
		title = fmt.Sprintf("[归档忽略] #%d", archiveId)
		text = fmt.Sprintf("会话: %s\n原因: %s", sessionId, truncateForNotify(reason, 200))
	default:
		return
	}

	if err := SendWebhookNotify(webhookCfg, title, text, ``); err != nil {
		gstool.FmtPrintlnLogTime("[butler-archive] 通知发送失败 id=%d status=%s err=%s", archiveId, status, err.Error())
	} else {
		gstool.FmtPrintlnLogTime("[butler-archive] 通知已发送 id=%d status=%s", archiveId, status)
	}
}

// truncateForNotify 截断字符串用于通知文本。
func truncateForNotify(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
