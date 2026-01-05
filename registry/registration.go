package registry

type RegistrationEntry struct {
	ServiceName      ServiceName // 自定义类型, 可以扩展功能
	ServiceURL       string
	RequiredServices []ServiceName // 依赖的服务, 在注册时请求这些服务
	ServiceUpdateURL string
	HeartbeatURL     string
}

type ServiceName string

const (
	LogService     = ServiceName("LogService")
	GradingService = ServiceName("GradingService")
)

// patchEntry 表示每次服务变更时, 注册中心发送的更新内容
type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
