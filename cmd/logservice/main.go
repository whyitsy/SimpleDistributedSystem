package main

import (
	"DistributedGo/log"
	"DistributedGo/registry"
	"DistributedGo/services"
	"context"
	"fmt"
	stlog "log"
)

func main() {
	log.Run("distributed_go.log") // 初始化日志服务, 指定日志文件路径
	host, port := "localhost", ":10001"
	serviceAddress := fmt.Sprintf("http://%v%v", host, port)
	re := registry.RegistrationEntry{
		ServiceName:      registry.LogService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/health",
	}
	ctx, err := services.Start(context.Background(), host, port, re, log.RegisterHandlers)
	if err != nil {
		stlog.Fatalln("启动服务失败:", err) // 此时自定义的日志服务还没有启动, 所以使用标准日志输出
		return
	}
	<-ctx.Done() // 等待服务结束的信号
	fmt.Printf("服务[%s]已关闭", re.ServiceName)
}
