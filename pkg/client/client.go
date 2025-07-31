package client

import (
	"context"
	"fmt"
	"time"

	"k8sgo/internal/types"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	clientset kubernetes.Interface
	config    *rest.Config
}

func NewKubernetesClient(kubeconfig string) (*KubernetesClient, error) {
	// Use default kubeconfig loading rules if no specific path provided
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		// Use the default loading rules which will check:
		// 1. KUBECONFIG environment variable
		// 2. $HOME/.kube/config (Linux/macOS) or %USERPROFILE%\.kube\config (Windows)
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesClient{
		clientset: clientset,
		config:    config,
	}, nil
}

func (kc *KubernetesClient) GetPods(namespace string) ([]types.Resource, error) {
	pods, err := kc.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []types.Resource
	for _, pod := range pods.Items {
		resource := types.Resource{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Type:      "Pod",
			Status:    string(pod.Status.Phase),
			Ready:     fmt.Sprintf("%d/%d", getReadyContainers(pod), len(pod.Spec.Containers)),
			Restarts:  getRestartCount(pod),
			Age:       time.Since(pod.CreationTimestamp.Time),
			CPU:       "N/A",
			Memory:    "N/A",
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

func (kc *KubernetesClient) GetServices(namespace string) ([]types.Resource, error) {
	services, err := kc.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []types.Resource
	for _, svc := range services.Items {
		resource := types.Resource{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      "Service",
			Status:    "Active",
			Ready:     "N/A",
			Restarts:  0,
			Age:       time.Since(svc.CreationTimestamp.Time),
			CPU:       "N/A",
			Memory:    "N/A",
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

func (kc *KubernetesClient) GetDeployments(namespace string) ([]types.Resource, error) {
	deployments, err := kc.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []types.Resource
	for _, deploy := range deployments.Items {
		resource := types.Resource{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Type:      "Deployment",
			Status:    getDeploymentStatus(deploy),
			Ready:     fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas),
			Restarts:  0,
			Age:       time.Since(deploy.CreationTimestamp.Time),
			CPU:       "N/A",
			Memory:    "N/A",
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

func (kc *KubernetesClient) SwitchContext(context string) error {
	return fmt.Errorf("context switching not implemented in client layer")
}

func (kc *KubernetesClient) GetContexts() ([]string, error) {
	return nil, fmt.Errorf("context listing not implemented in client layer")
}

func (kc *KubernetesClient) GetCurrentContext() (string, error) {
	return "", fmt.Errorf("current context not implemented in client layer")
}

func (kc *KubernetesClient) GetNamespaces() ([]string, error) {
	namespaces, err := kc.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nsNames []string
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)
	}

	return nsNames, nil
}

func getReadyContainers(pod v1.Pod) int {
	ready := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			ready++
		}
	}
	return ready
}

func getRestartCount(pod v1.Pod) int {
	restarts := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restarts += int(containerStatus.RestartCount)
	}
	return restarts
}

func getDeploymentStatus(deployment appsv1.Deployment) string {
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
		return "Ready"
	}
	if deployment.Status.ReadyReplicas == 0 {
		return "Not Ready"
	}
	return "Partial"
}
