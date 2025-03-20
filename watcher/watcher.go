package watcher

import (
	"context"
	"custom-ingress/model"
	coreV1 "k8s.io/api/core/v1"
	networkV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

type Watcher struct {
	k8sClient *kubernetes.Clientset
}

func NewWatcher() *Watcher {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Watcher{k8sClient: client}
}

func (w *Watcher) WatchIngress(resCh chan map[string]model.IngressRules) {
	watch, err := w.k8sClient.NetworkingV1().Ingresses("default").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for event := range watch.ResultChan() {
		ingresses := make(map[string]model.IngressRules)
		ingress, ok := event.Object.(*networkV1.Ingress)
		if !ok {
			log.Printf("unexpected type %T\n", event.Object)
			continue
		}
		for _, rule := range ingress.Spec.Rules {
			host := rule.Host
			rules := make([]model.IngressRule, 0)
			for _, path := range rule.HTTP.Paths {
				pathStr := path.Path
				serviceName := path.Backend.Service.Name
				servicePort := path.Backend.Service.Port.Number
				rules = append(rules, model.IngressRule{
					Path:    pathStr,
					Service: serviceName,
					Port:    servicePort,
				})
			}
			ingresses[host] = model.IngressRules{
				Host:  host,
				Rules: rules,
			}
			log.Printf("ingress event type: %s, host: %+v", event.Type, rule.Host)
		}
		resCh <- ingresses
	}
}

func (w *Watcher) WatchService() {
	defaultNs := "default"
	watch, err := w.k8sClient.CoreV1().Services(defaultNs).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for event := range watch.ResultChan() {
		service, ok := event.Object.(*coreV1.Service)
		if !ok {
			log.Printf("unexpected type %T\n", event.Object)
			continue
		}
		log.Printf("event type: %s, clusterId: %s, labels: %+v", event.Type, service.Spec.ClusterIP, service.ObjectMeta.Labels)
	}
}

func (w *Watcher) Stop() {

}
