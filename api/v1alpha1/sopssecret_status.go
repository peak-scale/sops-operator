// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"cmp"
	"slices"

	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/projectcapsule/capsule/pkg/api/meta"
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
	// Providers used on this secret
	Providers []*api.Origin `json:"providers,omitempty"`
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
func (ms *SopsSecretStatus) GetInstance(stat *SopsSecretItemStatus) *SopsSecretItemStatus {
	for _, source := range ms.Secrets {
		if ms.instancequal(source, stat) {
			return source
		}
	}

	ms.Normalize()

	return nil
}

// Add/Update the status for a single instance.
func (ms *SopsSecretStatus) UpdateInstance(stat *SopsSecretItemStatus) {
	// Check if the tenant is already present in the status
	for i, source := range ms.Secrets {
		if ms.instancequal(source, stat) {
			ms.Secrets[i] = stat
			ms.Normalize()

			return
		}
	}

	// If tenant not found, append it to the list
	ms.Secrets = append(ms.Secrets, stat)
	ms.Normalize()
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
	ms.Normalize()
}

// Normalize puts all status lists into a canonical order.
func (ms *SopsSecretStatus) Normalize() {
	slices.SortStableFunc(ms.Secrets, func(a, b *SopsSecretItemStatus) int {
		if a == nil {
			if b == nil {
				return 0
			}

			return 1
		}

		if b == nil {
			return -1
		}

		if order := cmp.Compare(a.Namespace, b.Namespace); order != 0 {
			return order
		}

		if order := cmp.Compare(a.Name, b.Name); order != 0 {
			return order
		}

		return cmp.Compare(a.UID, b.UID)
	})
	slices.SortStableFunc(ms.Providers, func(a, b *api.Origin) int {
		if a == nil {
			if b == nil {
				return 0
			}

			return 1
		}

		if b == nil {
			return -1
		}

		return compareOrigins(*a, *b)
	})
	slices.SortStableFunc(ms.Conditions, func(a, b meta.Condition) int {
		return cmp.Compare(a.Type, b.Type)
	})

	ms.Size = uint(len(ms.Secrets))
}

func compareOrigins(a, b api.Origin) int {
	if order := cmp.Compare(a.Namespace, b.Namespace); order != 0 {
		return order
	}

	if order := cmp.Compare(a.Name, b.Name); order != 0 {
		return order
	}

	return cmp.Compare(a.UID, b.UID)
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
