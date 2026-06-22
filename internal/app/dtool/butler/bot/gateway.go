package bot

import (
	"dev_tool/internal/app/dtool/define"
)

// Gateway 定义机器人网关接口，预留多平台扩展（当前仅钉钉实现）。
type Gateway interface {
	// Start 建立长连接并开始接收消息，非阻塞，内部起 goroutine。
	Start() error
	// Close 关闭连接。
	Close()
	// SendMarkdown 通过 Open API 主动发送 markdown 消息到指定用户（单聊）。
	// userId 为接收者内部员工 ID（senderStaffId），为空时跳过（无法确定接收者）。
	// 用于打招呼/休眠通知等无 incoming 消息上下文的场景。
	SendMarkdown(userId, title, text string) error
	// GetBotConfig 返回机器人配置，供外部获取 AppKey/AppSecret/RobotCode 等。
	GetBotConfig() *define.BotConfigItem
}

// GatewayProvider 网关提供者接口，管家核心通过此接口获取所有已连接的网关实例。
// 由 ButlerRuntime 实现，支持多机器人场景下遍历所有 Gateway 发送消息。
type GatewayProvider interface {
	// GetAllGateways 返回所有已连接的机器人网关（botConfigId → Gateway）。
	GetAllGateways() map[int]Gateway
	// GetGateway 返回指定机器人配置 ID 对应的网关实例。
	GetGateway(botConfigId int) Gateway
}
