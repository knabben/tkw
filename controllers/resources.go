package controllers

import (
	"context"
	"fmt"
	"github.com/knabben/tkw/api/v1alpha1"
	"github.com/knabben/tkw/controllers/assets"
	"github.com/knabben/tkw/pkg/config"
	"github.com/knabben/tkw/pkg/vsphere"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getCredentials fetch the vsphere-cloud-config cm and extract data in the mapper
func (r *OSImageReconciler) getCredentials(ctx context.Context, cmap *config.Mapper) error {
	vsphereCM, name := &v1.ConfigMap{}, "vsphere-cloud-config"
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: TKG_NAMESPACE}, vsphereCM); err != nil {
		return err
	}

	// Fetch vsphere-cloud-config and extract data
	var vsphereSM = &v1.Secret{}
	data := vsphereCM.Data["vsphere.conf"]
	cmap.Set(vsphere.VsphereServer, extractRValue(`\[VirtualCenter "(.*)"\]`, data))
	namespacedName := types.NamespacedName{
		Name:      extractRValue(`secret-name = "(.*)"`, data),
		Namespace: extractRValue(`secret-namespace = "(.*)"`, data),
	}

	// Set DataCenter from configMap
	cmap.Set(vsphere.VsphereDataCenter, extractRValue(`datacenters = "(.*)"`, data))

	if err := r.Get(ctx, namespacedName, vsphereSM); err != nil {
		return err
	}
	vcIP := cmap.Get(vsphere.VsphereServer)
	cmap.Set(vsphere.VsphereUsername, string(vsphereSM.Data[fmt.Sprintf("%s.%s", vcIP, "username")]))
	cmap.Set(vsphere.VspherePassword, string(vsphereSM.Data[fmt.Sprintf("%s.%s", vcIP, "password")]))
	return nil
}

func (r *OSImageReconciler) getOrCreate(ctx context.Context, object client.Object) (client.Object, error) {
	logger := log.FromContext(ctx)
	named := types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()}

	if err := r.Get(ctx, named, object); err != nil && errors.IsNotFound(err) {
		logger.Info("Creating object.", "object", named)
		if err := r.Create(ctx, object); err != nil {
			return object, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("Error trying to get object: %v", err)
	}

	return object, nil
}

type WindowsResourceBundle struct {
	Deployment *appsv1.Deployment
	Namespace  *v1.Namespace
	Service    *v1.Service
}

// getOrCreateWindowsResourceBundle returns the Windows resource bundle specification
func (r *OSImageReconciler) getOrCreateWindowsResourceBundle(ctx context.Context, ib *v1alpha1.OSImage) (*WindowsResourceBundle, error) {
	// Check for Windows resource bundle deployment and create
	deploy := assets.YAMLAccessor[*appsv1.Deployment]{}
	depObject, err := deploy.GetDecodedObject(assets.BUILDER_DEPLOYMENT, appsv1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}

	// Check for the Windows resource bundle service and create
	svc := assets.YAMLAccessor[*v1.Service]{}
	svcObject, err := svc.GetDecodedObject(assets.BUILDER_SERVICE, v1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}

	// Set controller reference and create the object
	for _, x := range []client.Object{depObject, svcObject} {
		if err := ctrl.SetControllerReference(ib, x, r.Scheme); err != nil {
			return nil, err
		}
		if _, err := r.getOrCreate(ctx, x); err != nil {
			return nil, err
		}
	}

	return &WindowsResourceBundle{
		Deployment: depObject,
		Service:    svcObject,
	}, nil
}

func (r *OSImageReconciler) getOrCreateWindowsImageBuilder(ctx context.Context, config string, ib *v1alpha1.OSImage) error {
	// Check for Windows resource bundle deployment and create
	configmap := assets.YAMLAccessor[*v1.ConfigMap]{}
	cmObject, err := configmap.GetDecodedObject(assets.IB_CONFIG, v1.SchemeGroupVersion)
	if err != nil {
		return err
	}

	// Save json data in the object and create the configmap.
	cmObject.Data = map[string]string{"windows.json": config}
	if _, err := r.getOrCreate(ctx, cmObject); err != nil {
		return err
	}

	// Creates the Job from spec file.
	job := assets.YAMLAccessor[*batchv1.Job]{}
	jobObject, err := job.GetDecodedObject(assets.IB_JOB, batchv1.SchemeGroupVersion)
	if err != nil {
		return err
	}

	if err := ctrl.SetControllerReference(ib, jobObject, r.Scheme); err != nil {
		return err
	}
	if _, err := r.getOrCreate(ctx, jobObject); err != nil {
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
