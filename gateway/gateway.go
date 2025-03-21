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
	c := g.watcher.WatchIngress()
	go func() {
		for {
			select {
			case rules, ok := <-c:
				if !ok {
					log.Println("watcher channel closed")
					return
				}
				for _, r := range rules {
					slices.SortFunc(r.Rules, func(a, b model.IngressRule) int {
						return len(b.Path) - len(a.Path)
					})
				}
				g.rules = rules
				log.Println("gateway update rules:", g.rules)
			}
		}
	}()
}

func (g *Gateway) Stop() {
	g.watcher.Stop()
}

func (g *Gateway) Route(writer http.ResponseWriter, request *http.Request) {
	log.Println("start routing request", request.URL.Path)
	host := request.Host
	rules, ok := g.rules[host]
	if !ok {
		log.Println("no rules found for host:", host)
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	ingressRules := rules.Rules
	for _, rule := range ingressRules {
		if strings.HasPrefix(request.URL.Path, rule.Path) {
			// service address pattern: <service-name>.<namespace-name>.svc.cluster.local
			target := fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", "http", rule.Service, rule.Namespace, rule.Port)
			g.request(request, writer, target)
			return
		}
	}
	log.Println("no path rule found for host: ", host)
}

func (g *Gateway) request(orgRequest *http.Request, orgResponseWriter http.ResponseWriter, target string) {
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
		orgResponseWriter.WriteHeader(http.StatusBadGateway)
		_, _ = orgResponseWriter.Write([]byte(err.Error()))
		return
	}

	if res != nil {
		_, err = io.Copy(orgResponseWriter, res.Body)
		if err != nil {
			log.Println("error when copying response body: ", err)
		}
		_ = res.Body.Close()
	}
}
