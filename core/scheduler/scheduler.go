package scheduler

const (
	IPHashName = "iphash"
	RandomName = "random"
)

var store = make(map[string]Scheduler)

// Scheduler 调度器
type Scheduler interface {
	Schedule(client string, servers []string) string
}

// Get 获取调度器
func Get(name string) Scheduler {
	return store[name]
}

// Register 注册调度器
func Register(name string, handle Scheduler) {
	store[name] = handle
}
