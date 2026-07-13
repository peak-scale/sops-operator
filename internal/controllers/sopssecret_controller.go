// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	errs "github.com/peak-scale/sops-operator/internal/api/errors"
	"github.com/peak-scale/sops-operator/internal/meta"
	"github.com/peak-scale/sops-operator/internal/metrics"
	capmeta "github.com/projectcapsule/capsule/pkg/api/meta"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SopsSecretReconcilerConfig struct {
	EnableStatus          bool
	ControllerName        string
	FailedSecretsInterval metav1.Duration
}

// SopsSecretReconciler reconciles a SopsSecret object.
type SopsSecretReconciler struct {
	client.Client

	Metrics  *metrics.Recorder
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	Config   SopsSecretReconcilerConfig
}

// SetupWithManager sets up the controller with the Manager.
func (r *SopsSecretReconciler) SetupWithManager(mgr ctrl.Manager, cfg SopsSecretReconcilerConfig) error {
	r.Config = cfg

	r.Log.V(7).Info("controller config", "config", r.Config)

	return ctrl.NewControllerManagedBy(mgr).
		Named(cfg.ControllerName).
		For(&sopsv1alpha1.SopsSecret{}, builder.WithPredicates(primaryResourcePredicate())).
		Watches(&corev1.Secret{},
			handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &sopsv1alpha1.SopsSecret{})).
		Watches(
			&sopsv1alpha1.SopsProvider{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, _ client.Object) []reconcile.Request {
				var list sopsv1alpha1.SopsSecretList
				if err := r.Client.List(ctx, &list); err != nil {
					r.Log.Error(err, "unable to list SopsSecrets")

					return nil
				}

				var requests []reconcile.Request
				for _, s := range list.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      s.Name,
							Namespace: s.Namespace,
						},
					})
				}

				return requests
			}),
			builder.WithPredicates(sopsProviderStatusPredicate()),
		).
		Complete(r)
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SopsSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("Request.Name", req.Name)

	instance := &sopsv1alpha1.SopsSecret{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Cleanup Metrics
			r.Metrics.DeleteSecret(instance)
			log.V(5).Info("Request object not found, could have been deleted after reconcile request")

			return reconcile.Result{}, nil
		}

		r.Log.Error(err, "Error reading the object")

		return reconcile.Result{}, nil
	}

	// Main Reconciler
	reconcileErr := r.reconcile(
		ctx,
		log,
		instance,
	)

	defer func() {
		r.Metrics.RecordSecretCondition(instance)

		if statusErr := r.updateStatus(ctx, reconcileErr, instance); statusErr != nil {
			statusErr = fmt.Errorf("cannot update SopsSecret status: %w", statusErr)

			if err == nil {
				err = statusErr
			} else {
				err = errors.Join(err, statusErr)
			}
		}
	}()

	if reconcileErr != nil {
		var sre *errs.SecretReconciliationError
		if !errors.As(reconcileErr, &sre) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{RequeueAfter: r.Config.FailedSecretsInterval.Duration}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SopsSecretReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	secret *sopsv1alpha1.SopsSecret,
) (err error) {
	// Load Decryption Provider (Keys)
	log.V(5).Info("loading secrets provider")

	sopsFormat, provider, cleanup, err := fetchDecryptionProviders(ctx, r.Client, log, r.Config, &secret.Status, secret)

	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	if err != nil {
		// Handle Cleanup
		return cleanupSecrets(
			ctx,
			r.Client,
			&secret.Status,
		)
	}

	// Iterate over Secrets
	selectedSecrets := make(map[string]bool)

	failed := false

	for _, sec := range secret.Spec.Secrets {
		slog := log.WithValues("secret", sec.Name)

		// Reconcile Secret
		target, serr := reconcileSecret(
			ctx,
			r.Client,
			slog,
			sopsFormat,
			provider,
			sec,
			secret.Namespace,
			secret.Spec.Metadata,
		)

		selectedSecrets[target.Name+"/"+target.Namespace] = true

		if serr != nil {
			failed = true

			secret.Status.UpdateInstance(
				meta.NewNotReadySecretStatusCondition(target, serr.Error()),
			)

			continue
		}

		secret.Status.UpdateInstance(
			meta.NewReadySecretStatusCondition(target),
		)
	}

	// Lifecycle Secrets
	for _, sec := range secret.Status.Secrets {
		if _, ok := selectedSecrets[sec.Name+"/"+sec.Namespace]; !ok {
			log.V(7).Info("garbage collection", "secret", sec.Name, "namespace", sec.Namespace)

			err := r.Delete(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sec.Name,
					Namespace: sec.Namespace,
				},
			})
			if err != nil && !apierrors.IsNotFound(err) {
				failed = true

				log.Error(err, "error removing secret")

				continue
			}

			// Remove Instance
			secret.Status.RemoveInstance(&sopsv1alpha1.SopsSecretItemStatus{
				Name:      sec.Name,
				Namespace: sec.Namespace,
			})
		}
	}

	if failed {
		log.V(7).Info("secrets had errors")

		return errs.NewSecretReconciliationError("Secret reconciliation failed")
	}

	return nil
}

func (r *SopsSecretReconciler) updateStatus(
	ctx context.Context,
	reconcileError error,
	instance *sopsv1alpha1.SopsSecret,
) (err error) {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		latest := &sopsv1alpha1.SopsSecret{}
		if err = r.Get(ctx, types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, latest); err != nil {
			return err
		}

		latest.Status = instance.Status
		latest.Status.ObservedGeneration = instance.GetGeneration()

		readyCondition := capmeta.NewReadyCondition(latest)
		readyCondition.ObservedGeneration = instance.GetGeneration()
		readyCondition.Status = metav1.ConditionTrue
		readyCondition.Reason = capmeta.SucceededReason
		readyCondition.Message = "reconciled"

		if reconcileError != nil {
			readyCondition.Message = reconcileError.Error()
			readyCondition.Status = metav1.ConditionFalse
			readyCondition.Reason = capmeta.FailedReason
		} else {
			readyCondition.Message = "Secrets Decrypted"
		}

		latest.Status.Conditions.UpdateConditionByType(readyCondition)
		latest.Status.Normalize()

		// Unset legacy Status.
		//nolint:staticcheck
		latest.Status.Condition = metav1.Condition{}

		return r.Client.Status().Update(ctx, latest)
	})
}
