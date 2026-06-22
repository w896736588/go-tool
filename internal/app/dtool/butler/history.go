package butler

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"time"

	"github.com/spf13/cast"
)

// History 历史对话存储，读写 tbl_butler_message。
type History struct {
	db          *common.CSqlite
	botConfigId int
}

// NewHistory 创建历史存储实例，botConfigId 用于关联机器人配置。
func NewHistory(db *common.CSqlite, botConfigId int) *History {
	return &History{db: db, botConfigId: botConfigId}
}

// Append 追加一条消息记录，botConfigId 为消息来源机器人的配置 ID。
func (h *History) Append(sessionId, role, content string, botConfigId int) error {
	_, err := h.db.Client.QuickCreate(`tbl_butler_message`, map[string]any{
		`session_id`:    sessionId,
		`role`:          role,
		`content`:       content,
		`token_count`:   0,
		`topic`:         ``,
		`bot_config_id`: botConfigId,
		`created_at`:    time.Now().Unix(),
	}).Exec()
	return err
}

// AppendWithTopic 追加一条带主题标记的消息记录，botConfigId 为消息来源机器人的配置 ID。
func (h *History) AppendWithTopic(sessionId, role, content, topic string, botConfigId int) error {
	_, err := h.db.Client.QuickCreate(`tbl_butler_message`, map[string]any{
		`session_id`:    sessionId,
		`role`:          role,
		`content`:       content,
		`token_count`:   0,
		`topic`:         topic,
		`bot_config_id`: botConfigId,
		`created_at`:    time.Now().Unix(),
	}).Exec()
	return err
}

// CountBySession 返回指定会话的消息条数。
func (h *History) CountBySession(sessionId string) (int, error) {
	rows, err := h.db.Client.QuickQuery(`tbl_butler_message`, `id`, map[string]any{
		`session_id`: sessionId,
	}).All()
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// CleanBySession 清除指定会话的全部历史消息。
func (h *History) CleanBySession(sessionId string) error {
	_, err := h.db.Client.ExecBySql(
		`DELETE FROM tbl_butler_message WHERE session_id = ?`, sessionId,
	).Exec()
	return err
}

// TrimBySession 保留指定会话最新的 maxLimit 条消息，删除多余的旧消息。
// maxLimit <= 0 时不执行任何操作。
func (h *History) TrimBySession(sessionId string, maxLimit int) error {
	if maxLimit <= 0 {
		return nil
	}
	_, err := h.db.Client.ExecBySql(
		`DELETE FROM tbl_butler_message WHERE session_id = ? AND id NOT IN (
			SELECT id FROM tbl_butler_message WHERE session_id = ? ORDER BY id DESC LIMIT ?
		)`, sessionId, sessionId, maxLimit,
	).Exec()
	return err
}

// ListBySession 返回指定会话的历史消息（按 id 升序），最多 limit 条。
func (h *History) ListBySession(sessionId string, limit int) ([]define.ButlerHistoryMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := h.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_message WHERE session_id = ? ORDER BY id ASC LIMIT ?`,
		sessionId, limit,
	).All()
	if err != nil {
		return nil, err
	}
	result := make([]define.ButlerHistoryMessage, 0, len(rows))
	for _, row := range rows {
		result = append(result, define.ButlerHistoryMessage{
			Id:        cast.ToInt(row[`id`]),
			SessionId: cast.ToString(row[`session_id`]),
			Role:      cast.ToString(row[`role`]),
			Content:   cast.ToString(row[`content`]),
			Topic:     cast.ToString(row[`topic`]),
			CreatedAt: cast.ToInt64(row[`created_at`]),
		})
	}
	return result, nil
}

// GetRecentTopic 获取指定会话最近一条消息的主题关键词。
// 返回空字符串表示无历史（新对话）。
// 注意：QueryBySql(...).One() 会自动追加 LIMIT 1，SQL 中无需再写 LIMIT。
func (h *History) GetRecentTopic(sessionId string) (string, error) {
	row, err := h.db.Client.QueryBySql(
		`SELECT topic FROM tbl_butler_message WHERE session_id = ? AND topic != '' ORDER BY id DESC`,
		sessionId,
	).One()
	if err != nil {
		return ``, err
	}
	if len(row) == 0 {
		return ``, nil
	}
	return cast.ToString(row[`topic`]), nil
}

// UpdateTopicBySession 更新指定会话所有消息的主题（用于新话题检测后的主题回填）。
func (h *History) UpdateTopicBySession(sessionId, topic string) error {
	_, err := h.db.Client.ExecBySql(
		`UPDATE tbl_butler_message SET topic = ? WHERE session_id = ? AND topic = ''`,
		topic, sessionId,
	).Exec()
	return err
}

// ToAiMessages 将历史消息列表转换为 AI chat 的 messages 格式。
// systemPrompt 作为第一条 system 消息，历史消息按时间顺序追加。
func ToAiMessages(systemPrompt string, historyMessages []define.ButlerHistoryMessage) []map[string]string {
	messages := make([]map[string]string, 0, len(historyMessages)+1)
	// 第一条：system prompt
	messages = append(messages, map[string]string{
		`role`:    define.ButlerRoleSystem,
		`content`: systemPrompt,
	})
	// 历史消息
	for _, msg := range historyMessages {
		// 只取 user 和 assistant 角色（过滤掉 system 消息，避免与 systemPrompt 冲突）
		if msg.Role == define.ButlerRoleUser || msg.Role == define.ButlerRoleAssistant {
			messages = append(messages, map[string]string{
				`role`:    msg.Role,
				`content`: msg.Content,
			})
		}
	}
	return messages
}
