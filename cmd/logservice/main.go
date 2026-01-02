package main

import (
	"DistributedGo/log"
	"DistributedGo/services"
	"context"
	stlog "log"
)

func main() {
	log.Run("distributed_go.log") // 初始化日志服务, 指定日志文件路径
	host, port := "localhost", ":10001"
	ctx, err := services.Start(context.Background(), "Log Service", host, port, log.RegisterHandlers)
	if err != nil {
		stlog.Fatalln("启动服务失败:", err) // 此时自定义的日志服务还没有启动, 所以使用标准日志输出
		return
	}
	<-ctx.Done() // 等待服务结束的信号
}
