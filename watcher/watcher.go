package watcher

import (
	"context"
	"crypto/tls"
	"custom-ingress/model"
	networkV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path"
)

const (
	defaultNs = "default"
)

const (
	ADDED    = "ADDED"
	MODIFIED = "MODIFIED"
	DELETED  = "DELETED"
)

type Watcher struct {
	k8sClient *kubernetes.Clientset
}

func localConfig() (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))

	if err != nil {
		panic(err.Error())
	}
	return config, nil
}

func NewWatcher() *Watcher {
	config, err := rest.InClusterConfig()
	//config, err := localConfig()
	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Watcher{k8sClient: client}
}

func (w *Watcher) WatchIngress() chan model.IngressEvent {
	watchEvent, err := w.k8sClient.NetworkingV1().Ingresses(defaultNs).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	c := make(chan model.IngressEvent)

	go func() {
		for event := range watchEvent.ResultChan() {
			ingresses := make(map[string]*model.IngressRules)
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
				ingresses[host] = &model.IngressRules{
					Host:  host,
					Rules: rules,
				}
				log.Printf("ingress event type: %s, host: %+v", event.Type, rule.Host)
			}

			for _, tlsRule := range ingress.Spec.TLS {
				hosts := tlsRule.Hosts
				secretName := tlsRule.SecretName
				secret, err := w.k8sClient.CoreV1().Secrets(ingress.Namespace).Get(context.Background(), secretName, metav1.GetOptions{})
				if err != nil {
					log.Printf("error getting secret %s: %v", secretName, err)
					continue
				}
				cert, err := tls.X509KeyPair(secret.Data["tls.crt"], secret.Data["tls.key"])
				for _, host := range hosts {
					existing := ingresses[host]
					existing.Cert = &cert
				}
			}

			c <- model.IngressEvent{Type: string(event.Type), Ingress: ingresses}
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
