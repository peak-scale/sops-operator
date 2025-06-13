// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NoDecryptionProviderError struct {
	Object client.Object
}

func (e *NoDecryptionProviderError) Error() string {
	return fmt.Sprintf("secret %s/%s has no decryption providers", e.Object.GetNamespace(), e.Object.GetName())
}

func NewNoDecryptionProviderError(obj client.Object) error {
	return &NoDecryptionProviderError{Object: obj}
}
