package business

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"

	"github.com/spf13/cast"
)

// GetWebhookConfigByAgentCliId 通过 agent_cli_id 获取关联的 webhook 配置，未配置返回 nil。
func GetWebhookConfigByAgentCliId(agentCliId int) *define.WebhookConfigItem {
	if agentCliId <= 0 {
		return nil
	}
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT webhook_config_id FROM tbl_agent_cli WHERE id = ?`, agentCliId,
	).One()
	if err != nil || len(row) == 0 {
		return nil
	}
	webhookConfigId := cast.ToInt(row["webhook_config_id"])
	if webhookConfigId <= 0 {
		return nil
	}
	wRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_webhook_config WHERE id = ?`, webhookConfigId,
	).One()
	if err != nil || len(wRow) == 0 {
		return nil
	}
	return &define.WebhookConfigItem{
		Id:         cast.ToInt(wRow["id"]),
		Name:       cast.ToString(wRow["name"]),
		Type:       cast.ToString(wRow["type"]),
		WebhookUrl: cast.ToString(wRow["webhook_url"]),
		Secret:     cast.ToString(wRow["secret"]),
	}
}

// SendWebhookNotify 根据 webhook 配置发送通知，目前支持钉钉。
func SendWebhookNotify(config *define.WebhookConfigItem, content string) error {
	if config == nil || config.WebhookUrl == "" {
		return nil
	}
	switch config.Type {
	case define.WebhookTypeDingtalk:
		return sendDingtalkNotify(config.WebhookUrl, config.Secret, content)
	default:
		log.Printf("[webhook-notify] 暂不支持的类型: %s", config.Type)
		return nil
	}
}

// sendDingtalkNotify 发送钉钉文本消息，支持 HMAC-SHA256 加签。
func sendDingtalkNotify(webhookUrl, secret, content string) error {
	url := strings.TrimSpace(webhookUrl)

	if secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		signStr := timestamp + "\n" + secret
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signStr))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		url = fmt.Sprintf("%s&timestamp=%s&sign=%s", url, timestamp, sign)
	}

	body := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json; charset=utf-8", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("response parse: %w", err)
	}
	if cast.ToInt(result["errcode"]) != 0 {
		return fmt.Errorf("dingtalk error: errcode=%v errmsg=%v", result["errcode"], result["errmsg"])
	}
	return nil
}
