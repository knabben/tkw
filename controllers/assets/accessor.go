package assets

import (
	"embed"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	//go:embed manifests/*
	manifests embed.FS

	appsSchema = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsSchema)
)

func init() {
	if err := appsv1.AddToScheme(appsSchema); err != nil {
		panic(err)
	}
}
