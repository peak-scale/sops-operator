# Version
GIT_HEAD_COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION         ?= $(or $(shell git describe --abbrev=0 --tags --match "v*" 2>/dev/null),$(GIT_HEAD_COMMIT))
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)

# Defaults
REGISTRY        ?= ghcr.io
REPOSITORY      ?= peak-scale/sops-operator
GIT_TAG_COMMIT  ?= $(shell git rev-parse --short $(VERSION))
GIT_MODIFIED_1  ?= $(shell git diff $(GIT_HEAD_COMMIT) $(GIT_TAG_COMMIT) --quiet && echo "" || echo ".dev")
GIT_MODIFIED_2  ?= $(shell git diff --quiet && echo "" || echo ".dirty")
GIT_MODIFIED    ?= $(shell echo "$(GIT_MODIFIED_1)$(GIT_MODIFIED_2)")
GIT_REPO        ?= $(shell git config --get remote.origin.url)
BUILD_DATE      ?= $(shell git log -1 --format="%at" | xargs -I{} sh -c 'if [ "$(shell uname)" = "Darwin" ]; then date -r {} +%Y-%m-%dT%H:%M:%S; else date -d @{} +%Y-%m-%dT%H:%M:%S; fi')
IMG_BASE        ?= $(REPOSITORY)
IMG             ?= $(IMG_BASE):$(VERSION)
FULL_IMG          ?= $(REGISTRY)/$(IMG_BASE)

## Kubernetes Version Support
KUBERNETES_SUPPORTED_VERSION ?= v1.32.0

## Tool Binaries
KUBECTL ?= kubectl
HELM ?= helm

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: golint
golint: golangci-lint
	$(GOLANGCI_LINT) run -c .golangci.yml --fix

manifests: controller-gen
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true paths="./..." output:crd:artifacts:config=charts/sops-operator/crds
	make apidocs

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) crd:generateEmbeddedObjectMeta=true object:headerFile="hack/boilerplate.go.txt" paths="./..."


apidocs: TARGET_DIR      := $(shell mktemp -d)
apidocs: apidocs-gen generate
	$(APIDOCS_GEN) crdoc --resources charts/sops-operator/crds --output docs/reference.md --template ./hack/templates/crds.tmpl

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: generate manifests
	@GO111MODULE=on go test -v $(shell go list ./... | grep -v "e2e") -coverprofile coverage.out

.PHONY: test-clean
test-clean: ## Clean tests cache
	@go clean -testcache

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run -c .golangci.yml

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run -c .golangci.yml --fix

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

####################
# -- Docker
####################

KO_PLATFORM     ?= linux/$(GOARCH)
KOCACHE         ?= /tmp/ko-cache
KO_REGISTRY     := ko.local
KO_TAGS         ?= "latest"
ifdef VERSION
KO_TAGS         := $(KO_TAGS),$(VERSION)
endif

BASE_DOCKERFILE ?= Dockerfile.base
BASE_IMAGE_TAG ?= ko.local/peak-scale/sops-operator:base
BASE_BUILD_ARGS ?= --load
KO_DEFAULTBASEIMAGE := $(BASE_IMAGE_TAG)

LD_FLAGS        := "-X main.Version=$(VERSION) \
					-X main.GitCommit=$(GIT_HEAD_COMMIT) \
					-X main.GitTag=$(VERSION) \
					-X main.GitTreeState=$(GIT_MODIFIED) \
					-X main.BuildDate=$(BUILD_DATE) \
					-X main.GitRepo=$(GIT_REPO)"

# Docker Image Build
# ------------------

.PHONY: build-base-image
build-base-image: ## Build base image using Docker Buildx
	@docker buildx build ${BASE_BUILD_ARGS} \
		--platform=$(KO_PLATFORM) \
		--tag ${BASE_IMAGE_TAG} \
		--file $(BASE_DOCKERFILE) .

.PHONY: ko-build-controller
ko-build-controller: ko build-base-image
	@echo Building Controller $(FULL_IMG) - $(KO_TAGS) >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(FULL_IMG) KO_DEFAULTBASEIMAGE=$(BASE_IMAGE_TAG) \
		$(KO) build ./cmd/ --bare --tags=$(KO_TAGS) --push=false --local --platform=$(KO_PLATFORM)

.PHONY: ko-build-all
ko-build-all:  ko-build-controller

# Docker Image Publish
# ------------------

REGISTRY_PASSWORD   ?= dummy
REGISTRY_USERNAME   ?= dummy

.PHONY: ko-login
ko-login: ko
	@$(KO) login $(REGISTRY) --username $(REGISTRY_USERNAME) --password $(REGISTRY_PASSWORD)

.PHONY: ko-publish-controller
ko-publish-controller: ko-login
	@echo Publishing Controller $(FULL_IMG) - $(KO_TAGS) >&2
	@LD_FLAGS=$(LD_FLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO=$(FULL_IMG) \
		$(KO) build ./cmd/ --bare --tags=$(KO_TAGS) --push=true

.PHONY: ko-publish-all
ko-publish-all: ko-publish-controller

####################
# -- Helm
####################

# Helm
SRC_ROOT = $(shell git rev-parse --show-toplevel)

helm-docs: helm-doc
	$(HELM_DOCS) --chart-search-root ./charts

helm-lint: ct
	@$(CT) lint --config .github/configs/ct.yaml --lint-conf .github/configs/lintconf.yaml --all --debug

helm-schema: helm-plugin-schema
	cd charts/sops-operator && $(HELM) schema -output values.schema.json

helm-test: kind ct
	@$(KIND) create cluster --wait=60s --name helm-sops-operator --image=kindest/node:$(KUBERNETES_SUPPORTED_VERSION)
	@$(MAKE) helm-test-exec
	@$(KIND) delete cluster --name helm-sops-operator

helm-test-exec: ct ko-build-all
	@$(KIND) load docker-image --name helm-sops-operator $(FULL_IMG):latest
	@$(CT) install --config $(SRC_ROOT)/.github/configs/ct.yaml --all --debug


####################
# -- Install E2E Tools
####################
CLUSTER_NAME ?= "sops-operator"

e2e: e2e-build e2e-exec e2e-destroy

e2e-build: kind
	$(KIND) create cluster --wait=60s --config e2e/kind.yaml --name $(CLUSTER_NAME) --image=kindest/node:$(KUBERNETES_SUPPORTED_VERSION)
	$(MAKE) e2e-install

e2e-exec: ginkgo e2e-init
	$(GINKGO) -r -vv ./e2e

.PHONY: e2e-init
e2e-init: sops openbao
	@VAULT_ADDR=http://openbao.openbao.svc.cluster.local:8200 VAULT_TOKEN=root bash -c '\
		$(OPENBAO) secrets enable -path=sops transit || true; \
		$(OPENBAO) write -force sops/keys/key-1; \
		$(OPENBAO) write -force sops/keys/key-2; \
		cd e2e/testdata/openbao; \
		$(SOPS) -e secret-key-1.yaml > secret-key-1.enc.yaml; \
		$(SOPS) -e secret-key-2.yaml > secret-key-2.enc.yaml; \
		$(SOPS) -e secret-multi.yaml > secret-multi.enc.yaml; \
		$(SOPS) -e secret-quorum.yaml > secret-quorum.enc.yaml';



e2e-destroy: kind
	$(KIND) delete cluster --name $(CLUSTER_NAME)

e2e-install: e2e-install-addon-helm e2e-install-distro


e2e-install-addon-helm: VERSION :=v0.0.0
e2e-install-addon-helm: KO_TAGS :=v0.0.0
e2e-install-addon-helm: e2e-load-image ko-build-all
	helm upgrade \
	    --dependency-update \
		--debug \
		--install \
		--namespace sops-operator \
		--create-namespace \
		--set 'image.pullPolicy=Never' \
		--set "image.tag=$(VERSION)" \
		--set args.logLevel=10 \
		--set args.pprof=true \
		sops-operator \
		./charts/sops-operator

e2e-install-distro:
	@$(KUBECTL) kustomize e2e/manifests/flux/ | kubectl apply -f -
	@$(KUBECTL) kustomize e2e/manifests/distro/ | kubectl apply -f -
	@$(MAKE) wait-for-helmreleases

.PHONY: e2e-load-image
e2e-load-image: ko-build-all
	kind load docker-image --name $(CLUSTER_NAME) $(FULL_IMG):$(VERSION)

wait-for-helmreleases:
	@ echo "Waiting for all HelmReleases to have observedGeneration >= 0..."
	@while [ "$$(kubectl get helmrelease -A -o jsonpath='{range .items[?(@.status.observedGeneration<0)]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}' | wc -l)" -ne 0 ]; do \
	  sleep 5; \
	done

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

####################
# -- Helm Plugins
####################
HELM_SCHEMA_VERSION   := ""
helm-plugin-schema:
	$(HELM) plugin install https://github.com/losisin/helm-values-schema-json.git --version $(HELM_SCHEMA_VERSION) || true

HELM_DOCS         := $(LOCALBIN)/helm-docs
HELM_DOCS_VERSION := v1.14.1
HELM_DOCS_LOOKUP  := norwoodj/helm-docs
helm-doc:
	@test -s $(HELM_DOCS) || \
	$(call go-install-tool,$(HELM_DOCS),github.com/$(HELM_DOCS_LOOKUP)/cmd/helm-docs@$(HELM_DOCS_VERSION))

####################
# -- Tools
####################
CONTROLLER_GEN         := $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION := v0.18.0
CONTROLLER_GEN_LOOKUP  := kubernetes-sigs/controller-tools
controller-gen:
	@test -s $(CONTROLLER_GEN) && $(CONTROLLER_GEN) --version | grep -q $(CONTROLLER_GEN_VERSION) || \
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION))

GINKGO := $(LOCALBIN)/ginkgo
ginkgo:
	$(call go-install-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo)

CT         := $(LOCALBIN)/ct
CT_VERSION := v3.13.0
CT_LOOKUP  := helm/chart-testing
ct:
	@test -s $(CT) && $(CT) version | grep -q $(CT_VERSION) || \
	$(call go-install-tool,$(CT),github.com/$(CT_LOOKUP)/v3/ct@$(CT_VERSION))

KIND         := $(LOCALBIN)/kind
KIND_VERSION := v0.29.0
KIND_LOOKUP  := kubernetes-sigs/kind
kind:
	@test -s $(KIND) && $(KIND) --version | grep -q $(KIND_VERSION) || \
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind/cmd/kind@$(KIND_VERSION))

OPENBAO         := $(LOCALBIN)/bao
OPENBAO_VERSION := v2.2.2
OPENBAO_STRIPPED  := $(subst v,,$(OPENBAO_VERSION))
OPENBAO_LOOKUP  := openbao/openbao
openbao:
	@if [ -s "$(OPENBAO)" ] && $(OPENBAO) --version | grep -q "$(OPENBAO_VERSION)"; then \
		echo "OpenBao $(OPENBAO_VERSION) already installed."; \
	else \
		mkdir -p $(LOCALBIN); \
		ARCH=$$(uname -m); \
		if [ "$$ARCH" = "x86_64" ]; then \
			curl -sL "https://github.com/$(OPENBAO_LOOKUP)/releases/download/$(OPENBAO_VERSION)/bao_$(OPENBAO_STRIPPED)_linux_amd64.pkg.tar.zst" -o bao.pkg.tar.zst; \
			mkdir -p bao && tar --zstd -xf bao.pkg.tar.zst -C bao; \
			mv bao/usr/bin/bao "$(OPENBAO)"; \
			chmod +x "$(OPENBAO)"; \
			rm -rf bao bao.pkg.tar.zst; \
		elif [ "$$ARCH" = "aarch64" ] || [ "$$ARCH" = "arm64" ]; then \
		    echo "HERE"; \
			curl -sL "https://github.com/$(OPENBAO_LOOKUP)/releases/download/$(OPENBAO_VERSION)/bao_$(OPENBAO_STRIPPED)_Linux_arm64.tar.gz" -o bao.pkg.tar; \
			mkdir -p bao && tar -xf bao.pkg.tar -C bao; \
			mv bao/bao "$(OPENBAO)"; \
			chmod +x "$(OPENBAO)"; \
			rm -rf bao bao.pkg.tar.zst; \
		else \
			echo "Unsupported architecture: $$ARCH" && exit 1; \
		fi; \
	fi

KO           := $(LOCALBIN)/ko
KO_VERSION   := v0.18.0
KO_LOOKUP    := google/ko
ko:
	@test -s $(KO) && $(KO) -h | grep -q $(KO_VERSION) || \
	$(call go-install-tool,$(KO),github.com/$(KO_LOOKUP)@$(KO_VERSION))

GOLANGCI_LINT          := $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION  := v2.1.6
GOLANGCI_LINT_LOOKUP   := golangci/golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	@test -s $(GOLANGCI_LINT) && $(GOLANGCI_LINT) -h | grep -q $(GOLANGCI_LINT_VERSION) || \
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/$(GOLANGCI_LINT_LOOKUP)/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))


APIDOCS_GEN         := $(LOCALBIN)/crdoc
APIDOCS_GEN_VERSION := v0.6.4
APIDOCS_GEN_LOOKUP  := fybrik/crdoc
apidocs-gen: ## Download crdoc locally if necessary.
	@test -s $(APIDOCS_GEN) && $(APIDOCS_GEN) --version | grep -q $(APIDOCS_GEN_VERSION) || \
	$(call go-install-tool,$(APIDOCS_GEN),fybrik.io/crdoc@$(APIDOCS_GEN_VERSION))

AGE_KEYGEN    := $(LOCALBIN)/age-keygen
AGE           := $(LOCALBIN)/age
AGE_VERSION   := v1.2.1
AGE_LOOKUP    := FiloSottile/age
age:
	@$(call go-install-tool,$(AGE_KEYGEN),filippo.io/age/cmd/age-keygen@$(AGE_VERSION))
	@$(call go-install-tool,$(AGE),filippo.io/age/cmd/age@$(AGE_VERSION))

SOPS          := $(LOCALBIN)/sops
SOPS_VERSION  := v3.10.2
SOPS_LOOKUP   := getsops/sops
sops:
	@$(call go-install-tool,$(SOPS),github.com/$(SOPS_LOOKUP)/v3/cmd/sops@$(SOPS_VERSION))

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
[ -f $(1) ] || { \
    set -e ;\
    GOBIN=$(LOCALBIN) go install $(2) ;\
}
endef
