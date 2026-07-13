// Copyright 2024-2026 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"reflect"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	capmeta "github.com/projectcapsule/capsule/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// primaryResourcePredicate prevents status-only updates from reconciling the
// resource which just wrote them. Labels are included because providers may
// select SopsSecret and GlobalSopsSecret resources by label.
func primaryResourcePredicate() predicate.Predicate {
	return predicate.Or(
		predicate.GenerationChangedPredicate{},
		predicate.LabelChangedPredicate{},
	)
}

// sopsProviderStatusPredicate only fans provider updates out to secret
// controllers when the usable provider set or readiness changes. In
// particular, condition timestamps, messages, reasons and observed generation
// do not cause every SopsSecret to be reconciled.
func sopsProviderStatusPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event.CreateEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldProvider, oldOK := e.ObjectOld.(*sopsv1alpha1.SopsProvider)
			newProvider, newOK := e.ObjectNew.(*sopsv1alpha1.SopsProvider)
			if !oldOK || !newOK {
				return false
			}

			return providerStatusChanged(&oldProvider.Status, &newProvider.Status)
		},
		DeleteFunc:  func(event.DeleteEvent) bool { return true },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}
}

type providerReadiness struct {
	status  metav1.ConditionStatus
	present bool
}

type providerEntryState struct {
	origin api.Origin
	ready  providerReadiness
}

func providerStatusChanged(oldStatus, newStatus *sopsv1alpha1.SopsProviderStatus) bool {
	if oldStatus.ProvidersAmount != newStatus.ProvidersAmount ||
		len(oldStatus.Providers) != len(newStatus.Providers) {
		return true
	}

	oldProviders := providerEntryStates(oldStatus.Providers)
	newProviders := providerEntryStates(newStatus.Providers)
	if !reflect.DeepEqual(oldProviders, newProviders) {
		return true
	}

	return readyCondition(oldStatus) != readyCondition(newStatus)
}

func providerEntryStates(providers []*sopsv1alpha1.SopsProviderItemStatus) map[api.Origin]providerReadiness {
	states := make(map[api.Origin]providerReadiness, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}

		states[provider.Origin] = providerReadiness{
			status:  provider.Condition.Status,
			present: true,
		}
	}

	return states
}

func readyCondition(status *sopsv1alpha1.SopsProviderStatus) providerReadiness {
	for _, condition := range status.Conditions {
		if condition.Type == capmeta.ReadyCondition {
			return providerReadiness{status: condition.Status, present: true}
		}
	}

	return providerReadiness{}
}
