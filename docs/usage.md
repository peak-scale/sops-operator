# Usage

Reference on how the Operator can be used.

## Providers

Providers are essentially connectors from **where** are the _private keys_ that can decrypt **which** [`SopsSecrets`](#sopssecrets). The following example matches providers with secrets with the given labels:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsProvider
metadata:
  name: sample-provider
spec:
  keys:
  - matchLabels:
      "sops-private-key": "true"
  sops:
  - namespaceSelector:
      matchLabels:
        capsule.clastix.io/tenant: solar
```

### Selection

For both selecting `keys` and `sops` the same selector implementation is used. Each entry can be viewed as dedicated aggregation for selecting secrets:

With this statement, `keys` are loaded from `Secret` in namespaces which match the label `capsule.clastix.io/tenant: solar`. In Addition, the `Secret` must match the label `"sops-private-key": "true"`:

```yaml
  keys:
  - matchLabels:
      "sops-private-key": "true"
    namespaceSelector:
      matchLabels:
        capsule.clastix.io/tenant: solar
```

All items defined are `OR` operations.

Not setting a selector, allows you to select any, so this is selecting all `Secrets`:

```yaml
  keys:
  - matchLabels: {}
  sops:
  - matchLabels: {}
```

### Provider Secrets

> [!IMPORTANT]
> Currently we only support, as we can reliable test them:
> * PGP
> * AGE
> * Openbao
>
> [Generally, the key-management is the same as with FluxCD](https://fluxcd.io/flux/guides/mozilla-sops/)

Providers load decryption keys from `secrets`, which match any condition in the `spec.providers` block of a `SopsProvider`. For `secrets` to be generally considered as key provider, they must have the following specific label:

* `sops.addons.projectcapsule.dev`

It's verified if the label exists, the value is not relevant. So a skeleton secret would look like this:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-private-keys
  labels:
    sops.addons.projectcapsule.dev: "true"
data:
```

### Key-Groups

[Key-Groups](https://github.com/getsops/sops?tab=readme-ov-file#216key-groups) are supported. All the required private-keys may even be distributed amongst different `SopsProviders`. As long as a `SopsSecret` is allowed to collect all the required keys from these `SopsProviders`, it will be able to decrypt.

### SOPS-Configuration

The operator only decrypts fields `.data` and `.stringData` in `.spec.secrets`. All the other fields must not be encrypted, otherwise you will encounter a non-functional behavior. This also allows for customization without possesing the private key of meta-data.

Here a skeleton of such a config:

```shell
cat <<EOF > ./.sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
EOF
```

This is best stored at the root of your repository.

### GPG

Creating a new PGP-Key which can be used from this provider. You may also use





**SOPS**


creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    pgp: CE411B68660C33B0F83A4EBD56FDA28155A45CB1

### AGE


## SopsSecrets

In this approach we post sops encrypted secrets directly to the Kubernetes API. This requires to have the sops encryption marker as additional property. Let's try to use the Provider we created previously to decrypt a new secret.

This is our new `SopsSecret`, we would like to push to git (or twitter):

__secret-key-1.yaml__
```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsSecret
metadata:
  name: example-secret
spec:
  secrets:
    - name: my-secret-name-1
      labels:
        label1: value1
      stringData:
        data-name0: data-value0
      data:
        data-name1: ZGF0YS12YWx1ZTE=
    - name: jenkins-secret
      labels:
        "jenkins.io/credentials-type": "usernamePassword"
      annotations:
        "jenkins.io/credentials-description": "credentials from Kubernetes"
      stringData:
        username: myUsername
        password: 'Pa$$word'
    - name: docker-login
      type: 'kubernetes.io/dockerconfigjson'
      stringData:
        .dockerconfigjson: '{"auths":{"index.docker.io":{"username":"imyuser","password":"mypass","email":"myuser@abc.com","auth":"aW15dXNlcjpteXBhc3M="}}}'
```

Now currently we have the base64 encrypted values or just the plain values in there, we want to change that. We can now simply encrypt that file:

```shell
# For dedicated file
sops -e secret-key-1.yaml  > secret-key-1.enc.yaml

# In-Place
sops -e -i secret-key-1.yaml
```

The new file is encrypted:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsSecret
metadata:
    name: example-secret
spec:
    secrets:
        - name: my-secret-name-1
          labels:
            label1: value1
          stringData:
            data-name0: ENC[AES256_GCM,data:m/VG12vSPVP4NS0=,iv:tB3zBrRdtkaB9SdyWfOH5/BT0fH6QMRLtch7aFOLI/E=,tag:y+o4BuN5bRMdI0wCeA01Rw==,type:str]
          data:
            data-name1: ENC[AES256_GCM,data:MbgjZ1pvDgofEJh/p5SQbA==,iv:2e/DZXxQCDNfHt1zxCEAXFeVgtbCLdqDx4Y0JjfJH4g=,tag:s5D5s8CuWw9qHYwAy7PJPA==,type:str]
        - name: jenkins-secret
          labels:
            jenkins.io/credentials-type: usernamePassword
          annotations:
            jenkins.io/credentials-description: credentials from Kubernetes
          stringData:
            username: ENC[AES256_GCM,data:CSZ8A/b9d21tbA==,iv:9zpDqBp5MIVqFrKKGGiSQg0InlSw5O/shv86LftPzg0=,tag:E1So8yXTjiu4KwwNXztXsA==,type:str]
            password: ENC[AES256_GCM,data:3S4sOAS2bEQ=,iv:RNDuXJpVtBo8NiZr4/g6Zjjp9Gq+e9yF3tukRTA7leU=,tag:F5aAIw6KNMv+GJv2XEgYBw==,type:str]
        - name: docker-login
          type: kubernetes.io/dockerconfigjson
          stringData:
            .dockerconfigjson: ENC[AES256_GCM,data:ikWS88qwtt2i+sFbT1QtkLbV3bzloAwKskDLd3ypJVglVwLVmm+0CJ1VnyemAHcLRM56M/k0/AM76gz0HBQ+RnAKuq9IqJc8My6gOLv35TDX39a+U5iH+5cvtgCa1k7Q4CjGrv2b4PrcAtWaG+esWsoFww6v4/WBcaZWsIvfzg==,iv:Re+0yieLq0dW6V35Rt3rrliWWX07voRCLUawwZ7FoOo=,tag:Vkkaro72aXwNj5BYWyfkFw==,type:str]
sops:
    kms: []
    gcp_kms: []
    azure_kv: []
    hc_vault: []
    age: []
    lastmodified: "2025-02-04T01:38:26Z"
    mac: ENC[AES256_GCM,data:CjgzOe3YFaxCj9PkKMceIpQJTgNcth625xMtKptsnNMMg7MR9VdSOORqFaw4lDXUXdGs9QvPNgTz7YKX3RwDMZTLrUnmwUm9YLpOe3/rRyY/E1pKgqr43W0E8pNnWtjQlmgbRdLd4yNDnvwLRnL66aoa9WvHqNr4CoQXtDhAf2M=,iv:6ftsNfk3DpHovrqBs4h7vbP0UCqnYI7cYrbXJlwQkHg=,tag:2hgv8q5fNxe61Maxy7uzKg==,type:str]
    pgp:
        - created_at: "2025-02-04T01:38:26Z"
          enc: |-
            -----BEGIN PGP MESSAGE-----

            hQIMA4Zyy+rN8BAMARAAgTsH6dIWYQVGsGz40KHEngftxfRssqeQmSfH1IqWIUpi
            2bBqyI3d4WGFzv3WhZNQUHL3tGclTm/zyrKfaJWTuB2mAkaEiExM/Ee8ArKLJPng
            Z/xysuJvYDqVgKe+Ul/VMb7N4y0MniUjjVpR52C5a6z3WAknXO3ai/1WGrD1bLiE
            lZ3bY8k17bIpct/y1NR0X2EaxvQKyS/SE5/eKrj7W32ryrySKHbXcizGMaKle35x
            kalyETfOxYgyfNc5sBrsOQlMY8emFnMmfNuOOyON7Qde14s91YU1Vx49u5wl7UlD
            uKaGY9KuWdyXnpHYX7XavHgPHEbhtnySIcwGbxJs9KrzQoL47+AUUT9AaZM3bsep
            sddP/KM0q9P8aYO61aYViI42KqLJBfsvb5IN+7Qf5/7iO8SBEds6tH52gJ34nBMO
            YB7GGbkfyjb5VD1bM5Ys5EP0sXnl/kTreWd1ZDEF6iWQMXzQg9DMVPkTdy7HN+4h
            9Siha4EX9pLpreCrh1xq3DO8h6rTvER5d/kEQMUFRh8AxnlqXCgNfNkqkM5vSE5r
            E7Zg8CcxT5eSzYiRCO+rw5fAPJrSJ7PS+RGQPOBiQZEECewsSgRRwaH9LWeC9OGJ
            iqc+kHU3t/qc6uathtIf3lappyn53DQGrCkupxBxSo8pq9ibQ9r41Z0WSnZgXZnU
            ZgEJAhAXUFslMGbzR3lynTL8SSQ8UKXpq1RAylon1E4CSkKnKBW9JcXVPHvlDqH7
            p/+Ht/2+q/GX13w0hIPwgt+AWZm/iUEZLn8Mn6B3JZpQxEBq1fMziVj7RZ5MsjNV
            6WlhMCHj3w==
            =3iNP
            -----END PGP MESSAGE-----
          fp: CE411B68660C33B0F83A4EBD56FDA28155A45CB1
    encrypted_regex: ^(data|stringData)$
    version: 3.8.1
```

**IMPORTANT**: The operator only decrypts fields `.data` and `.stringData` in `.spec.secrets`. All the other fields must not be encrypted. This allows for customization without possesing the private key.

Let's apply the new secret:

```bash
kubectl apply -f secret-key-1.enc.yaml
sopssecret.addons.projectcapsule.dev/example-secret created
```

If we look at the secret, we can instantly see if everything is alright or not

```bash
$ kubectl get sopssecret
NAME             SECRETS   STATUS     AGE   MESSAGE
example-secret   0         NotReady   50s   secret default/example-secret has no decryption providers
```

Currently, this secret can not be encrypted, because no provider is selecting it. To change that, we have to label the secret with `sops-secret=true`, because that's what we are selecting with the provider.

```bash
kubectl label sopssecret example-secret sops-secret=true
sopssecret.addons.projectcapsule.dev/example-secret labeled
```

Now since our provider selects the secret, it was decrypted successfully:

```bash
kubectl get sopssecret
NAME             SECRETS   STATUS   AGE     MESSAGE
example-secret   3         Ready    2m56s   Reconcilation Succeded
```

You can now also see the secrets being replicated in the namespace the `SopsSecret` was created:

```bash
kubectl get secret
NAME               TYPE     DATA   AGE
docker-login       Opaque   1      105s
jenkins-secret     Opaque   2      105s
my-secret-name-1   Opaque   2      106s
```
