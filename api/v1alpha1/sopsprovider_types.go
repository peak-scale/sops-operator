/*
Copyright 2025.

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
	"github.com/peak-scale/sops-operator/internal/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SopsProviderSpec defines the desired state of SopsProvider.
type SopsProviderSpec struct {
	// Selector Referencing which Secrets can be encrypted by this provider
	// This selects effective SOPS Secrets
	SOPSSelectors []*api.NamespacedSelector `json:"sops"`

	// Select namespaces or secrets where decryption information for this
	// provider can be sourced from
	ProviderSecrets []*SopsProviderSelector `json:"providers"`
}

type SopsProviderSelector struct {
	// Select namespaces or secrets where decryption information for this
	// provider can be sourced from
	*api.NamespacedSelector `json:",omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// SopsProvider is the Schema for the sopsproviders API.
type SopsProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SopsProviderSpec   `json:"spec,omitempty"`
	Status SopsProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SopsProviderList contains a list of SopsProvider.
type SopsProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SopsProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SopsProvider{}, &SopsProviderList{})
}
