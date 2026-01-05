package main

import (
	"DistributedGo/grades"
	"DistributedGo/log"
	"DistributedGo/registry"
	"DistributedGo/services"
	"context"
	"fmt"
	stlog "log"
)

func main() {
	host, port := "localhost", ":10002"
	serviceAddress := fmt.Sprintf("http://%v%v", host, port)
	re := registry.RegistrationEntry{
		ServiceName:      registry.GradingService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/health",
	}
	ctx, err := services.Start(context.Background(), host, port, re, grades.RegisterHandler)
	if err != nil {
		stlog.Fatalf("failed to start service: %v", err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Println("Log service provider found: ", logProvider)
		log.SetClientLogger(logProvider+"/log", re.ServiceName)
		stlog.Println("Log service provider found: ", re.ServiceName)
	}

	<-ctx.Done()
	fmt.Printf("服务[%s]已关闭", re.ServiceName)
}
