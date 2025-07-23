# Monitoring

Via the `/metrics` endpoint and the dedicated port you can scrape Prometheus Metrics. Amongst the standard [Kubebuilder Metrics](https://book-v1.book.kubebuilder.io/beyond_basics/controller_metrics) we provide metrics, to give you oversight of what's currently working and what's broken. This way you can always be informed, when something is not working as expected. Our custom metrics are prefixed with `sops_`:

```shell
# HELP sops_provider_condition The current condition status of a Provider.
# TYPE sops_provider_condition gauge
sops_provider_condition{name="sample-provider",status="NotReady"} 0
sops_provider_condition{name="sample-provider",status="Ready"} 1

# HELP sops_secret_condition The current condition status of a Secret.
# TYPE sops_secret_condition gauge
sops_secret_condition{name="secret-key-1",namespace="default",status="NotReady"} 0
sops_secret_condition{name="secret-key-1",namespace="default",status="Ready"} 1

# HELP sops_global_secret_condition The current condition status of a Global Secret.
# TYPE sops_global_secret_condition gauge
sops_global_secret_condition{name="global-secret-key-1",status="NotReady"} 1
sops_global_secret_condition{name="global-secret-key-1",status="Ready"} 0
```

The Helm-Chart comes with a [ServiceMonitor](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#servicemonitor) and [PrometheusRules](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.PrometheusRule)
