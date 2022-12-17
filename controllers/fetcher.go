package controllers

import (
	"fmt"
	"context"
	"github.com/knabben/tkw/controllers/assets"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"github.com/knabben/tkw/pkg/config"
	errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
)

// getCredentials fetch the vsphere-cloud-config cm and extract data in the mapper
func (r *OSImageReconciler) getCredentials(ctx context.Context, cmap *config.Mapper) error {
	var (
		vsphereSM  = &v1.Secret{}
		namedspace = types.NamespacedName{Name: cmap.Get("secret-name"), Namespace: cmap.Get("secret-ns")}
	)

	vsphereCM, name := &v1.ConfigMap{}, "vsphere-cloud-config"
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: TKG_NAMESPACE}, vsphereCM); err != nil {
		return err
	}

	// Fetch vsphere-cloud-config and extract data
	data := vsphereCM.Data["vsphere.conf"]
	cmap.Set("vc", extractRValue(`\[VirtualCenter "(.*)"\]`, data))
	cmap.Set("secret-name", extractRValue(`secret-name = "(.*)"`, data))
	cmap.Set("secret-ns", extractRValue(`secret-namespace = "(.*)"`, data))

	if err := r.Get(ctx, namedspace, vsphereSM); err != nil {
		return err
	}
	vcIP := cmap.Get("vc")
	for _, s := range []string{"username", "password"} {
		cmap.Set(s, string(vsphereSM.Data[fmt.Sprintf("%s.%s", vcIP, s)]))
	}

	return nil
}

func (r *OSImageReconciler) getOrCreateWindowsResourceBundle(ctx context.Context, name types.NamespacedName) error {
	var (
		deployment *appsv1.Deployment
		err error
	)

	err = r.Get(ctx, name, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment := assets.YAMLAccessor[*appsv1.Deployment]{
			FileName: assets.BUILDER_DEPLOYMENT,
			SchemaGV: appsv1.SchemeGroupVersion,
		}
		var obj *appsv1.Deployment
		if obj, err = deployment.GetDecodedObject(); err != nil {
			return err
		}
		fmt.Println("do not exist getting now")
		fmt.Println(fmt.Sprintf("----- %v", obj))
	} else if err != nil {
		return fmt.Errorf("Error getting existing deployment.")
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
