package k8s

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
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

// func (k *KubeManager) CreateDeploymentWatcher(ctx context.Context, namespace string, labelSelector string) (watch.Interface, error) {
// 	opts := metav1.ListOptions{
// 		LabelSelector: labelSelector,
// 	}
// 	return k.ClientSet.AppsV1().Deployments(namespace).Watch(ctx, opts)
// }

// func (k *KubeManager) WaitDeploymentDeleted(ctx context.Context, namespace string, labelSelector string) error {
// 	watcher, err := k.CreateDeploymentWatcher(ctx, namespace, labelSelector)
// 	if err != nil {
// 		return err
// 	}
// 	defer watcher.Stop()
// 	for {
// 		select {
// 		case event := <-watcher.ResultChan():
// 			if event.Type == watch.Deleted {
// 				log.Printf("Deployment has been deleted")
// 				continue
// 			}
// 		case <-ctx.Done():
// 			log.Printf("All deployments deleted")
// 			return nil
// 		}
// 	}
// }

func (k *KubeManager) UpdateWebDeploymentEnv(ctx context.Context, namespace string, name string, maintenance string) {

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deploy, err := k.ClientSet.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			panic(fmt.Errorf("failed to get latest version of deployment: %v", err))
		}
		currentEnv := deploy.Spec.Template.Spec.Containers[0].Env
		for _, env := range currentEnv {
			if env.Name == "MAINTENANCE" {
				env.Value = maintenance
				log.Printf("Updated maintenance to %s", maintenance)
			}
		}
		_, err = k.ClientSet.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment...")
}
