// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/api"
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

	// Create reconciler
	reconciler := &SopsSecretReconciler{
		Client: mgr.GetClient(),
		Scheme: scheme,
		Log:    ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Setup controller
	err = reconciler.SetupWithManager(mgr, SopsSecretReconcilerConfig{
		EnableStatus:   true,
		ControllerName: "sopssecret-status-enabled",
	})
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

	// Create reconciler
	reconciler := &SopsSecretReconciler{
		Client: mgr.GetClient(),
		Scheme: scheme,
		Log:    ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Setup controller
	err = reconciler.SetupWithManager(mgr, SopsSecretReconcilerConfig{
		EnableStatus:   false,
		ControllerName: "sopssecret-status-disabled",
	})
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
						"key": "ENC[AES256_GCM,data:value,iv:test,type:str]",
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
		EnableStatus:   true,
		ControllerName: "sopssecret-reconcile-status-enabled",
	})
	require.NoError(t, err)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for cache to sync
	require.True(t, mgr.GetCache().WaitForCacheSync(ctx))

	// Reconcile
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
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
						"key": "ENC[AES256_GCM,data:value,iv:test,type:str]",
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
		EnableStatus:   false,
		ControllerName: "sopssecret-reconcile-status-disabled",
	})
	require.NoError(t, err)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for cache to sync
	require.True(t, mgr.GetCache().WaitForCacheSync(ctx))

	// Reconcile
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
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

func TestSopsSecretReconciler_ProviderStatusEnabled(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Create test objects - matching the working test format exactly
	testSecret := &sopsv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Spec: sopsv1alpha1.SopsSecretSpec{
			Secrets: []*sopsv1alpha1.SopsSecretItem{
				&sopsv1alpha1.SopsSecretItem{
					Name: "test-secret",
					Data: map[string]string{
						"key": "ENC[AES256_GCM,data:value,iv:test,type:str]",
					},
				},
			},
		},
	}

	// Create a test provider that should match our secret via label selector
	testProvider := &sopsv1alpha1.SopsProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider",
			Namespace: "default",
		},
		Spec: sopsv1alpha1.SopsProviderSpec{
			SOPSSelectors: []*api.NamespacedSelector{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "true", // This should match our secret's label
						},
					},
				},
			},
		},
		Status: sopsv1alpha1.SopsProviderStatus{
			Providers: []*sopsv1alpha1.SopsProviderItemStatus{
				{
					Origin: api.Origin{
						Name:      "test-provider-secret",
						Namespace: "default",
					},
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
	}

	// Create client with test objects - same pattern as working test
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(testSecret, testProvider).
		WithStatusSubresource(testSecret, testProvider).
		Build()

	// Create manager - same pattern as working test
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Disable metrics and webhook server to avoid port conflicts
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer: nil,
	})
	require.NoError(t, err)

	// Create reconciler with unique metrics recorder - same pattern as working test
	metrics := metrics.NewRecorder()
	reconciler := &SopsSecretReconciler{
		Client:  client,
		Scheme:  scheme,
		Metrics: metrics,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Setup controller with EnableStatus: true
	err = reconciler.SetupWithManager(mgr, SopsSecretReconcilerConfig{
		EnableStatus:   true,
		ControllerName: "sopssecret-provider-status-enabled",
	})
	require.NoError(t, err)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for cache to sync
	require.True(t, mgr.GetCache().WaitForCacheSync(ctx))

	// Get the initial state of the secret (should have empty provider status)
	initialSecret := &sopsv1alpha1.SopsSecret{}
	err = client.Get(ctx, types.NamespacedName{
		Name:      testSecret.Name,
		Namespace: testSecret.Namespace,
	}, initialSecret)
	require.NoError(t, err)

	// Verify initial state - provider status should be empty
	assert.Empty(t, initialSecret.Status.Providers, "Provider status should be empty initially")

	// Reconcile - this should populate the provider status when EnableStatus is true
	result, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testSecret.Name,
			Namespace: testSecret.Namespace,
		},
	})

	// Log reconcile result for debugging
	t.Logf("Reconcile result: %+v, error: %v", result, err)

	// The reconcile might fail due to SOPS decryption, but we still want to check
	// if the provider status was updated before the failure

	// Fetch the updated secret
	updatedSecret := &sopsv1alpha1.SopsSecret{}
	err = client.Get(ctx, types.NamespacedName{
		Name:      testSecret.Name,
		Namespace: testSecret.Namespace,
	}, updatedSecret)
	require.NoError(t, err)

	// Log the actual status for debugging
	t.Logf("Updated secret status providers: %+v", updatedSecret.Status.Providers)

	// Test the main functionality: When EnableStatus is true,
	// the reconciler should populate the Status.Providers field with matching providers
	if assert.NotEmpty(t, updatedSecret.Status.Providers, "Provider status should be populated when EnableStatus is true") {
		// Verify that the correct provider is referenced
		found := false
		for _, provider := range updatedSecret.Status.Providers {
			if provider.Name == "test-provider" && provider.Namespace == "default" {
				found = true
				break
			}
		}
		assert.True(t, found, "The test-provider should be referenced in the status when labels match")
	}
}

func TestSopsSecretReconciler_ProviderStatusDisabled(t *testing.T) {
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
						"key": "ENC[AES256_GCM,data:value,iv:test,type:str]",
					},
				},
			},
		},
	}

	// Create a test provider
	testProvider := &sopsv1alpha1.SopsProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-provider",
		},
		Status: sopsv1alpha1.SopsProviderStatus{
			Providers: []*sopsv1alpha1.SopsProviderItemStatus{
				{
					Origin: api.Origin{
						Name:      "test-provider-secret",
						Namespace: "default",
					},
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
	}

	// Create client with test objects
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(testSecret, testProvider).
		WithStatusSubresource(testSecret, testProvider).
		Build()

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		WebhookServer: nil,
	})
	require.NoError(t, err)

	// Create reconciler
	metrics := metrics.NewRecorder()
	reconciler := &SopsSecretReconciler{
		Client:  client,
		Scheme:  scheme,
		Metrics: metrics,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Setup controller
	err = reconciler.SetupWithManager(mgr, SopsSecretReconcilerConfig{
		EnableStatus:   false,
		ControllerName: "sopssecret-provider-status-disabled",
	})
	require.NoError(t, err)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the manager in a goroutine
	go func() {
		err := mgr.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for cache to sync
	require.True(t, mgr.GetCache().WaitForCacheSync(ctx))

	// Reconcile
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      testSecret.Name,
			Namespace: testSecret.Namespace,
		},
	})
	assert.NoError(t, err)

	// Fetch the updated secret
	updatedSecret := &sopsv1alpha1.SopsSecret{}
	err = client.Get(ctx, types.NamespacedName{
		Name:      testSecret.Name,
		Namespace: testSecret.Namespace,
	}, updatedSecret)
	require.NoError(t, err)

	// Verify provider status is hidden
	assert.Empty(t, updatedSecret.Status.Providers, "Provider status should be hidden when disabled")
}

func TestSopsSecretReconciler_decryptionProvider(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	require.NoError(t, sopsv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	// Create test objects
	testSecret := &sopsv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
			Labels: map[string]string{
				"test": "true",
			},
		},
		Spec: sopsv1alpha1.SopsSecretSpec{
			Secrets: []*sopsv1alpha1.SopsSecretItem{
				{
					Name: "test-secret",
					Data: map[string]string{
						"key": "ENC[AES256_GCM,data:value,iv:test,type:str]",
					},
				},
			},
		},
	}

	// Create a test provider
	testProvider := &sopsv1alpha1.SopsProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider",
			Namespace: "default",
		},
		Spec: sopsv1alpha1.SopsProviderSpec{
			SOPSSelectors: []*api.NamespacedSelector{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"test": "true",
						},
					},
				},
			},
		},
		Status: sopsv1alpha1.SopsProviderStatus{
			Providers: []*sopsv1alpha1.SopsProviderItemStatus{
				{
					Origin: api.Origin{
						Name:      "test-provider-secret",
						Namespace: "default",
					},
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
	}

	// Create client with test objects
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(testSecret, testProvider).
		WithStatusSubresource(testSecret, testProvider).
		Build()

	// Create reconciler
	metrics := metrics.NewRecorder()
	reconciler := &SopsSecretReconciler{
		Client:  client,
		Scheme:  scheme,
		Metrics: metrics,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}

	// Test cases
	tests := []struct {
		name          string
		secret        *sopsv1alpha1.SopsSecret
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid provider configuration",
			secret:      testSecret,
			expectError: false,
		},
		{
			name: "no matching provider",
			secret: &sopsv1alpha1.SopsSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-match-secret",
					Namespace: "default",
					Labels: map[string]string{
						"test": "false", // This won't match any provider
					},
				},
			},
			expectError:   true,
			errorContains: "has no decryption providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call decryptionProvider
			provider, cleanup, err := reconciler.decryptionProvider(context.Background(), reconciler.Log, tt.secret)

			// Check cleanup function
			if cleanup != nil {
				cleanup()
			}

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}
