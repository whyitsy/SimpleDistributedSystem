package main

import (
	"DistributedGo/log"
	"DistributedGo/services"
	"context"
	stlog "log"
)

func main() {
	log.Run("distributed_go.log")
	host, port := "localhost", "10900"
	ctx, err := services.Start(context.Background(), "Log Service", host, port, log.RegisterHandlers)
	if err != nil {
		stlog.Fatalln("启动服务失败:", err) // 此时自定义的日志服务还没有启动, 所以使用标准日志输出
		return
	}
	<-ctx.Done() // 等待服务结束的信号
}
