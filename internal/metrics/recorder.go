// Copyright 2024-2025 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/meta"
	"github.com/prometheus/client_golang/prometheus"
	crtlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Recorder struct {
	providerConditionGauge     *prometheus.GaugeVec
	secretConditionGauge       *prometheus.GaugeVec
	globalSecretConditionGauge *prometheus.GaugeVec
}

func MustMakeRecorder() *Recorder {
	metricsRecorder := NewRecorder()
	crtlmetrics.Registry.MustRegister(metricsRecorder.Collectors()...)

	return metricsRecorder
}

func NewRecorder() *Recorder {
	namespace := "sops"

	return &Recorder{
		providerConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "provider_condition",
				Help:      "The current condition status of a Provider.",
			},
			[]string{"name", "status"},
		),
		secretConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "secret_condition",
				Help:      "The current condition status of a Secret.",
			},
			[]string{"name", "namespace", "status"},
		),
		globalSecretConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "global_secret_condition",
				Help:      "The current condition status of a Global Secret.",
			},
			[]string{"name", "status"},
		),
	}
}

func (r *Recorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.providerConditionGauge,
		r.secretConditionGauge,
		r.globalSecretConditionGauge,
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordProviderCondition(provider *sopsv1alpha1.SopsProvider) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if provider.Status.Condition.Type == status {
			value = 1
		}

		r.providerConditionGauge.WithLabelValues(provider.Name, status).Set(value)
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordSecretCondition(secret *sopsv1alpha1.SopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if secret.Status.Condition.Type == status {
			value = 1
		}

		r.secretConditionGauge.WithLabelValues(secret.Name, secret.Namespace, status).Set(value)
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordGlobalSecretCondition(secret *sopsv1alpha1.GlobalSopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if secret.Status.Condition.Type == status {
			value = 1
		}

		r.globalSecretConditionGauge.WithLabelValues(secret.Name, status).Set(value)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteProvider(provider *sopsv1alpha1.SopsProvider) {
	r.providerConditionGauge.DeletePartialMatch(map[string]string{
		"name": provider.Name,
	})
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteProviderCondition(provider *sopsv1alpha1.SopsProvider) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.providerConditionGauge.DeleteLabelValues(provider.Name, status)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteSecret(secret *sopsv1alpha1.SopsSecret) {
	r.secretConditionGauge.DeletePartialMatch(map[string]string{
		"name":      secret.Name,
		"namespace": secret.Namespace,
	})
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteSecretCondition(secret *sopsv1alpha1.SopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.secretConditionGauge.DeleteLabelValues(secret.Name, secret.Namespace, status)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteGlobalSecret(secret *sopsv1alpha1.GlobalSopsSecret) {
	r.globalSecretConditionGauge.DeletePartialMatch(map[string]string{
		"name": secret.Name,
	})
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteGlobalSecretCondition(secret *sopsv1alpha1.GlobalSopsSecret) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.globalSecretConditionGauge.DeleteLabelValues(secret.Name, status)
	}
}
