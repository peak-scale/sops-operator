// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	errs "github.com/peak-scale/sops-operator/internal/api/errors"
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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SopsSecretReconciler reconciles a SopsSecret object.
type GlobalSopsSecretReconciler struct {
	client.Client
	Metrics  *metrics.Recorder
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	Config   SopsSecretReconcilerConfig
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlobalSopsSecretReconciler) SetupWithManager(mgr ctrl.Manager, cfg SopsSecretReconcilerConfig) error {
	r.Config = cfg

	r.Log.V(7).Info("controller config", "config", r.Config)

	return ctrl.NewControllerManagedBy(mgr).
		Named(cfg.ControllerName).
		For(&sopsv1alpha1.GlobalSopsSecret{}).
		Watches(&corev1.Secret{},
			handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &sopsv1alpha1.GlobalSopsSecret{})).
		Watches(
			&sopsv1alpha1.SopsProvider{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, _ client.Object) []reconcile.Request {
				var list sopsv1alpha1.GlobalSopsSecretList
				if err := r.Client.List(ctx, &list); err != nil {
					r.Log.Error(err, "unable to list GlobalSopsSecrets")

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
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(event.CreateEvent) bool {
					return true
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldObj, okOld := e.ObjectOld.(*sopsv1alpha1.SopsProvider)
					newObj, okNew := e.ObjectNew.(*sopsv1alpha1.SopsProvider)
					if !okOld || !okNew {
						return false
					}

					return !reflect.DeepEqual(oldObj.Status, newObj.Status)
				},
				DeleteFunc: func(event.DeleteEvent) bool {
					return true
				},
			}),
		).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				namespace := obj.GetName()

				var list sopsv1alpha1.GlobalSopsSecretList
				if err := r.Client.List(ctx, &list); err != nil {
					r.Log.Error(err, "unable to list GlobalSopsSecrets")

					return nil
				}

				var requests []reconcile.Request
				for _, gss := range list.Items {
					for _, secret := range gss.Status.Secrets {
						if secret.Namespace == namespace &&
							secret.Condition.Type == "NotReady" &&
							secret.Condition.Status == metav1.ConditionFalse {
							requests = append(requests, reconcile.Request{
								NamespacedName: types.NamespacedName{
									Name:      gss.Name,
									Namespace: gss.Namespace,
								},
							})

							break
						}
					}
				}

				return requests
			}),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(event.CreateEvent) bool { return true },
				DeleteFunc: func(event.DeleteEvent) bool { return false },
				UpdateFunc: func(event.UpdateEvent) bool { return false },
			}),
		).
		Complete(r)
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *GlobalSopsSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Name", req.Name)
	// Fetch the Tenant instance
	instance := &sopsv1alpha1.GlobalSopsSecret{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Cleanup Metrics
			r.Metrics.DeleteGlobalSecretCondition(instance)
			log.V(5).Info("Request object not found, could have been deleted after reconcile request")

			return reconcile.Result{}, nil
		}

		r.Log.Error(err, "Error reading the object")

		return reconcile.Result{}, nil
	}

	defer func() {
		r.Metrics.RecordGlobalSecretCondition(instance)
	}()

	// Main Reconciler
	reconcileErr := r.reconcile(
		ctx,
		log,
		instance,
	)

	// Always Post Status
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		current := &sopsv1alpha1.GlobalSopsSecret{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(instance), current); err != nil {
			return fmt.Errorf("failed to refetch instance before update: %w", err)
		}

		current.Status = instance.Status

		log.V(7).Info("updating status", "status", current.Status)

		return r.Client.Status().Update(ctx, current)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileErr != nil {
		var sre *errs.SecretReconciliationError
		if errors.As(reconcileErr, &sre) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{RequeueAfter: r.Config.FailedSecretsInterval.Duration}, nil
	}

	return ctrl.Result{}, nil
}

func (r *GlobalSopsSecretReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	secret *sopsv1alpha1.GlobalSopsSecret,
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
		secret.Status.Condition = meta.NewNotReadyCondition(secret, err.Error())
		secret.Status.Condition.Reason = meta.DecryptionFailedReason

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
			&sec.SopsSecretItem,
			sec.Namespace,
		)

		selectedSecrets[string(target.GetUID())] = true

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
		if _, ok := selectedSecrets[string(sec.UID)]; !ok {
			log.V(7).Info("garbage collection", "secret", sec.Name)

			var orphanSecret corev1.Secret

			err := r.Get(ctx, types.NamespacedName{
				Name:      sec.Name,
				Namespace: sec.Namespace,
			}, &orphanSecret)
			if err != nil && !apierrors.IsNotFound(err) {
				// Error Removing
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
		secret.Status.Condition = meta.NewNotReadyCondition(secret, "Secret reconciliation failed")

		return errs.NewSecretReconciliationError("Secret reconciliation failed")
	}

	// Everything alright!
	secret.Status.Condition = meta.NewReadyCondition(secret)
	secret.Status.Condition.Message = "Secrets Decrypted"

	return nil
}
