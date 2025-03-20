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
	c := make(chan map[string]model.IngressRules)
	go func() {
		for {
			select {
			case rules := <-c:
				g.rules = rules
			}
		}
	}()
	g.watcher.WatchIngress(c)
}

func (g *Gateway) Stop() {
	g.watcher.Stop()
}

func (g *Gateway) Route(writer http.ResponseWriter, request *http.Request) {
	log.Println("start routing request", request.URL.Path)

	host := request.Host
	rules, ok := g.rules[host]
	if !ok {
		log.Println("no rules found for host: ", host)
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
			target := fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", "http", rule.Service, "default", rule.Port)
			g.request(request, writer, target)
			return
		}
	}
}

func (g *Gateway) request(orgRequest *http.Request, orgResponseWriter http.ResponseWriter, target string) {
	log.Println("forwarding request to: ", target)
	req, err := http.NewRequestWithContext(context.Background(), orgRequest.Method, target+orgRequest.URL.Path, nil)
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
		if err != nil {
			log.Println("error when copying response body: ", err)
		}
		_ = res.Body.Close()
	}
}
