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
	"github.com/knabben/tkw/pkg/windows"
	"github.com/vmware/govmomi/vim25/mo"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const (
	TKG_NAMESPACE = "kube-system"

	ReasonCRNotAvailable         = "OperatorResourceNotAvailable"
	ReasonDeploymentNotAvailable = "DeploymentNotAvailable"
	ReasonSucceeded              = "OperatorSucceeded"
)

// OSImageReconciler reconciles a OSImage object
type OSImageReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Credentials *config.Mapper
}

// todo(knabben): review the correct required RBACs
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=imagebuilder.tanzu.opssec.in,resources=osimages/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;get;list
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;read;list;watch;create
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;read;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs="*"
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs="*"
//+kubebuilder:rbac:groups="batch",resources=jobs,verbs="*"

func (r *OSImageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling object.", "req", req.NamespacedName)

	var cmap = &config.Mapper{}
	var o imagebuilderv1alpha1.OSImage

	if err := r.Get(ctx, req.NamespacedName, &o); err != nil && errors.IsNotFound(err) {
		logger.Info("Resource not found.")
		return ctrl.Result{}, nil
	} else if err != nil {
		logger.Error(err, "Error getting object resource.")
		meta.SetStatusCondition(&o.Status.Conditions, metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionTrue,
			Reason:             ReasonCRNotAvailable,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            fmt.Sprintf("unable to get CR: %s", err.Error()),
		})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, &o)})
	}

	if err := r.getCredentials(ctx, cmap); err != nil {
		logger.Error(err, "unable to get configmap, create the required objects.")
		return ctrl.Result{}, nil
	}

	logger.Info("Checking assets deployment and execute.")
	if err := r.checkAssetsDeployment(ctx, cmap, &o); err != nil {
		logger.Error(err, "Error getting assets objects.")
		meta.SetStatusCondition(&o.Status.Conditions, metav1.Condition{
			Type:               "OperatorDegraded",
			Status:             metav1.ConditionTrue,
			Reason:             ReasonDeploymentNotAvailable,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Message:            fmt.Sprintf("unable to get deployment: %s", err.Error()),
		})
		return ctrl.Result{}, utilerrors.NewAggregate([]error{err, r.Status().Update(ctx, &o)})
	}

	// reconcile the status with the machine find
	if err := r.reconcileStatus(ctx, &o, cmap); err != nil {
		logger.Error(err, "unable to set OSImage object status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OSImageReconciler) checkAssetsDeployment(ctx context.Context, cmap *config.Mapper, imagebuilder *imagebuilderv1alpha1.OSImage) error {
	logger := log.FromContext(ctx)

	// Create the Windows resource bundle objects
	wrb, err := r.getOrCreateWindowsResourceBundle(ctx, imagebuilder)
	if err != nil {
		return err
	}

	// Populate Windows configuration and save on a temporary file
	logger.Info("Building windows.json file on memory.")

	// Manage the configuration based on mgmt parameters and specs
	// this configMap will be mounted in the Job as a volume.
	settings, err := windows.NewWindowsSettings(
		imagebuilder.Spec.WindowsISOPath,
		imagebuilder.Spec.VMToolsPath,
		wrb.Service.Name,
		wrb.Service.Namespace,
		wrb.Service.Spec.Ports[0].Port,
		imagebuilder,
	).GenerateJSONConfig(cmap)
	if err != nil {
		return err
	}

	return r.getOrCreateWindowsImageBuilder(ctx, string(settings), imagebuilder)
}

// SetupWithManager sets up the controller with the Manager.
func (r *OSImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&imagebuilderv1alpha1.OSImage{}).
		Owns(&appsv1.Deployment{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

func (r *OSImageReconciler) reconcileStatus(ctx context.Context, o *imagebuilderv1alpha1.OSImage, cmap *config.Mapper) error {
	var vms []mo.VirtualMachine

	if len(o.Status.OSTemplates) < 1 {
		// Connect and filter DataCenter.
		vc, dc, err := vsphere.ConnectFilterDC(ctx,
			cmap.Get(vsphere.VsphereServer),
			cmap.Get(vsphere.VsphereUsername),
			cmap.Get(vsphere.VspherePassword),
			cmap.Get(vsphere.VsphereDataCenter),
		)
		if err != nil {
			return err
		}

		// Get templates from vSphere and DC.
		if vms, err = vc.GetImportedVirtualMachinesImages(ctx, dc.Moid); err != nil {
			return err
		}

		// Iterate on VMS and print table by VM
		var osTemplates = make([]imagebuilderv1alpha1.OSImageTemplates, len(vms))
		for i, vm := range vms {
			osTemplates[i].Name = vm.Name
			properties := vc.GetVMMetadata(&vm)
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
	}

	meta.SetStatusCondition(&o.Status.Conditions, metav1.Condition{
		Type:               "OperatorDegraded",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             ReasonSucceeded,
		Message:            "operator successfully reconciling.",
	})

	return r.Status().Update(ctx, o)
}
