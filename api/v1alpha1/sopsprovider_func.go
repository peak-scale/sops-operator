/*
Copyright 2024 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package v1alpha1

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GatherProviderSecrets selects unique secrets based on ProviderSelectors.
func (s *SopsProvider) GatherProviderSecrets(ctx context.Context, client client.Client) ([]corev1.Secret, error) {
	secretList := &corev1.SecretList{}
	if err := client.List(ctx, secretList); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	uniqueSecrets := make(map[string]*corev1.Secret)

	for _, selector := range s.Spec.ProviderSecrets {
		if selector == nil || selector.NamespacedSelector == nil {
			continue
		}

		matchingSecrets, err := selector.MatchObjects(ctx, client, toObjectList(secretList.Items))
		if err != nil {
			return nil, fmt.Errorf("error matching secrets: %w", err)
		}

		for _, sec := range matchingSecrets {
			secret, ok := sec.(*corev1.Secret)
			if ok {
				uniqueSecrets[secret.Namespace+"/"+secret.Name] = secret
			}
		}
	}

	finalSecrets := make([]corev1.Secret, 0, len(uniqueSecrets))
	for _, sec := range uniqueSecrets {
		finalSecrets = append(finalSecrets, *sec)
	}

	return finalSecrets, nil
}

// Helper function to convert []corev1.Secret to []metav1.Object.
func toObjectList(secrets []corev1.Secret) []metav1.Object {
	objectList := make([]metav1.Object, len(secrets))
	for i := range secrets {
		objectList[i] = &secrets[i]
	}

	return objectList
}
