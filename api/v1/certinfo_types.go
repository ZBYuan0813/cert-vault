/*
Copyright 2020 ms.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertInfoSpec defines the desired state of CertInfo
type CertInfoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CertInfo. Edit CertInfo_types.go to remove/update
	RoleName         string   `json:"role"`
	Allowed_Domains  []string `json:"allowed_domain,omitempty"`
	Allow_subdomains bool     `json:"allow_subdomains,omitempty"`
	Allow_Any_Name   bool     `json:"allow_any_name,omitempty"`
	Organization     string   `json:"organization,omitempty"`
	Ou               string   `json:"ou"`
	Max_TTL          string   `json:"max_ttl"`
	CommonName       string   `json:"common_name"`
}

// CertInfoStatus defines the observed state of CertInfo
type CertInfoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// CertInfo is the Schema for the certinfoes API
type CertInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertInfoSpec   `json:"spec,omitempty"`
	Status CertInfoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertInfoList contains a list of CertInfo
type CertInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertInfo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertInfo{}, &CertInfoList{})
}
