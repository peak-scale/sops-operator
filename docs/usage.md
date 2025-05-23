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
  - matchLabels:
       "sops-secret": "true"
```

**important:** In this case the namespace has the value `secrets: sure`.

### Create AGE key

Fist you need to create a keypair with age.

```bash
age-keygen -o key.txt
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
> * Openbao/Vault
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

### Gnu OpenPGP

> [!NOTE]
> [Upstream source for some of the material](https://fluxcd.io/flux/guides/mozilla-sops/#encrypting-secrets-using-age)

For pgp-keys to be considered, they must have the file extensions `.asc` within the secret, otherwise they are not recognized. So something like this:

```yaml
apiVersion: v1
data:
  sops.asc: LS0tLS1CRUdJTiBQR1AgUFJJVkFURSBLRVkgQkx...
kind: Secret
metadata:
  labels:
    sops.addons.projectcapsule.dev: "true"
  name: pgp-key-1
  namespace: user-ns-1
```

Use openPGP to generate a key-pair:

```shell
export KEY_NAME="key.dev-team"
export KEY_COMMENT="sops secret key"

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
uid        [ ultimativ ] key.dev-team (sops secret key)
ssb   rsa4096 2025-05-16 [SEA]
```

Export the key fingerprint:

```shell
export KEY_FP="02D183E768A118979D338F3D61BFB7FAE4690165"
```

Generate a Key-Secret for pgp:

```shell
gpg --export-secret-keys --armor "${KEY_FP}" |
kubectl create secret generic sops-gpg \
--from-file=sops.asc=/dev/stdin \
--namespace=default\
```

Label secret correctly, to be considered by the operator:

```shell
kubectl label secret sops-pgp sops.addons.projectcapsule.dev=true
```

Use the public key the generate a [Sops-Configuration](#sops-configuration):

```shell
cat <<EOF > ./.sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    pgp: ${KEY_FP}
EOF
```

Or if you would like to use multiple keys to decrypt secrets via [key-groups](#key-groups):

```yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    shamir_threshold: 1
    key_groups:
      - pgp:
          - CE411B68660C33B0F83A4EBD56FDA28155A45CB1
          - 60684ED5F92EA3FD960E83E6CB8BC811D17A58DE
```

#### Public Key

> [!NOTE]
> Optional

Store the public key in the repository. This allows anyone to encrypt with the public key, as it need to be imported into the local keyring.

```shell
gpg --export --armor "${KEY_FP}" > .sops.pub.asc
```

Other users ay import it to their local keyring with:

```shell
gpg --import .sops.pub.asc
```

### AGE

For age-keys to be considered, they must have the file extensions `.agekey` within the secret, otherwise they are not recognized. So something like this:

```yaml
apiVersion: v1
data:
  age.agekey: IyBjcmVhdGVkOiAyMDI1LTA1LTE1VDE1OjM2OjQ5KzAyOjAwCiMgcHVibGljIGtleTogYWdlMXM3dDJ2azJjcmx4YXVtZ203Y2FjczU2OHh3dXRranM1MzVwbGE2OWt0NncwMDZ0N3dnenFoa2Z3dnAKQUdFLVNFQ1JFVC1LRVktMUFQUFdFS0VTRkRHMlhWQVhYMzgzOUdBN1FDVkw4UURKV0dRVzBQUzNQNjY1RERHVkFNSFNXRUtLS04=
kind: Secret
metadata:
  name: age-key-1
  namespace: user-ns-1
  labels:
    sops.addons.projectcapsule.dev: "true"
```

Use AGE to generate a key-pair:

```shell
age-keygen -o age.agekey

Public key: age15ts05pwkfhm339ym9f2tpe3kpc97aawmsyep293a6scverreyakq889cpd
```

Generate a Key-Secret for Age:

```shell
cat age.agekey |
kubectl create secret generic sops-age \
--from-file=age.agekey=/dev/stdin \
--namespace=default
```

Label secret correctly, to be considered by the operator:

```shell
kubectl label secret sops-age sops.addons.projectcapsule.dev=true
```

Use the public key the generate a [Sops-Configuration](#sops-configuration):

```shell
export AGE_PUB_KEY="age15ts05pwkfhm339ym9f2tpe3kpc97aawmsyep293a6scverreyakq889cpd"
cat <<EOF > ./.sops.yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    age: >-
      ${AGE_PUB_KEY}
EOF
```

Or if you would like to use multiple keys to decrypt secrets:

```yaml
creation_rules:
  - path_regex: .*.yaml
    encrypted_regex: ^(data|stringData)$
    age: >-
      age15ts05pwkfhm339ym9f2tpe3kpc97aawmsyep293a6scverreyakq889cpd,
      age1dffcwct9zstd038u8f4a33jey3d04gwrpnznc0xwfc3n0ec8nyeq2jvhyr
```

Encrypt relevant [SopsSecrets](#sopssecrets). You must be in the same directory, where the `.sops.yaml` resides, for the rules to apply:

```shell
sops -e -i deploy/prod/secret-env.yaml
```

Secret encrypted and ready to be applied and pushed.

### Vault/Openbao

Initialize relevant client environments. The `VAULT_ADDR` should be the public vault address. In this example it's for a local setup.

```shell
export VAULT_ADDR=http://openbao.openbao.svc.cluster.local:8200
export VAULT_TOKEN=root
```

Verify the connection with instance is successfull:

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

Generate a Key-Secret for Vault-Token, Property must be named ``:

```shell
echo $VAULT_TOKEN |
kubectl create secret generic sops-hcvault \
--from-file=sops.vault-token=/dev/stdin \
--namespace=default
```

Label secret correctly, to be considered by the operator:

```shell
kubectl label secret sops-hcvault sops.addons.projectcapsule.dev=true
```


Use the public key the generate a [Sops-Configuration](#sops-configuration):

```shell
cat <<EOF > ./.sops.yaml
creation_rules:
    - path_regex: .*.yaml
      encrypted_regex: ^(data|stringData)$
      hc_vault_transit_uri: "${VAULT_ADDR}$/v1/sops/keys/key-1"

    - path_regex: .*prod.yaml
      encrypted_regex: ^(data|stringData)$
      hc_vault_transit_uri: "${VAULT_ADDR}/v1/sops/keys/key-2"
```

Or if you would like to use multiple keys to decrypt secrets:

```shell
cat <<EOF > ./.sops.yaml
creation_rules:
    - path_regex: *.yaml
      encrypted_regex: ^(data|stringData)$
      shamir_threshold: 1
      key_groups:
        - hc_vault:
            - "${VAULT_ADDR}/v1/sops/keys/key-1"
            - "${VAULT_ADDR}/v1/sops/keys/key-2"
```

#### Setup Keys

Enable Transit:

```shell
bao secrets enable -path=sops transit
```

Create Encryption-Keys:

```shell
bao write -f sops/keys/key-1
bao write -f sops/keys/key-2
```

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
    labels:
      "sops-secret": "true"
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

with age it looks like this:
```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: SopsSecret
metadata:
    name: example-secret
    labels:
      "sops-secret": "true"
spec:
    secrets:
        - name: my-secret-name-1
          labels:
            label1: value1
          stringData:
            data-name0: ENC[AES256_GCM,data:rzeUm9qWZZoZPo8=,iv:VYKdM8RYW5ksLWdGiq3GF4g9GQDwyBVSsujf/SaqmO4=,tag:5+PHfnV+269GmG4nBmLWMA==,type:str]
          data:
            data-name1: ENC[AES256_GCM,data:2JWdH24EMdKkBjlvFbHlRg==,iv:H1wRXMjXmF4ZPn8h3SxSWmQDvwcGh3KErXHUxbkz6PM=,tag:HnV79rychvI4CZJotp8mNQ==,type:str]
        - name: jenkins-secret
          labels:
            jenkins.io/credentials-type: usernamePassword
          annotations:
            jenkins.io/credentials-description: credentials from Kubernetes
          stringData:
            username: ENC[AES256_GCM,data:FJzExzetwQKWhA==,iv:kT2DpN+fuhAmLN1FtgPR6JjC5uQtUnpUYRHz1Q/9hJs=,tag:R+WyLU0R6kGE8/6buwcN7Q==,type:str]
            password: ENC[AES256_GCM,data:v4+8eyfUw5A=,iv:ib0VCmSTs6alRot3MVl5fa0x3jN/xTkiLghzOPrxKB8=,tag:l+fjDZEhCNO6uc6b145Emw==,type:str]
        - name: docker-login
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
