package gateway

import (
	"custom-ingress/model"
	"custom-ingress/watcher"
	"log"
	"net/http"
	"net/http/httputil"
	"slices"
	"time"
)

type Gateway struct {
	watcher    *watcher.Watcher
	route      *Route
	httpClient *http.Client
}

func NewGateway(watcher *watcher.Watcher, httpClient *http.Client) *Gateway {
	return &Gateway{
		watcher:    watcher,
		route:      NewRoute(),
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
				if rules.Type == watcher.ADDED || rules.Type == watcher.MODIFIED {
					for _, r := range rules.Ingress {
						slices.SortFunc(r.Rules, func(a, b model.IngressRule) int {
							return len(b.Path) - len(a.Path)
						})
						g.route.UpdateRoute(r.Host, *r)
					}
				} else if rules.Type == watcher.DELETED {
					for host := range rules.Ingress {
						g.route.DeleteRoute(host)
					}
				}
				log.Println("gateway update rules:", rules)
			}
		}
	}()
}

func (g *Gateway) Stop() {
	g.watcher.Stop()
}

func (g *Gateway) GetRoute() *Route {
	return g.route
}

func (g *Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("start routing request", request.URL.Path)
	backendURL, err := g.route.GetRoute(request)
	if err != nil {
		log.Println("error when getting route: ", err)
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
	p := httputil.NewSingleHostReverseProxy(backendURL)
	p.Transport = transport()
	p.ServeHTTP(writer, request)
}

func transport() http.RoundTripper {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = true
	t.MaxConnsPerHost = 1
	t.IdleConnTimeout = time.Duration(5) * time.Second
	t.MaxIdleConns = 1
	t.MaxIdleConnsPerHost = 1
	return t
}
