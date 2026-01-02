package registry

type RegistrationEntry struct {
	ServiceName ServiceName // 自定义类型, 可以扩展功能
	ServiceURL  string
}

type ServiceName string

const (
	LogServiceName = ServiceName("LogService")
)
