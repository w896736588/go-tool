package gstool

import "net"

func NetIsPortAvailable(host string) bool {
	addr := host
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	closeErr := listener.Close()
	if closeErr != nil {
		return false
	}
	return true
}
