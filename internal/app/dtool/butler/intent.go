package butler

import (
	"dev_tool/internal/app/dtool/common"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

// IntentResult 意图分析结果。
type IntentResult struct {
	Clear     bool     // 意图是否清晰
	Topic     string   // 主题关键词
	NewTopic  bool     // 是否为新话题（与最近对话主题不同）
	Questions []string // 澄清问题（意图不清晰时，2-3 个）
}

// intentAnalysisSystemPrompt 意图分析器的 system prompt，要求 AI 以 JSON 格式返回分析结果。
const intentAnalysisSystemPrompt = `你是意图分析器。分析用户消息的意图清晰度和话题归属，以 JSON 格式返回。

返回格式（严格遵循）：
{"clear": true/false, "topic": "主题关键词", "new_topic": true/false, "questions": ["问题1", "问题2"]}

规则：
- clear：用户意图是否明确（清楚要做什么则为 true）
- topic：用 2-5 个字概括消息主题
- new_topic：是否与最近对话主题不同（开启新话题）
- questions：如果意图模糊，列出 2-3 个简短澄清问题；意图清晰时为空数组

只返回 JSON，不要添加任何额外文字。`

// AnalyzeIntent 使用轻量 AI 模型分析用户消息的意图。
// modelId 优先使用 fc_model_id（轻量模型），为 0 时回落 model_id。
// recentTopic 为最近对话的主题关键词（空表示无历史）。
func AnalyzeIntent(db *common.CSqlite, modelId int, userMessage string, recentTopic string) (*IntentResult, error) {
	if modelId <= 0 {
		// 模型未配置，默认意图清晰，跳过分析
		gstool.FmtPrintlnLogTime(`[butler-intent] 意图分析模型未配置 model_id=%d，跳过分析`, modelId)
		return &IntentResult{Clear: true, Topic: ``, NewTopic: recentTopic == ``}, nil
	}
	// 构建用户提示：包含消息和最近主题信息
	userPrompt := buildIntentUserPrompt(userMessage, recentTopic)
	// 调用 AI（非流式，意图分析需要完整 JSON 结果）
	rawReply, _, err := db.AIChatByModel(modelId, intentAnalysisSystemPrompt, userPrompt)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-intent] AI 请求失败 %s，默认意图清晰`, err.Error())
		return &IntentResult{Clear: true, Topic: ``, NewTopic: false}, nil
	}
	// 解析 JSON 结果
	result := parseIntentJson(rawReply)
	if result == nil {
		gstool.FmtPrintlnLogTime(`[butler-intent] JSON 解析失败 raw=%s，默认意图清晰`, truncateForLog(rawReply, 200))
		return &IntentResult{Clear: true, Topic: ``, NewTopic: false}, nil
	}
	gstool.FmtPrintlnLogTime(`[butler-intent] 分析结果 clear=%v topic=%s new_topic=%v`, result.Clear, result.Topic, result.NewTopic)
	return result, nil
}

// buildIntentUserPrompt 构建意图分析的用户提示。
func buildIntentUserPrompt(userMessage string, recentTopic string) string {
	if recentTopic == `` {
		return fmt.Sprintf(`用户消息：%s\n（这是新对话，无历史主题）`, userMessage)
	}
	return fmt.Sprintf(`最近对话主题：%s\n用户消息：%s`, recentTopic, userMessage)
}

// parseIntentJson 从 AI 回复中提取 JSON 并解析为 IntentResult。
func parseIntentJson(raw string) *IntentResult {
	// 尝试直接解析
	raw = strings.TrimSpace(raw)
	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		// 尝试从文本中提取 JSON（AI 可能包裹了额外文字）
		jsonStr := extractJsonBlock(raw)
		if jsonStr == `` {
			return nil
		}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return nil
		}
	}
	result := &IntentResult{
		Clear:     castToBool(data[`clear`]),
		Topic:     castToString(data[`topic`]),
		NewTopic:  castToBool(data[`new_topic`]),
		Questions: castToStringSlice(data[`questions`]),
	}
	return result
}

// extractJsonBlock 从可能包含额外文字的 AI 回复中提取 JSON 块。
func extractJsonBlock(raw string) string {
	// 找到第一个 { 和最后一个 }
	start := strings.Index(raw, `{`)
	end := strings.LastIndex(raw, `}`)
	if start < 0 || end < 0 || end <= start {
		return ``
	}
	return raw[start : end+1]
}

// castToBool 将 JSON 值转为 bool。
func castToBool(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.ToLower(val) == `true`
	case float64:
		return val != 0
	case int:
		return val != 0
	default:
		return false
	}
}

// castToString 将 JSON 值转为 string。
func castToString(v any) string {
	if v == nil {
		return ``
	}
	return fmt.Sprintf(`%v`, v)
}

// castToStringSlice 将 JSON 值转为 []string。
func castToStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	slice, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		result = append(result, fmt.Sprintf(`%v`, item))
	}
	return result
}

// truncateForLog 截断字符串用于日志输出。
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + `...`
}
