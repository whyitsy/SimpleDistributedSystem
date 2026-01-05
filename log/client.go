package log

import (
	"DistributedGo/registry"
	"bytes"
	"fmt"
	"io"
	stlog "log"
	"net/http"
)

// 提供一个方法供客户端使用

func SetClientLogger(serviceUrl string, clientService registry.ServiceName) {
	//return stlog.New(&clientLogger{url: serviceUrl}, fmt.Sprintf("[%v] - ", clientService), 0)
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	stlog.SetFlags(0)
	stlog.SetOutput(&clientLogger{url: serviceUrl})
}

type clientLogger struct {
	url string
}

func (c *clientLogger) Write(data []byte) (n int, err error) {
	// 将日志发送到远程日志服务
	b := bytes.NewBuffer(data)
	res, err := http.Post(c.url, "text/plain", b)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log: %d, message: %s", res.StatusCode, string(data))
	}

	return len(data), nil
}
