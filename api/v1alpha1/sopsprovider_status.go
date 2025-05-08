/*
Copyright 2024 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"github.com/peak-scale/sops-operator/internal/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SopsProviderStatus defines the observed state of SopsProvider.
type SopsProviderStatus struct {
	// Amount of providers
	//+kubebuilder:default=0
	ProvidersAmount uint `json:"size,omitempty"`
	// List Validated Providers
	Providers []*SopsProviderItemStatus `json:"providers,omitempty"`
	// Conditions represent the latest available observations of an instances state
	Condition metav1.Condition `json:"condition,omitempty"`
}

// Get an instance current status.
func (ms *SopsProviderStatus) GetInstance(stat *SopsProviderItemStatus) *SopsProviderItemStatus {
	for _, source := range ms.Providers {
		if ms.instancequal(source, stat) {
			return source
		}
	}

	return nil
}

// Add/Update the status for a single instance.
func (ms *SopsProviderStatus) UpdateInstance(stat *SopsProviderItemStatus) {
	for i, source := range ms.Providers {
		if ms.instancequal(source, stat) {
			ms.Providers[i] = stat

			return
		}
	}

	ms.Providers = append(ms.Providers, stat)
	ms.updateStats()
}

// Removes an instance.
func (ms *SopsProviderStatus) RemoveInstance(stat *SopsProviderItemStatus) {
	filter := []*SopsProviderItemStatus{}

	for _, source := range ms.Providers {
		if !ms.instancequal(source, stat) {
			filter = append(filter, source)
		}
	}

	ms.Providers = filter
	ms.updateStats()
}

// Get an instance current status.
func (ms *SopsProviderStatus) updateStats() *SopsProviderItemStatus {
	ms.ProvidersAmount = uint(len(ms.Providers))

	return nil
}

func (ms *SopsProviderStatus) instancequal(a, b *SopsProviderItemStatus) bool {
	if a.Origin == b.Origin {
		return true
	}

	return false
}

type SopsProviderItemStatus struct {
	// Conditions represent the latest available observations of an instances state
	metav1.Condition `json:"condition,omitempty"`
	// The Origin this Provider originated from
	api.Origin `json:",inline"`
}
