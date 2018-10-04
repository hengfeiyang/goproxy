package server

import "errors"

// Server 代理服务器接口
type Server interface {
	Start()
	Stop()
}

// ErrNetClosing is returned when a network descriptor is used after
// it has been closed. Keep this string consistent because of issue
// #4373: since historically programs have not been able to detect
// this error, they look for the string.
var ErrNetClosing = errors.New("use of closed network connection")
