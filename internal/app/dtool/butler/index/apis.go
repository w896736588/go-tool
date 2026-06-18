package index

import (
	"fmt"
	"strings"
)

// ApiEntry dtool HTTP 接口条目，用于 apis.md 生成。
type ApiEntry struct {
	Path        string // 接口路径
	Method      string // 请求方法（POST）
	Description string // 功能描述
}

// dtoolApis 预定义的 dtool HTTP 接口列表（管家可能需要调用的关键接口）。
// 但这里列出全部可用接口供管家参考，便于 FC 工具调用时定位。
var dtoolApis = []ApiEntry{
	// 基础接口
	{Path: "/api/BaseLogin", Method: "POST", Description: "登录认证"},
	{Path: "/api/BaseLoginStatus", Method: "POST", Description: "登录状态检查"},
	// 配置管理（Set 类）
	{Path: "/api/Set/SshList", Method: "POST", Description: "SSH 配置列表"},
	{Path: "/api/Set/SshAdd", Method: "POST", Description: "新增/更新 SSH 配置"},
	{Path: "/api/Set/SshDelete", Method: "POST", Description: "删除 SSH 配置"},
	{Path: "/api/Set/GitList", Method: "POST", Description: "Git 配置列表"},
	{Path: "/api/Set/GitAdd", Method: "POST", Description: "新增/更新 Git 配置"},
	{Path: "/api/Set/GitDelete", Method: "POST", Description: "删除 Git 配置"},
	{Path: "/api/Set/GlobalList", Method: "POST", Description: "全局配置列表"},
	{Path: "/api/Set/GlobalCreate", Method: "POST", Description: "新增全局配置"},
	{Path: "/api/Set/GlobalDelete", Method: "POST", Description: "删除全局配置"},
	{Path: "/api/Set/AiProviderList", Method: "POST", Description: "AI 服务商列表"},
	{Path: "/api/Set/AiProviderAdd", Method: "POST", Description: "新增/更新 AI 服务商"},
	{Path: "/api/Set/AiProviderDelete", Method: "POST", Description: "删除 AI 服务商"},
	{Path: "/api/Set/AiModelList", Method: "POST", Description: "AI 模型列表"},
	{Path: "/api/Set/AiModelAdd", Method: "POST", Description: "新增/更新 AI 模型"},
	{Path: "/api/Set/AiModelDelete", Method: "POST", Description: "删除 AI 模型"},
	{Path: "/api/Set/MemoryConfigGet", Method: "POST", Description: "记忆库配置查询"},
	{Path: "/api/Set/MemoryConfigSave", Method: "POST", Description: "记忆库配置保存"},
	{Path: "/api/Set/CronConfigGet", Method: "POST", Description: "定时任务配置查询"},
	{Path: "/api/Set/CronConfigSave", Method: "POST", Description: "定时任务配置保存"},
	{Path: "/api/Set/HomeTaskConfigGet", Method: "POST", Description: "首页任务配置查询"},
	{Path: "/api/Set/HomeTaskConfigSave", Method: "POST", Description: "首页任务配置保存"},
	// Butler 管家配置
	{Path: "/api/Set/ButlerBotConfigList", Method: "POST", Description: "管家机器人配置列表"},
	{Path: "/api/Set/ButlerBotConfigAdd", Method: "POST", Description: "新增/更新管家机器人配置"},
	{Path: "/api/Set/ButlerBotConfigDelete", Method: "POST", Description: "删除管家机器人配置"},
	{Path: "/api/Set/ButlerBotConfigTest", Method: "POST", Description: "测试管家机器人Stream连接"},
	{Path: "/api/Set/ButlerMessageList", Method: "POST", Description: "管家机器人消息日志列表"},
	{Path: "/api/Set/ButlerRoleList", Method: "POST", Description: "管家角色列表"},
	{Path: "/api/Set/ButlerRoleAdd", Method: "POST", Description: "新增/更新管家角色"},
	{Path: "/api/Set/ButlerRoleDelete", Method: "POST", Description: "删除管家角色"},
	{Path: "/api/Set/ButlerConfigList", Method: "POST", Description: "管家运行参数列表"},
	{Path: "/api/Set/ButlerConfigAdd", Method: "POST", Description: "新增/更新管家运行参数"},
	{Path: "/api/Set/ButlerConfigDelete", Method: "POST", Description: "删除管家运行参数"},
	// Git 操作
	{Path: "/api/GitQueryCurrentBranch", Method: "POST", Description: "查询当前分支"},
	{Path: "/api/GitChangeBranch", Method: "POST", Description: "切换分支"},
	{Path: "/api/GitPullBranchOrigin", Method: "POST", Description: "拉取最新分支"},
	{Path: "/api/GitRemoteBranchList", Method: "POST", Description: "查询远程分支列表"},
	{Path: "/api/GitQueryStatus", Method: "POST", Description: "查询本地状态"},
	{Path: "/api/GitCommitLog", Method: "POST", Description: "查询提交日志"},
	// 首页任务
	{Path: "/api/HomeTaskList", Method: "POST", Description: "首页任务列表"},
	{Path: "/api/HomeTaskInfo", Method: "POST", Description: "首页任务详情"},
	{Path: "/api/HomeTaskSave", Method: "POST", Description: "保存首页任务"},
	{Path: "/api/HomeTaskDelete", Method: "POST", Description: "删除首页任务"},
	// 任务工作流
	{Path: "/api/task/workflow/create_or_get", Method: "POST", Description: "创建或获取工作流"},
	{Path: "/api/task/workflow/info", Method: "POST", Description: "工作流详情"},
	{Path: "/api/task/workflow/chat/send", Method: "POST", Description: "工作流对话发送"},
	{Path: "/api/task/workflow/chat/list", Method: "POST", Description: "工作流对话列表"},
	// 知识片段
	{Path: "/api/MemoryFragmentList", Method: "POST", Description: "知识片段列表"},
	{Path: "/api/MemoryFragmentInfo", Method: "POST", Description: "知识片段详情"},
	{Path: "/api/MemoryFragmentSave", Method: "POST", Description: "保存知识片段"},
	{Path: "/api/MemoryFragmentDelete", Method: "POST", Description: "删除知识片段"},
	{Path: "/api/MemoryFragmentSearch", Method: "POST", Description: "知识片段搜索"},
	// Redis
	{Path: "/api/RedisAvailableList", Method: "POST", Description: "可用的 Redis 列表"},
	{Path: "/api/RedisSearch", Method: "POST", Description: "查询 Redis Key"},
	{Path: "/api/RedisKeys", Method: "POST", Description: "模糊搜索 Key"},
	{Path: "/api/RedisSaveString", Method: "POST", Description: "保存 Redis String"},
	{Path: "/api/RedisDelKey", Method: "POST", Description: "删除 Redis Key"},
	// MySQL
	{Path: "/api/MysqlTables", Method: "POST", Description: "查询 MySQL 表列表"},
	{Path: "/api/MysqlTableStructure", Method: "POST", Description: "查询 MySQL 表结构"},
	{Path: "/api/MysqlQuery", Method: "POST", Description: "执行 MySQL 查询"},
	{Path: "/api/MysqlExec", Method: "POST", Description: "执行 MySQL 写入"},
	// API 管理
	{Path: "/api/Collections", Method: "POST", Description: "API 集合列表"},
	{Path: "/api/Apis", Method: "POST", Description: "API 列表"},
	{Path: "/api/ApiRun", Method: "POST", Description: "执行 API 请求"},
	// Agent CLI
	{Path: "/api/AgentCliList", Method: "POST", Description: "Agent CLI 列表"},
	{Path: "/api/AgentCliSave", Method: "POST", Description: "保存 Agent CLI"},
	{Path: "/api/AgentCliDelete", Method: "POST", Description: "删除 Agent CLI"},
	// Smart Link
	{Path: "/api/SmartLinkItemList", Method: "POST", Description: "Smart Link 列表"},
	{Path: "/api/SmartLinkRun", Method: "POST", Description: "执行 Smart Link Playwright"},
	// Shell
	{Path: "/api/shellOut", Method: "POST", Description: "执行远程 Shell 命令"},
	// 变量
	{Path: "/api/VariableList", Method: "POST", Description: "变量列表"},
	{Path: "/api/VariableRun", Method: "POST", Description: "执行变量命令"},
	// Docker
	{Path: "/api/DockerComposeList", Method: "POST", Description: "Docker Compose 列表"},
	{Path: "/api/DockerComposeRestart", Method: "POST", Description: "重启 Docker Compose"},
	{Path: "/api/DockerComposeStatus", Method: "POST", Description: "查询 Docker Compose 状态"},
	// MCP
	{Path: "/api/McpTypeList", Method: "POST", Description: "MCP 类型列表"},
	{Path: "/api/McpBindingList", Method: "POST", Description: "MCP 绑定列表"},
	{Path: "/api/McpBindingAdd", Method: "POST", Description: "新增 MCP 绑定"},
	// Webhook
	{Path: "/api/WebhookConfigList", Method: "POST", Description: "Webhook 配置列表"},
	{Path: "/api/WebhookConfigTest", Method: "POST", Description: "测试 Webhook"},
	// 工具类
	{Path: "/api/Ip", Method: "POST", Description: "获取外网 IP"},
	{Path: "/api/GetLocalIP", Method: "POST", Description: "获取局域网 IP"},
	{Path: "/api/Upload", Method: "POST", Description: "上传文件"},
}

// GenerateApisIndex 生成 apis.md 内容——dtool HTTP 接口索引。
func GenerateApisIndex() string {
	var sb strings.Builder
	sb.WriteString("# dtool HTTP 接口索引\n\n")
	sb.WriteString(fmt.Sprintf("共 %d 个可用接口。所有接口均为 POST 方法，需携带 Token 鉴权。\n\n", len(dtoolApis)))

	// 按功能分组
	groups := groupApisByPrefix(dtoolApis)
	for _, group := range groups {
		sb.WriteString(fmt.Sprintf("## %s\n\n", group.Name))
		sb.WriteString("| 接口路径 | 说明 |\n|----------|------|\n")
		for _, api := range group.Apis {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", api.Path, api.Description))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// ApiGroup 接口分组。
type ApiGroup struct {
	Name string     // 分组名称
	Apis []ApiEntry // 分组内的接口列表
}

// groupApisByPrefix 按接口路径前缀分组。
func groupApisByPrefix(apis []ApiEntry) []ApiGroup {
	groupMap := map[string][]ApiEntry{}
	groupOrder := []string{}
	for _, api := range apis {
		prefix := extractGroupPrefix(api.Path)
		if _, exists := groupMap[prefix]; !exists {
			groupOrder = append(groupOrder, prefix)
		}
		groupMap[prefix] = append(groupMap[prefix], api)
	}
	groups := make([]ApiGroup, 0, len(groupOrder))
	for _, prefix := range groupOrder {
		groups = append(groups, ApiGroup{
			Name: prefix,
			Apis: groupMap[prefix],
		})
	}
	return groups
}

// extractGroupPrefix 从接口路径提取分组前缀。
// /api/Set/Xxx → "配置管理 (Set)"
// /api/task/workflow/xxx → "任务工作流"
// /api/Xxx → "基础功能"
func extractGroupPrefix(path string) string {
	parts := strings.Split(path, "/")
	// /api/ 之后的第一个部分决定分组
	if len(parts) >= 3 {
		segment := parts[2]
		switch segment {
		case "Set":
			return "配置管理 (Set)"
		case "task":
			return "任务工作流"
		case "MemoryFragment":
			return "知识片段"
		default:
			// 大写开头的一般是功能模块
			if len(segment) > 0 && segment[0] >= 'A' && segment[0] <= 'Z' {
				return segment + " 模块"
			}
			return "基础功能"
		}
	}
	return "其他"
}
