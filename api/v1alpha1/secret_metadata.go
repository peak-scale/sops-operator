/*
Copyright 2024-2025 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

// SopsSecretSpec defines the desired state of SopsSecret.
type SecretMetadata struct {
	// Prefix added to all generated Secrets names
	Prefix string `json:"prefix,omitempty"`
	// Suffix added to all generated Secrets names
	Suffix string `json:"suffix,omitempty"`
	// Labels added to all generated Secrets
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations added to all generated Secrets
	Annotations map[string]string `json:"annotations,omitempty"`
}
