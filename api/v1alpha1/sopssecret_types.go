// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/peak-scale/sops-operator/internal/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SopsSecretSpec defines the desired state of SopsSecret.
type SopsSecretSpec struct {
	// Define Secrets to replicate, when secret is decrypted
	Secrets []*SopsSecretItem `json:"secrets"`
	// Define additional Metadata for the generated secrets
	Metadata SecretMetadata `json:"metadata,omitempty"`
}

// SopsSecretTemplate defines the map of secrets to create
// +kubebuilder:object:root=false
type SopsSecretItem struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,11,rep,name=labels"`
	// Kubernetes secret type.
	// Defaults to Opaque.
	// Allowed values:
	// - Opaque
	// - kubernetes.io/service-account-token
	// - kubernetes.io/dockercfg
	// - kubernetes.io/dockerconfigjson
	// - kubernetes.io/basic-auth
	// - kubernetes.io/ssh-auth
	// - kubernetes.io/tls
	// - bootstrap.kubernetes.io/token
	// +kubebuilder:validation:Enum=Opaque;kubernetes.io/service-account-token;kubernetes.io/dockercfg;kubernetes.io/dockerconfigjson;kubernetes.io/basic-auth;kubernetes.io/ssh-auth;kubernetes.io/tls;bootstrap.kubernetes.io/token
	Type corev1.SecretType `json:"type,omitempty"`
	// Data map to use in Kubernetes secret (equivalent to Kubernetes Secret object data, please see for more
	// information: https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets)
	//+optional
	Data map[string]string `json:"data,omitempty"`
	// stringData map to use in Kubernetes secret (equivalent to Kubernetes Secret object stringData, please see for more
	// information: https://kubernetes.io/docs/concepts/configuration/secret/#overview-of-secrets)
	//+optional
	StringData map[string]string `json:"stringData,omitempty"`
	// Immutable, if set to true, ensures that data stored in the Secret cannot
	// be updated (only object metadata can be modified).
	// If not set to true, the field can be modified at any time.
	// Defaulted to nil.
	// +optional
	Immutable *bool `json:"immutable,omitempty" protobuf:"varint,5,opt,name=immutable"`
}

func (s *SopsSecret) GetSopsMetadata() *api.Metadata {
	return s.Sops
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Secrets",type="integer",JSONPath=".status.size",description="The amount of secrets being managed"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition.type",description="The actual state of the SopsSecret"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.condition.message",description="Condition Message"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"

// SopsSecret is the Schema for the sopssecrets API.
type SopsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SopsSecretSpec   `json:"spec,omitempty"`
	Status SopsSecretStatus `json:"status,omitempty"`
	Sops   *api.Metadata    `json:"sops"`
}

// +kubebuilder:object:root=true

// SopsSecretList contains a list of SopsSecret.
type SopsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SopsSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SopsSecret{}, &SopsSecretList{})
}
