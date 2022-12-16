/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	imagebuilderv1alpha1 "github.com/knabben/tkw/api/v1alpha1"
	"github.com/knabben/tkw/pkg/config"
	"github.com/knabben/tkw/pkg/vsphere"
	"github.com/vmware/govmomi/vim25/mo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const TKGNamespace = "kube-system"

// OSImageReconciler reconciles a OSImage object
type OSImageReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Credentials *config.Mapper
}

//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages/finalizers,verbs=update

func (r *OSImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cmap = &config.Mapper{}
	logger := log.FromContext(ctx)

	var o imagebuilderv1alpha1.OSImage
	if err := r.Get(ctx, req.NamespacedName, &o); err != nil {
		logger.Error(err, "unable to get image object")
		return ctrl.Result{}, err
	}

	if err := r.getConfigMapCredentials(ctx, cmap); err != nil {
		logger.Error(err, "unable to get configmap")
		return ctrl.Result{}, err
	}
	if err := r.getSecretCredentials(ctx, cmap); err != nil {
		logger.Error(err, "unable to get configmap")
		return ctrl.Result{}, err
	}

	// reconcile the status with the machine find
	if err := r.reconcileStatus(ctx, &o, cmap); err != nil {
		logger.Error(err, "unable to get object status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OSImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagebuilderv1alpha1.OSImage{}).
		//Owns(&v1.Job{}).
		Complete(r)
}

// getConfigMapCredentials fetch the vsphere-cloud-config cm and extract data in the mapper
func (r *OSImageReconciler) getConfigMapCredentials(ctx context.Context, cmap *config.Mapper) error {
	vsphereCM, name := &v1.ConfigMap{}, "vsphere-cloud-config"
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: TKGNamespace}, vsphereCM); err != nil {
		return err
	}

	// Fetch vsphere-cloud-config and extract data
	data := vsphereCM.Data["vsphere.conf"]
	cmap.Set("vc", extractRValue(`\[VirtualCenter "(.*)"\]`, data))
	cmap.Set("secret-name", extractRValue(`secret-name = "(.*)"`, data))
	cmap.Set("secret-ns", extractRValue(`secret-namespace = "(.*)"`, data))

	return nil
}

// getSecretCredentials dump the secret user and pass in the mapper
func (r *OSImageReconciler) getSecretCredentials(ctx context.Context, cmap *config.Mapper) error {
	var (
		vsphereSM  = &v1.Secret{}
		namedspace = types.NamespacedName{Name: cmap.Get("secret-name"), Namespace: cmap.Get("secret-ns")}
	)

	if err := r.Get(ctx, namedspace, vsphereSM); err != nil {
		return err
	}

	vcIP := cmap.Get("vc")
	for _, s := range []string{"username", "password"} {
		cmap.Set(s, string(vsphereSM.Data[fmt.Sprintf("%s.%s", vcIP, s)]))
	}

	return nil
}

func (r *OSImageReconciler) reconcileStatus(ctx context.Context, o *imagebuilderv1alpha1.OSImage, cmap *config.Mapper) error {
	var vms []mo.VirtualMachine

	// Connect and filter DataCenter.
	client, dc, err := vsphere.ConnectFilterDC(ctx, cmap.Get("vc"), cmap.Get("username"), cmap.Get("password"))
	if err != nil {
		return err
	}

	// Get templates from vSphere and DC.
	if vms, err = client.GetImportedVirtualMachinesImages(ctx, dc.Moid); err != nil {
		return err
	}

	// Iterate on VMS and print table by VM
	var osTemplates = make([]imagebuilderv1alpha1.OSImageTemplates, len(vms))
	for i, vm := range vms {
		osTemplates[i].Name = vm.Name
		properties := client.GetVMMetadata(&vm)
		if properties != nil {
			osTemplates[i].BuildDate = properties["BUILD_DATE"]
			osTemplates[i].BuildTimestamp = properties["BUILD_TIMESTAMP"]
			osTemplates[i].CNIVersion = properties["CNI_VERSION"]
			osTemplates[i].ContainerDVersion = properties["CONTAINERD_VERSION"]
			osTemplates[i].DistroArch = properties["DISTRO_ARCH"]
			osTemplates[i].DistroName = properties["DISTRO_NAME"]
			osTemplates[i].DistroVersion = properties["DISTRO_VERSION"]
			osTemplates[i].ImageBuilderVersion = properties["IMAGE_BUILDER_VERSION"]
			osTemplates[i].KubernetesSemVer = properties["KUBERNETES_SEMVER"]
			osTemplates[i].KubernetesSourceType = properties["KUBERNETES_SOURCE_TYPE"]
		}
	}

	o.Status.OSTemplates = osTemplates
	if err := r.Status().Update(context.Background(), o); err != nil {
		return err
	}
	return nil
}

func extractRValue(v, d string) string {
	var (
		re  *regexp.Regexp
		err error
	)
	if re, err = regexp.Compile(v); err != nil {
		return ""
	}
	submatch := re.FindStringSubmatch(d)
	if len(submatch) < 1 {
		return ""
	}
	return submatch[1]
}
