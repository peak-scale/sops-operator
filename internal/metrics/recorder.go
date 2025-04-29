// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/meta"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crtlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Recorder struct {
	providerConditionGauge *prometheus.GaugeVec
	secretConditionGauge   *prometheus.GaugeVec
}

func MustMakeRecorder() *Recorder {
	metricsRecorder := NewRecorder()
	crtlmetrics.Registry.MustRegister(metricsRecorder.Collectors()...)

	return metricsRecorder
}

func NewRecorder() *Recorder {
	return &Recorder{
		providerConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sops_provider_condition",
				Help: "The current condition status of a Provider.",
			},
			[]string{"name", "status"},
		),

		secretConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sops_secret_condition",
				Help: "The current condition status of a Secret.",
			},
			[]string{"name", "namespace", "status"},
		),
	}
}

func (r *Recorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.providerConditionGauge,
		r.secretConditionGauge,
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordProviderCondition(provider *sopsv1alpha1.SopsProvider) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if provider.Status.Condition.Status == metav1.ConditionTrue {
			value = 1
		}

		r.providerConditionGauge.WithLabelValues(provider.Name, status).Set(value)
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordSecretCondition(secret *sopsv1alpha1.SopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if secret.Status.Condition.Status == metav1.ConditionTrue {
			value = 1
		}

		r.secretConditionGauge.WithLabelValues(secret.Name, secret.Namespace, status).Set(value)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteProviderCondition(provider *sopsv1alpha1.SopsProvider) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.providerConditionGauge.DeleteLabelValues(provider.Name, status)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteSecretCondition(secret *sopsv1alpha1.SopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.secretConditionGauge.DeleteLabelValues(secret.Name, secret.Namespace, status)
	}
}
