/*
Copyright 2024 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"github.com/peak-scale/sops-operator/internal/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// SopsProvider is the Schema for the sopsproviders API.
type SopsProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SopsProviderSpec   `json:"spec,omitempty"`
	Status SopsProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SopsProviderList contains a list of SopsProvider.
type SopsProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SopsProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SopsProvider{}, &SopsProviderList{})
}
