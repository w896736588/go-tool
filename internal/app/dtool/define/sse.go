package define

const (
	SseAiCode                = `ai_code`                 //废弃
	SseGitLab                = `gitlab`                  //固定全局唯一
	SseShellConnections      = `shell_connections`       //Shell连接状态推送
	SseMemoryFragmentUpdates = `memory_fragment_updates` //知识片段变更推送
	SseAsyncTasks            = `async_tasks`             //异步任务状态推送
)

const (
	SseEventClean = `[CLEAN]`                   //清除前端的数据
	SseEventLogin = `[LOGIN_USERNAME_PASSWORD]` //通知前端弹窗输入账号密码
	SseDown       = `[DONE]`                    //前端换个行
	SseConnect    = `[CONNECT]`                 //链接已建立
)

type SseEvent string
