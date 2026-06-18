package worker

import (
	"dev_tool/internal/app/dtool/common"
	"fmt"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// fcMaxLoopIterations FC 循环最大迭代次数，防止无限循环。
const fcMaxLoopIterations = 10

// FCLoopResult FC 循环执行结果。
type FCLoopResult struct {
	Content  string   // 最终 AI 回复文本
	Success  bool     // 任务是否成功完成
	ToolUsed []string // 使用过的工具名称列表
}

// RunFCLoop 执行 Function Calling 循环。
// 反复调用 AI，执行工具调用，直到 AI 返回最终文本回复（不再调用工具）或达到最大迭代次数。
// 使用 fcModelId（Function Calling 用模型），为 0 时回落 modelId。
func RunFCLoop(db *common.CSqlite, modelId int, systemPrompt string, historyMessages []map[string]string, userMessage string) *FCLoopResult {
	// 构建 messages 列表
	messages := buildFCMessages(systemPrompt, historyMessages, userMessage)
	// 获取工具定义
	tools := ToolDefinitions()
	toolsUsed := make([]string, 0)

	for i := 0; i < fcMaxLoopIterations; i++ {
		// 调用 AI（非流式，需解析完整 tool_calls）
		content, toolCalls, _, err := db.AIChatByModelWithTools(modelId, messages, tools)
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler-fc] AI 请求失败 %s`, err.Error())
			return &FCLoopResult{Content: fmt.Sprintf(`任务执行失败：%s`, err.Error()), Success: false}
		}

		// 没有工具调用 → AI 已给出最终回复
		if len(toolCalls) == 0 {
			return &FCLoopResult{Content: content, Success: true, ToolUsed: toolsUsed}
		}

		// 记录 assistant 消息（含 tool_calls）
		assistantMsg := map[string]any{
			`role`:       `assistant`,
			`content`:    content,
			`tool_calls`: toolCalls,
		}
		messages = append(messages, assistantMsg)

		// 逐个执行工具调用
		for _, tc := range toolCalls {
			tcMap, ok := tc.(map[string]any)
			if !ok {
				continue
			}
			callID := cast.ToString(tcMap[`id`])
			fnMap, _ := tcMap[`function`].(map[string]any)
			fnName := cast.ToString(fnMap[`name`])
			fnArgs := cast.ToString(fnMap[`arguments`])

			gstool.FmtPrintlnLogTime(`[butler-fc] 执行工具 %s(%s)`, fnName, truncateForLog(fnArgs, 100))
			result := ExecuteTool(fnName, fnArgs)
			toolsUsed = append(toolsUsed, fnName)
			gstool.FmtPrintlnLogTime(`[butler-fc] 工具结果 %s → %s`, fnName, truncateForLog(result, 200))

			// 添加工具结果消息
			messages = append(messages, map[string]any{
				`role`:         `tool`,
				`tool_call_id`: callID,
				`content`:      result,
			})
		}
	}

	// 超过最大迭代次数
	gstool.FmtPrintlnLogTime(`[butler-fc] FC 循环超过最大迭代次数 %d`, fcMaxLoopIterations)
	return &FCLoopResult{Content: `任务执行超时：工具调用次数过多`, Success: false, ToolUsed: toolsUsed}
}

// buildFCMessages 构建 FC 循环的初始 messages 列表。
func buildFCMessages(systemPrompt string, historyMessages []map[string]string, userMessage string) []map[string]any {
	messages := make([]map[string]any, 0, len(historyMessages)+2)
	// system prompt
	messages = append(messages, map[string]any{
		`role`:    `system`,
		`content`: systemPrompt,
	})
	// 历史消息
	for _, msg := range historyMessages {
		messages = append(messages, map[string]any{
			`role`:    msg[`role`],
			`content`: msg[`content`],
		})
	}
	// 当前用户消息
	messages = append(messages, map[string]any{
		`role`:    `user`,
		`content`: userMessage,
	})
	return messages
}

// truncateForLog 截断字符串用于日志输出。
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + `...`
}
