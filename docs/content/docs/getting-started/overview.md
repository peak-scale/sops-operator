---
title: Overview
weight: 10
---

SOPS Operator decrypts encrypted Kubernetes secrets inside the cluster. It uses
standard Kubernetes labels and namespace selectors to connect keys with the secrets
that may use them.

- **SopsProvider** selects private key resources and the secrets allowed to use them.
- **SopsSecret** holds encrypted values and creates Kubernetes Secrets in its namespace.
- **GlobalSopsSecret** is cluster scoped and can create Secrets in multiple namespaces.

<img class="sops-architecture" src="{{< relurl "images/sops-operator.drawio.png" >}}" alt="SOPS Operator architecture">

Start with [installation]({{< relref "/docs/getting-started/installation" >}}), then
follow the [usage guide]({{< relref "/docs/guides/usage" >}}) to configure age,
OpenPGP, or Vault/OpenBao keys and encrypt your first secret.

The [SOPS Checker]({{< relref "/docs/guides/checker" >}}) can also verify encrypted
files locally or in commit hooks before they reach the cluster.

## Attribution

The SOPS implementation builds on the
[Flux kustomize-controller decryptor](https://github.com/fluxcd/kustomize-controller/blob/main/internal/decryptor/decryptor.go).
The original license headers are preserved. The project was inspired by
[sops-secrets-operator](https://github.com/isindir/sops-secrets-operator).
