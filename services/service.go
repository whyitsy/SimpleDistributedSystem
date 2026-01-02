package services

import (
	"context"
	"fmt"
	"net/http"

	"log"
)

// 这里的服务是公共服务, 供其他模块调用的

// Start 启动一个http服务, 并注册处理器. 这是一个通用的服务启动函数, 所以单独放在service包中
func Start(ctx context.Context, serviceName, host, port string, registerHandler func()) (context.Context, error) {
	// 1. 注册处理器
	registerHandler()

	// 2. 启动服务
	ctx = startService(ctx, serviceName, host, port)
	return ctx, nil
}

func startService(ctx context.Context, name string, host string, port string) context.Context {
	// 因为后面需要启动两个goroutine来管理服务的生成周期, 所以使用可需要的ctx. 应该是一个典型的应用场景
	// 返回这个可取消的ctx, 让调用方可以等待服务的结束
	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	srv.Addr = host + port

	// 1. 启动该服务, 如果启动失败, 直接结束该服务
	go func() {
		log.Println(srv.ListenAndServe()) // 启动失败, 直接结束该服务
		cancel()
	}()

	// 2. 手动关闭该服务
	go func() {
		fmt.Printf("服务[%s]已启动, 监听地址%s%s\n", name, host, port)
		fmt.Printf("按任意键退出[%s]服务...\n", name)
		var s string
		_, _ = fmt.Scanln(&s)
		fmt.Println("正在关闭服务[", name, "]...")
		_ = srv.Shutdown(ctx)
		cancel()
		fmt.Println("服务[", name, "]已关闭")
	}()

	return ctx
}
