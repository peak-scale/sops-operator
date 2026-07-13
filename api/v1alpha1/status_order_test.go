// Copyright 2024-2026 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/peak-scale/sops-operator/internal/api"
	capmeta "github.com/projectcapsule/capsule/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestSopsProviderStatusNormalize(t *testing.T) {
	t.Parallel()

	status := SopsProviderStatus{
		ProvidersAmount: 99,
		Providers: []*SopsProviderItemStatus{
			providerStatusItem("a", "z", "uid-3"),
			nil,
			providerStatusItem("z", "a", "uid-2"),
			providerStatusItem("z", "a", "uid-1"),
		},
		Conditions: capmeta.ConditionList{
			{Type: "Synced", Status: metav1.ConditionTrue},
			{Type: "Ready", Status: metav1.ConditionTrue},
		},
	}

	status.Normalize()

	require.Equal(t, uint(4), status.ProvidersAmount)
	require.Equal(t, []types.UID{"uid-1", "uid-2", "uid-3"}, []types.UID{
		status.Providers[0].UID,
		status.Providers[1].UID,
		status.Providers[2].UID,
	})
	require.Nil(t, status.Providers[3])
	require.Equal(t, "Ready", status.Conditions[0].Type)
	require.Equal(t, "Synced", status.Conditions[1].Type)
}

func TestSopsProviderStatusUpdateInstanceNormalizesExistingStatus(t *testing.T) {
	t.Parallel()

	first := providerStatusItem("z", "z", "uid-2")
	second := providerStatusItem("a", "a", "uid-1")
	status := SopsProviderStatus{Providers: []*SopsProviderItemStatus{first, second}}

	// Updating with an equivalent entry must still normalize status loaded from
	// an older, non-canonical version of the controller.
	status.UpdateInstance(first.DeepCopy())

	require.Equal(t, types.UID("uid-1"), status.Providers[0].UID)
	require.Equal(t, types.UID("uid-2"), status.Providers[1].UID)
	require.Equal(t, uint(2), status.ProvidersAmount)
}

func TestSopsSecretStatusNormalize(t *testing.T) {
	t.Parallel()

	status := SopsSecretStatus{
		Size: 99,
		Secrets: []*SopsSecretItemStatus{
			{Name: "z", Namespace: "z", UID: "uid-3"},
			nil,
			{Name: "z", Namespace: "a", UID: "uid-2"},
			{Name: "z", Namespace: "a", UID: "uid-1"},
		},
		Providers: []*api.Origin{
			{Name: "z", Namespace: "z", UID: "provider-3"},
			nil,
			{Name: "z", Namespace: "a", UID: "provider-2"},
			{Name: "z", Namespace: "a", UID: "provider-1"},
		},
		Conditions: capmeta.ConditionList{
			{Type: "Synced", Status: metav1.ConditionTrue},
			{Type: "Ready", Status: metav1.ConditionTrue},
		},
	}

	status.Normalize()

	require.Equal(t, uint(4), status.Size)
	require.Equal(t, []types.UID{"uid-1", "uid-2", "uid-3"}, []types.UID{
		status.Secrets[0].UID,
		status.Secrets[1].UID,
		status.Secrets[2].UID,
	})
	require.Nil(t, status.Secrets[3])
	require.Equal(t, []types.UID{"provider-1", "provider-2", "provider-3"}, []types.UID{
		status.Providers[0].UID,
		status.Providers[1].UID,
		status.Providers[2].UID,
	})
	require.Nil(t, status.Providers[3])
	require.Equal(t, "Ready", status.Conditions[0].Type)
	require.Equal(t, "Synced", status.Conditions[1].Type)
}

func TestSopsSecretStatusMutationMaintainsCanonicalOrder(t *testing.T) {
	t.Parallel()

	status := SopsSecretStatus{}
	status.UpdateInstance(&SopsSecretItemStatus{Name: "z", Namespace: "b", UID: "uid-2"})
	status.UpdateInstance(&SopsSecretItemStatus{Name: "a", Namespace: "a", UID: "uid-1"})
	status.UpdateInstance(&SopsSecretItemStatus{Name: "m", Namespace: "a", UID: "uid-3"})

	require.Equal(t, []string{"a/a", "a/m", "b/z"}, []string{
		status.Secrets[0].Namespace + "/" + status.Secrets[0].Name,
		status.Secrets[1].Namespace + "/" + status.Secrets[1].Name,
		status.Secrets[2].Namespace + "/" + status.Secrets[2].Name,
	})
	require.Equal(t, uint(3), status.Size)

	status.RemoveInstance(&SopsSecretItemStatus{Name: "m", Namespace: "a"})
	require.Equal(t, []string{"a", "z"}, []string{
		status.Secrets[0].Name,
		status.Secrets[1].Name,
	})
	require.Equal(t, uint(2), status.Size)
}

func providerStatusItem(name, namespace string, uid types.UID) *SopsProviderItemStatus {
	return &SopsProviderItemStatus{
		Origin: api.Origin{Name: name, Namespace: namespace, UID: uid},
		Condition: metav1.Condition{
			Type:   "Ready",
			Status: metav1.ConditionTrue,
			Reason: "Loaded",
		},
	}
}
