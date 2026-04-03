package common

import (
	"bufio"
	"bytes"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// AIChatByModel 使用模型发起一次 AI 请求。
func (h *CSqlite) AIChatByModel(modelID int, systemPrompt, userPrompt string) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`: cast.ToString(modelInfo[`model`]),
		`messages`: []map[string]string{
			{`role`: `system`, `content`: systemPrompt},
			{`role`: `user`, `content`: userPrompt},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
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

// AIChatStreamByModel 使用模型发起流式 AI 请求。
func (h *CSqlite) AIChatStreamByModel(modelID int, systemPrompt, userPrompt string, onChunk func(string)) (string, map[string]any, error) {
	modelInfo, requestURL, apiKey, err := h.aiChatBuildRequest(modelID)
	if err != nil {
		return ``, nil, err
	}
	bodyMap := map[string]any{
		`model`:  cast.ToString(modelInfo[`model`]),
		`stream`: true,
		`messages`: []map[string]string{
			{`role`: `system`, `content`: systemPrompt},
			{`role`: `user`, `content`: userPrompt},
		},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return ``, nil, err
	}
	request.Header.Set(`Authorization`, `Bearer `+apiKey)
	request.Header.Set(`Content-Type`, `application/json`)
	client := &http.Client{Timeout: 10 * time.Minute}
	response, err := client.Do(request)
	if err != nil {
		return ``, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(response.Body)
		return ``, nil, errors.New(`AI 请求失败: ` + string(responseBody))
	}
	reader := bufio.NewReader(response.Body)
	contentBuilder := strings.Builder{}
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `data:`) {
			payload := strings.TrimSpace(strings.TrimPrefix(line, `data:`))
			if payload == `[DONE]` {
				break
			}
			chunk := h.aiChatExtractStreamContent(payload)
			if chunk != `` {
				contentBuilder.WriteString(chunk)
				if onChunk != nil {
					onChunk(chunk)
				}
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return strings.TrimSpace(contentBuilder.String()), modelInfo, readErr
		}
	}
	return strings.TrimSpace(contentBuilder.String()), modelInfo, nil
}

// AiModelInfo 查询 AI 模型配置。
func (h *CSqlite) AiModelInfo(id int) (map[string]any, error) {
	info, err := h.Client.QueryBySql(`
select m.*,p.name as provider_name,p.provider_type,p.base_url,p.api_key
from tbl_ai_model m
left join tbl_ai_provider p on p.id = m.provider_id
where m.id = ? and m.status = 1 and p.status = 1`, id).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`AI模型不存在或已停用`)
	}
	return info, nil
}

func (h *CSqlite) aiChatBuildRequest(modelID int) (map[string]any, string, string, error) {
	modelInfo, err := h.AiModelInfo(modelID)
	if err != nil {
		return nil, ``, ``, err
	}
	if strings.ToLower(cast.ToString(modelInfo[`provider_type`])) != `openai` {
		return nil, ``, ``, errors.New(`当前仅支持 openai 兼容服务商`)
	}
	baseURL := strings.TrimSpace(cast.ToString(modelInfo[`base_url`]))
	if baseURL == `` {
		return nil, ``, ``, errors.New(`AI 服务商 base_url 不能为空`)
	}
	requestURI := strings.TrimSpace(cast.ToString(modelInfo[`uri`]))
	if requestURI == `` {
		requestURI = `/v1/chat/completions`
	}
	apiKey := strings.TrimSpace(cast.ToString(modelInfo[`api_key`]))
	if apiKey == `` {
		return nil, ``, ``, errors.New(`AI 服务商 api_key 不能为空`)
	}
	return modelInfo, joinAIRequestURL(baseURL, requestURI), apiKey, nil
}

func (h *CSqlite) aiChatExtractStreamContent(payload string) string {
	if strings.TrimSpace(payload) == `` {
		return ``
	}
	dataMap := make(map[string]any)
	if err := json.Unmarshal([]byte(payload), &dataMap); err != nil {
		return ``
	}
	choiceList, ok := dataMap[`choices`].([]any)
	if !ok || len(choiceList) == 0 {
		return ``
	}
	choiceMap, ok := choiceList[0].(map[string]any)
	if !ok {
		return ``
	}
	if deltaMap, ok := choiceMap[`delta`].(map[string]any); ok {
		if chunk := cast.ToString(deltaMap[`content`]); chunk != `` {
			return chunk
		}
	}
	if messageMap, ok := choiceMap[`message`].(map[string]any); ok {
		if chunk := cast.ToString(messageMap[`content`]); chunk != `` {
			return chunk
		}
	}
	return ``
}

func joinAIRequestURL(baseURL, requestURI string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), `/`)
	requestURI = strings.TrimSpace(requestURI)
	if requestURI == `` {
		return baseURL
	}
	if !strings.HasPrefix(requestURI, `/`) {
		requestURI = `/` + requestURI
	}
	return baseURL + requestURI
}
