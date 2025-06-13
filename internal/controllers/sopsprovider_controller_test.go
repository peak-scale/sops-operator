package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSopsProviderReconciler_SetupWithManager_StatusEnabled(t *testing.T) {
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

	reconciler := &SopsProviderReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Metrics: metrics.NewRecorder(),
		Log:     ctrl.Log.WithName("controllers").WithName("SopsProvider"),
	}

	config := SopsProviderReconcilerConfig{
		EnableStatus: true,
	}

	err = reconciler.SetupWithManager(mgr, config, "sopsprovider-status-enabled")
	assert.NoError(t, err)
}

func TestSopsProviderReconciler_SetupWithManager_StatusDisabled(t *testing.T) {
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

	reconciler := &SopsProviderReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Metrics: metrics.NewRecorder(),
		Log:     ctrl.Log.WithName("controllers").WithName("SopsProvider"),
	}

	config := SopsProviderReconcilerConfig{
		EnableStatus: false,
	}

	err = reconciler.SetupWithManager(mgr, config, "sopsprovider-status-disabled")
	assert.NoError(t, err)
}
