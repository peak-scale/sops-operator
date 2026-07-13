// Copyright 2024-2026 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"

	sopsv1alpha1 "github.com/peak-scale/sops-operator/api/v1alpha1"
	"github.com/peak-scale/sops-operator/internal/meta"
	capmeta "github.com/projectcapsule/capsule/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRecordConditionMetricLabelCardinality(t *testing.T) {
	t.Parallel()

	recorder := NewRecorder()
	ready := capmeta.ConditionList{{
		Type:   meta.ReadyCondition,
		Status: metav1.ConditionTrue,
	}}

	require.NotPanics(t, func() {
		recorder.RecordProviderCondition(&sopsv1alpha1.SopsProvider{
			ObjectMeta: metav1.ObjectMeta{Name: "provider"},
			Status:     sopsv1alpha1.SopsProviderStatus{Conditions: ready.DeepCopy()},
		})
		recorder.RecordSecretCondition(&sopsv1alpha1.SopsSecret{
			ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "tenant"},
			Status:     sopsv1alpha1.SopsSecretStatus{Conditions: ready.DeepCopy()},
		})
		recorder.RecordGlobalSecretCondition(&sopsv1alpha1.GlobalSopsSecret{
			ObjectMeta: metav1.ObjectMeta{Name: "global-secret"},
			Status:     sopsv1alpha1.SopsSecretStatus{Conditions: ready.DeepCopy()},
		})
	})

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(recorder.Collectors()...)
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	secretMetric := findMetricFamily(t, metricFamilies, "sops_secret_condition")
	require.Len(t, secretMetric.Metric, 1)
	require.Equal(t, map[string]string{
		"name":      "secret",
		"namespace": "tenant",
		"status":    meta.ReadyCondition,
	}, metricLabels(secretMetric.Metric[0]))
	require.Equal(t, float64(1), secretMetric.Metric[0].GetGauge().GetValue())
}

func TestDeleteConditionMetricLabelCardinality(t *testing.T) {
	t.Parallel()

	recorder := NewRecorder()
	require.NotPanics(t, func() {
		recorder.DeleteProviderCondition(&sopsv1alpha1.SopsProvider{
			ObjectMeta: metav1.ObjectMeta{Name: "provider"},
		})
		recorder.DeleteSecretCondition(&sopsv1alpha1.SopsSecret{
			ObjectMeta: metav1.ObjectMeta{Name: "secret", Namespace: "tenant"},
		})
		recorder.DeleteGlobalSecretCondition(&sopsv1alpha1.GlobalSopsSecret{
			ObjectMeta: metav1.ObjectMeta{Name: "global-secret"},
		})
	})
}

func findMetricFamily(
	t *testing.T,
	metricFamilies []*dto.MetricFamily,
	name string,
) *dto.MetricFamily {
	t.Helper()

	for _, family := range metricFamilies {
		if family.GetName() == name {
			return family
		}
	}

	t.Fatalf("metric family %q not found", name)

	return nil
}

func metricLabels(metric *dto.Metric) map[string]string {
	labels := make(map[string]string, len(metric.Label))
	for _, label := range metric.Label {
		labels[label.GetName()] = label.GetValue()
	}

	return labels
}
