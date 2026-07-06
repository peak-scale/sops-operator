# SOPS Checker

`sops-checker` is a small validation binary for pre-commit style workflows. It checks that files which should be SOPS encrypted are not committed as plaintext.

The checker does not decrypt secrets. It verifies that matching files can be parsed as SOPS encrypted files.

# Installation

Download the standalone binary from the GitHub release assets:

```shell
curl -L -o sops-checker \
  https://github.com/peak-scale/sops-operator/releases/download/vX.Y.Z/sops-checker-linux-amd64
chmod +x sops-checker
sudo mv sops-checker /usr/local/bin/sops-checker
```

Use `sops-checker-linux-arm64` on ARM64 systems.

The checker is also published as a container image:

```shell
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  ghcr.io/peak-scale/sops-checker:vX.Y.Z \
  --glob 'secrets/*.yaml'
```

# Usage

Check files against the closest `.sops.yaml` discovered from each file path:

```shell
sops-checker secrets/app.yaml clusters/dev/secret.sops.yaml
```

Use a specific SOPS config:

```shell
sops-checker --config .sops.yaml secrets/app.yaml
```

Check files selected by glob patterns:

```shell
sops-checker \
  --glob 'secrets/*.yaml' \
  --glob 'clusters/*/*.sops.yaml'
```

When one or more `--glob` flags are provided, positional file arguments are ignored. This is useful for hook systems that normally pass changed files but where the repository should be checked from a fixed set of patterns.

Require every selected file to be encrypted, without consulting `.sops.yaml` creation rules:

```shell
sops-checker --require-all --glob 'secrets/*.sops.yaml'
```

Exit codes:

- `0`: all required files are encrypted
- `1`: one or more required files are not encrypted
- `2`: invalid input, unreadable files, or invalid SOPS configuration

The glob syntax is Go `filepath.Glob` syntax. `*`, `?`, and character classes are supported. Recursive `**` is not treated specially, so prefer explicit patterns such as `clusters/*/*.sops.yaml`.

# pre-commit

## Standalone Binary

Use the installed binary and let `pre-commit` pass changed files:

```yaml
repos:
  - repo: local
    hooks:
      - id: sops-checker
        name: sops-checker
        entry: sops-checker
        language: system
        files: \.sops\.(ya?ml|json|env|ini)$
```

Use fixed glob patterns and ignore the file list from `pre-commit`:

```yaml
repos:
  - repo: local
    hooks:
      - id: sops-checker
        name: sops-checker
        entry: sops-checker --glob 'secrets/*.sops.yaml' --glob 'clusters/*/*.sops.yaml'
        language: system
        pass_filenames: false
        always_run: true
```

## Container

Run the checker image while still letting `pre-commit` pass changed files:

```yaml
repos:
  - repo: local
    hooks:
      - id: sops-checker
        name: sops-checker
        entry: bash -c 'docker run --rm -v "$PWD:/work" -w /work ghcr.io/peak-scale/sops-checker:vX.Y.Z "$@"' --
        language: system
        files: \.sops\.(ya?ml|json|env|ini)$
```

Run the checker image with fixed glob patterns:

```yaml
repos:
  - repo: local
    hooks:
      - id: sops-checker
        name: sops-checker
        entry: bash -c 'docker run --rm -v "$PWD:/work" -w /work ghcr.io/peak-scale/sops-checker:vX.Y.Z --glob "secrets/*.sops.yaml'" --glob "clusters/*/*.sops.yaml"'
        language: system
        pass_filenames: false
        always_run: true
```

# prek

`prek` can use the same local hook definitions as `pre-commit`. Install the checker binary or use the container image, then keep the hook in `.pre-commit-config.yaml`.

Standalone binary:

```yaml
repos:
  - repo: local
    hooks:
      - id: sops-checker
        name: sops-checker
        entry: sops-checker --glob 'secrets/*.sops.yaml'
        language: system
        pass_filenames: false
        always_run: true
```

Container:

```yaml
repos:
  - repo: local
    hooks:
      - id: sops-checker
        name: sops-checker
        entry: bash -c 'docker run --rm -v "$PWD:/work" -w /work ghcr.io/peak-scale/sops-checker:vX.Y.Z --glob "secrets/*.sops.yaml'"'
        language: system
        pass_filenames: false
        always_run: true
```

Run it with:

```shell
prek run sops-checker --all-files
```

For changed-file mode, remove `--glob`, `pass_filenames: false`, and `always_run: true` from the hook definition.
