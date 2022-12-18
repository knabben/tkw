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
	imagebuilderv1alpha1 "github.com/knabben/tkw/api/v1alpha1"
	"github.com/knabben/tkw/pkg/config"
	"github.com/knabben/tkw/pkg/vsphere"
	"github.com/vmware/govmomi/vim25/mo"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	TKG_NAMESPACE = "kube-system"
)

// OSImageReconciler reconciles a OSImage object
type OSImageReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Credentials *config.Mapper
}

//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;get;list
//+kubebuilder:rbac:groups="",namespace="kube-system",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",namespace="kube-system",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=service,namespace=windows,verbs="*"
//+kubebuilder:rbac:groups="",resources=deployments,namespace=windows,verbs="*"

func (r *OSImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var cmap = &config.Mapper{}
	logger := log.FromContext(ctx)
	logger.Info("Reconciling object.", "req", req.NamespacedName)

	var o imagebuilderv1alpha1.OSImage
	if err := r.Get(ctx, req.NamespacedName, &o); err != nil {
		logger.Error(err, "unable to get OSImage object")
		return ctrl.Result{}, nil
	}

	if err := r.getCredentials(ctx, cmap); err != nil {
		logger.Error(err, "unable to get configmap")
		return ctrl.Result{}, err
	}

	logger.Info("Checking assets deployment and execute.")
	if err := r.checkAssetsDeployment(ctx, &o); err != nil {
		logger.Error(err, "--- unable to create deployment objects")
		return ctrl.Result{}, err
	}

	// reconcile the status with the machine find
	if err := r.reconcileStatus(ctx, &o, cmap); err != nil {
		logger.Error(err, "unable to set OSImage object status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: false}, nil
}

func (r *OSImageReconciler) checkAssetsDeployment(ctx context.Context, imagebuilder *imagebuilderv1alpha1.OSImage) error {
	_, err := r.getOrCreateWindowsResourceBundle(ctx, imagebuilder)
	if err != nil {
		return err
	}
	/*
			// 3. Populate Windows configuration and save on a temporary file
			klog.Info(template.Info("Generate windows.json file with parameters"))
			winSettings := windows.NewWindowsSettings(
				viper.GetString("isopath"),
				viper.GetString("vmtoolspath"),
				nodeIP,
			)

			// Manage the configuration based on mgmt parameters
			data, err := winSettings.GenerateJSONConfig(mapper)
			config.ExplodeGraceful(err)
			windowsFile, err := winSettings.SaveTempJSON(data)
			config.ExplodeGraceful(err)

			// 4. Image builder running on a docker
			klog.Info(template.Info("Running Docker container with Image builder, be ready!"))
			cli, err := docker.NewDockerClient(windowsFile)
			config.ExplodeGraceful(err)

			// Run the image-builder container.
			var containerID string
			containerID, err = cli.Run(ctx)
			config.ExplodeGraceful(err)

			// Iterate on logs and print output, monitor for errors.
			err = monitorOutput(cli, containerID)
			config.ExplodeGraceful(err)
		},
	*/
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OSImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagebuilderv1alpha1.OSImage{}).
		Complete(r)
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
