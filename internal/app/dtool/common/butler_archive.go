package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// CreateArchiveRecord 主管家任务完成后提交归档记录。
func (c *CSqlite) CreateArchiveRecord(configId, taskId int, sessionId string, files []string, conversation string) (int, error) {
	// 读取各文件的内容拼入对话记录
	var sb strings.Builder
	sb.WriteString(conversation)
	sb.WriteString("\n\n=== 产生的文件内容 ===\n")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			sb.WriteString(fmt.Sprintf("[%s] 读取失败: %s\n", f, err.Error()))
		} else {
			sb.WriteString(fmt.Sprintf("[%s]\n%s\n", f, string(data)))
		}
	}
	filesJSON, _ := json.Marshal(files)
	now := time.Now().Unix()
	_, err := c.Client.QuickCreate(`tbl_butler_archive`, map[string]any{
		`config_id`:    configId,
		`task_id`:      taskId,
		`session_id`:   sessionId,
		`files`:        string(filesJSON),
		`conversation`: sb.String(),
		`status`:       `pending`,
		`created_at`:   now,
		`updated_at`:   now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	// 使用 last_insert_rowid() 获取刚插入的自增 ID，比 WHERE session_id 查询更可靠
	one, qErr := c.Client.QueryBySql(`SELECT last_insert_rowid() as id`).One()
	if qErr == nil && len(one) > 0 {
		id := cast.ToInt(one[`id`])
		if id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf(`创建归档记录后无法获取自增ID session=%s`, sessionId)
}

// ListPendingArchives 查询待处理的归档记录。
func (c *CSqlite) ListPendingArchives(limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}
	return c.Client.QueryBySql(
		`SELECT * FROM tbl_butler_archive WHERE status = 'pending' ORDER BY id ASC LIMIT ?`, limit,
	).All()
}

// UpdateArchiveStatus 更新归档记录的状态、日志和结果。
func (c *CSqlite) UpdateArchiveStatus(id int, status, logContent, result, resultFile, resultIndex string) error {
	updateData := map[string]any{
		`status`:       status,
		`log`:          logContent,
		`result`:       result,
		`result_file`:  resultFile,
		`result_index`: resultIndex,
		`updated_at`:   time.Now().Unix(),
	}
	_, err := c.Client.QuickUpdate(`tbl_butler_archive`, map[string]any{`id`: id}, updateData).Exec()
	return err
}

// WriteArchiveScript 将归档管家生成的脚本内容写入 skills/dtool-butler/scripts/ 目录。
func WriteArchiveScript(rootPath, scriptName, content string) (string, error) {
	scriptsDir := filepath.Join(rootPath, `skills`, `dtool-butler`, `scripts`)
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return ``, err
	}
	filePath := filepath.Join(scriptsDir, scriptName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return ``, err
	}
	return filePath, nil
}

// AppendArchiveIndex 向 scripts.md 追加一条归档脚本索引条目。
func AppendArchiveIndex(rootPath, skillName, scriptName, description string) error {
	indexPath := filepath.Join(rootPath, `skills`, `dtool-butler`, `index`, `scripts.md`)
	// 读取现有内容
	existing, err := os.ReadFile(indexPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	var sb strings.Builder
	sb.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("\n## [%s] %s\n\n- 脚本: %s\n- 来源: 归档管家自进化\n", skillName, description, scriptName))
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}
