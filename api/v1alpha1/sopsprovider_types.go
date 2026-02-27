// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

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
	ProviderSecrets []*api.NamespacedSelector `json:"keys"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition.type",description="Status"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.condition.message",description="Message"
// +kubebuilder:printcolumn:name="Providers",type="integer",JSONPath=".status.size",description="Amount of providers"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// SopsProvider is the Schema for the sopsproviders API.
type SopsProvider struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec SopsProviderSpec `json:"spec"`
	// +optional
	Status SopsProviderStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// SopsProviderList contains a list of SopsProvider.
type SopsProviderList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitzero"`

	Items []SopsProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SopsProvider{}, &SopsProviderList{})
}
