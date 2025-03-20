package main

import (
	"context"
	"custom-ingress/gateway"
	"custom-ingress/http"
	"custom-ingress/server"
	"custom-ingress/watcher"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	httpClient := http.NewHttpClient()
	k8sWatcher := watcher.NewWatcher()
	k8sGateway := gateway.NewGateway(k8sWatcher, httpClient)
	go k8sGateway.Start()

	handler := server.NewGatewayHandler(k8sGateway)
	gatewayServer := server.NewServer(":8085", handler.BuildGatewayHandler(), 10*time.Second)
	gatewayServer.Listen()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("Shutting down gateway...")
}
