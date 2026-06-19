package index

import (
	"dev_tool/internal/app/dtool/component"
	"fmt"
	"strings"
)

// routeDesc 路由描述元数据，用于生成 apis.md。
type routeDesc struct {
	Path        string
	Method      string
	Description string
}

// GenerateApisIndex 生成 apis.md 内容——dtool HTTP 接口索引。
// 从 Gin 引擎中内省所有已注册路由，自动生成接口路径和描述，无需手动维护。
// 描述由路由路径和 Handler 函数名自动推断。
func GenerateApisIndex() string {
	routes := collectRoutes()

	var sb strings.Builder
	sb.WriteString("# dtool HTTP 接口索引\n\n")
	sb.WriteString(fmt.Sprintf("共 %d 个已注册接口。所有接口均为 POST 方法，需携带 Token 鉴权。\n", len(routes)))
	sb.WriteString("此索引由 `/init` 自动根据当前代码生成，始终与代码同步。\n\n")

	// 按功能域分组
	groups := groupRoutes(routes)
	for _, group := range groups {
		sb.WriteString(fmt.Sprintf("## %s\n\n", group.Name))
		sb.WriteString("| 接口路径 | 说明 |\n|----------|------|\n")
		for _, r := range group.Routes {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", r.Path, r.Description))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// collectRoutes 从 Gin 引擎收集所有已注册路由。
func collectRoutes() []routeDesc {
	routes := make([]routeDesc, 0)
	if len(component.TGins) == 0 || component.TGins[0] == nil {
		return routes
	}
	ginRoutes := component.TGins[0].GinGetRoutes()
	seen := make(map[string]bool)
	for _, r := range ginRoutes {
		path := normalizePath(r.Path)
		if path == `` || seen[path] {
			continue
		}
		seen[path] = true
		desc := deriveDesc(r.Path, r.Handler)
		routes = append(routes, routeDesc{
			Path:        path,
			Method:      r.Method,
			Description: desc,
		})
	}
	return routes
}

// normalizePath 规范化路由路径。
func normalizePath(path string) string {
	if path == `` {
		return ``
	}
	// 跳过 SSE/WebSocket 内部路由和不适合管家调用的路径
	skipPrefixes := []string{`/sse`, `/share/`, `/memory/images/`, `/web/download/`, `/download/`}
	for _, p := range skipPrefixes {
		if strings.HasPrefix(path, p) {
			return ``
		}
	}
	return path
}

// deriveDesc 根据路由路径和 Handler 字符串推断中文描述。
func deriveDesc(path, handler string) string {
	funcName := extractHandlerName(handler)
	if funcName == `` {
		return path
	}
	desc := inferDesc(funcName)
	if desc != `` {
		return desc
	}
	return funcName
}

// extractHandlerName 从 Gin handler 名提取函数名。
// "dev_tool/.../controller.GitConfigList-fm" → "GitConfigList"
func extractHandlerName(handler string) string {
	dotIdx := -1
	for i := len(handler) - 1; i >= 0; i-- {
		if handler[i] == '.' {
			dotIdx = i
			break
		}
	}
	if dotIdx < 0 || dotIdx >= len(handler)-1 {
		return ``
	}
	name := handler[dotIdx+1:]
	if idx := strings.Index(name, `-`); idx >= 0 {
		name = name[:idx]
	}
	return strings.TrimSpace(name)
}

// inferDesc 根据函数名后缀和前缀推断中文描述。
func inferDesc(funcName string) string {
	// 操作后缀 → 中文
	suffixMap := map[string]string{
		`List`:          `列表`,
		`Save`:          `保存`,
		`Add`:           `新增`,
		`Delete`:        `删除`,
		`Del`:           `删除`,
		`Create`:        `创建`,
		`Info`:          `详情`,
		`Search`:        `搜索`,
		`Query`:         `查询`,
		`Run`:           `执行`,
		`Status`:        `状态`,
		`Restart`:       `重启`,
		`Stop`:          `停止`,
		`Start`:         `启动`,
		`Test`:          `测试`,
		`Check`:         `检查`,
		`Update`:        `更新`,
		`Sort`:          `排序`,
		`Upload`:        `上传`,
		`Download`:      `下载`,
		`Toggle`:        `切换`,
		`Restore`:       `恢复`,
		`Remove`:        `移除`,
		`Preview`:       `预览`,
		`Logs`:          `日志`,
		`Generate`:      `生成`,
		`Fetch`:         `拉取`,
		`Organize`:      `整理`,
		`Import`:        `导入`,
		`Send`:          `发送`,
		`Continue`:      `继续`,
		`Login`:         `登录`,
		`Logout`:        `登出`,
		`Register`:      `注册`,
		`Reset`:         `重置`,
		`Migrate`:       `迁移`,
		`Change`:        `变更`,
		`Action`:        `操作`,
		`Retry`:         `重试`,
		`Vacuum`:        `回收`,
		`Analysis`:      `分析`,
		`Recycle`:       `回收`,
		`Truncate`:      `清空`,
		`Detail`:        `详情`,
		`Basic`:         `基础信息`,
		`Batch`:         `批量`,
		`Push`:          `推送`,
		`Pull`:          `拉取`,
		`Cleanup`:       `清理`,
		`Clean`:         `清理`,
		`MarkRead`:      `标记已读`,
		`EnsureRunning`: `保活`,
		`Tail`:          `尾行`,
		`Share`:         `分享`,
		`Page`:          `页面`,
		`All`:           `全部`,
		`Sub`:           `子项`,
		`Type`:          `类型`,
		`Instruction`:   `说明`,
		`Version`:       `版本`,
		`Forward`:       `转发`,
		`Set`:           `设置`,
	}

	// 尝试前缀+后缀组合
	for suffix, action := range suffixMap {
		if strings.HasSuffix(funcName, suffix) {
			prefix := strings.TrimSuffix(funcName, suffix)
			domain := prefixToDomain(prefix)
			if domain != `` {
				return domain + action
			}
			return prefix + action
		}
	}
	return funcName
}

// prefixToDomain 将函数名前缀转为功能域中文名。
func prefixToDomain(prefix string) string {
	prefix = strings.TrimPrefix(prefix, `Set`)
	if strings.HasPrefix(prefix, `Set`) {
		prefix = strings.TrimPrefix(prefix, `Set`)
	}
	domainMap := map[string]string{
		`Git`:                     `Git `,
		`GitLab`:                  `GitLab `,
		`GitGroup`:                `Git分组 `,
		`GitQuick`:                `Git快捷 `,
		`GitPending`:              `Git待提交 `,
		`Mysql`:                   `MySQL `,
		`Redis`:                   `Redis `,
		`PgSql`:                   `PgSQL `,
		`Docker`:                  `Docker `,
		`DockerCompose`:           `Docker `,
		`DockerImage`:             `Docker镜像 `,
		`DockerContainer`:         `Docker容器 `,
		`DockerService`:           `Docker服务 `,
		`Supervisor`:              `Supervisor `,
		`Supervisorctl`:           `Supervisor `,
		`Ssh`:                     `SSH `,
		`AiProvider`:              `AI服务商 `,
		`AiModel`:                 `AI模型 `,
		`AiRequestLog`:            `AI日志 `,
		`AgentCli`:                `AgentCLI `,
		`AgentCliGroup`:           `AgentCLI分组 `,
		`AgentCliPromptTemplate`:  `AgentCLI模板 `,
		`Mcp`:                     `MCP `,
		`McpType`:                 `MCP类型 `,
		`McpBinding`:              `MCP绑定 `,
		`McpAgentTarget`:          `MCP目标 `,
		`McpChromeDevtoolsConfig`: `MCP工具 `,
		`McpConfig`:               `MCP配置 `,
		`Butler`:                  `管家 `,
		`ButlerBotConfig`:         `管家机器人 `,
		`ButlerMessage`:           `管家消息 `,
		`ButlerRole`:              `管家角色 `,
		`ButlerConfig`:            `管家参数 `,
		`ButlerApi`:               `管家API `,
		`MemoryFragment`:          `记忆片段 `,
		`MemoryConfig`:            `记忆库配置 `,
		`Memory`:                  `记忆库 `,
		`HomeTask`:                `首页任务 `,
		`HomeTaskConfig`:          `首页任务配置 `,
		`HomeTaskDailyReport`:     `首页日报 `,
		`HomeTaskLastDevConfig`:   `首页开发配置 `,
		`HomeTaskBranchName`:      `首页分支名 `,
		`HomeTaskZcode`:           `首页Zcode `,
		`HomeTaskPageData`:        `首页页面数据 `,
		`HomeTaskUnused`:          `首页清理 `,
		`TaskWorkflow`:            `任务工作流 `,
		`TaskStatus`:              `任务状态 `,
		`WorkflowTemplate`:        `工作流模板 `,
		`WorkflowSkill`:           `工作流技能 `,
		`SmartLink`:               `智能链接 `,
		`SmartLinkItem`:           `智能链接项 `,
		`SmartLinkGroup`:          `智能链接分组 `,
		`SmartLinkChrome`:         `智能链接Chrome `,
		`SmartLinkDownload`:       `智能链接下载 `,
		`SmartLinkOpen`:           `智能链接打开 `,
		`SmartLinkLocator`:        `智能链接定位器 `,
		`SmartProcess`:            `智能流程 `,
		`SmartProcessItem`:        `智能流程项 `,
		`Variable`:                `变量 `,
		`VariableGroup`:           `变量分组 `,
		`VariableCmd`:             `变量命令 `,
		`Webhook`:                 `Webhook `,
		`WebhookConfig`:           `Webhook `,
		`Cron`:                    `定时任务 `,
		`CronConfig`:              `定时任务 `,
		`Global`:                  `全局配置 `,
		`Group`:                   `分组 `,
		`Account`:                 `账号 `,
		`AccountGroup`:            `账号分组 `,
		`CmdGroup`:                `命令分组 `,
		`RuntimeConfig`:           `运行时配置 `,
		`MainDB`:                  `主库 `,
		`PromptChangeLog`:         `Prompt变更 `,
		`LocalDir`:                `本地目录 `,
		`LocalBranch`:             `本地分支 `,
		`RemoteBranch`:            `远程分支 `,
		`Api`:                     `API `,
		`ApiFolder`:               `API目录 `,
		`ApiCollection`:           `API集合 `,
		`ApiCollectionEnv`:        `API环境 `,
		`Tool`:                    `工具 `,
		`ToolPort`:                `端口进程 `,
		`ToolManaged`:             `托管进程 `,
		`Base`:                    `基础 `,
		`BaseLogin`:               `登录 `,
		`BaseSsh`:                 `SSH `,
		`Markdown`:                `Markdown `,
		`Star`:                    `收藏 `,
		`ShellOut`:                `Shell `,
		`ShellOutRuleSet`:         `Shell规则集 `,
		`Php`:                     `PHP `,
		`Sse`:                     `SSE `,
		`AsyncTask`:               `异步任务 `,
		`Screenshot`:              `截图 `,
		`File`:                    `文件 `,
		`Collection`:              `API集合 `,
		`Archive`:                 `归档 `,
		`KnowledgeBase`:           `知识库 `,
		`Prompt`:                  `Prompt `,
		`Open`:                    `打开 `,
	}
	// 精确匹配
	if ch, ok := domainMap[prefix]; ok {
		return ch
	}
	// 模糊匹配：尝试去掉末尾数字和常见词
	for key, ch := range domainMap {
		if strings.HasPrefix(prefix, key) {
			return ch
		}
	}
	return ``
}

// routeGroup 路由分组。
type routeGroup struct {
	Name   string
	Routes []routeDesc
}

// groupRoutes 按路径前缀分组。
func groupRoutes(routes []routeDesc) []routeGroup {
	// 分组规则：按路径第二段（/api/xxx/...）分组
	groupMap := make(map[string][]routeDesc)
	groupOrder := make([]string, 0)

	for _, r := range routes {
		groupName := extractGroup(r.Path)
		if _, exists := groupMap[groupName]; !exists {
			groupOrder = append(groupOrder, groupName)
		}
		groupMap[groupName] = append(groupMap[groupName], r)
	}

	groups := make([]routeGroup, 0, len(groupOrder))
	for _, name := range groupOrder {
		groups = append(groups, routeGroup{Name: name, Routes: groupMap[name]})
	}
	return groups
}

// extractGroup 从路径提取分组名。
func extractGroup(path string) string {
	parts := strings.Split(path, `/`)
	if len(parts) < 3 {
		return `其他`
	}
	// /api/Set/Xxx → 配置管理
	// /api/task/workflow/xxx → 任务工作流
	// /api/xxx → 按 xxx 分组
	if parts[2] == `Set` && len(parts) >= 4 {
		return prefixToDomain(parts[3]) + `配置`
	}
	if parts[2] == `task` && len(parts) >= 4 {
		return `任务` + parts[3]
	}
	if parts[2] == `workflow` {
		return `工作流`
	}
	if parts[2] == `agent` {
		return `Agent`
	}
	if parts[2] == `ai` {
		return `AI浏览器`
	}
	if parts[2] == `smart-link` {
		return `智能链接`
	}
	domain := prefixToDomain(parts[2])
	if domain != `` {
		return strings.TrimSpace(domain)
	}
	return parts[2]
}

// VerifyPathsInApisMd 在 apis.md 内容中验证给定的 API 路径列表是否存在。
// 返回 allFound（是否全部找到）和详情说明字符串。
func VerifyPathsInApisMd(apisContent string, paths []string) (allFound bool, detail string) {
	if apisContent == `` || len(paths) == 0 {
		return false, ``
	}
	// 解析 apis.md 中的路径表格，收集所有已注册的接口路径
	registered := make(map[string]bool, 400)
	lines := strings.Split(apisContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 匹配表格行：| /api/xxx | ... |
		if !strings.HasPrefix(line, `| /api/`) {
			continue
		}
		parts := strings.SplitN(line, `|`, 3)
		if len(parts) >= 2 {
			path := strings.TrimSpace(parts[1])
			if path != `` && strings.HasPrefix(path, `/api/`) {
				registered[path] = true
			}
		}
	}

	found := make([]string, 0)
	missing := make([]string, 0)
	for _, p := range paths {
		if registered[p] {
			found = append(found, p)
		} else {
			missing = append(missing, p)
		}
	}

	allFound = len(missing) == 0
	var sb strings.Builder
	if len(found) > 0 {
		sb.WriteString(fmt.Sprintf(`已确认 %d 个路径存在于 apis.md：%s。`, len(found), strings.Join(found, `, `)))
	}
	if len(missing) > 0 {
		sb.WriteString(fmt.Sprintf(`以下 %d 个路径不存在于 apis.md：%s。`, len(missing), strings.Join(missing, `, `)))
	}
	detail = sb.String()
	return
}
