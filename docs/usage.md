# Table of Contents

- [Table of Contents](#table-of-contents)
- [Usage](#usage)
- [Overview](#overview)
- [SopsProvider Custom Resource](#sopsprovider-custom-resource)
- [Generate Key Pair](#generate-key-pair)
  - [Prerequisites](#prerequisites)
  - [Option 1: Age key-pair](#option-1-age-key-pair)
    - [Generate key-pair](#generate-key-pair-1)
    - [Deploy private key](#deploy-private-key)
    - [Generate Sops Configuration](#generate-sops-configuration)
    - [Optional: Share public key](#optional-share-public-key)
  - [Option 2: Gnu OpenPGP key-pair](#option-2-gnu-openpgp-key-pair)
    - [Prerequisites](#prerequisites-1)
    - [Generate key pair](#generate-key-pair-2)
    - [Deploy private key](#deploy-private-key-1)
    - [Generate Sops Configuration](#generate-sops-configuration-1)
    - [Optional: Share public key](#optional-share-public-key-1)
  - [Option 3: Vault/Openbao key-pair](#option-3-vaultopenbao-key-pair)
    - [Prerequisites](#prerequisites-2)
    - [Put Vault token as secret in the cluster](#put-vault-token-as-secret-in-the-cluster)
    - [Configure Vault](#configure-vault)
    - [Generate Sops configuration](#generate-sops-configuration-2)
- [SopsSecret Custom Resource](#sopssecret-custom-resource)
  - [Spec](#spec)
  - [Encrypt](#encrypt)
  - [Deploy sops secret](#deploy-sops-secret)
  - [Debugging](#debugging)
- [GlobalSopsSecret Custom Resource](#globalsopssecret-custom-resource)
  - [Spec](#spec-1)
  - [Encrypt](#encrypt-1)
  - [Deploy sops secret](#deploy-sops-secret-1)
- [Recommendations](#recommendations)
  - [Mac Encryption](#mac-encryption)
  - [Key Groups](#key-groups)

# Usage

These docs describe how you can configure and use the sops-operator, mainly to use in combination with [Capsule](https://projectcapsule.dev/); although it can also be deployed stand-alone.

# Overview

The setup contains three components in order to work:
- A `SopsProvider`. This resource maps which private key can decrypt which `SopsSecrets`.
- A `SopsSecret`. This resource contains the encrypted password with the public key attached to it.
- A public/private key pair. This can be one of the following, up to your own preference:
    - **[age](#option-1-age-key-pair)**
    - **[Gnu OpenPGP](#option-2-gnu-openpgp-key-pair)**
    - **[Vault/OpenBao](#option-3-vaultopenbao-key-pair)**

Only one key pair is needed, so only create the key pair that you prefer to use.

# SopsProvider Custom Resource

The `SopsProvider` Custom Resource is essentially a connector that determines which private key can decrypt which `SopsSecrets`. In the following example, a `SopsProvider` is shown with a selector for how a private key is matched, and which `SopsSecrets` these private keys can decrypt. So a provider is basically a matcher: Where is the key which can decrypt which `SopsSecrets`, which is based on `namespaceSelectors` and `matchLabels`. When used in combination with Capsule, it is very likely to select a tenant as namespaceSelector.

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsProvider
metadata:
  name: solar-provider
spec:
  keys:
  - namespaceSelector:
      matchLabels:
        capsule.clastix.io/tenant: solar
  sops:
  - namespaceSelector:
      matchLabels:
        capsule.clastix.io/tenant: solar
```

It is also possible to not add a specific selector for the `keys` and `SopsSecrets`. In that case, it doesn't matter where the resources are located and which labels they have (apart from the required label that a private key must have: `sops.addons.projectcapsule.dev: "true"`):

```yaml
spec:
  keys:
  - matchLabels: {}
  sops:
  - matchLabels: {}
```

# Generate Key Pair

A key pair needs to be generated to encrypt/decrypt secrets.

## Prerequisites
The `sops` binary is needed. On Mac/Linux, install with `brew install sops`. For other platforms, see the [official instructions](https://github.com/getsops/sops/releases).

## Option 1: Age key-pair

### Generate key-pair

The `age` binary is needed. On Mac/Linux, install with `brew install age`. For other platforms see the [official instructions](https://github.com/FiloSottile/age?tab=readme-ov-file#installation).

Generate a key pair with `age`:
```bash
age-keygen -o key.txt
```

### Deploy private key

This key needs to be deployed to a namespace where you want to use this keypair. This must match the selector that is set in the `SopsProvider` `.spec.key` configuration, so in this case this secret should be deployed in a namespace that is part of the solar tenant. The secret should have the key of `age.agekey`:

```shell
export NAMESPACE=solar-namespace-1
export SECRETNAME=sops-age-solar
cat key.txt |
kubectl create secret generic $SECRETNAME \
  --from-file=age.agekey=/dev/stdin \
  --namespace=$NAMESPACE
```

In order to use this secret as a private key in the sops provider, the label `sops.addons.projectcapsule.dev=true` must be added:
```shell
kubectl label secret $SECRETNAME \
  --namespace=$NAMESPACE \
  sops.addons.projectcapsule.dev=true
```

### Generate Sops Configuration

Configure `sops` on your local machine with the correct public key. Extract the public key from `key.txt` and create a `.sops.yaml` configuration file:

```shell
export AGE_PUB_KEY=$(grep '^# public key:' ./key.txt | awk '{print $4}')
cat <<EOF > ./.sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    mac_only_encrypted: true
    age: >-
      ${AGE_PUB_KEY}
EOF
```
This configuration file describes that every `data` and `stringData` object should be encrypted; all the other settings will be left untouched. Also, the `age` public key is added.

### Optional: Share public key

The `.sops.yaml` contains the public key. You can safely distribute this configuration file with your team members (for example on git) so they can encrypt secrets with the same public key.

## Option 2: Gnu OpenPGP key-pair

### Prerequisites

> **Note:**
> [See the upstream source for more instructions](https://fluxcd.io/flux/guides/mozilla-sops/#encrypting-secrets-using-age)

> **Note**
> It is recommended to use `age` prior to `openPGP`.

The `gnupg` binary is needed. On Mac/Linux, install with `brew install gnupg`. For other platforms see the [official instructions](https://www.gnupg.org/download/index.html).

### Generate key pair

Use openPGP to generate a key-pair:

```shell
export KEY_NAME="key.solar"
export KEY_COMMENT="sops solar secret key"

gpg --batch --full-generate-key <<EOF
%no-protection
Key-Type: 1
Key-Length: 4096
Subkey-Type: 1
Subkey-Length: 4096
Expire-Date: 0
Name-Comment: ${KEY_COMMENT}
Name-Real: ${KEY_NAME}
EOF
```

Gather the fingerprint for your key:

```shell
gpg --list-secret-keys "${KEY_NAME}"

sec   rsa4096 2025-05-16 [SCEAR]
      02D183E768A118979D338F3D61BFB7FAE4690165
uid        [ ultimate ] key.solar (sops solar secret key)
ssb   rsa4096 2025-05-16 [SEA]
```

Export the key fingerprint:

```shell
export KEY_FP="02D183E768A118979D338F3D61BFB7FAE4690165"
```

### Deploy private key

This key needs to be deployed to a namespace where you want to use this keypair. This must match the selector that is set in the `SopsProvider` `.spec.key` configuration, so in this case this secret should be deployed in a namespace that is part of the solar tenant. The secret should have the key of `sops.asc`:

```shell
export NAMESPACE=solar-namespace-1
export SECRETNAME=sops-gpg-solar
gpg --export-secret-keys --armor "${KEY_FP}" |
kubectl create secret generic $SECRETNAME \
--from-file=sops.asc=/dev/stdin \
--namespace=$NAMESPACE
```

In order to use this secret as a private key in the sops provider, the label `sops.addons.projectcapsule.dev=true` must be added:

```shell
kubectl label secret $SECRETNAME \
  --namespace=$NAMESPACE \
  sops.addons.projectcapsule.dev=true
```

### Generate Sops Configuration

Use the public key that was gathered in the previous steps to create a `.sops.yaml` configuration file:

```shell
cat <<EOF > ./.sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    mac_only_encrypted: true
    pgp: ${KEY_FP}
EOF
```
This configuration says that every `data` and `stringData` object should be encrypted; all the other settings will be left untouched. Also, the `pgp` public key is added.

### Optional: Share public key

This public key can be shared with team members, so they can encrypt secrets with the same public key. For this to work, the public key needs to be exported. This can be published to (for example) a git repository, where team members can download this public key.

```shell
gpg --export --armor "${KEY_FP}" > .sops.pub.asc
```

Other team members can import it to their local keyring with:

```shell
gpg --import .sops.pub.asc
```

## Option 3: Vault/Openbao key-pair

### Prerequisites

Set the relevant client environments. The `VAULT_ADDR` should be the public vault address, and set also the `VAULT_TOKEN`. This `VAULT_TOKEN` will also used in the cluster to decrypt the secrets. In this example, we use a local test setup:

```shell
export VAULT_ADDR=http://openbao.openbao.svc.cluster.local:8200
export VAULT_TOKEN=root
```

Verify the connection with the instance is successful:

```shell
bao status


Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         2.2.0
Build Date      2025-03-05T13:07:08Z
Storage Type    inmem
Cluster Name    vault-cluster-f768a190
Cluster ID      9b6d0949-5c71-b180-04b8-f066ce36749d
HA Enabled      false
```

### Put Vault token as secret in the cluster

The Vault token needs to be deployed to a namespace where you want to use this keypair. This must match the selector that is set in the `SopsProvider` `.spec.key` configuration, so in this case this secret should be deployed in a namespace that is part of the solar tenant. The secret should have the key of `sops.vault-token`:

```shell
export NAMESPACE=solar-namespace-1
export SECRETNAME=sops-hcvault-solar
echo $VAULT_TOKEN |
kubectl create secret generic $SECRETNAME \
--from-file=sops.vault-token=/dev/stdin \
--namespace=$NAMESPACE
```

In order to use this secret as a private key in the sops provider, the label `sops.addons.projectcapsule.dev=true` must be added:

```shell
kubectl label secret $SECRETNAME \
  --namespace=$NAMESPACE \
  sops.addons.projectcapsule.dev=true
```

### Configure Vault

Enable transit in Bao:

```shell
bao secrets enable -path=sops transit
```

Create Encryption-Keys which are used for decryption:

```shell
bao write -f sops/keys/key-1
bao write -f sops/keys/key-2
```

### Generate Sops configuration

Use the public key to generate a [Sops-Configuration](#generate-sops-configuration):

```shell
cat <<EOF > ./.sops.yaml
creation_rules:
    - path_regex: .*.yaml
      encrypted_regex: ^(data|stringData)$
      hc_vault_transit_uri: "${VAULT_ADDR}/v1/sops/keys/key-1"

    - path_regex: .*prod.yaml
      encrypted_regex: ^(data|stringData)$
      hc_vault_transit_uri: "${VAULT_ADDR}/v1/sops/keys/key-2"
EOF
```

# SopsSecret Custom Resource

## Spec

To create a secret, use the following apiSpec with `apiVersion: addons.projectcapsule.dev/v1alpha1` and `kind: SopsSecret`. Multiple secrets can be defined in one `SopsSecret`.
The `spec` follows the following schema:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsSecret
metadata:
  name: example-secret
spec:
  secrets:
    - name: [secret-name]
      labels:
        my-label: value1
      annotations:
        my-annotation: value2
      stringData: [Plain text string to be encrypted]
      data: [base64 encoded string to be encrypted]
```

For example, this secret below will result in 3 separate Kubernetes secrets, called `my-secret-name-1`, `jenkins-test-secret`, and `docker-test-login`. Of course, it is also possible to provide one secret in the `.spec.secrets` part.

__secret.yaml__
```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsSecret
metadata:
  name: example-secret
  namespace: solar-namespace-2
spec:
  secrets:
    - name: my-secret-name-1
      labels:
        label1: value1
      stringData:
        data-name0: data-value0
      data:
        data-name1: ZGF0YS12YWx1ZTE=
    - name: jenkins-test-secret
      labels:
        "jenkins.io/credentials-type": "usernamePassword"
      annotations:
        "jenkins.io/credentials-description": "credentials from Kubernetes"
      stringData:
        username: myUsername
        password: 'Pa$$word'
    - name: docker-test-login
      type: 'kubernetes.io/dockerconfigjson'
      stringData:
        .dockerconfigjson: '{"auths":{"index.docker.io":{"username":"imyuser","password":"mypass","email":"myuser@abc.com","auth":"aW15dXNlcjpteXBhc3M="}}}'
```

## Encrypt

To encrypt the `sops-secret`, use the command `sops`. Make sure that the Sops Configuration file (`.sops.yaml`) is in the current directory.

```shell
# Encrypt to a new file
sops -e secret.yaml  > secret-encrypted.yaml

# Or encrypt in-place
sops -e -i secret.yaml
```

The `secret-encrypted.yaml` file is encrypted, resulting in encrypted strings in every `data` and `stringData` field, and additional information about the encryption method and public key in the `.spec.sops` part. In this case, the encryption was done with `age`:

```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsSecret
metadata:
  name: example-secret
  namespace: solar-namespace-2
spec:
    secrets:
        - name: my-secret-name-1
          labels:
            label1: value1
          stringData:
            data-name0: ENC[AES256_GCM,data:rzeUm9qWZZoZPo8=,iv:VYKdM8RYW5ksLWdGiq3GF4g9GQDwyBVSsujf/SaqmO4=,tag:5+PHfnV+269GmG4nBmLWMA==,type:str]
          data:
            data-name1: ENC[AES256_GCM,data:2JWdH24EMdKkBjlvFbHlRg==,iv:H1wRXMjXmF4ZPn8h3SxSWmQDvwcGh3KErXHUxbkz6PM=,tag:HnV79rychvI4CZJotp8mNQ==,type:str]
        - name: jenkins-test-secret
          labels:
            jenkins.io/credentials-type: usernamePassword
          annotations:
            jenkins.io/credentials-description: credentials from Kubernetes
          stringData:
            username: ENC[AES256_GCM,data:FJzExzetwQKWhA==,iv:kT2DpN+fuhAmLN1FtgPR6JjC5uQtUnpUYRHz1Q/9hJs=,tag:R+WyLU0R6kGE8/6buwcN7Q==,type:str]
            password: ENC[AES256_GCM,data:v4+8eyfUw5A=,iv:ib0VCmSTs6alRot3MVl5fa0x3jN/xTkiLghzOPrxKB8=,tag:l+fjDZEhCNO6uc6b145Emw==,type:str]
        - name: docker-test-login
          type: kubernetes.io/dockerconfigjson
          stringData:
            .dockerconfigjson: ENC[AES256_GCM,data:d4/wjjm43GD/dUU2aVvSQf8BANBq3Y++DKFqHWyRFC5QVG5gC1EU8GIHn1N1IGgbSM+cX3G4M3OVQlDNzjmH6TmIID6yiqnSt5XhVocoWHRiBFE8KFqphkrIqLqOKZxJMfZWvbQ7ncuV9Jv1/mo6vpG8B4dqeWC9sUi4URH40A==,iv:wXcp/hD9OPOw0s0kFiGeRyaZZt9ffST/rikS9qp6tYo=,tag:1WWHAjq1lRgfUd9HUS5bkg==,type:str]
sops:
    age:
        - recipient: age10t4z6kr0nfl7xxwrwtj9ehfl7wkp7kdy2whlpmzannppqhvfu3lsyjxqjm
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSB6VDZnMUJ5YXlndStRWlRu
            QnNGWmtkd016MjhOMTFXQURaRTg0cXRLNWc0CmNCRUxqdDRjQkNTWWw2RFdMZXJW
            SHNpWTZvWlQ4ZnpLdnVlblF5YW44eUEKLS0tIHJ3akJjeGRCTmJETlRqVmtjTTY3
            SmdPTms3TnZqc2ZDdm1KclhNWnJhOWcKwWXCTacYOynueHUeQX5ByTmajItT8NnJ
            Hfe3I4NZ72p/MbnfzmZWBFOR5ANJZ+we6vUnz1fair9MdyvQV+uhxA==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2025-05-15T07:01:38Z"
    mac: ENC[AES256_GCM,data:KxCP0JXws5+u2c7F1Hdek8mn51Ld5su+meB0nLUzPZoOR0VfSm2mTveGkz8/OsO3u8Uo9OM4dUbd+zsnYjhL6t11Eok8ePVvzkYthYQBpPtWXFLnkobpOTMWVP7FUlmTVwFIwGuUC4Wh8LaPF/jYkXowF9mylhjJLURRVM1u+3U=,iv:u3hgRmvhHB84HR4bNuPUHfYHktGXzbe4zerXftOoY54=,tag:zJTpxyJJ532DkPHSwhorog==,type:str]
    version: 3.10.2
```

## Deploy sops secret

Let's apply the new secret:

```shell
kubectl apply -f secret-encrypted.yaml
sopssecret.addons.projectcapsule.dev/example-secret created
```

If we look at the secret, we can immediately see if everything is alright or not:

```shell
kubectl get sopssecret -n solar-namespace-2
NAME             SECRETS   STATUS   AGE     MESSAGE
example-secret   3         Ready    2m56s   Reconciliation succeeded
```

You can now also see the secrets being created in the namespace where the `SopsSecret` was created:

```shell
kubectl get secret -n solar-namespace-2
NAME                TYPE     DATA   AGE
docker-test-login   Opaque   1      105s
jenkins-test-secret Opaque   2      105s
my-secret-name-1    Opaque   2      106s
```

## Debugging

If something is wrong with the decryption, it will be added as the `message` as well as to the `.status` field of the sopssecret resource:

```shell
$ kubectl get sopssecret
NAME             SECRETS   STATUS     AGE   MESSAGE
example-secret   0         NotReady   50s   secret solar-namespace-2/example-secret has no decryption providers
```
In this case, the decryption provider has not been found. That could mean a few possible things:
  - There is no `SopsProvider` created
  - The secret isn't in the correct namespace that is selected in the `SopsProvider`
  - The secret doesn't have the labels that are configured for secrets in the `SopsProvider`

```shell
kubectl label sopssecret example-secret sops-secret=true
sopssecret.addons.projectcapsule.dev/example-secret labeled
```

# GlobalSopsSecret Custom Resource

> [!IMPORTANT]
> Providers disregard the `namespaceSelector` alltogether for `GlobalSopsSecrets`. If the labels match, it's valid.

Is essentially identical to [SopsSecret](#sopssecret-custom-resource) but a cluster-scoped resource. Therefor you must provide a `namespace` for every secret item.

## Spec

To create a secret, use the following apiSpec with `apiVersion: addons.projectcapsule.dev/v1alpha1` and `kind: SopsSecret`. Multiple secrets can be defined in one `SopsSecret`.
The `spec` follows the following schema:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: GlobalSopsSecret
metadata:
  name: example-secret
spec:
  secrets:
    - name: [secret-name]
      namespace: [secret-name]
      labels:
        my-label: value1
      annotations:
        my-annotation: value2
      stringData: [Plain text string to be encrypted]
      data: [base64 encoded string to be encrypted]
```

For example, this secret below will result in 3 separate Kubernetes secrets, called `my-secret-name-1`, `jenkins-test-secret`, and `docker-test-login`. Of course, it is also possible to provide one secret in the `.spec.secrets` part.

__secret.yaml__
```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: GlobalSopsSecret
metadata:
  name: example-secret
spec:
  secrets:
    - name: my-secret-name-1
      namespace: solar-namespace-1
      labels:
        label1: value1
      stringData:
        data-name0: data-value0
      data:
        data-name1: ZGF0YS12YWx1ZTE=
    - name: jenkins-test-secret
      namespace: solar-namespace-2
      labels:
        "jenkins.io/credentials-type": "usernamePassword"
      annotations:
        "jenkins.io/credentials-description": "credentials from Kubernetes"
      stringData:
        username: myUsername
        password: 'Pa$$word'
    - name: docker-test-login
      namespace: solar-namespace-3
      type: 'kubernetes.io/dockerconfigjson'
      stringData:
        .dockerconfigjson: '{"auths":{"index.docker.io":{"username":"imyuser","password":"mypass","email":"myuser@abc.com","auth":"aW15dXNlcjpteXBhc3M="}}}'
```

## Encrypt

To encrypt the `sops-secret`, use the command `sops`. Make sure that the Sops Configuration file (`.sops.yaml`) is in the current directory.

```shell
# Encrypt to a new file
sops -e secret.yaml  > secret-encrypted.yaml

# Or encrypt in-place
sops -e -i secret.yaml
```

The `secret-encrypted.yaml` file is encrypted, resulting in encrypted strings in every `data` and `stringData` field, and additional information about the encryption method and public key in the `.spec.sops` part. In this case, the encryption was done with `age`:

```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: GlobalSopsSecret
metadata:
  name: example-secret
spec:
    secrets:
        - name: my-secret-name-1
          namespace: solar-namespace-1
          labels:
            label1: value1
          stringData:
            data-name0: ENC[AES256_GCM,data:rzeUm9qWZZoZPo8=,iv:VYKdM8RYW5ksLWdGiq3GF4g9GQDwyBVSsujf/SaqmO4=,tag:5+PHfnV+269GmG4nBmLWMA==,type:str]
          data:
            data-name1: ENC[AES256_GCM,data:2JWdH24EMdKkBjlvFbHlRg==,iv:H1wRXMjXmF4ZPn8h3SxSWmQDvwcGh3KErXHUxbkz6PM=,tag:HnV79rychvI4CZJotp8mNQ==,type:str]
        - name: jenkins-test-secret
          namespace: solar-namespace-2
          labels:
            jenkins.io/credentials-type: usernamePassword
          annotations:
            jenkins.io/credentials-description: credentials from Kubernetes
          stringData:
            username: ENC[AES256_GCM,data:FJzExzetwQKWhA==,iv:kT2DpN+fuhAmLN1FtgPR6JjC5uQtUnpUYRHz1Q/9hJs=,tag:R+WyLU0R6kGE8/6buwcN7Q==,type:str]
            password: ENC[AES256_GCM,data:v4+8eyfUw5A=,iv:ib0VCmSTs6alRot3MVl5fa0x3jN/xTkiLghzOPrxKB8=,tag:l+fjDZEhCNO6uc6b145Emw==,type:str]
        - name: docker-test-login
          namespace: solar-namespace-3
          type: kubernetes.io/dockerconfigjson
          stringData:
            .dockerconfigjson: ENC[AES256_GCM,data:d4/wjjm43GD/dUU2aVvSQf8BANBq3Y++DKFqHWyRFC5QVG5gC1EU8GIHn1N1IGgbSM+cX3G4M3OVQlDNzjmH6TmIID6yiqnSt5XhVocoWHRiBFE8KFqphkrIqLqOKZxJMfZWvbQ7ncuV9Jv1/mo6vpG8B4dqeWC9sUi4URH40A==,iv:wXcp/hD9OPOw0s0kFiGeRyaZZt9ffST/rikS9qp6tYo=,tag:1WWHAjq1lRgfUd9HUS5bkg==,type:str]
sops:
    age:
        - recipient: age10t4z6kr0nfl7xxwrwtj9ehfl7wkp7kdy2whlpmzannppqhvfu3lsyjxqjm
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSB6VDZnMUJ5YXlndStRWlRu
            QnNGWmtkd016MjhOMTFXQURaRTg0cXRLNWc0CmNCRUxqdDRjQkNTWWw2RFdMZXJW
            SHNpWTZvWlQ4ZnpLdnVlblF5YW44eUEKLS0tIHJ3akJjeGRCTmJETlRqVmtjTTY3
            SmdPTms3TnZqc2ZDdm1KclhNWnJhOWcKwWXCTacYOynueHUeQX5ByTmajItT8NnJ
            Hfe3I4NZ72p/MbnfzmZWBFOR5ANJZ+we6vUnz1fair9MdyvQV+uhxA==
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2025-05-15T07:01:38Z"
    mac: ENC[AES256_GCM,data:KxCP0JXws5+u2c7F1Hdek8mn51Ld5su+meB0nLUzPZoOR0VfSm2mTveGkz8/OsO3u8Uo9OM4dUbd+zsnYjhL6t11Eok8ePVvzkYthYQBpPtWXFLnkobpOTMWVP7FUlmTVwFIwGuUC4Wh8LaPF/jYkXowF9mylhjJLURRVM1u+3U=,iv:u3hgRmvhHB84HR4bNuPUHfYHktGXzbe4zerXftOoY54=,tag:zJTpxyJJ532DkPHSwhorog==,type:str]
    version: 3.10.2
```

## Deploy sops secret

Let's apply the new secret:

```shell
kubectl apply -f secret-encrypted.yaml
globalsopssecret.addons.projectcapsule.dev/example-secret created
```

If we look at the secret, we can immediately see if everything is alright or not:

```shell
kubectl get globalsopssecret example-secret
NAME             SECRETS   STATUS   AGE     MESSAGE
example-secret   3         Ready    2m56s   Reconciliation succeeded
```

You can now also see the secrets being created in the namespace where the `SopsSecret` was created:

```shell
kubectl get secret -n solar-namespace-2
NAME                TYPE     DATA   AGE
jenkins-test-secret Opaque   2      105s
```


# Recommendations

## Mac Encryption

By default the entire mac of the file is used when encrypting. This means you can not change anything about the encrypted file, as it will always result in a MAC-Mistmatch. In this case it's recommended to only mac the encrypted values, this is done while encrypting secrets. Either via flag:

```shell
sops --mac-only-encrypted -e -i secret.sops.yaml
```

Or in the **sops.yaml** (This was added for all examples above already):

```yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    mac_only_encrypted: true
    pgp: KEY
```

## Key Groups

[Key-Groups](https://github.com/getsops/sops?tab=readme-ov-file#216key-groups) are supported. All the required private-keys may even be distributed amongst different `SopsProviders`. As long as a `SopsSecret` is allowed to collect all the required keys from these `SopsProviders`, it will be able to decrypt. Just add the extra public key to the `.sops.yaml` configuration.

> **Note:**
> The `shamir_threshold` field specifies the minimum number of keys required to decrypt the secret.

For `age`:
```yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    age:
      - ${AGE_PUB_KEY_1}
      - ${AGE_PUB_KEY_2}
```

For `pgp`:
```yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    shamir_threshold: 1
    key_groups:
      - pgp:
          - ${PGP_PUB_KEY_1}
          - ${PGP_PUB_KEY_2}
```

For `Vault`/`OpenBao`:
```yaml
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    shamir_threshold: 1
    key_groups:
      - hc_vault:
          - "${VAULT_ADDR}/v1/sops/keys/key-1"
          - "${VAULT_ADDR}/v1/sops/keys/key-2"
```
