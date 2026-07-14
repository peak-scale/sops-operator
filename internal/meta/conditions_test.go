// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	"testing"

	"github.com/peak-scale/sops-operator/internal/meta"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewNotReadyConditionUsesReadyType(t *testing.T) {
	t.Parallel()

	obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Generation: 3}}
	c := meta.NewNotReadyCondition(obj, "something failed")

	require.Equal(t, meta.ReadyCondition, c.Type,
		"NewNotReadyCondition must emit type %q, not a separate NotReady type", meta.ReadyCondition)
	require.Equal(t, metav1.ConditionFalse, c.Status)
	require.Equal(t, meta.FailedReason, c.Reason)
	require.Equal(t, "something failed", c.Message)
	require.Equal(t, int64(3), c.ObservedGeneration)
}

func TestNewNotReadySecretStatusConditionUsesReadyType(t *testing.T) {
	t.Parallel()

	obj := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:       "my-secret",
		Namespace:  "default",
		Generation: 5,
	}}
	item := meta.NewNotReadySecretStatusCondition(obj, "decryption failed")

	require.Equal(t, meta.ReadyCondition, item.Condition.Type,
		"NewNotReadySecretStatusCondition must emit type %q, not a separate NotReady type", meta.ReadyCondition)
	require.Equal(t, metav1.ConditionFalse, item.Condition.Status)
	require.Equal(t, meta.FailedReason, item.Condition.Reason)
	require.Equal(t, "decryption failed", item.Condition.Message)
	require.Equal(t, int64(5), item.Condition.ObservedGeneration)
	require.Equal(t, "my-secret", item.Name)
	require.Equal(t, "default", item.Namespace)
}
