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
}

// OSImageStatus defines the observed state of OSImage
type OSImageStatus struct {
	OSTemplates []OSImageTemplates
}

type OSImageTemplates struct {
	BuildDate            string `json:"build_date,omitempty"`
	BuildTimestamp       string `json:"build_timestamp,omitempty"`
	CNIVersion           string `json:"cni_version,omitempty"`
	ContainerDVersion    string `json:"containerd_version,omitempty"`
	DistroArch           string `json:"distro_arch,omitempty"`
	DistroName           string `json:"distro_name,omitempty"`
	DistroVersion        string `json:"distro_version,omitempty"`
	ImageBuilderVersion  string `json:"image_builder_version,omitempty"`
	KubernetesSemVer     string `json:"kubernetes_sem_ver,omitempty"`
	KubernetesSourceType string `json:"kubernetes_source_type,omitempty"`
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
