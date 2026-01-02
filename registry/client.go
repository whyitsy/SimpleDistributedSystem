package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// 需要注册服务到服务中心的服务调用这里提供的方法进行注册, DRY.
// 该方法会对服务注册中心发送一个HTTP.POST请求进行服务注册.

func RegisterService(re RegistrationEntry) error {
	// POST请求需要一个io.Reader类型的body参数.可以这样构造:
	// buffer是一个实现了io.Writer接口和io.Reader接口的类型.使用json.Encoder可以直接将结构体编码到buffer中.
	// 然后将buffer作为POST请求的body参数传递, 作为io.Reader使用.
	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(re); err != nil {
		return err
	}
	// 2. 发送POST请求到服务注册中心.
	res, err := http.Post(ServicesURL, "application/json", buffer)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("关闭服务注册响应Body失败, %s:%s\n", re.ServiceName, re.ServiceURL)
		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("服务注册失败, 状态码: %d, 服务: %s:%s", res.StatusCode, re.ServiceName, re.ServiceURL)
	}
	return nil
}
