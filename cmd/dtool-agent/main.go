package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

var defaultServerURL = "http://localhost:17170"

// Config 客户端配置
type Config struct {
	ServerURL         string
	ClientID          string
	ClientVersion     string
	HeartbeatInterval time.Duration
	TaskPollInterval  time.Duration
}

// Agent 本地客户端代理
type Agent struct {
	config     Config
	httpClient *http.Client
	stopChan   chan bool
}

// NewAgent 创建新代理
func NewAgent(cfg Config) *Agent {
	return &Agent{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		stopChan:   make(chan bool),
	}
}

// Register 向服务端注册
func (a *Agent) Register() error {
	data := map[string]any{
		"client_id":      a.config.ClientID,
		"client_version": a.config.ClientVersion,
		"hostname":       getHostname(),
		"os":             runtime.GOOS,
		"arch":           runtime.GOARCH,
		"user_name":      getUsername(),
		"token":          "", // 第一版不强制鉴权
	}

	resp, err := a.post("/api/agent/register", data)
	if err != nil {
		return fmt.Errorf("请求注册接口失败: %w", err)
	}

	errCode := 0
	if v, ok := resp["ErrCode"]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			errCode = int(val)
		case int:
			errCode = val
		}
	}
	if errCode != 0 {
		errMsg := "未知错误"
		if resp["ErrMsg"] != nil {
			errMsg = fmt.Sprintf("%v", resp["ErrMsg"])
		}
		return fmt.Errorf("服务端返回错误: %s (错误码: %d)", errMsg, errCode)
	}

	fmt.Printf("注册成功，服务端要求版本: %v\n", resp["Data"].(map[string]any)["required_client_version"])
	return nil
}

// Heartbeat 发送心跳
func (a *Agent) Heartbeat() error {
	data := map[string]any{
		"client_id":       a.config.ClientID,
		"client_version":  a.config.ClientVersion,
		"status":          "online",
		"hostname":        getHostname(),
		"current_task_id": "",
	}

	_, err := a.post("/api/agent/heartbeat", data)
	return err
}

// PollTask 轮询任务
func (a *Agent) PollTask() error {
	url := fmt.Sprintf("%s/api/agent/task/pull?client_id=%s", a.config.ServerURL, a.config.ClientID)
	resp, err := a.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result["ErrCode"] != 0 {
		return fmt.Errorf("拉取任务失败: %v", result["ErrMsg"])
	}

	data := result["Data"]
	if data == nil {
		return nil // 没有任务
	}

	taskData := data.(map[string]any)
	taskID := taskData["task_id"].(string)

	fmt.Printf("收到任务: %s\n", taskID)

	// 执行任务
	go a.executeTask(taskID, taskData)
	return nil
}

// executeTask 执行任务
func (a *Agent) executeTask(taskID string, taskData map[string]any) {
	fmt.Printf("开始执行任务: %s\n", taskID)

	// 上报任务开始
	a.reportTask(taskID, "running", "", "")

	// 检查 Node.js 和 Playwright 环境
	if !a.checkRuntime() {
		a.reportTask(taskID, "failed", "", "Node.js 或 Playwright 环境检查失败")
		return
	}

	// 解析运行参数
	runParams := taskData["run_params"].(string)

	// 执行任务（调用 Playwright）
	logOutput, err := a.runPlaywright(runParams)

	if err != nil {
		a.reportTask(taskID, "failed", logOutput, err.Error())
	} else {
		a.reportTask(taskID, "success", logOutput, "")
	}
}

// checkRuntime 检查运行环境
func (a *Agent) checkRuntime() bool {
	// 检查 Node.js
	cmd := exec.Command("node", "--version")
	if err := cmd.Run(); err != nil {
		fmt.Println("Node.js 未安装")
		return false
	}
	return true
}

// runPlaywright 运行 Playwright
func (a *Agent) runPlaywright(runParams string) (string, error) {
	// 这里简化处理，实际应该调用 Playwright 执行
	// 第一版可以先输出日志
	output := fmt.Sprintf("执行参数: %s\n", runParams)

	// TODO: 实现真正的 Playwright 调用
	// 可以使用 exec.Command 调用 node 脚本

	return output, nil
}

// reportTask 上报任务结果
func (a *Agent) reportTask(taskID, status, logText, errorMsg string) error {
	data := map[string]any{
		"task_id":        taskID,
		"status":         status,
		"log_append":     logText,
		"result_payload": "{}",
		"error_message":  errorMsg,
	}

	_, err := a.post("/api/agent/task/report", data)
	return err
}

// post 发送 POST 请求
func (a *Agent) post(path string, data map[string]any) (map[string]any, error) {
	jsonData, _ := json.Marshal(data)
	url := a.config.ServerURL + path

	resp, err := a.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Run 启动代理
func (a *Agent) Run() {
	fmt.Printf("dtool-agent 启动\n")
	fmt.Printf("服务端地址: %s\n", a.config.ServerURL)
	fmt.Printf("客户端ID: %s\n", a.config.ClientID)
	fmt.Printf("版本: %s\n", a.config.ClientVersion)

	// 注册
	if err := a.Register(); err != nil {
		gstool.FmtPrintlnLogTime(`注册失败 %s`, err.Error())
		return
	}

	// 启动心跳和任务轮询
	ticker := time.NewTicker(a.config.HeartbeatInterval)
	taskTicker := time.NewTicker(a.config.TaskPollInterval)

	defer ticker.Stop()
	defer taskTicker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.Heartbeat(); err != nil {
				fmt.Printf("心跳失败: %v\n", err)
			}
		case <-taskTicker.C:
			if err := a.PollTask(); err != nil {
				fmt.Printf("拉取任务失败: %v\n", err)
			}
		case <-a.stopChan:
			return
		}
	}
}

// Stop 停止代理
func (a *Agent) Stop() {
	close(a.stopChan)
}

// 辅助函数
func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getUsername() string {
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	return username
}

func main() {
	// 从环境变量或命令行参数读取配置
	serverURL := os.Getenv("DTOOL_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	clientID := os.Getenv("DTOOL_CLIENT_ID")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%s_%d", getHostname(), time.Now().Unix())
	}

	clientVersion := os.Getenv("DTOOL_CLIENT_VERSION")
	if clientVersion == "" {
		clientVersion = "1.0.0"
	}

	config := Config{
		ServerURL:         serverURL,
		ClientID:          clientID,
		ClientVersion:     clientVersion,
		HeartbeatInterval: 5 * time.Second,
		TaskPollInterval:  3 * time.Second,
	}

	agent := NewAgent(config)
	agent.Run()
}
