package watcher

import (
	"context"
	"custom-ingress/model"
	networkV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

const (
	defaultNs = "default"
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

func (w *Watcher) WatchIngress() chan map[string]model.IngressRules {
	watchEvent, err := w.k8sClient.NetworkingV1().Ingresses(defaultNs).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	c := make(chan map[string]model.IngressRules)
	go func() {
		for event := range watchEvent.ResultChan() {
			ingresses := make(map[string]model.IngressRules)
			ingress, ok := event.Object.(*networkV1.Ingress)
			if !ok {
				log.Printf("unexpected type %T\n", event.Object)
				continue
			}
			namespace := ingress.Namespace
			for _, rule := range ingress.Spec.Rules {
				host := rule.Host
				rules := make([]model.IngressRule, 0)
				for _, path := range rule.HTTP.Paths {
					pathStr := path.Path
					serviceName := path.Backend.Service.Name
					servicePort := path.Backend.Service.Port.Number
					rules = append(rules, model.IngressRule{
						Namespace: namespace,
						Path:      pathStr,
						Service:   serviceName,
						Port:      servicePort,
					})
				}
				ingresses[host] = model.IngressRules{
					Host:  host,
					Rules: rules,
				}
				log.Printf("ingress event type: %s, host: %+v", event.Type, rule.Host)
			}
			c <- ingresses
		}
	}()
	return c
}

func (w *Watcher) WatchService() <-chan watch.Event {
	defaultNs := "default"
	watchEvent, err := w.k8sClient.CoreV1().Services(defaultNs).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return watchEvent.ResultChan()
	//for event := range watchEvent.ResultChan() {
	//	service, ok := event.Object.(*coreV1.Service)
	//	if !ok {
	//		log.Printf("unexpected type %T\n", event.Object)
	//		continue
	//	}
	//	log.Printf("event type: %s, clusterId: %s, labels: %+v", event.Type, service.Spec.ClusterIP, service.ObjectMeta.Labels)
	//}
}

func (w *Watcher) Stop() {

}
