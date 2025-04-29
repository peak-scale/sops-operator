/*
Copyright 2024 Peak Scale
SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/meta"
	"github.com/peak-scale/sops-operator/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SopsProviderReconciler reconciles a SopsProvider object.
type SopsProviderReconciler struct {
	client.Client
	Metrics  *metrics.Recorder
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
// SetupWithManager sets up the controller with the Manager.
func (r *SopsProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sopsv1alpha1.SopsProvider{}).
		Complete(r)
}

func (r *SopsProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Name", req.Name)
	// Fetch the Tenant instance
	instance := &sopsv1alpha1.SopsProvider{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Request object not found, could have been deleted after reconcile request")

			return reconcile.Result{}, nil
		}

		log.Error(err, "Error reading the object")

		return reconcile.Result{}, nil
	}

	// Main Reconciler
	err := r.reconcile(ctx, log, instance)

	// Always Record Metric
	r.Metrics.RecordProviderCondition(instance)

	// Always Post Status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		log.V(10).Info("updating", "status", instance.Status)
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, instance.DeepCopy(), func() error {
			return r.Client.Status().Update(ctx, instance, &client.SubResourceUpdateOptions{})
		})

		return
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SopsProviderReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	provider *sopsv1alpha1.SopsProvider,
) error {
	// Collect Namespaces (Matching)
	secretList := &corev1.SecretList{}
	if err := r.Client.List(ctx, secretList); err != nil {
		r.Log.Error(err, "Failed to list secrets")

		return err
	}

	log.V(10).Info("listing secrets", "secrets", secretList.Items)

	selectedSecrets := make(map[string]*corev1.Secret)

	for _, selector := range provider.Spec.ProviderSecrets {
		matchingSecrets, err := selector.MatchSecrets(ctx, r.Client, secretList.Items)
		if err != nil {
			log.Error(err, "error creating selector")

			continue
		}

		log.V(7).Info("loading secrets", "total", len(matchingSecrets))

		// Iterate over matched secrets
		for _, secret := range matchingSecrets {
			// Disregard Deleting Secrets
			secret := secret
			if !secret.ObjectMeta.DeletionTimestamp.IsZero() {
				continue
			}

			// Index under unique key
			uniqueKey := secret.Namespace + "/" + secret.Name
			selectedSecrets[uniqueKey] = &secret
		}
	}

	log.V(7).Info("selected secrets", "total", len(selectedSecrets))

	// Run Garbage Collection (Removes items which are no longer selected)
	for _, secret := range provider.Status.Providers {
		uniqueKey := secret.Origin.Namespace + "/" + secret.Name
		if _, ok := selectedSecrets[uniqueKey]; !ok {
			provider.Status.RemoveInstance(&sopsv1alpha1.SopsProviderItemStatus{
				Origin: secret.Origin,
			})
		}
	}

	// Update Each Secret
	for _, secret := range selectedSecrets {
		r.reconcileProvider(
			ctx,
			log,
			provider,
			secret,
		)
	}

	provider.Status.Condition = meta.NewReadyCondition(provider)

	return nil
}

func (r *SopsProviderReconciler) reconcileProvider(
	ctx context.Context,
	log logr.Logger,
	provider *sopsv1alpha1.SopsProvider,
	secret *corev1.Secret,
) {
	// Initialize Status
	status := &sopsv1alpha1.SopsProviderItemStatus{
		Origin: *api.NewOrigin(secret),
	}

	// Skip if namespace is being deleted
	if !secret.ObjectMeta.DeletionTimestamp.IsZero() {
		provider.Status.RemoveInstance(status)
	}

	// Currently No validation present, therefor always ready
	status.Condition = meta.NewReadyCondition(secret)
	provider.Status.UpdateInstance(status)
}
