package assets_test

import (
	"github.com/knabben/tkw/controllers/assets"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestAccessor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accessor Tests")
}

var _ = Describe("Object decoding", func() {
	Describe("Having a manifest", func() {
		Context("Of type deployment", func() {
			It("it should decode the object correctly", func() {
				accessor := assets.YAMLAccessor[*appsv1.Deployment]{}
				deployment, err := accessor.GetDecodedObject(assets.BUILDER_DEPLOYMENT, appsv1.SchemeGroupVersion)

				Expect(err).To(BeNil())
				Expect(deployment.Name).To(Equal("windows-resource-kit"))
				Expect(deployment.Namespace).To(Equal("windows"))
				Expect(len(deployment.Spec.Template.Spec.Containers)).To(Equal(1))
			})
		})
		Context("Of type service", func() {
			It("it should decode the object correctly", func() {
				accessor := assets.YAMLAccessor[*v1.Service]{}
				service, err := accessor.GetDecodedObject(assets.BUILDER_SERVICE, v1.SchemeGroupVersion)

				Expect(err).To(BeNil())
				Expect(service.Name).To(Equal("windows-resource"))
				Expect(service.Namespace).To(Equal("windows"))
				Expect(len(service.Spec.Ports)).To(Equal(1))
			})
		})
	})
})
