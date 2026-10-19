---
title: SOPS Operator
description: Manage encryption keys and decrypt SOPS secrets inside Kubernetes, with or without Capsule.
---

## Kubernetes native secret decryption

SOPS Operator brings the [Flux approach to SOPS](https://fluxcd.io/flux/guides/mozilla-sops/)
to a standalone Kubernetes operator. Keep encrypted secrets in Git, distribute keys
through Kubernetes resources, and decrypt secrets inside the cluster.

A `SopsProvider` connects private keys to the `SopsSecrets` allowed to use them.
Cluster administrators control those connections, while teams can manage their own
keys and encrypted secrets. `GlobalSopsSecret` supports secrets across namespaces.

Use it with [Capsule](https://projectcapsule.dev/) or on its own, independently of
your GitOps engine.

## Try it out

Explore the [interactive Killercoda demonstration](https://killercoda.com/peakscale/course/solutions/multi-tenant-sops).
For enterprise support, contact [Peak Scale](https://peakscale.ch/en/).
