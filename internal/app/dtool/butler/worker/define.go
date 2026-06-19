package worker

// 工具名称常量
const (
	ToolFileRead   = `file_read`
	ToolFileWrite  = `file_write`
	ToolFileModify = `file_modify`
	ToolFileDelete = `file_delete`
	ToolHttpCall   = `http_call`  // 调用 dtool HTTP API
	ToolRunScript  = `run_script` // 执行 Python 脚本
	ToolAskUser    = `ask_user`   // 向用户发起确认问题
)

// askUserMarker ask_user 工具返回结果的特殊标记，FC 循环通过此标记识别需要暂停等待用户回复。
const AskUserMarker = `__ASK_USER__`
