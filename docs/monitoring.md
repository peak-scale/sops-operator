# Monitoring

Via the `/metrics` endpoint and the dedicated port you can scrape Prometheus Metrics. Amongst the standard [Kubebuilder Metrics](https://book-v1.book.kubebuilder.io/beyond_basics/controller_metrics) we provide metrics, to give you oversight of what's currently working and what's broken. This way you can always be informed, when something is not working as expected. Our custom metrics are prefixed with `sops_`:

```shell
sops_provider_condition{name="default-onboarding",status="NotReady"} 0
sops_provider_condition{name="default-onboarding",status="Ready"} 1
sops_secret_condition{name="dev-onboarding",namespace="secret-namespace",status="NotReady"} 0
sops_secret_condition{name="dev-onboarding",namespace="secret-namespace",status="Ready"} 1
```

The Helm-Chart comes with a [ServiceMonitor](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#servicemonitor) and [PrometheusRules](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.PrometheusRule)
