// Copyright 2024-2026 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	capmeta "github.com/projectcapsule/capsule/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSopsSecretUpdateStatusPostsReadiness(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		reconcileError error
		wantStatus     metav1.ConditionStatus
		wantReason     string
		wantMessage    string
	}{
		"ready": {
			wantStatus:  metav1.ConditionTrue,
			wantReason:  capmeta.SucceededReason,
			wantMessage: "Secrets Decrypted",
		},
		"not ready": {
			reconcileError: errors.New("decryption failed"),
			wantStatus:     metav1.ConditionFalse,
			wantReason:     capmeta.FailedReason,
			wantMessage:    "decryption failed",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			scheme := runtime.NewScheme()
			require.NoError(t, sopsv1alpha1.AddToScheme(scheme))

			stored := &sopsv1alpha1.SopsSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "secret",
					Namespace:  "default",
					Generation: 7,
				},
			}
			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithStatusSubresource(&sopsv1alpha1.SopsSecret{}).
				WithObjects(stored).
				Build()

			instance := stored.DeepCopy()
			instance.Status.Condition = metav1.Condition{Type: "NotReady", Status: metav1.ConditionFalse}
			instance.Status.Secrets = []*sopsv1alpha1.SopsSecretItemStatus{
				{Name: "z", Namespace: "default"},
				{Name: "a", Namespace: "default"},
			}

			reconciler := &SopsSecretReconciler{Client: client}
			require.NoError(t, reconciler.updateStatus(context.Background(), tt.reconcileError, instance))

			updated := &sopsv1alpha1.SopsSecret{}
			require.NoError(t, client.Get(context.Background(), types.NamespacedName{
				Name: "secret", Namespace: "default",
			}, updated))

			condition := updated.Status.Conditions.GetConditionByType(capmeta.ReadyCondition)
			require.NotNil(t, condition)
			require.Equal(t, tt.wantStatus, condition.Status)
			require.Equal(t, tt.wantReason, condition.Reason)
			require.Equal(t, tt.wantMessage, condition.Message)
			require.Equal(t, int64(7), condition.ObservedGeneration)
			require.Equal(t, int64(7), updated.Status.ObservedGeneration)
			require.Empty(t, updated.Status.Condition)
			require.Equal(t, []string{"a", "z"}, []string{
				updated.Status.Secrets[0].Name,
				updated.Status.Secrets[1].Name,
			})
		})
	}
}

func TestSopsSecretUpdateStatusReturnsFetchError(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))

	reconciler := &SopsSecretReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
	}
	err := reconciler.updateStatus(context.Background(), nil, &sopsv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{Name: "missing", Namespace: "default"},
	})
	require.Error(t, err)
}
