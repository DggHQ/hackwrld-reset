package k8s

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeManager struct {
	Config    *rest.Config
	ClientSet *kubernetes.Clientset
}

// Load config from inside cluster
func (k *KubeManager) Init() *KubeManager {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	k.Config = clusterConfig
	return k
}

// Load client set from config
func (k *KubeManager) LoadClientSet() {
	clientSet, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		panic(err)
	}
	k.ClientSet = clientSet
}

func (k *KubeManager) DeletePlayers(ctx context.Context, namespace string, labelSelector string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := k.ClientSet.AppsV1().Deployments(namespace).DeleteCollection(ctx, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	},
		metav1.ListOptions{
			LabelSelector: labelSelector,
		},
	)
	if err != nil {
		log.Fatalf("Could not delete deployments %v", err)
		return err
	}
	return nil
}

func (k *KubeManager) CreateDeploymentWatcher(ctx context.Context, namespace string, labelSelector string) (watch.Interface, error) {
	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	return k.ClientSet.AppsV1().Deployments(namespace).Watch(ctx, opts)
}

func (k *KubeManager) WaitDeploymentDeleted(ctx context.Context, namespace string, labelSelector string) error {
	watcher, err := k.CreateDeploymentWatcher(ctx, namespace, labelSelector)
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Deleted {
				log.Printf("Deployment has been deleted")
			}
		case <-ctx.Done():
			log.Printf("All deployments deleted")
			return nil
		}
	}
}
