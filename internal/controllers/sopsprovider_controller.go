// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

func (r *SopsProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sopsv1alpha1.SopsProvider{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, _ client.Object) []reconcile.Request {
				var list sopsv1alpha1.SopsProviderList
				if err := r.Client.List(ctx, &list); err != nil {
					r.Log.Error(err, "unable to list SopsProvider objects")

					return nil
				}

				var requests []reconcile.Request
				for _, sp := range list.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      sp.Name,
							Namespace: sp.Namespace,
						},
					})
				}

				return requests
			}),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					_, ok := e.Object.GetLabels()[meta.KeySecretLabel]

					return ok
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					_, oldOk := e.ObjectOld.GetLabels()[meta.KeySecretLabel]
					_, newOk := e.ObjectNew.GetLabels()[meta.KeySecretLabel]

					return oldOk || newOk
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					_, ok := e.Object.GetLabels()[meta.KeySecretLabel]

					return ok
				},
			}),
		).
		Complete(r)
}

func (r *SopsProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Name", req.Name)
	// Fetch the Tenant instance
	instance := &sopsv1alpha1.SopsProvider{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Request object not found, could have been deleted after reconcile request")

			return reconcile.Result{}, nil
		}

		log.Error(err, "Error reading the object")

		return reconcile.Result{}, nil
	}

	// Main Reconciler
	reconcileErr := r.reconcile(ctx, log, instance)

	// Always Record Metric
	r.Metrics.RecordProviderCondition(instance)

	// Always Post Status
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		log.V(10).Info("updating", "status", instance.Status)
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, instance.DeepCopy(), func() error {
			return r.Client.Status().Update(ctx, instance, &client.SubResourceUpdateOptions{})
		})

		return
	})

	if reconcileErr != nil || err != nil {
		return ctrl.Result{}, reconcileErr
	}

	return ctrl.Result{}, nil
}

func (r *SopsProviderReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	provider *sopsv1alpha1.SopsProvider,
) error {
	labelSelector := &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      meta.KeySecretLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		r.Log.Error(err, "Failed to convert label selector")

		return err
	}

	secretList := &corev1.SecretList{}
	if err := r.List(ctx, secretList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		r.Log.Error(err, "Failed to list secrets")

		return err
	}

	log.V(10).Info("listing secrets", "secrets", secretList.Items)

	secretPtrs := make([]*corev1.Secret, 0)
	for i := range secretList.Items {
		secretPtrs = append(secretPtrs, &secretList.Items[i])
	}

	selectedSecrets := make(map[string]*corev1.Secret)

	for _, selector := range provider.Spec.ProviderSecrets {
		matchingSecrets, err := api.MatchTypedObjects(ctx, r.Client, selector, secretPtrs)
		if err != nil {
			log.Error(err, "error creating selector")

			continue
		}

		log.V(4).Info("loading secrets", "total", len(matchingSecrets))

		// Iterate over matched secrets
		for _, secret := range matchingSecrets {
			// Disregard Deleting Secrets
			if !secret.DeletionTimestamp.IsZero() {
				continue
			}

			// Index under unique key
			uniqueKey := secret.Namespace + "/" + secret.Name
			selectedSecrets[uniqueKey] = secret
		}
	}

	log.V(4).Info("selected secrets", "total", len(selectedSecrets))

	for key, secret := range selectedSecrets {
		log.V(7).Info("selected secret", "key", key, "type", secret.Type)
	}

	// Run Garbage Collection (Removes items which are no longer selected)
	for _, secret := range provider.Status.Providers {
		uniqueKey := secret.Namespace + "/" + secret.Name
		if _, ok := selectedSecrets[uniqueKey]; !ok {
			provider.Status.RemoveInstance(&sopsv1alpha1.SopsProviderItemStatus{
				Origin: secret.Origin,
			})
		}
	}

	// Update Each Secret
	for _, secret := range selectedSecrets {
		r.reconcileProvider(
			provider,
			secret,
		)
	}

	provider.Status.Condition = meta.NewReadyCondition(provider)

	return nil
}

func (r *SopsProviderReconciler) reconcileProvider(
	provider *sopsv1alpha1.SopsProvider,
	secret *corev1.Secret,
) {
	// Initialize Status
	status := &sopsv1alpha1.SopsProviderItemStatus{
		Origin: *api.NewOrigin(secret),
	}

	// Skip if namespace is being deleted
	if !secret.DeletionTimestamp.IsZero() {
		provider.Status.RemoveInstance(status)
	}

	// Currently No validation present, therefor always ready
	status.Condition = meta.NewReadyCondition(secret)
	provider.Status.UpdateInstance(status)
}
