package watcher

import (
	"context"
	coreV1 "k8s.io/api/core/v1"
	networkV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path"
)

type Watcher struct {
	k8sClient *kubernetes.Clientset
}

func NewWatcher() *Watcher {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path.Join(home, ".kube/config"))

	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Watcher{k8sClient: client}
}

func (w *Watcher) WatchIngress() {
	watch, err := w.k8sClient.NetworkingV1().Ingresses("default").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for event := range watch.ResultChan() {
		ingress, ok := event.Object.(*networkV1.Ingress)
		if !ok {
			log.Printf("unexpected type %T\n", event.Object)
			continue
		}
		log.Printf("event type: %s, rules: %+v, labels: %+v", event.Type, ingress.Spec.Rules[0], ingress.ObjectMeta.Labels)
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
