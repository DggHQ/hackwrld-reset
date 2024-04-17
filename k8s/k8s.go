package k8s

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (k *KubeManager) LoadClientSet() *KubeManager {
	clientSet, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		panic(err)
	}
	k.ClientSet = clientSet
	return k
}

func (k *KubeManager) DeletePlayers(namespace string, labelSelector string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := k.ClientSet.AppsV1().Deployments(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{
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
