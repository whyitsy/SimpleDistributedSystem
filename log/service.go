package log

import (
	"io"
	stlog "log" // 项目自定义的log变量和标准库的log会有命名冲突, 所以做一个别名
	"net/http"
	"os"
)

var log *stlog.Logger

// 定义一个类型，实现io.Writer接口，用于写入日志文件. 初始化logger时使用这个自定义的io.Writer
type fileLog string

func (fl fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) // 0600表示文件权限，拥有者可读写
	if err != nil {
		return 0, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			stlog.Println("关闭文件失败:", err)
		}
	}(f)
	return f.Write(data)
}

// Run 初始化日志写入路径
func Run(destination string) {
	log = stlog.New(fileLog(destination), "[Go]: ", stlog.LstdFlags)
}

// RegisterHandlers 注册日志处理器, 用于单独启动的日志Web服务
func RegisterHandlers() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			msg, err := io.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			write(string(msg))
		default:
			_, _ = w.Write([]byte("method not allowed"))
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}
	})
}

func write(message string) {
	log.Printf("%v\n", message) // %v 表示按默认格式输出, 是一个通用的占位符
}
