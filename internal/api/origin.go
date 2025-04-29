// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Origin struct {
	// Name of Object
	Name string `json:"name"`
	// namespace of Object
	Namespace string `json:"namespace,omitempty"`
}

func NewOrigin(obj metav1.Object) *Origin {
	return &Origin{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}
