// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package decryptor

import "fmt"

type MissingKubernetesSecretError struct {
	Secret    string
	Namespace string
}

func (e *MissingKubernetesSecretError) Error() string {
	return fmt.Sprintf("Secret not found: %s/%s", e.Namespace, e.Secret)
}
