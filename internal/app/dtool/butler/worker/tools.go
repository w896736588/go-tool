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
	}
}
