package windows

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	NodePort  = 30008
	BURRITO   = "projects.registry.vmware.com/tkg/windows-resource-bundle:v1.23.8_vmware.2-tkg.1"
	namespace = "windows"
)

type Kubernetes struct {
	Clientset *kubernetes.Clientset
	Config    *restclient.Config

	Namespace  *apiv1.Namespace
	Service    *apiv1.Service
	Deployment *appsv1.Deployment
}

// GetFirstNodeIP returns the IP (internal or external) for the first node.
func (k *Kubernetes) GetFirstNodeIP(ctx context.Context) (string, error) {
	nodes, err := k.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == apiv1.NodeExternalIP || addr.Type == apiv1.NodeInternalIP {
				return addr.Address, nil
			}
		}
	}
	return "", nil // not found
}
