package config

import (
	"fmt"
	"strings"

	"github.com/safeie/goproxy/core/scheduler"
)

// Config 配置
type Config struct {
	Debug     bool
	Scheduler string
	Protocol  string
	Local     string
	Servers   []string
}

// New 创建配置
func New(protocol string, local string, server string) (*Config, error) {
	t := new(Config)
	t.Scheduler = scheduler.IPHashName
	if protocol == "" {
		protocol = "tcp"
	}
	if protocol == "tcp" || protocol == "udp" {
		t.Protocol = protocol
	} else {
		return nil, fmt.Errorf("仅支持tcp/udp协议")
	}
	if local == "" {
		return nil, fmt.Errorf("本地监听端口不能为空")
	}
	t.Local = local
	servers := strings.Split(server, ",")
	if len(servers) == 0 {
		return nil, fmt.Errorf("真实服务器地址不能为空")
	}
	t.Servers = servers

	return t, nil
}
