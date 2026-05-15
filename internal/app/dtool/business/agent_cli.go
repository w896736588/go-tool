package business

import (
	"dev_tool/internal/app/dtool/common"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cast"
)

// ReadAgentCliSettings 读取 settings.json 文件内容，若文件不存在返回空
func ReadAgentCliSettings(settingsPath string) (exists bool, content string, err error) {
	data, readErr := os.ReadFile(settingsPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("读取配置文件失败 %s: %w", settingsPath, readErr)
	}
	return true, string(data), nil
}

// WriteMcpServersToSettings 从 DB 读取全部 ChromeDevtools 端口写入 settings.json 的 mcpServers
func WriteMcpServersToSettings(settingsPath string) error {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_chrome_devtools_config ORDER BY id`,
	).All()
	if err != nil {
		return fmt.Errorf("查询端口配置失败: %w", err)
	}

	// 读取已有配置
	configData := make(map[string]any)
	content, readErr := os.ReadFile(settingsPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &configData); err != nil {
			return fmt.Errorf("解析配置文件失败 %s: %w", settingsPath, err)
		}
	}

	// 构建 mcpServers
	mcpServers := make(map[string]any)
	for i, row := range rows {
		port := cast.ToInt(row["port"])
		key := fmt.Sprintf("devtool_%d", i)
		mcpServers[key] = map[string]any{
			"command": "npx",
			"args": []string{
				"chrome-devtools-mcp@latest",
				fmt.Sprintf("--browser-url=http://127.0.0.1:%d", port),
			},
		}
	}
	configData["mcpServers"] = mcpServers

	// 确保目录存在
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	// 写回
	newContent, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	newContent = append(newContent, '\n')

	if err := os.WriteFile(settingsPath, newContent, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %w", settingsPath, err)
	}
	return nil
}

// WriteDeepSeekToSettings 写入 DeepSeek 配置到 settings.json
func WriteDeepSeekToSettings(settingsPath string, modelName string, apiKey string, baseUrl string) error {
	// 读取已有配置
	configData := make(map[string]any)
	content, readErr := os.ReadFile(settingsPath)
	if readErr == nil && len(content) > 0 {
		if err := json.Unmarshal(content, &configData); err != nil {
			return fmt.Errorf("解析配置文件失败 %s: %w", settingsPath, err)
		}
	}

	if baseUrl == "" {
		baseUrl = "https://api.deepseek.com/anthropic"
	}

	// 设置 env
	configData["env"] = map[string]string{
		"ANTHROPIC_BASE_URL":                       baseUrl,
		"API_TIMEOUT_MS":                           "3000000",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":            modelName,
		"ANTHROPIC_DEFAULT_SONNET_MODEL":           modelName,
		"ANTHROPIC_DEFAULT_OPUS_MODEL":             modelName,
		"ANTHROPIC_AUTH_TOKEN":                     apiKey,
		"ANTHROPIC_API_KEY":                        "",
		"ANTHROPIC_REASONING_MODEL":                modelName,
	}

	// 设置 model
	configData["model"] = modelName

	// 设置 enabledPlugins
	configData["enabledPlugins"] = map[string]bool{
		"skill-creator@claude-plugins-official": true,
	}

	// 设置 extraKnownMarketplaces
	configData["extraKnownMarketplaces"] = map[string]any{
		"anthropic-agent-skills": map[string]any{
			"source": map[string]any{
				"source": "git",
				"url":    "https://github.com/anthropics/skills.git",
			},
		},
	}

	// 确保目录存在
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败 %s: %w", dir, err)
	}

	// 写回
	newContent, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	newContent = append(newContent, '\n')

	if err := os.WriteFile(settingsPath, newContent, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %w", settingsPath, err)
	}
	return nil
}

// GetAgentCliSettingsSummary 从 settings.json 内容中提取摘要信息
func GetAgentCliSettingsSummary(content string) (model string, mcpCount int) {
	if content == "" {
		return "", 0
	}
	var configData map[string]any
	if err := json.Unmarshal([]byte(content), &configData); err != nil {
		return "", 0
	}
	if m, ok := configData["model"]; ok {
		model = cast.ToString(m)
	}
	if servers, ok := configData["mcpServers"]; ok {
		if serverMap, ok := servers.(map[string]any); ok {
			mcpCount = len(serverMap)
		}
	}
	return
}
