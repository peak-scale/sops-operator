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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// SopsSecretStatus defines the observed state of SopsSecret
type SopsSecretStatus struct {
	// Amount of tenants selected by this translator
	//+kubebuilder:default=0
	Size uint `json:"size,omitempty"`
	// Secrets being replicated by this SopsSecret
	Secrets []*SopsSecretItemStatus `json:"secrets,omitempty"`
	// Conditions represent the latest available observations of an instances state
	Condition metav1.Condition `json:"condition,omitempty"`
}

// Get an instance current status
func (ms *SopsSecretStatus) updateStats() {
	ms.Size = uint(len(ms.Secrets))
}

// Get an instance current status
func (ms *SopsSecretStatus) GetInstance(stat *SopsSecretItemStatus) *SopsSecretItemStatus {
	for _, source := range ms.Secrets {
		if ms.instancequal(source, stat) {
			return source
		}
	}
	ms.updateStats()

	return nil
}

// Add/Update the status for a single instance
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

// Removes an instance
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
