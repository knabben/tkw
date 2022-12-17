package assets

import (
	"embed"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	//go:embed manifests/*
	manifests embed.FS

	appsSchema = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsSchema)
)

const (
	BUILDER_DEPLOYMENT = "manifests/builder-deployment.yaml"
	BUILDER_SERVICE    = "manifests/builder-svc.yaml"
	BUILDER_NAMESPACE  = "manifests/builder-ns.yaml"
)

func init() {
	if err := v1.AddToScheme(appsSchema); err != nil {
		panic(err)
	}
	if err := appsv1.AddToScheme(appsSchema); err != nil {
		panic(err)
	}
}

// ObjectTypes defines the generic types available
type ObjectTypes interface {
	*appsv1.Deployment | *v1.Service | *v1.Namespace
}

// YAMLAccessor implement the definition of YAML accessor
type YAMLAccessor[O ObjectTypes] struct {
	FileName string
	SchemaGV schema.GroupVersion
}

// GetDecodedObject returns the generic unmarshalled object from filename.
func (y *YAMLAccessor[O]) GetDecodedObject(fileName string, sc schema.GroupVersion) (O, error) {
	if y.FileName == "" {
		y.FileName = fileName
		y.SchemaGV = sc
	}

	bytes, err := manifests.ReadFile(y.FileName)
	if err != nil {
		return nil, err
	}

	obj, err := runtime.Decode(appsCodecs.UniversalDecoder(y.SchemaGV), bytes)
	if err != nil {
		return nil, err
	}
	return obj.(O), nil
}