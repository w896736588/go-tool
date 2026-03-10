package common

import (
	"bytes"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// InfoCrawlChatByModel 使用模型发起一次 AI 请求。
func (h *CSqlite) InfoCrawlChatByModel(modelID int, systemPrompt, userPrompt string) (string, map[string]any, error) {
	modelInfo, err := h.InfoCrawlAiModelInfo(modelID)
	if err != nil {
		return ``, nil, err
	}
	if strings.ToLower(cast.ToString(modelInfo[`provider_type`])) != `openai` {
		return ``, nil, errors.New(`当前仅支持 openai 兼容服务商`)
	}
	baseURL := strings.TrimSpace(cast.ToString(modelInfo[`base_url`]))
	if baseURL == `` {
		return ``, nil, errors.New(`AI 服务商 base_url 不能为空`)
	}
	if !strings.Contains(baseURL, `/chat/completions`) {
		baseURL = strings.TrimRight(baseURL, `/`) + `/v1/chat/completions`
	}
	apiKey := strings.TrimSpace(cast.ToString(modelInfo[`api_key`]))
	if apiKey == `` {
		return ``, nil, errors.New(`AI 服务商 api_key 不能为空`)
	}
	bodyMap := map[string]any{
		`model`: cast.ToString(modelInfo[`model`]),
		`messages`: []map[string]string{
			{
				`role`:    `system`,
				`content`: systemPrompt,
			},
			{
				`role`:    `user`,
				`content`: userPrompt,
			},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: 120 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return ``, nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ``, nil, err
	}
	if response.StatusCode >= 300 {
		return ``, nil, errors.New(`AI 请求失败: ` + string(responseBody))
	}
	content := p_common.ExtractOpenAiMessage(string(responseBody))
	if strings.TrimSpace(content) == `` {
		content = string(responseBody)
	}
	return strings.TrimSpace(content), modelInfo, nil
}

// InfoCrawlPlannerSystemPrompt 返回 AI 规划系统提示词。
func (h *CSqlite) InfoCrawlPlannerSystemPrompt() string {
	return "你是一个网页抓取规划助手，只负责为 Playwright 生成结构化抓取计划。\n" +
		"要求：\n" +
		"1. 只输出合法 JSON，不要输出 markdown，不要输出解释。\n" +
		"2. 每个网页最多输出 8 个动作。\n" +
		"3. 只允许使用 wait、click、exist_wait、no_exist_wait、text_content、bool_result。\n" +
		"4. 不允许输出 input、goto、evaluate、press、hover 等未授权动作。\n" +
		"5. locator 必须尽量简短稳定。\n" +
		"6. 如无法确定区域，可直接抓取 body 的 text_content。\n" +
		"输出格式：{\"pages\":[{\"task_page_id\":1,\"goal\":\"一句话说明抓取目标\",\"actions\":[{\"type\":\"wait\",\"locator\":\"\",\"value\":\"1500\",\"out_key\":\"\",\"tip\":\"等待页面稳定\"}]}]}"
}

// InfoCrawlBuildPlannerUserPrompt 构建 AI 规划用户提示词。
func (h *CSqlite) InfoCrawlBuildPlannerUserPrompt(taskInfo map[string]any, pageList []map[string]any) string {
	builder := strings.Builder{}
	builder.WriteString("任务名称：")
	builder.WriteString(cast.ToString(taskInfo[`name`]))
	builder.WriteString("\n\n任务目标：")
	builder.WriteString(cast.ToString(taskInfo[`prompt`]))
	builder.WriteString("\n\n网页列表：\n")
	for _, page := range pageList {
		builder.WriteString("网页ID: ")
		builder.WriteString(cast.ToString(page[`id`]))
		builder.WriteString("\n网页名称: ")
		builder.WriteString(cast.ToString(page[`name`]))
		builder.WriteString("\nURL: ")
		builder.WriteString(cast.ToString(page[`url`]))
		builder.WriteString("\n网页说明: ")
		builder.WriteString(cast.ToString(page[`note`]))
		builder.WriteString("\n\n")
	}
	builder.WriteString("请为每个网页生成抓取计划。")
	return builder.String()
}

// InfoCrawlParsePlannerResult 解析规划结果。
func (h *CSqlite) InfoCrawlParsePlannerResult(content string) (_struct.InfoCrawlPlannerResult, error) {
	result := _struct.InfoCrawlPlannerResult{}
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}
	err := gstool.JsonDecode(content, &result)
	return result, err
}

// InfoCrawlSummarySystemPrompt 返回 AI 汇总系统提示词。
func (h *CSqlite) InfoCrawlSummarySystemPrompt() string {
	return "你是一个信息整理助手，请根据多个网页抓取结果输出中文汇总结论。\n" +
		"要求：1. 先给整体摘要。2. 再按网页来源分别总结。3. 明确标注不确定信息。4. 不要编造未抓取到的内容。"
}

// InfoCrawlBuildSummaryUserPrompt 构建汇总提示词。
func (h *CSqlite) InfoCrawlBuildSummaryUserPrompt(taskInfo map[string]any, runTime string, runPageList []map[string]any) string {
	builder := strings.Builder{}
	builder.WriteString("任务名称：")
	builder.WriteString(cast.ToString(taskInfo[`name`]))
	builder.WriteString("\n\n任务提示词：\n")
	builder.WriteString(cast.ToString(taskInfo[`prompt`]))
	builder.WriteString("\n\n执行时间：")
	builder.WriteString(runTime)
	builder.WriteString("\n\n网页抓取结果：\n")
	totalLength := 0
	for _, page := range runPageList {
		builder.WriteString("网页：")
		builder.WriteString(cast.ToString(page[`page_name`]))
		builder.WriteString("\nURL: ")
		builder.WriteString(cast.ToString(page[`url`]))
		builder.WriteString("\n状态: ")
		builder.WriteString(cast.ToString(page[`status`]))
		builder.WriteString("\n")
		if cast.ToString(page[`error_message`]) != `` {
			builder.WriteString("错误: ")
			builder.WriteString(cast.ToString(page[`error_message`]))
			builder.WriteString("\n")
		}
		pageText := cast.ToString(page[`raw_text`])
		if len(pageText) > define.InfoCrawlPageTextMaxLength {
			pageText = pageText[:define.InfoCrawlPageTextMaxLength] + "\n[内容已截断]"
		}
		if totalLength+len(pageText) > define.InfoCrawlSummaryInputMaxLength {
			remain := define.InfoCrawlSummaryInputMaxLength - totalLength
			if remain > 0 {
				pageText = pageText[:remain] + "\n[内容已截断]"
			} else {
				pageText = `[内容已截断]`
			}
		}
		builder.WriteString("内容：\n")
		builder.WriteString(pageText)
		builder.WriteString("\n\n")
		totalLength += len(pageText)
		if totalLength >= define.InfoCrawlSummaryInputMaxLength {
			break
		}
	}
	builder.WriteString("请输出：\n1. 整体摘要\n2. 各网页核心信息\n3. 需要重点关注的变化\n4. 不确定或疑似噪音的信息")
	return builder.String()
}
