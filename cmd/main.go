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
	k8sGateway.Start()

	handler := server.NewGatewayHandler(k8sGateway)

	// TODO: load tls cert to remove this workaround
	time.Sleep(1 * time.Second)
	gatewayServer := server.NewServer(":8085", handler.BuildGatewayHandler(), 10*time.Second, k8sGateway)
	gatewayServer.Listen()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("Shutting down gateway...")
}
