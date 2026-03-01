package dtool

import (
	"dev_tool/internal/app/dtool/component"
	"fmt"
)

// GetPrimaryPort 返回第一个可用端口，供桌面端拼接启动地址。
func GetPrimaryPort() string {
	if component.EnvClient == nil || len(component.EnvClient.Ports) == 0 {
		return `17170`
	}
	return component.EnvClient.Ports[0]
}

// GetPrimaryURL 返回桌面端默认加载的本地地址。
func GetPrimaryURL() string {
	return fmt.Sprintf(`http://localhost:%s/`, GetPrimaryPort())
}
