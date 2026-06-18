package bot

import (
	"context"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// IncomingMessage 接收到的机器人消息，统一格式后投递给管家核心处理。
type IncomingMessage struct {
	ConversationId   string
	ConversationType string // 1 单聊 2 群聊
	SenderNick       string
	SenderStaffId    string // 发送者 userId（内部员工 ID），用于 Open API 单聊回复
	Text             string
	SessionWebhook   string // 用于回复该消息的临时 webhook
	BotConfigId      int    // 来源机器人配置 ID，用于多机器人场景下定位 Gateway
}

// DingTalkGateway 钉钉 Stream 模式网关实现。
type DingTalkGateway struct {
	botConfig   *define.BotConfigItem
	botConfigId int // 机器人配置 ID，用于消息投递时标识来源
	cli         *client.StreamClient
	msgChan     chan<- IncomingMessage
}

// NewDingTalkGateway 创建钉钉网关，msgChan 为消息投递通道。
func NewDingTalkGateway(botConfig *define.BotConfigItem, msgChan chan<- IncomingMessage) *DingTalkGateway {
	return &DingTalkGateway{
		botConfig:   botConfig,
		botConfigId: botConfig.Id,
		msgChan:     msgChan,
	}
}

// Start 建立 Stream 长连接并注册机器人消息回调，非阻塞。
func (g *DingTalkGateway) Start() error {
	if g.botConfig == nil || g.botConfig.AppKey == `` || g.botConfig.AppSecret == `` {
		return fmt.Errorf(`钉钉机器人配置缺失 app_key/app_secret`)
	}
	logger.SetLogger(logger.NewStdTestLoggerWithDebug())
	g.cli = client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(g.botConfig.AppKey, g.botConfig.AppSecret)),
	)
	g.cli.RegisterChatBotCallbackRouter(g.onChatBotMessage)
	if err := g.cli.Start(context.Background()); err != nil {
		return fmt.Errorf(`钉钉 Stream 启动失败 %w`, err)
	}
	gstool.FmtPrintlnLogTime(`[butler-bot] 钉钉 Stream 连接成功`)
	return nil
}

// Close 关闭 Stream 连接。
func (g *DingTalkGateway) Close() {
	if g.cli != nil {
		g.cli.Close()
	}
}

// SendText 通过钉钉 Open API 单聊消息发送接口主动推送文本消息。
// userId 为接收者内部员工 ID（senderStaffId）；为空时跳过（无法确定接收者）。
// 用于打招呼/休眠通知等无 incoming 消息上下文的场景。
func (g *DingTalkGateway) SendText(userId, text string) error {
	if g.botConfig == nil {
		return fmt.Errorf(`机器人配置为空`)
	}
	if userId == `` {
		gstool.FmtPrintlnLogTime(`[butler-bot] userId 为空，跳过主动推送（无法确定接收者）`)
		return nil
	}
	robotCode := g.botConfig.RobotCode
	if robotCode == `` {
		gstool.FmtPrintlnLogTime(`[butler-bot] robot_code 未配置，跳过主动推送`)
		return nil
	}
	return SendDingtalkSingleChatMsg(g.botConfig.AppKey, g.botConfig.AppSecret, robotCode, userId, text)
}

// GetBotConfig 返回机器人配置。
func (g *DingTalkGateway) GetBotConfig() *define.BotConfigItem {
	return g.botConfig
}

// onChatBotMessage 钉钉机器人消息回调，解析后投递到消息通道。
func (g *DingTalkGateway) onChatBotMessage(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	if data == nil {
		return []byte(``), nil
	}
	msg := IncomingMessage{
		ConversationId:   data.ConversationId,
		ConversationType: data.ConversationType,
		SenderNick:       data.SenderNick,
		SenderStaffId:    data.SenderStaffId,
		Text:             cast.ToString(data.Text.Content),
		SessionWebhook:   data.SessionWebhook,
		BotConfigId:      g.botConfigId,
	}
	gstool.FmtPrintlnLogTime(`[butler-bot] 收到消息 会话=%s 发送者=%s 内容=%s`,
		msg.ConversationId, msg.SenderNick, msg.Text)
	go func() {
		g.msgChan <- msg
	}()
	return []byte(``), nil
}

// ========== 钉钉 Open API 交互 ==========

// dingtalkApiBase 钉钉新版 Open API 基础 URL。
const dingtalkApiBase = `https://api.dingtalk.com`

// GetDingtalkAccessToken 通过钉钉 OAuth2 API 获取 access_token。
// 使用新版 api.dingtalk.com 端点。
func GetDingtalkAccessToken(appKey, appSecret string) (string, error) {
	// 构建请求
	body := map[string]string{
		`appKey`:    appKey,
		`appSecret`: appSecret,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return ``, fmt.Errorf(`json marshal: %w`, err)
	}
	url := dingtalkApiBase + `/v1.0/oauth2/accessToken`
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Post(url, "application/json; charset=utf-8", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return ``, fmt.Errorf(`http post: %w`, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return ``, fmt.Errorf(`response parse: %w`, err)
	}
	token := cast.ToString(result[`accessToken`])
	if token == `` {
		return ``, fmt.Errorf(`获取 access_token 失败: %v`, result)
	}
	return token, nil
}

// SendDingtalkSingleChatMsg 通过钉钉 Open API 发送单聊文本消息。
// robotCode 为机器人编码，userId 为接收者内部员工 ID。
func SendDingtalkSingleChatMsg(appKey, appSecret, robotCode, userId, text string) error {
	// 1. 获取 access_token
	token, err := GetDingtalkAccessToken(appKey, appSecret)
	if err != nil {
		return fmt.Errorf(`获取 access_token 失败: %w`, err)
	}
	// 2. 调用单聊发送 API
	msgParam := map[string]string{`content`: text}
	msgParamBytes, _ := json.Marshal(msgParam)
	body := map[string]any{
		`robotCode`: robotCode,
		`userIds`:   []string{userId},
		`msgKey`:    `sampleText`,
		`msgParam`:  string(msgParamBytes),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf(`json marshal: %w`, err)
	}
	url := dingtalkApiBase + `/v1.0/robot/oToO/batchSend`
	req, _ := http.NewRequest(`POST`, url, strings.NewReader(string(bodyBytes)))
	req.Header.Set(`Content-Type`, `application/json; charset=utf-8`)
	req.Header.Set(`x-acs-dingtalk-access-token`, token)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(`http post: %w`, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf(`response parse: %w`, err)
	}
	// 检查是否有错误
	if errMsg, ok := result[`message`]; ok && cast.ToString(errMsg) != `ok` {
		return fmt.Errorf(`dingtalk send error: %v`, errMsg)
	}
	return nil
}

// GetDingtalkAppAdmins 通过钉钉 Open API 获取应用管理员列表（返回 userId 列表）。
// 用于测试场景，确定测试消息的接收者。
func GetDingtalkAppAdmins(accessToken, appKey string) ([]string, error) {
	url := dingtalkApiBase + `/v1.0/app/internalApps/` + appKey + `/admins`
	req, _ := http.NewRequest(`GET`, url, nil)
	req.Header.Set(`x-acs-dingtalk-access-token`, accessToken)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(`http get: %w`, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf(`response parse: %w`, err)
	}
	// 解析管理员列表
	adminList, ok := result[`adminList`]
	if !ok {
		return nil, fmt.Errorf(`未找到 adminList 字段: %v`, result)
	}
	adminsRaw, ok := adminList.([]any)
	if !ok {
		return nil, fmt.Errorf(`adminList 格式错误: %v`, adminList)
	}
	userIds := make([]string, 0, len(adminsRaw))
	for _, admin := range adminsRaw {
		adminMap, ok := admin.(map[string]any)
		if !ok {
			continue
		}
		userId := cast.ToString(adminMap[`userId`])
		if userId != `` {
			userIds = append(userIds, userId)
		}
	}
	if len(userIds) == 0 {
		return nil, fmt.Errorf(`管理员列表为空`)
	}
	return userIds, nil
}
