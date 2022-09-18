package windows

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"tkw/pkg/template"
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

// NewKubernetesClient returns a Kubernetes object with clientset connected.
func NewKubernetesClient(kubeconfig string) (*Kubernetes, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{Clientset: clientset, Config: config}, nil
}

// CreateWindowsResources generates the Windows resource bundle assets - aka burrito.
func (k *Kubernetes) CreateWindowsResources(ctx context.Context) error {
	// see if ns is already create and skip if so.
	ns, err := k.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil && ns.GetName() == namespace {
		klog.Info(template.Warning(fmt.Sprintf("Namespace %s already exists, skipping resource creation.", namespace)))
		return nil
	}

	// create namespace.
	k8sns, err := k.Clientset.CoreV1().Namespaces().Create(ctx, NamespaceSpec(namespace), metav1.CreateOptions{})
	if err != nil {
		return err
	}
	k.Namespace = k8sns

	// create deployment with burrito image.
	k8sdpl, err := k.Clientset.AppsV1().Deployments(namespace).Create(ctx, DeploymentSpec(namespace), metav1.CreateOptions{})
	if err != nil {
		return err
	}
	k.Deployment = k8sdpl

	// create the service with NodePort.
	k8ssvc, err := k.Clientset.CoreV1().Services(namespace).Create(ctx, ServiceSpec(namespace), metav1.CreateOptions{})
	if err != nil {
		return err
	}
	k.Service = k8ssvc
	return nil
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

// NamespaceSpec defines the Kubernetes specification for the namespace.
func NamespaceSpec(ns string) *apiv1.Namespace {
	return &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
}

// ServiceSpec defines the Kubernetes specification for the service.
func ServiceSpec(ns string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "image-builder-wrs",
			Namespace: ns,
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": "image-builder-resource-kit",
			},
			Type: apiv1.ServiceTypeNodePort,
			Ports: []apiv1.ServicePort{
				{
					Port:     3000,
					NodePort: NodePort,
				},
			},
		},
	}
}

// DeploymentSpec defines the Kubernetes specification for the deployment.
func DeploymentSpec(ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "image-builder-resource-kit",
			Namespace: ns,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "image-builder-resource-kit"},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "image-builder-resource-kit"},
				},
				Spec: apiv1.PodSpec{
					NodeSelector: map[string]string{"kubernetes.io/os": "linux"},
					Containers: []apiv1.Container{
						{
							Name:            "windows-image-builder",
							Image:           BURRITO,
							ImagePullPolicy: apiv1.PullAlways,
							Ports: []apiv1.ContainerPort{
								{
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 3000,
								},
							},
						},
					},
				},
			},
		},
	}
}
