package main

import (
	"context"
	"custom-ingress/gateway"
	"custom-ingress/server"
	"custom-ingress/watcher"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	k8sWatcher := watcher.NewWatcher()
	k8sGateway := gateway.NewGateway(k8sWatcher)
	k8sGateway.Start()

	handler := server.NewGatewayHandler(k8sGateway)

	gatewayServer := server.NewServer(":8085", handler.BuildGatewayHandler(), 30*time.Second, k8sGateway)
	gatewayServer.Listen()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("Shutting down gateway...")
}
