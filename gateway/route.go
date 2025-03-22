package gateway

import (
	"custom-ingress/model"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Route struct {
	table map[string]model.IngressRules
}

func NewRoute() *Route {
	return &Route{table: make(map[string]model.IngressRules)}
}

func (r *Route) GetRoute(request *http.Request) (*url.URL, error) {
	host := request.Host
	routes, ok := r.table[host]
	if !ok {
		return nil, errors.New("no upstream found")
	}

	ingressRules := routes.Rules
	backend := ""
	for _, rule := range ingressRules {
		if strings.HasPrefix(request.URL.Path, rule.Path) {
			// service address pattern: <service-name>.<namespace-name>.svc.cluster.local
			backend = fmt.Sprintf("%s.%s.svc.cluster.local:%d", rule.Service, rule.Namespace, rule.Port)
			break
		}
	}
	if backend == "" {
		return nil, errors.New("no route found")
	}
	return &url.URL{
		Scheme: "http",
		Host:   backend,
	}, nil
}

func (r *Route) UpdateRoute(host string, rules model.IngressRules) {
	r.table[host] = rules
}

func (r *Route) DeleteRoute(host string) {
	delete(r.table, host)
}
