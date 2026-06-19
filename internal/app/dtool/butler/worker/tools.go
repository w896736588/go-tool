package worker

// ToolDefinitions 返回 OpenAI 格式的工具定义列表，供 FC 循环使用。
func ToolDefinitions() []map[string]any {
	return []map[string]any{
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolFileRead,
				`description`: `读取文件内容。返回文件的完整文本内容。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `要读取的文件路径`,
						},
					},
					`required`: []string{`path`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolFileWrite,
				`description`: `将内容写入文件。如果文件不存在则创建（包括父目录），如果已存在则覆盖。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `要写入的文件路径`,
						},
						`content`: map[string]any{
							`type`:        `string`,
							`description`: `要写入的文件内容`,
						},
					},
					`required`: []string{`path`, `content`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolFileModify,
				`description`: `修改文件内容，通过查找并替换指定文本。如果不指定 replacement，则删除匹配的文本。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `要修改的文件路径`,
						},
						`search`: map[string]any{
							`type`:        `string`,
							`description`: `要查找的文本`,
						},
						`replacement`: map[string]any{
							`type`:        `string`,
							`description`: `替换后的文本（为空则删除匹配文本）`,
						},
					},
					`required`: []string{`path`, `search`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolFileDelete,
				`description`: `删除文件。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `要删除的文件路径`,
						},
					},
					`required`: []string{`path`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolHttpCall,
				`description`: `调用 dtool 的 HTTP API 接口。所有接口均为 POST 方法，基地址已自动拼接，只需传接口路径和 JSON 请求体。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `API 接口路径，如 /api/GitConfigList、/api/GitRemoteBranchList`,
						},
						`body`: map[string]any{
							`type`:        `string`,
							`description`: `JSON 格式的请求体，如 {}、{"ssh_id":"5","code_path":"/var/www/common3"}`,
						},
					},
					`required`: []string{`path`, `body`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolRunScript,
				`description`: `执行本地 Python 脚本，返回 stdout 和 stderr 输出。脚本路径基于 skills/ 目录。优先使用已有脚本完成用户任务。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`path`: map[string]any{
							`type`:        `string`,
							`description`: `脚本路径，如 skills/dtool-git/scripts/git_api.py`,
						},
						`args`: map[string]any{
							`type`:        `string`,
							`description`: `命令行参数（空格分隔），如 --repo_name common3 --branch develop`,
						},
						`timeout`: map[string]any{
							`type`:        `string`,
							`description`: `超时秒数，默认 60 秒`,
						},
					},
					`required`: []string{`path`},
				},
			},
		},
		{
			`type`: `function`,
			`function`: map[string]any{
				`name`:        ToolAskUser,
				`description`: `向用户提问确认，暂停当前任务等待用户回复。仅当缺少必要信息（操作对象不明确、参数不足）或需要确认危险操作时使用。只读查询无需确认。`,
				`parameters`: map[string]any{
					`type`: `object`,
					`properties`: map[string]any{
						`question`: map[string]any{
							`type`:        `string`,
							`description`: `向用户提问的内容，应清晰列出选项或需要补充的信息`,
						},
						`options`: map[string]any{
							`type`:        `string`,
							`description`: `可选选项列表，用逗号分隔，如 common3-web,common3-api,common3-admin。为空则用户自由回答`,
						},
						`reason`: map[string]any{
							`type`:        `string`,
							`description`: `需要确认的原因，如 操作对象不明确、危险操作确认、参数不足`,
						},
					},
					`required`: []string{`question`, `reason`},
				},
			},
		},
	}
}
