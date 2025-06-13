// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/controllers"
	"github.com/peak-scale/sops-operator/internal/metrics"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	//+kubebuilder:scaffold:scheme
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(sopsv1alpha1.AddToScheme(scheme))
}

// Options extends the controller-runtime Options with additional configuration
type Options struct {
	ctrl.Options
	EnableStatus bool
}

func main() {
	var metricsAddr string
	var enableLeaderElection, enablePprof, enableStatus bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":10080", "The address the probe endpoint binds to.")
	flag.BoolVar(&enablePprof, "enable-pprof", false, "Enables Pprof endpoint for profiling (not recommend in production)")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableStatus, "enable-status", true, "Enables printing the status of the controller")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrlConfig := Options{
		Options: ctrl.Options{
			Scheme:                 scheme,
			Metrics:                metricsserver.Options{BindAddress: metricsAddr},
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         enableLeaderElection,
			LeaderElectionID:       "2e0ffcfb.peakscale.ch",
		},
		EnableStatus: enableStatus,
	}

	if enablePprof {
		ctrlConfig.PprofBindAddress = ":8082"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlConfig.Options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	metricsRecorder := metrics.MustMakeRecorder()

	if err = (&controllers.SopsSecretReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Metrics: metricsRecorder,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsSecret"),
	}).SetupWithManager(mgr, controllers.SopsSecretReconcilerConfig{
		EnableStatus: enableStatus,
	}, "sopssecret"); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SopsSecret")
		os.Exit(1)
	}

	if err = (&controllers.SopsProviderReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Metrics: metricsRecorder,
		Log:     ctrl.Log.WithName("controllers").WithName("SopsProvider"),
	}).SetupWithManager(mgr, controllers.SopsProviderReconcilerConfig{
		EnableStatus: enableStatus,
	}, "sopsprovider"); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SopsProvider")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
