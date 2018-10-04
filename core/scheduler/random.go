package scheduler

import (
	"time"
)

// Random 随机调度器
type Random struct {
}

// Schedule 调度
func (strategy *Random) Schedule(client string, servers []string) string {
	length := len(servers)
	server := servers[int(time.Now().UnixNano())%length]
	return server
}

func init() {
	Register(RandomName, new(Random))
}
