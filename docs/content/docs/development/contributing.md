---
title: Contributing
weight: 10
aliases: ["/development.md"]
---

Enter the locked development environment and install the Git hooks:

```shell
devbox shell
task hooks:install
task --list
```

The repository uses `prek.toml` for native file checks and Task commands for Go,
Helm, and documentation checks. CI runs the same commands through Devbox.
Run `task hooks:check` to check all tracked files.

The whitespace hooks automatically fix missing final newlines and trailing spaces.
When a hook changes files, review and re-stage those changes, then rerun the hooks
or commit. Use `task lint:files` for the read-only file checks used in CI.

Getting started locally is pretty easy. You can execute:

```shell
task e2e:build
```

This installs the required operators and SOPS Operator in a [KinD Cluster](https://kind.sigs.k8s.io/). Devbox provides the pinned tool versions. Docker must be running.

If you wish to test against a specific Kubernetes version, you can pass that via variable:

```shell
KIND_NODE_IMAGE="kindest/node:v1.35.0" task e2e:build
```

When you want to quickly develop, you can scale down the operator within the cluster:

```shell
kubectl scale deploy sops-operator --replicas=0 -n sops-operator
```

And then execute the binary:

```shell
task run -- -zap-log-level=10
```

You might need to first export the Kubeconfig for the cluster (If you are using multiple clusters at the same time):

```shell
kind get kubeconfig --name sops-operator  > /tmp/sops-operator
export KUBECONFIG="/tmp/sops-operator"
```

## Testing

When you are done with the development run the following commands.

For linting

```shell
task lint
```

For unit tests (including Docker-backed AWS KMS and Vault fixtures):

```shell
task test
```

For end-to-end testing against the development Kind cluster:

```shell
task e2e:exec
```

## Helm Chart

When making changes to the Helm-Chart, Update the documentation by running:

```shell
task helm:docs
```

Linting and Testing the chart:

```shell
task helm:lint
task helm:test
```

## Performance

Use [PProf](https://book.kubebuilder.io/reference/pprof-tutorial) for profiling:

```shell
curl -s "http://127.0.0.1:8082/debug/pprof/profile" > ./cpu-profile.out

go tool pprof -http=:8080 ./cpu-profile.out
```

## Documentation

The Hugo site uses the Doks layout, with content in `docs/content/docs/`.

```shell
task docs:serve
task docs:generate
task docs:check
task docs:build
```

`docs:generate` builds the API reference from Helm CRDs with crdoc. `docs:check`
checks generated-reference drift and local links and fragments under both root
and subpath deployments. `docs:build` writes to `dist/docs`; set `DOCS_BASE_URL`
for a different deployment URL. Keep npm dependencies locked and installed with
`npm ci --ignore-scripts`.

Run `task e2e` for a disposable cluster with automatic cleanup. The standalone
`e2e:build`, `e2e:exec`, and `e2e:destroy` tasks support a longer lived development
cluster. Set `KIND_CLUSTER` to select its name. The OpenBao tests need
`openbao.openbao.svc.cluster.local` to resolve to `127.0.0.1` on the host.
