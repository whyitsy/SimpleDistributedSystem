package main

import (
	"DistributedGo/registry"
	"context"
	"fmt"
	"log"
	"net/http"
)

// 服务注册这个服务与其他被注册服务不一样. 服务注册类似于后端的服务, 被注册的服务类似客户端的服务.
// 这里的逻辑类似于service.service.go中的逻辑.
func main() {
	http.Handle("/services", &registry.RegistryService{}) // 注册服务注册处理器

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var srv http.Server
	srv.Addr = registry.ServerPort

	// 1. 启动该服务, 如果启动失败, 直接结束该服务
	go func() {
		log.Println(srv.ListenAndServe())
		cancel()
	}()

	// 2. 手动关闭该服务
	go func() {
		fmt.Printf("服务注册中心已启动, 监听地址%s\n", registry.ServerPort)
		fmt.Printf("按任意键退出服务注册中心...\n")
		var s string
		_, _ = fmt.Scanln(&s)
		fmt.Println("正在关闭服务注册中心...")
		_ = srv.Shutdown(ctx)
		cancel()

	}()
	<-ctx.Done()
	fmt.Println("服务注册中心已关闭")
}
