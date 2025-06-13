// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func TestSopsSecretReconciler_SetupWithManager_StatusEnabled(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Disable metrics and webhook server to avoid port conflicts
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer: nil,
	})
	require.NoError(t, err)

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctrl.SetupSignalHandler())
		require.NoError(t, err)
	}()

	reconciler := &SopsSecretReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Metrics: metrics.NewRecorder(),
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	config := SopsSecretReconcilerConfig{
		EnableStatus: true,
	}

	err = reconciler.SetupWithManager(mgr, config, "sopssecret-status-enabled")
	assert.NoError(t, err)
}

func TestSopsSecretReconciler_SetupWithManager_StatusDisabled(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Disable metrics and webhook server to avoid port conflicts
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer: nil,
	})
	require.NoError(t, err)

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctrl.SetupSignalHandler())
		require.NoError(t, err)
	}()

	reconciler := &SopsSecretReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Metrics: metrics.NewRecorder(),
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	config := SopsSecretReconcilerConfig{
		EnableStatus: false,
	}

	err = reconciler.SetupWithManager(mgr, config, "sopssecret-status-disabled")
	assert.NoError(t, err)
}

func TestSopsSecretReconciler_Reconcile_StatusEnabled(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Create test objects
	testSecret := &sopsv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: sopsv1alpha1.SopsSecretSpec{
			Secrets: []*sopsv1alpha1.SopsSecretItem{
				&sopsv1alpha1.SopsSecretItem{
					Name: "test-secret",
					Data: map[string]string{
						"key": "value",
					},
				},
			},
		},
	}

	// Create client with test objects
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(testSecret).
		WithStatusSubresource(testSecret).
		Build()

	// Create manager with the same scheme
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Disable metrics and webhook server to avoid port conflicts
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer: nil,
	})
	require.NoError(t, err)

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctrl.SetupSignalHandler())
		require.NoError(t, err)
	}()

	// Create reconciler with unique metrics recorder for each test
	metrics := metrics.NewRecorder()
	reconciler := &SopsSecretReconciler{
		Client:  client,
		Scheme:  scheme,
		Metrics: metrics,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Setup controller
	err = reconciler.SetupWithManager(mgr, SopsSecretReconcilerConfig{
		EnableStatus: true,
	}, "sopssecret-status-enabled")
	require.NoError(t, err)

	// Wait for cache to sync
	mgr.GetCache().WaitForCacheSync(context.Background())

	// Reconcile
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testSecret.Name,
			Namespace: testSecret.Namespace,
		},
	})
	assert.NoError(t, err)
}

func TestSopsSecretReconciler_Reconcile_StatusDisabled(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Create test objects
	testSecret := &sopsv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Spec: sopsv1alpha1.SopsSecretSpec{
			Secrets: []*sopsv1alpha1.SopsSecretItem{
				&sopsv1alpha1.SopsSecretItem{
					Name: "test-secret",
					Data: map[string]string{
						"key": "value",
					},
				},
			},
		},
	}

	// Create client with test objects
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(testSecret).
		WithStatusSubresource(testSecret).
		Build()

	// Create manager with the same scheme
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Disable metrics and webhook server to avoid port conflicts
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer: nil,
	})
	require.NoError(t, err)

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctrl.SetupSignalHandler())
		require.NoError(t, err)
	}()

	// Create reconciler with unique metrics recorder for each test
	metrics := metrics.NewRecorder()
	reconciler := &SopsSecretReconciler{
		Client:  client,
		Scheme:  scheme,
		Metrics: metrics,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Setup controller
	err = reconciler.SetupWithManager(mgr, SopsSecretReconcilerConfig{
		EnableStatus: false,
	}, "sopssecret-status-disabled")
	require.NoError(t, err)

	// Wait for cache to sync
	mgr.GetCache().WaitForCacheSync(context.Background())

	// Reconcile
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testSecret.Name,
			Namespace: testSecret.Namespace,
		},
	})
	assert.NoError(t, err)
}

func TestSopsSecretReconciler_cleanupSecrets(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Test cases
	tests := []struct {
		name           string
		sopsSecret     *sopsv1alpha1.SopsSecret
		existingSecret *corev1.Secret
		expectError    bool
	}{
		{
			name: "successful cleanup of existing secret",
			sopsSecret: &sopsv1alpha1.SopsSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Status: sopsv1alpha1.SopsSecretStatus{
					Secrets: []*sopsv1alpha1.SopsSecretItemStatus{
						{
							Name:      "test-secret",
							Namespace: "default",
						},
					},
				},
			},
			existingSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
			},
			expectError: false,
		},
		{
			name: "cleanup of non-existent secret",
			sopsSecret: &sopsv1alpha1.SopsSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Status: sopsv1alpha1.SopsSecretStatus{
					Secrets: []*sopsv1alpha1.SopsSecretItemStatus{
						{
							Name:      "non-existent-secret",
							Namespace: "default",
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with initial state
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)
			if tt.existingSecret != nil {
				clientBuilder = clientBuilder.WithObjects(tt.existingSecret)
			}
			client := clientBuilder.Build()

			reconciler := &SopsSecretReconciler{
				Client: client,
				Log:    logr.Discard(),
			}

			// Execute cleanup
			err := reconciler.cleanupSecrets(context.Background(), tt.sopsSecret)

			// Verify the result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check if secret was deleted
				if tt.existingSecret != nil {
					var secret corev1.Secret
					err := client.Get(context.Background(), types.NamespacedName{
						Name:      tt.existingSecret.Name,
						Namespace: tt.existingSecret.Namespace,
					}, &secret)
					assert.True(t, apierrors.IsNotFound(err))
				}
			}
		})
	}
}
