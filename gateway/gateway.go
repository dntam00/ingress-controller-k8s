package gateway

import (
	"context"
	"custom-ingress/model"
	"custom-ingress/watcher"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
)

type Gateway struct {
	watcher    *watcher.Watcher
	rules      map[string]model.IngressRules
	httpClient *http.Client
}

func NewGateway(watcher *watcher.Watcher, httpClient *http.Client) *Gateway {
	return &Gateway{
		watcher:    watcher,
		rules:      make(map[string]model.IngressRules),
		httpClient: httpClient,
	}
}

func (g *Gateway) Start() {
	g.watcher.WatchIngress()
}

func (g *Gateway) Stop() {
	g.watcher.Stop()
}

func (g *Gateway) Route(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	rules, ok := g.rules[host]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	ingressRules := rules.Rules
	clone := slices.Clone(ingressRules)
	slices.SortFunc(clone, func(i, j model.IngressRule) int {
		return len(j.Path) - len(i.Path)
	})
	for _, rule := range clone {
		if strings.HasPrefix(request.URL.Path, rule.Path) {
			// <service-name>.<namespace-name>.svc.cluster.local
			target := fmt.Sprintf("%s.%s.svc.cluster.local", rule.Service, "default")
			g.request(request, writer, target)
			// proxy to service
			return
		}
	}
}

func (g *Gateway) request(orgRequest *http.Request, orgResponseWriter http.ResponseWriter, target string) {
	// TODO: support GET request
	req, err := http.NewRequestWithContext(context.Background(), orgRequest.Method, target, nil)
	if err != nil {
		log.Println("error when forwarding request: ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	g.doRequest(req, orgResponseWriter)
}

func (g *Gateway) doRequest(req *http.Request, orgResponseWriter http.ResponseWriter) {
	res, err := g.httpClient.Do(req)
	if err != nil {
		fmt.Println("error call api: ", err)
	}

	if res != nil {
		_, err = io.Copy(orgResponseWriter, res.Body)
		_ = res.Body.Close()
	}
}
