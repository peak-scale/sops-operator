// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/peak-scale/sops-operator/internal/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// SopsSecretStatus defines the observed state of SopsSecret.
type SopsSecretStatus struct {
	// Amount of Secrets
	//+kubebuilder:default=0
	Size uint `json:"size,omitempty"`
	// Secrets being replicated by this SopsSecret
	Secrets []*SopsSecretItemStatus `json:"secrets,omitempty"`
	// Conditions represent the latest available observations of an instances state
	Condition metav1.Condition `json:"condition,omitempty"`
	// Providers used on this secret
	Providers []*api.Origin `json:"providers,omitempty"`
}

// Get an instance current status.
func (ms *SopsSecretStatus) GetInstance(stat *SopsSecretItemStatus) *SopsSecretItemStatus {
	for _, source := range ms.Secrets {
		if ms.instancequal(source, stat) {
			return source
		}
	}

	ms.updateStats()

	return nil
}

// Add/Update the status for a single instance.
func (ms *SopsSecretStatus) UpdateInstance(stat *SopsSecretItemStatus) {
	// Check if the tenant is already present in the status
	for i, source := range ms.Secrets {
		if ms.instancequal(source, stat) {
			ms.Secrets[i] = stat

			return
		}
	}

	// If tenant not found, append it to the list
	ms.Secrets = append(ms.Secrets, stat)
	ms.updateStats()
}

// Removes an instance.
func (ms *SopsSecretStatus) RemoveInstance(stat *SopsSecretItemStatus) {
	// Filter out the datasource with given UID
	filter := []*SopsSecretItemStatus{}

	for _, source := range ms.Secrets {
		if !ms.instancequal(source, stat) {
			filter = append(filter, source)
		}
	}

	// Update the tenants and adjust the size
	ms.Secrets = filter
	ms.updateStats()
}

// Get an instance current status.
func (ms *SopsSecretStatus) updateStats() {
	ms.Size = uint(len(ms.Secrets))
}

func (ms *SopsSecretStatus) instancequal(a, b *SopsSecretItemStatus) bool {
	if a.Name == b.Name && a.Namespace == b.Namespace {
		return true
	}

	return false
}

type SopsSecretItemStatus struct {
	Condition metav1.Condition `json:"condition"`
	Name      string           `json:"name"`
	Namespace string           `json:"namespace"`
	UID       k8stypes.UID     `json:"uid,omitempty"`
}
