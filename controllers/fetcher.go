package controllers

import (
	"context"
	"fmt"
	"github.com/knabben/tkw/controllers/assets"
	"github.com/knabben/tkw/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
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
	data := vsphereCM.Data["vsphere.conf"]
	cmap.Set("vc", extractRValue(`\[VirtualCenter "(.*)"\]`, data))
	cmap.Set("secret-name", extractRValue(`secret-name = "(.*)"`, data))
	cmap.Set("secret-ns", extractRValue(`secret-namespace = "(.*)"`, data))

	var vsphereSM = &v1.Secret{}
	namespacedName := types.NamespacedName{Name: cmap.Get("secret-name"), Namespace: cmap.Get("secret-ns")}
	if err := r.Get(ctx, namespacedName, vsphereSM); err != nil {
		return err
	}
	vcIP := cmap.Get("vc")
	for _, s := range []string{"username", "password"} {
		cmap.Set(s, string(vsphereSM.Data[fmt.Sprintf("%s.%s", vcIP, s)]))
	}

	return nil
}

func (r *OSImageReconciler) getOrCreate(ctx context.Context, object client.Object) (client.Object, error) {
	logger := log.FromContext(ctx)

	named := types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()}
	logger.Info("Fetching object.", "object", named)

	if err := r.Get(ctx, named, object); err != nil && errors.IsNotFound(err) {
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

func (r *OSImageReconciler) getOrCreateWindowsResourceBundle(ctx context.Context) (*WindowsResourceBundle, error) {
	wrb := &WindowsResourceBundle{}

	// Check for Windows resource bundle namespace and create
	ns := assets.YAMLAccessor[*v1.Namespace]{}
	if nsObject, err := ns.GetDecodedObject(assets.BUILDER_NAMESPACE, v1.SchemeGroupVersion); err != nil {
		return nil, err
	} else {
		if _, err := r.getOrCreate(ctx, nsObject); err != nil {
			return nil, err
		}
		wrb.Namespace = nsObject
	}

	// Check for Windows resource bundle deployment and create
	deploy := assets.YAMLAccessor[*appsv1.Deployment]{}
	if deployObject, err := deploy.GetDecodedObject(assets.BUILDER_DEPLOYMENT, appsv1.SchemeGroupVersion); err != nil {
		return nil, err
	} else {
		if _, err := r.getOrCreate(ctx, deployObject); err != nil {
			return nil, err
		}
		wrb.Deployment = deployObject
	}

	// Check for the Windows resource bundle service and create
	svc := assets.YAMLAccessor[*v1.Service]{}
	if svcObject, err := svc.GetDecodedObject(assets.BUILDER_SERVICE, v1.SchemeGroupVersion); err != nil {
		return nil, err
	} else {
		if _, err := r.getOrCreate(ctx, svcObject); err != nil {
			return nil, err
		}
		wrb.Service = svcObject
	}

	return wrb, nil
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