package ai_model

import "dev_tool/internal/pkg/ai/ai_define"

type AiModel interface {
	Api(messageList []ai_define.Message, tools []ai_define.Tool) (string, error)
	MessageList() []ai_define.Message
}
