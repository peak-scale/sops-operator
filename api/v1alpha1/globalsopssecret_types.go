/*
Copyright 2024-2025 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"github.com/peak-scale/sops-operator/internal/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SopsSecretSpec defines the desired state of SopsSecret.
type GlobalSopsSecretSpec struct {
	// Define Secrets to replicate, when secret is decrypted
	Secrets []*GlobalSopsSecretItem `json:"secrets"`
}

// GlobalSopsSecretItem defines the desired state of GlobalSopsSecret.
type GlobalSopsSecretItem struct {
	// Namespace must be declared since this is a cluster scoped resource
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`

	SopsSecretItem `json:",inline"`
}

func (s *GlobalSopsSecret) GetSopsMetadata() *api.Metadata {
	return s.Sops
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Secrets",type="integer",JSONPath=".status.size",description="The amount of secrets being managed"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition.type",description="The actual state of the GlobalSopsSecret"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.condition.message",description="Condition Message"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"

// GlobalSopsSecret is the Schema for the globalsopssecrets API.
type GlobalSopsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalSopsSecretSpec `json:"spec,omitempty"`
	Status SopsSecretStatus     `json:"status,omitempty"`
	Sops   *api.Metadata        `json:"sops"`
}

// +kubebuilder:object:root=true

// GlobalSopsSecretList contains a list of GlobalSopsSecret.
type GlobalSopsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalSopsSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlobalSopsSecret{}, &GlobalSopsSecretList{})
}
