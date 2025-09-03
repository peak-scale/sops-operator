// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/api/errors"
	"github.com/peak-scale/sops-operator/internal/decryptor"
	"github.com/peak-scale/sops-operator/internal/meta"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Retrieve a Decryption Provider.
func fetchDecryptionProviders(
	ctx context.Context,
	c client.Client,
	log logr.Logger,
	cfg SopsSecretReconcilerConfig,
	status *sopsv1alpha1.SopsSecretStatus,
	secret client.Object,
) (sopsFile api.SopsImplementation, sops *decryptor.SOPSDecryptor, cleanup func(), err error) {
	// Reset previous providers
	status.Providers = make([]*api.Origin, 0)

	// Gather all Providers
	providerList := &sopsv1alpha1.SopsProviderList{}
	if err := c.List(ctx, providerList); err != nil {
		log.Error(err, "Failed to list providers")

		return nil, nil, nil, err
	}

	// Evaluate the Providers, which are matching
	matchingProviders := []sopsv1alpha1.SopsProvider{}

	for _, provider := range providerList.Items {
		// match state for provider
		providerMatch := false

		for _, selector := range provider.Spec.SOPSSelectors {
			match, err := selector.SingleMatch(ctx, c, secret)
			if err != nil {
				continue
			}

			if match {
				providerMatch = true

				break
			}
		}

		// Append Provider if matched
		if providerMatch {
			matchingProviders = append(matchingProviders, provider)
		}
	}

	log.V(5).Info("evaluated providers", "matching", len(matchingProviders))

	// No providers throws an error
	if len(matchingProviders) == 0 {
		return nil, nil, nil, errors.NewNoDecryptionProviderError(secret)
	}

	// Initialize Temporary Decryptor
	decryptor, cleanup, err := decryptor.NewSOPSTempDecryptor()
	if err != nil {
		return nil, nil, nil, err
	}

	if !cfg.EnableStatus && len(status.Providers) > 0 {
		status.Providers = []*api.Origin{}
	}

	// Gather Secrets from Provider
	for _, provider := range matchingProviders {
		if cfg.EnableStatus {
			status.Providers = append(status.Providers,
				api.NewOrigin(&provider),
			)
		}

		for _, sec := range provider.Status.Providers {
			if sec.Status == metav1.ConditionTrue {
				log.V(7).Info("adding secret from provider", "secret", sec.Name)

				if err := decryptor.KeysFromSecret(ctx, c, sec.Name, sec.Namespace); err != nil {
					log.Error(err, "provider secret error")
				}
			}
		}
	}

	sopsFormat, encrypted, err := sops.IsEncrypted(secret)
	if err != nil {
		return nil, nil, nil, err
	}

	// Reject unencrypted secrets
	if !encrypted {
		err = fmt.Errorf("secret missing SOPS encryption marker (not encrypted)")

		status.Condition = meta.NewNotReadyCondition(secret, err.Error())
		status.Condition.Reason = meta.NotSopsEncryptedReason

		return nil, nil, nil, err
	}

	return sopsFormat, decryptor, cleanup, nil
}

// Reconcile a single Secret Item.
func reconcileSecret(
	ctx context.Context,
	c client.Client,
	log logr.Logger,
	origin api.SopsImplementation,
	decryptor *decryptor.SOPSDecryptor,
	item *sopsv1alpha1.SopsSecretItem,
	itemNamespace string,
	metadata sopsv1alpha1.SecretMetadata,
) (target *corev1.Secret, err error) {
	// Target for Replication
	target = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metadata.Prefix + item.Name + metadata.Suffix,
			Namespace: itemNamespace,
		},
	}

	err = c.Get(ctx, types.NamespacedName{Name: target.Name, Namespace: target.Namespace}, target)
	if err == nil {
		if y, _ := controllerutil.HasOwnerReference(target.OwnerReferences, origin, c.Scheme()); !y {
			err = fmt.Errorf("secret %s/%s already present, but not provisioned by sops-controller", target.Name, target.Namespace)

			return target, err
		}
	}

	if err := decryptor.Decrypt(origin.GetSopsMetadata(), item, log); err != nil {
		return target, fmt.Errorf("secret could not be decrypted")
	}

	// Replicate Secret
	_, cerr := controllerutil.CreateOrUpdate(ctx, c, target, func() error {
		labels := target.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}

		for k, v := range metadata.Labels {
			labels[k] = v
		}

		for k, v := range item.Labels {
			labels[k] = v
		}

		target.SetLabels(labels)

		annotations := target.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		for k, v := range metadata.Annotations {
			annotations[k] = v
		}

		for k, v := range item.Annotations {
			annotations[k] = v
		}

		target.SetAnnotations(annotations)

		target.Data = make(map[string][]byte, len(item.Data))

		for k, v := range item.Data {
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return fmt.Errorf("failed to decode secret data key %s: %w", k, err)
			}

			target.Data[k] = decoded
		}

		target.StringData = item.StringData
		target.Type = item.Type

		log.V(7).Info("patching secret", "manifest", "secret")

		// We set owner reference to the secret
		return controllerutil.SetOwnerReference(origin, target, c.Scheme())
	})
	if cerr != nil {
		return target, cerr
	}

	return target, nil
}

// Delete all decrypted secrets.
func cleanupSecrets(
	ctx context.Context,
	c client.Client,
	status *sopsv1alpha1.SopsSecretStatus,
) (err error) {
	for _, sec := range status.Secrets {
		err := c.Delete(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sec.Name,
				Namespace: sec.Namespace,
			},
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		status.RemoveInstance(&sopsv1alpha1.SopsSecretItemStatus{
			Name:      sec.Name,
			Namespace: sec.Namespace,
		})
	}

	return nil
}
