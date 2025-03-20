package server

import (
	"custom-ingress/gateway"
	"github.com/gorilla/mux"
	"net/http"
)

type GatewayHandler struct {
	k8sGateway *gateway.Gateway
}

func NewGatewayHandler(k8sGateway *gateway.Gateway) *GatewayHandler {
	return &GatewayHandler{k8sGateway: k8sGateway}
}

func (handler *GatewayHandler) BuildGatewayHandler() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		handler.k8sGateway.Route(writer, request)
	}).Methods(http.MethodGet, http.MethodPost, http.MethodHead, http.MethodOptions)
	return r
}
