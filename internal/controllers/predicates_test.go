// Copyright 2024-2026 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	capmeta "github.com/projectcapsule/capsule/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestSopsProviderStatusPredicate(t *testing.T) {
	t.Parallel()

	base := providerWithStatus(
		providerItem("key", "keys", "uid-1", metav1.ConditionTrue),
		metav1.ConditionTrue,
	)

	tests := map[string]struct {
		mutate  func(*sopsv1alpha1.SopsProvider)
		changed bool
	}{
		"unchanged": {
			changed: false,
		},
		"timestamps messages and reasons are ignored": {
			mutate: func(provider *sopsv1alpha1.SopsProvider) {
				provider.Status.Providers[0].LastTransitionTime = metav1.NewTime(time.Now().Add(time.Hour))
				provider.Status.Providers[0].Message = "new item message"
				provider.Status.Providers[0].Reason = "new item reason"
				provider.Status.Conditions[0].LastTransitionTime = metav1.NewTime(time.Now().Add(time.Hour))
				provider.Status.Conditions[0].Message = "new overall message"
				provider.Status.Conditions[0].Reason = "new overall reason"
				provider.Status.ObservedGeneration++
			},
			changed: false,
		},
		"reported amount changed": {
			mutate: func(provider *sopsv1alpha1.SopsProvider) {
				provider.Status.ProvidersAmount++
			},
			changed: true,
		},
		"provider added even with stale amount": {
			mutate: func(provider *sopsv1alpha1.SopsProvider) {
				provider.Status.Providers = append(provider.Status.Providers,
					providerItem("key-2", "keys", "uid-2", metav1.ConditionTrue))
			},
			changed: true,
		},
		"provider uid changed": {
			mutate: func(provider *sopsv1alpha1.SopsProvider) {
				provider.Status.Providers[0].UID = types.UID("uid-2")
			},
			changed: true,
		},
		"provider readiness changed": {
			mutate: func(provider *sopsv1alpha1.SopsProvider) {
				provider.Status.Providers[0].Status = metav1.ConditionFalse
			},
			changed: true,
		},
		"overall readiness changed": {
			mutate: func(provider *sopsv1alpha1.SopsProvider) {
				provider.Status.Conditions[0].Status = metav1.ConditionFalse
			},
			changed: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			oldProvider := base.DeepCopy()
			newProvider := base.DeepCopy()
			if tt.mutate != nil {
				tt.mutate(newProvider)
			}

			got := sopsProviderStatusPredicate().Update(event.UpdateEvent{
				ObjectOld: oldProvider,
				ObjectNew: newProvider,
			})
			require.Equal(t, tt.changed, got)
		})
	}
}

func TestSopsProviderStatusPredicateLifecycleEvents(t *testing.T) {
	t.Parallel()

	p := &sopsv1alpha1.SopsProvider{}
	predicate := sopsProviderStatusPredicate()
	require.True(t, predicate.Create(event.CreateEvent{Object: p}))
	require.True(t, predicate.Delete(event.DeleteEvent{Object: p}))
	require.False(t, predicate.Generic(event.GenericEvent{Object: p}))
}

func TestPrimaryResourcePredicate(t *testing.T) {
	t.Parallel()

	oldProvider := &sopsv1alpha1.SopsProvider{ObjectMeta: metav1.ObjectMeta{
		Generation: 1,
		Labels:     map[string]string{"environment": "prod"},
	}}

	t.Run("status only update", func(t *testing.T) {
		newProvider := oldProvider.DeepCopy()
		newProvider.Status.ObservedGeneration = 1
		require.False(t, primaryResourcePredicate().Update(event.UpdateEvent{
			ObjectOld: oldProvider,
			ObjectNew: newProvider,
		}))
	})

	t.Run("generation update", func(t *testing.T) {
		newProvider := oldProvider.DeepCopy()
		newProvider.Generation++
		require.True(t, primaryResourcePredicate().Update(event.UpdateEvent{
			ObjectOld: oldProvider,
			ObjectNew: newProvider,
		}))
	})

	t.Run("label update", func(t *testing.T) {
		newProvider := oldProvider.DeepCopy()
		newProvider.Labels["environment"] = "staging"
		require.True(t, primaryResourcePredicate().Update(event.UpdateEvent{
			ObjectOld: oldProvider,
			ObjectNew: newProvider,
		}))
	})
}

func providerWithStatus(
	item *sopsv1alpha1.SopsProviderItemStatus,
	ready metav1.ConditionStatus,
) *sopsv1alpha1.SopsProvider {
	return &sopsv1alpha1.SopsProvider{
		Status: sopsv1alpha1.SopsProviderStatus{
			ProvidersAmount: 1,
			Providers:       []*sopsv1alpha1.SopsProviderItemStatus{item},
			Conditions: capmeta.ConditionList{{
				Type:               capmeta.ReadyCondition,
				Status:             ready,
				Reason:             capmeta.SucceededReason,
				Message:            "reconciled",
				LastTransitionTime: metav1.Now(),
			}},
		},
	}
}

func providerItem(
	name string,
	namespace string,
	uid types.UID,
	ready metav1.ConditionStatus,
) *sopsv1alpha1.SopsProviderItemStatus {
	return &sopsv1alpha1.SopsProviderItemStatus{
		Origin: api.Origin{Name: name, Namespace: namespace, UID: uid},
		Condition: metav1.Condition{
			Type:               "Ready",
			Status:             ready,
			Reason:             "Loaded",
			Message:            "loaded",
			LastTransitionTime: metav1.Now(),
		},
	}
}
