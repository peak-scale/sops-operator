// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package errors

type SecretReconciliationError struct {
	Message string
}

func (e *SecretReconciliationError) Error() string {
	return e.Message
}

func NewSecretReconciliationError(message string) error {
	return &SecretReconciliationError{
		Message: message,
	}
}
