package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":10000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

// 包中的变量都是包级的, 私有的, 不需要对外暴露. 因为这些服务是通过http请求handler来调用的.
// 将handler实现在当前包中.

// 创建包级的类型, 保存注册服务的信息
type registry struct {
	services []RegistrationEntry
	// 上面的slice字段是线程不安全的, 需要加锁保护
	mutex *sync.RWMutex
}

// 注册服务的方法
func (r *registry) addService(re RegistrationEntry) error {
	r.mutex.Lock()
	r.services = append(r.services, re)
	r.mutex.Unlock()
	err := r.sendRequiredServices(re)
	r.notify(&patch{
		Added: []patchEntry{
			{
				Name: re.ServiceName,
				URL:  re.ServiceURL,
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// 取消注册服务的方法
func (r *registry) removeService(entry RegistrationEntry) error {
	for i, e := range r.services {
		if e.ServiceName == entry.ServiceName && e.ServiceURL == entry.ServiceURL {
			// 找到匹配的服务, 删除它
			r.mutex.Lock()
			r.services = append(r.services[:i], r.services[i+1:]...)
			r.mutex.Unlock()
			r.notify(&patch{
				Removed: []patchEntry{
					{
						Name: entry.ServiceName,
						URL:  entry.ServiceURL,
					},
				},
			})
			return nil
		}
	}
	return fmt.Errorf("service with name %s and URL %s not found", entry.ServiceName, entry.ServiceURL)
}

func (r *registry) notify(fullPatch *patch) {
	for _, entry := range r.services {
		go func(re RegistrationEntry) {
			for _, reqServiceName := range re.RequiredServices {
				p := patch{
					Added:   []patchEntry{},
					Removed: []patchEntry{},
				}
				sendUpdate := false
				for _, added := range fullPatch.Added {
					if added.Name == reqServiceName {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqServiceName {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, re.ServiceUpdateURL)
					if err != nil {
						log.Printf("Failed to send patch to %s: %v\n", re.ServiceUpdateURL, err)
					}
				}
			}
		}(entry)
	}
}

func (r *registry) sendRequiredServices(re RegistrationEntry) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var p patch
	for _, reqService := range re.RequiredServices {
		for _, registeredService := range r.services {
			if registeredService.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: registeredService.ServiceName,
					URL:  registeredService.ServiceURL,
				})
			}
		}
	}
	err := r.sendPatch(p, re.ServiceUpdateURL)
	if err != nil {
		log.Printf("Failed to send patch to %s: %v\n", re.ServiceUpdateURL, err)
		return err
	}
	return nil
}

func (r *registry) sendPatch(p patch, url string) error {
	pj, err := json.Marshal(p)
	if err != nil {
		log.Printf("Failed to marshal patch: %v\n", err)
		return err
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(pj))
	if err != nil {
		log.Printf("Failed to send patch to %s: %v\n", url, err)
		return err
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v\n", err)
		}
	}(res.Body)

	return nil
}

// reg var声明并实例化一个包级的registry变量
// Attention:  := 这种声明方式称为短变量声明, 只能在局部作用域中使用, 如函数体内, if/for块内等.
var reg = registry{
	services: make([]RegistrationEntry, 0),
	mutex:    &sync.RWMutex{},
}

// RegistryService 实现http.Handler接口, 用于http.Handle的第二个接口参数
type RegistryService struct{}

func (rs *RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request to register service received.")

	switch r.Method {
	case http.MethodPost:
		// 解析请求体中的字节数组注册信息
		var entry RegistrationEntry
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %+v\n", entry)
		err = reg.addService(entry)
		if err != nil {
			http.Error(w, "Failed to register service", http.StatusInternalServerError)
			return
		}
	// 这段结束后会自动返回200 OK 并关闭连接
	case http.MethodDelete:
		var entry RegistrationEntry
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("Removing service: %+v\n", entry)
		err = reg.removeService(entry)
		if err != nil {
			http.Error(w, "Failed to unregister service", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}
