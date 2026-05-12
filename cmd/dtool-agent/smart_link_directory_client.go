package main

import (
	"bytes"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type agentSmartLinkDirectoryStore struct {
	serverURL string
	safeToken string
	client    *http.Client
}

// newAgentSmartLinkDirectoryStore 创建 agent 侧的固定目录映射存储适配器。
// agent 不直连配置库，所有 tbl_smart_link_directory_mapping 读写都通过服务端接口完成。
func newAgentSmartLinkDirectoryStore(serverURL, safeToken string) *agentSmartLinkDirectoryStore {
	return &agentSmartLinkDirectoryStore{
		serverURL: strings.TrimRight(serverURL, "/"),
		safeToken: strings.TrimSpace(safeToken),
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

// GetByMappingKey 根据 mapping_key 查询固定目录索引。
func (h *agentSmartLinkDirectoryStore) GetByMappingKey(mappingKey string) (int, error) {
	resp, err := h.do(define.AgentSmartLinkDirectoryRequest{
		Action:     define.AgentSmartLinkDirectoryActionGetByMappingKey,
		MappingKey: mappingKey,
	})
	if err != nil {
		return 0, err
	}
	return resp.UserDataIndex, nil
}

// ExistsUserDataIndex 判断指定目录索引是否已被固定映射占用。
func (h *agentSmartLinkDirectoryStore) ExistsUserDataIndex(userDataIndex int) (bool, error) {
	resp, err := h.do(define.AgentSmartLinkDirectoryRequest{
		Action:        define.AgentSmartLinkDirectoryActionExistsIndex,
		UserDataIndex: userDataIndex,
	})
	if err != nil {
		return false, err
	}
	return resp.Exists, nil
}

// UpsertMapping 写入或更新固定目录映射关系。
func (h *agentSmartLinkDirectoryStore) UpsertMapping(mappingKey string, smartLinkID int, label, accountKey string, userDataIndex int) error {
	_, err := h.do(define.AgentSmartLinkDirectoryRequest{
		Action:        define.AgentSmartLinkDirectoryActionUpsert,
		MappingKey:    mappingKey,
		SmartLinkID:   smartLinkID,
		Label:         label,
		AccountKey:    accountKey,
		UserDataIndex: userDataIndex,
	})
	return err
}

// do 统一调用服务端代理接口，并解析项目标准响应结构。
func (h *agentSmartLinkDirectoryStore) do(payload define.AgentSmartLinkDirectoryRequest) (define.AgentSmartLinkDirectoryResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return define.AgentSmartLinkDirectoryResponse{}, err
	}
	request, err := http.NewRequest(http.MethodPost, h.serverURL+"/api/smart-link/agent/directory-mapping", bytes.NewReader(body))
	if err != nil {
		return define.AgentSmartLinkDirectoryResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	if h.safeToken != "" {
		request.Header.Set("Token", h.safeToken)
	}
	response, err := h.client.Do(request)
	if err != nil {
		return define.AgentSmartLinkDirectoryResponse{}, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return define.AgentSmartLinkDirectoryResponse{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return define.AgentSmartLinkDirectoryResponse{}, fmt.Errorf("directory-mapping status=%d body=%s", response.StatusCode, string(responseBody))
	}
	result := struct {
		ErrCode int                                    `json:"ErrCode"`
		ErrMsg  string                                 `json:"ErrMsg"`
		Data    define.AgentSmartLinkDirectoryResponse `json:"Data"`
	}{}
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return define.AgentSmartLinkDirectoryResponse{}, err
	}
	if result.ErrCode != 0 {
		if result.ErrMsg == "" {
			result.ErrMsg = "directory-mapping接口返回失败"
		}
		return define.AgentSmartLinkDirectoryResponse{}, errors.New(result.ErrMsg)
	}
	return result.Data, nil
}
