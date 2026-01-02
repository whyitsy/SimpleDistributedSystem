package main

import (
	"DistributedGo/grades"
	"DistributedGo/registry"
	"DistributedGo/services"
	"context"
	"fmt"
	"log"
)

func main() {
	host, port := "localhost", ":10002"
	serviceAddress := fmt.Sprintf("%v%v", host, port)
	re := registry.RegistrationEntry{
		ServiceName: registry.GradingService,
		ServiceURL:  serviceAddress,
	}
	ctx, err := services.Start(context.Background(), host, port, re, grades.RegisterHandler)
	if err != nil {
		log.Fatalf("failed to start service: %v", err)
	}
	<-ctx.Done()
	fmt.Printf("服务[%s]已关闭", re.ServiceName)
}
