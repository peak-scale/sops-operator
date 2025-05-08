// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
	"github.com/peak-scale/sops-operator/internal/api/errors"
	"github.com/peak-scale/sops-operator/internal/decryptor"
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

// SopsSecretReconciler reconciles a SopsSecret object.
type SopsSecretReconciler struct {
	client.Client
	Metrics  *metrics.Recorder
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *SopsSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sopsv1alpha1.SopsSecret{}).
		Watches(&corev1.Secret{},
			handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &sopsv1alpha1.SopsSecret{})).
		Watches(
			&sopsv1alpha1.SopsProvider{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
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
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldStatus := e.ObjectOld.(*sopsv1alpha1.SopsProvider).Status
					newStatus := e.ObjectNew.(*sopsv1alpha1.SopsProvider).Status
					return !reflect.DeepEqual(oldStatus, newStatus)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
			}),
		).
		Complete(r)
}

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SopsSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Request.Name", req.Name)
	// Fetch the Tenant instance
	instance := &sopsv1alpha1.SopsSecret{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			// Cleanup Metrics
			r.Metrics.DeleteSecretCondition(instance)

			r.Log.Info("Request object not found, could have been deleted after reconcile request")

			return reconcile.Result{}, nil
		}

		r.Log.Error(err, "Error reading the object")

		return reconcile.Result{}, nil
	}

	// Main Reconciler
	err := r.reconcile(
		ctx,
		log,
		instance,
	)

	// Always Record Metric
	r.Metrics.RecordSecretCondition(instance)

	// Always Post Status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		log.V(7).Info("updating", "status", instance.Status)
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

func (r *SopsSecretReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	secret *sopsv1alpha1.SopsSecret,
) (err error) {
	// Load Decryption Provider (Keys)
	log.V(5).Info("loading secrets provider")

	provider, err := r.decryptionProvider(ctx, log, secret)
	if err != nil {
		secret.Status.Condition = meta.NewNotReadyCondition(secret, err.Error())
		secret.Status.Condition.Reason = meta.DecryptionFailedReason

		// Handle Cleanup
		return r.cleanupSecrets(
			ctx,
			log,
			secret,
		)
	}

	encrypted, err := provider.IsEncrypted(secret)
	if err != nil {
		return err
	}

	// Reject unencrypted secrets
	if !encrypted {
		secret.Status.Condition = meta.NewNotReadyCondition(secret, "Secret missing SOPS encryption marker (not encrypted)")
		secret.Status.Condition.Reason = meta.NotSopsEncryptedReason
	}

	// Iterate over Secrets
	selectedSecrets := make(map[string]bool)

	for _, sec := range secret.Spec.Secrets {
		// Index under unique key
		uniqueKey := sec.Name
		selectedSecrets[uniqueKey] = true

		slog := log.WithValues("secret", sec.Name)

		// Reconcile Secret
		serr := r.reconcileSecret(
			ctx,
			slog,
			secret,
			provider,
			sec,
		)
		if serr != nil {
			slog.Error(serr, "failed to reconcile secret")
		}
	}

	// Lifecycle Secrets
	for _, sec := range secret.Status.Secrets {
		uniqueKey := sec.Name
		if _, ok := selectedSecrets[uniqueKey]; !ok {
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

	// Everything alright!
	secret.Status.Condition = meta.NewReadyCondition(secret)
	secret.Status.Condition.Message = "Secrets Decrypted"

	return err
}

// Decrypt SOPS Secret.
func (r *SopsSecretReconciler) reconcileSecret(
	ctx context.Context,
	log logr.Logger,
	origin *sopsv1alpha1.SopsSecret,
	decryptor *decryptor.SOPSDecryptor,
	secret *sopsv1alpha1.SopsSecretItem,
) (err error) {
	// Target for Replication
	target := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: origin.Namespace,
		},
	}

	// Status
	status := meta.NewReadySecretStatusCondition(target)
	defer func() {
		log.V(7).Info("updating instance", "status", status)
		origin.Status.UpdateInstance(status)
	}()

	err = r.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: origin.Namespace}, target)
	//// Check if Ownerreference set, if not return
	if err == nil {
		if y, _ := controllerutil.HasOwnerReference(target.OwnerReferences, origin, r.Scheme); !y {
			err = fmt.Errorf("secret %s/%s already present, but not provisioned by sops-controller", target.Name, target.Namespace)

			return err
		}
	}

	if err = decryptor.Decrypt(origin, secret, log); err != nil {
		log.Error(err, "encryption failed")

		return err
	}

	// Replicate Secret
	_, cerr := controllerutil.CreateOrUpdate(ctx, r.Client, target, func() error {
		labels := target.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}

		for k, v := range secret.Labels {
			labels[k] = v
		}

		annotations := target.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		for k, v := range secret.Annotations {
			annotations[k] = v
		}

		target.Data = map[string][]byte{}
		for k, v := range secret.Data {
			target.Data[k] = []byte(v)
		}

		target.StringData = secret.StringData

		log.V(7).Info("patching secret", "manifest", "secret")

		// We set owner reference to the secret
		return controllerutil.SetOwnerReference(origin, target, r.Client.Scheme())
	})
	if cerr != nil {
		status = meta.NewNotReadySecretStatusCondition(target, cerr.Error())

		return cerr
	}

	return nil
}

// Delete all decrypted secret
func (r *SopsSecretReconciler) cleanupSecrets(
	ctx context.Context,
	log logr.Logger,
	secret *sopsv1alpha1.SopsSecret,
) (err error) {
	for _, sec := range secret.Status.Secrets {
		err := r.Delete(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sec.Name,
				Namespace: sec.Namespace,
			},
		})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}

		secret.Status.RemoveInstance(&sopsv1alpha1.SopsSecretItemStatus{
			Name:      sec.Name,
			Namespace: sec.Namespace,
		})
	}

	return nil
}

// Initialize SOPS Decryption Provider.
func (r *SopsSecretReconciler) decryptionProvider(
	ctx context.Context,
	log logr.Logger,
	secret *sopsv1alpha1.SopsSecret,
) (sops *decryptor.SOPSDecryptor, err error) {
	// Reset previous providers
	secret.Status.Providers = make([]*api.Origin, 0)

	// Gather all Providers
	providerList := &sopsv1alpha1.SopsProviderList{}
	if err := r.List(ctx, providerList); err != nil {
		r.Log.Error(err, "Failed to list providers")

		return nil, err
	}

	// Evaluate the Providers, which are matching
	matchingProviders := []sopsv1alpha1.SopsProvider{}

	for _, provider := range providerList.Items {
		// match state for provider
		providerMatch := false

		for _, selector := range provider.Spec.SOPSSelectors {
			match, err := selector.SingleMatch(ctx, r.Client, secret)
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
		return nil, errors.NewNoDecryptionProviderError(secret)
	}

	// Initialize Temporary Decryptor
	decryptor, _, err := decryptor.NewSOPSTempDecryptor()
	if err != nil {
		return nil, err
	}

	// Gather Secrets from Provider
	for _, provider := range matchingProviders {
		secret.Status.Providers = append(secret.Status.Providers,
			api.NewOrigin(&provider),
		)

		for _, sec := range provider.Status.Providers {
			if sec.Status == metav1.ConditionTrue {
				log.V(5).Info("adding secret from provider", "secret", sec.Name)

				if err := decryptor.KeysFromSecret(ctx, r.Client, sec.Name, sec.Namespace); err != nil {
					log.Error(err, "adding provider secret")
				}
			} else {
				log.V(5).Info("security not ready", "secret", sec.Name)
			}
		}
	}

	return decryptor, nil
}
