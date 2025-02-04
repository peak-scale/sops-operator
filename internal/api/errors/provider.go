// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NoDecryptionProvider struct {
	Object client.Object
}

func (e *NoDecryptionProvider) Error() string {
	return fmt.Sprintf("secret %s/%s has no decryption providers", e.Object.GetNamespace(), e.Object.GetName())
}

func NewNoDecryptionProviderError(obj client.Object) error {
	return &NoDecryptionProvider{Object: obj}
}
