package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
)

// 需要注册服务到服务中心的服务调用这里提供的方法进行注册, DRY.
// 该方法会对服务注册中心发送一个HTTP.POST请求进行服务注册.

func RegisterService(re RegistrationEntry) error {
	// 在注册服务时, 添加回调接收服务更新通知的handler.
	serviceUpdateUrl, err := url.Parse(re.ServiceUpdateURL)
	if err != nil {
		return fmt.Errorf("服务更新URL解析失败: %s, 错误: %v", re.ServiceUpdateURL, err)
	}
	http.Handle(serviceUpdateUrl.Path, new(serviceUpdateHandler))
	// 添加健康检查的 handler
	heartbeatUrl, err := url.Parse(re.HeartbeatURL)
	if err != nil {
		return fmt.Errorf("服务更新URL解析失败: %s, 错误: %v", re.HeartbeatURL, err)
	}
	http.HandleFunc(heartbeatUrl.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// 可以返回一些服务的状态信息, 这里简单起见, 只返回200状态码.
	})
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

func DeregisterService(re RegistrationEntry) error {
	// http包没有直接提供DELETE方法, 需要通过NewRequest来创建请求.
	buffer := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buffer)
	if err := encoder.Encode(re); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, ServicesURL, buffer)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("关闭服务注销响应Body失败, %s:%s\n", re.ServiceName, re.ServiceURL)
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("服务注销失败, 状态码: %d, 服务: %s:%s", res.StatusCode, re.ServiceName, re.ServiceURL)
	}
	return nil
}

// 更新 Provider的http逻辑
type serviceUpdateHandler struct{}

func (s *serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	var p patch
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}
	fmt.Printf("接收到服务更新通知: %+v\n", p)
	prov.Update(p)
}

type providers struct {
	services map[ServiceName][]string
	mutex    *sync.RWMutex
}

func (p *providers) Update(pat patch) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, entry := range pat.Added {
		if _, ok := p.services[entry.Name]; !ok {
			// 如果服务还不存在, 先创建一个空的切片.
			p.services[entry.Name] = []string{}
		}
		p.services[entry.Name] = append(p.services[entry.Name], entry.URL)
	}

	// 遍历通知的移除服务列表, 如果存在, 则遍历Provider找到对应的URL并移除.
	for _, entry := range pat.Removed {
		if providedUrls, ok := p.services[entry.Name]; ok {
			for i, providedUrl := range providedUrls {
				if providedUrl == entry.URL {
					p.services[entry.Name] = append(providedUrls[:i], providedUrls[i+1:]...)
				}
			}
		}
	}
}

// get 根据服务名获取对应的服务提供者URL列表. 如果存在多个url可以使用, 则可以使用负载均衡策略选择一个.
// 这里简单起见, 直接随机返回一个URL.
func (p *providers) get(name ServiceName) (string, error) {
	providers, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("服务不存在: %s", name)
	}
	// 生成一个在[0.0, len(providers))范围内的随机浮点数, 然后转换为整数索引.
	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil
}

var prov = providers{
	services: make(map[ServiceName][]string),
	mutex:    &sync.RWMutex{},
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}
