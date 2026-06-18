package butler

import (
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"strings"
)

// BuildSystemPrompt 根据角色配置拼装最终的 system prompt。
// 如果角色已配置完整的 system_prompt，直接使用；
// 否则用 persona（定位）+ tone（语气）组合生成。
func BuildSystemPrompt(role *define.RoleItem) string {
	if role == nil {
		return defaultSystemPrompt
	}
	// 优先使用角色自带的完整 system_prompt
	if strings.TrimSpace(role.SystemPrompt) != `` {
		return role.SystemPrompt
	}
	// 用 persona + tone 组合
	parts := make([]string, 0, 2)
	if strings.TrimSpace(role.Persona) != `` {
		parts = append(parts, fmt.Sprintf(`你是%s。`, role.Persona))
	}
	if strings.TrimSpace(role.Tone) != `` {
		parts = append(parts, fmt.Sprintf(`回复语气：%s。`, role.Tone))
	}
	if len(parts) == 0 {
		return defaultSystemPrompt
	}
	// 加上简洁约束
	parts = append(parts, `回复简洁专业，避免冗长，不使用表情符号。`)
	return strings.Join(parts, `\n`)
}

// defaultSystemPrompt 角色配置为空时的兜底 system prompt。
const defaultSystemPrompt = `你是 dtool 智能管家，负责协助开发者完成日常工作。
回复简洁专业，避免冗长，不使用表情符号。
如果用户意图不明确，请简短追问澄清。`
