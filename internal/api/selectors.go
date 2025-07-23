// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Selector for resources and their labels or selecting origin namespaces
// +kubebuilder:object:generate=true
type NamespacedSelector struct {
	// Select Items based on their labels. If the namespaceSelector is also set, the selector is applied
	// to items within the selected namespaces. Otherwise for all the items.
	*metav1.LabelSelector `json:",inline"`
	// NamespaceSelector for filtering namespaces by labels where items can be located in
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

// GetMatchingNamespaces retrieves the list of namespaces that match the NamespaceSelector.
func (s *NamespacedSelector) GetMatchingNamespaces(
	ctx context.Context,
	client client.Client,
) ([]corev1.Namespace, error) {
	if s.NamespaceSelector == nil {
		return nil, nil // No namespace selector means all namespaces
	}

	nsSelector, err := metav1.LabelSelectorAsSelector(s.NamespaceSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid namespace selector: %w", err)
	}

	namespaceList := &corev1.NamespaceList{}
	if err := client.List(ctx, namespaceList); err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var matchingNamespaces []corev1.Namespace

	for _, ns := range namespaceList.Items {
		if nsSelector.Matches(labels.Set(ns.Labels)) {
			matchingNamespaces = append(matchingNamespaces, ns)
		}
	}

	return matchingNamespaces, nil
}

// Pass A Kubernetes Object to verify it matches.
func (s *NamespacedSelector) SingleMatch(
	ctx context.Context,
	client client.Client,
	obj metav1.Object,
) (state bool, err error) {
	if s == nil {
		return true, nil
	}

	if obj.GetNamespace() != "" {
		// Get namespaces matching NamespaceSelector
		matchingNamespaces, err := s.GetMatchingNamespaces(ctx, client)
		if err != nil {
			return false, fmt.Errorf("return 1: %w", err)
		}

		namespaceSet := make(map[string]bool)
		for _, ns := range matchingNamespaces {
			namespaceSet[ns.Name] = true
		}

		// If NamespaceSelector is set, ensure the object's namespace is included
		if len(namespaceSet) > 0 && !namespaceSet[obj.GetNamespace()] {
			return false, nil
		}
	}

	var objSelector labels.Selector
	if s.LabelSelector != nil {
		objSelector, err = metav1.LabelSelectorAsSelector(s.LabelSelector)
		if err != nil {
			return false, fmt.Errorf("invalid object selector: %w", err)
		}
	}

	if objSelector == nil {
		return true, nil
	}

	// If Selector is set, ensure the object matches the labels
	if !objSelector.Matches(labels.Set(obj.GetLabels())) {
		return false, nil
	}

	return true, nil
}

func (s *NamespacedSelector) MatchObjects(
	ctx context.Context,
	client client.Client,
	objects []metav1.Object,
) ([]metav1.Object, error) {
	if s == nil {
		return nil, nil
	}

	// Convert LabelSelector to a Selector object (precompiled for efficiency)
	var objSelector labels.Selector

	if s.LabelSelector != nil {
		var err error

		objSelector, err = metav1.LabelSelectorAsSelector(s.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid namespace selector: %w", err)
		}
	}

	// ✅ First filter by label selector (if provided)
	var labelFilteredObjects []metav1.Object

	for _, obj := range objects {
		if objSelector != nil && !objSelector.Matches(labels.Set(obj.GetLabels())) {
			continue // Skip non-matching objects
		}

		labelFilteredObjects = append(labelFilteredObjects, obj)
	}

	// ✅ If no NamespaceSelector is set, return the label-filtered objects
	if s.NamespaceSelector == nil {
		return labelFilteredObjects, nil
	}

	// Get namespaces matching NamespaceSelector
	matchingNamespaces, err := s.GetMatchingNamespaces(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("error fetching matching namespaces: %w", err)
	}

	// Convert matching namespaces to a fast lookup map
	namespaceSet := make(map[string]struct{})
	for _, ns := range matchingNamespaces {
		namespaceSet[ns.Name] = struct{}{}
	}

	// ✅ Second filter: Ensure the objects' namespaces are in the allowed set
	var finalMatchingObjects []metav1.Object

	for _, obj := range labelFilteredObjects {
		if len(namespaceSet) > 0 {
			if _, exists := namespaceSet[obj.GetNamespace()]; !exists {
				continue // Skip objects in disallowed namespaces
			}
		}

		finalMatchingObjects = append(finalMatchingObjects, obj)
	}

	return finalMatchingObjects, nil
}

func MatchTypedObjects[T client.Object](
	ctx context.Context,
	cl client.Client,
	selector *NamespacedSelector,
	list []T,
) ([]T, error) {
	if selector == nil {
		return list, nil
	}

	// Precompile object label selector
	var objSelector labels.Selector

	var err error
	if selector.LabelSelector != nil {
		objSelector, err = metav1.LabelSelectorAsSelector(selector.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid object selector: %w, selector: %v", err, selector.LabelSelector)
		}
	}

	// Compile namespace selector
	namespaceSet := make(map[string]struct{})

	if selector.NamespaceSelector != nil {
		namespaces, err := selector.GetMatchingNamespaces(ctx, cl)
		if err != nil {
			return nil, fmt.Errorf("error fetching matching namespaces: %w", err)
		}

		for _, ns := range namespaces {
			namespaceSet[ns.Name] = struct{}{}
		}
	}

	var result []T

	for _, obj := range list {
		lbls := obj.GetLabels()
		namespace := obj.GetNamespace()

		if objSelector != nil && !objSelector.Matches(labels.Set(lbls)) {
			continue
		}

		if selector.NamespaceSelector != nil {
			if _, ok := namespaceSet[namespace]; !ok {
				continue
			}
		}

		result = append(result, obj)
	}

	return result, nil
}

func (s *NamespacedSelector) MatchSecrets(
	ctx context.Context,
	cl client.Client,
	secrets []corev1.Secret,
) ([]corev1.Secret, error) {
	// Convert []corev1.Secret to []metav1.Object.
	objects := make([]metav1.Object, 0, len(secrets))
	for i := range secrets {
		// Taking the address of each secret makes it implement metav1.Object.
		objects = append(objects, &secrets[i])
	}

	// Call the generic MatchObjects function.
	matchedObjs, err := s.MatchObjects(ctx, cl, objects)
	if err != nil {
		return nil, err
	}

	// Convert matchedObjs (which are metav1.Object) back to []corev1.Secret.
	var matchedSecrets []corev1.Secret

	for _, obj := range matchedObjs {
		// Type assertion to *corev1.Secret.
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			// Skip any objects that are not secrets.
			continue
		}

		matchedSecrets = append(matchedSecrets, *secret)
	}

	return matchedSecrets, nil
}
