package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const configMapData = `
	[Global]
		secret-name = "cloud-provider-vsphere-credentials"
		secret-namespace = "kube-system"
		insecure-flag = "1"
	[VirtualCenter "10.0.0.1"]
		datacenters = "/dc0"
		insecure-flag = "1"
`

var _ = Describe("ConfigMap parsing", func() {
	Describe("Having a defined TOML file", func() {
		Context("with separated categories", func() {
			It("should find the VC", func() {
				value := extractRValue(`\[VirtualCenter "(.*)"\]`, configMapData)
				Expect(value).To(Equal("10.0.0.1"))
			})
			It("should find the secret namespace", func() {
				value := extractRValue(`secret-namespace = "(.*)"`, configMapData)
				Expect(value).To(Equal("kube-system"))
			})
			It("should find the secret name", func() {
				value := extractRValue(`secret-name = "(.*)"`, configMapData)
				Expect(value).To(Equal("cloud-provider-vsphere-credentials"))
			})
			It("not find value must be empty", func() {
				value := extractRValue(`not-existing = "(.*)"`, configMapData)
				Expect(value).To(Equal(""))
			})
		})
	})
})
