// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"cmp"
	"slices"

	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/projectcapsule/capsule/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SopsProviderStatus defines the observed state of SopsProvider.
type SopsProviderStatus struct {
	// Amount of providers
	//+kubebuilder:default=0
	ProvidersAmount uint `json:"size,omitempty"`
	// List Validated Providers
	Providers []*SopsProviderItemStatus `json:"providers,omitempty"`
	// Conditions
	Conditions meta.ConditionList `json:"conditions"`
	// ObservedGeneration is the most recent generation the controller has observed.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Deprecated: use conditions as list
	//
	// Conditions represent the latest available observations of an instances state
	// +optional
	Condition metav1.Condition `json:"condition,omitzero"`
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
			if source.Type == stat.Type &&
				source.Status == stat.Status &&
				source.Reason == stat.Reason && source.Message == stat.Message {
				ms.Normalize()

				return
			}

			ms.Providers[i] = stat
			ms.Normalize()

			return
		}
	}

	ms.Providers = append(ms.Providers, stat)
	ms.Normalize()
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
	ms.Normalize()
}

// Normalize puts all status lists into a canonical order.
func (ms *SopsProviderStatus) Normalize() {
	slices.SortStableFunc(ms.Providers, func(a, b *SopsProviderItemStatus) int {
		if a == nil {
			if b == nil {
				return 0
			}

			return 1
		}

		if b == nil {
			return -1
		}

		return compareOrigins(a.Origin, b.Origin)
	})
	slices.SortStableFunc(ms.Conditions, func(a, b meta.Condition) int {
		return cmp.Compare(a.Type, b.Type)
	})

	ms.ProvidersAmount = uint(len(ms.Providers))
}

func (ms *SopsProviderStatus) instancequal(a, b *SopsProviderItemStatus) bool {
	return a.Origin == b.Origin
}

type SopsProviderItemStatus struct {
	// Conditions represent the latest available observations of an instances state
	// +optional
	metav1.Condition `json:"condition,omitzero"`
	// The Origin this Provider originated from
	api.Origin `json:",inline"`
}
