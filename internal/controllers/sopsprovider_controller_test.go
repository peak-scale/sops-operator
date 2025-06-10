// Copyright (C) 2022 The Flux authors
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package controllers

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	cache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type fakeManager struct {
	client client.Client
	scheme *runtime.Scheme
}

func (f *fakeManager) GetClient() client.Client                                 { return f.client }
func (f *fakeManager) GetScheme() *runtime.Scheme                               { return f.scheme }
func (f *fakeManager) GetFieldIndexer() client.FieldIndexer                     { return nil }
func (f *fakeManager) GetCache() cache.Cache                                    { return nil }
func (f *fakeManager) GetConfig() *rest.Config                                  { return &rest.Config{} }
func (f *fakeManager) GetControllerOptions() config.Controller                  { return config.Controller{} }
func (f *fakeManager) GetHTTPClient() *http.Client                              { return http.DefaultClient }
func (f *fakeManager) GetWebhookServer() webhook.Server                         { return nil }
func (f *fakeManager) GetEventRecorderFor(name string) record.EventRecorder     { return nil }
func (f *fakeManager) GetRESTMapper() meta.RESTMapper                           { return nil }
func (f *fakeManager) GetAPIReader() client.Reader                              { return f.client }
func (f *fakeManager) Add(runnable manager.Runnable) error                      { return nil }
func (f *fakeManager) Elected() <-chan struct{}                                 { return make(chan struct{}) }
func (f *fakeManager) SetFields(interface{}) error                              { return nil }
func (f *fakeManager) GetLogger() logr.Logger                                   { return log.Log }
func (f *fakeManager) Start(ctx context.Context) error                          { return nil }
func (f *fakeManager) AddHealthzCheck(name string, check healthz.Checker) error { return nil }
func (f *fakeManager) AddMetricsServerExtraHandler(path string, handler http.Handler) error {
	return nil
}
func (f *fakeManager) AddReadyzCheck(name string, check healthz.Checker) error { return nil }

func TestSopsSecretReconciler_SetupWithManager(t *testing.T) {
	g := NewWithT(t)

	scheme := runtime.NewScheme()
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())
	g.Expect(sopsv1alpha1.AddToScheme(scheme)).To(Succeed())

	sopsSecret := &sopsv1alpha1.SopsSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-secret",
			Namespace: "default",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(sopsSecret).
		Build()

	mgr := &fakeManager{
		client: client,
		scheme: scheme,
	}

	r := &SopsSecretReconciler{
		Client: client,
		Scheme: scheme,
		Log:    log.Log.WithName("test"),
	}

	// SetupWithManager aufrufen und Fehler prüfen
	err := r.SetupWithManager(mgr)
	g.Expect(err).ToNot(HaveOccurred())
}

/*func TestMasterKey_EncryptedDataKey(t *testing.T) {
	g := NewWithT(t)
	key := MasterKey{EncryptedKey: encryptedData}
	g.Expect(key.EncryptedDataKey()).To(BeEquivalentTo(encryptedData))
}
*/
