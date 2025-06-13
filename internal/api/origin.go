// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

type Origin struct {
	// Name of Object
	Name string `json:"name"`
	// namespace of Object
	Namespace string `json:"namespace,omitempty"`
	// namespace of Object
	UID k8stypes.UID `json:"uid,omitempty"`
}

func NewOrigin(obj metav1.Object) *Origin {
	return &Origin{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		UID:       obj.GetUID(),
	}
}
