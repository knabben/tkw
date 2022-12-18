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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OSImageSpec defines the desired state of OSImage
type OSImageSpec struct {
	WindowsISOPath      string `json:"windowsISOPath"`
	VMToolsPath         string `json:"vmtoolsPath"`

	// +kubebuilder:default:=dc0
	VsphereDatacenter   string `json:"vsphereDatacenter"`

	// +kubebuilder:default:=folder0
	VSphereFolder       string `json:"vsphereFolder"`

	// +kubebuilder:default:=sharedVmfs-0
	VSphereDataStore    string `json:"vsphereDatastore"`

	// +kubebuilder:default:=VM Network
	VSphereNetwork      string `json:"vsphereNetwork"`

	// +kubebuilder:default:=rp0
	VSphereResourcePool string `json:"vsphereResourcePool"`

	// +kubebuilder:default:=cluster0
	VSphereCluster      string `json:"vsphereCluster"`
}

// OSImageStatus defines the observed state of OSImage
type OSImageStatus struct {
	// OSTemplates are the OVA templates in the vSphere
	OSTemplates   []OSImageTemplates `json:"templates"`

	// Conditions holds a list of internal conditions of the operator
	Conditions []metav1.Condition `json:"conditions"`
}

type OSImageTemplates struct {
	Name                 string `json:"name,omitempty"`
	BuildDate            string `json:"buildDate,omitempty"`
	BuildTimestamp       string `json:"buildTimestamp,omitempty"`
	CNIVersion           string `json:"cniVersion,omitempty"`
	ContainerDVersion    string `json:"containerdVersion,omitempty"`
	DistroArch           string `json:"distroArch,omitempty"`
	DistroName           string `json:"distroName,omitempty"`
	DistroVersion        string `json:"distroVersion,omitempty"`
	ImageBuilderVersion  string `json:"imageBuilder,omitempty"`
	KubernetesSemVer     string `json:"kubernetesSemver,omitempty"`
	KubernetesSourceType string `json:"kubernetesSource,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OSImage is the Schema for the osimages API
type OSImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OSImageSpec   `json:"spec,omitempty"`
	Status OSImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OSImageList contains a list of OSImage
type OSImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OSImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OSImage{}, &OSImageList{})
}
