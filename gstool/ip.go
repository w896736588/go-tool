package gstool

import (
	"net"
)

// IpGetLocalIp 获取内网IP 注意有多个IP
func IpGetLocalIp() []string {
	localIpList := make([]string, 0)
	address, err := net.InterfaceAddrs()
	if err != nil {
		return localIpList
	}
	for _, address := range address {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				localIpList = append(localIpList, ipNet.IP.String())
			}
		}
	}
	return localIpList
}
