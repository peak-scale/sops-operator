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

// SopsProviderStatus defines the observed state of SopsProvider
type SopsProviderStatus struct {
	// List Validated Providers
	Providers []*SopsProviderItemStatus `json:"providers,omitempty"`
	// Conditions represent the latest available observations of an instances state
	Condition metav1.Condition `json:"condition,omitempty"`
}

// Get an instance current status
func (ms *SopsProviderStatus) GetInstance(stat *SopsProviderItemStatus) *SopsProviderItemStatus {
	for _, source := range ms.Providers {
		if ms.instancequal(source, stat) {
			return source
		}
	}
	return nil
}

// Add/Update the status for a single instance
func (ms *SopsProviderStatus) UpdateInstance(stat *SopsProviderItemStatus) {
	// Check if the tenant is already present in the status
	for i, source := range ms.Providers {
		if ms.instancequal(source, stat) {
			ms.Providers[i] = stat
			return
		}
	}

	// If tenant not found, append it to the list
	ms.Providers = append(ms.Providers, stat)
}

// Removes an instance
func (ms *SopsProviderStatus) RemoveInstance(stat *SopsProviderItemStatus) {
	// Filter out the datasource with given UID
	filter := []*SopsProviderItemStatus{}
	for _, source := range ms.Providers {
		if !ms.instancequal(source, stat) {
			filter = append(filter, source)
		}
	}

	// Update the tenants and adjust the size
	ms.Providers = filter
}

func (ms *SopsProviderStatus) instancequal(a, b *SopsProviderItemStatus) bool {
	if a.Origin == b.Origin {
		return true
	}

	return false
}

type SopsProviderItemStatus struct {
	// Conditions represent the latest available observations of an instances state
	metav1.Condition `json:",inline"`
	// The Origin this Provider origaniated from
	api.Origin `json:",inline"`
}
