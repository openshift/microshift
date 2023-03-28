# Export shell defined to support Ubuntu
export SHELL := $(shell which bash)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

# Include openshift build-machinery-go libraries
include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

# TIMESTAMP is defined here, and only here, and propagated through out the build flow.  This ensures that every artifact
# (binary version and image tag) all have the exact same build timestamp.  Because kubectl/oc expect
# a timestamp composed with ':'s we must adjust the string so that it is still compliant with image tag format.
export BIN_TIMESTAMP ?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
export TIMESTAMP ?=$(shell echo $(BIN_TIMESTAMP) | tr -d ':' | tr 'T' '-' | tr -d 'Z')
SOURCE_GIT_COMMIT_TIMESTAMP ?= $(shell TZ=UTC0 git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")

OCP_VERSION := $(shell jq -r '.release.base' ${PROJECT_DIR}/assets/release/release-$(shell uname -i).json)
MICROSHIFT_VERSION ?= $(subst -clean,,$(shell echo '${OCP_VERSION}-${SOURCE_GIT_COMMIT_TIMESTAMP}-${SOURCE_GIT_COMMIT}-${SOURCE_GIT_TREE_STATE}'))

# Overload SOURCE_GIT_TAG value set in vendor/github.com/openshift/build-machinery-go/make/lib/golang.mk
# because since it doesn't work with our version scheme.
SOURCE_GIT_TAG := ${MICROSHIFT_VERSION}
EMBEDDED_GIT_TAG ?= ${SOURCE_GIT_TAG}
EMBEDDED_GIT_COMMIT ?= ${SOURCE_GIT_COMMIT}
EMBEDDED_GIT_TREE_STATE ?= ${SOURCE_GIT_TREE_STATE}
MAJOR := $(shell echo $(SOURCE_GIT_TAG) | cut -f1 -d.)
MINOR := $(shell echo $(SOURCE_GIT_TAG) | cut -f2 -d.)

SRC_ROOT :=$(shell pwd)

OUTPUT_DIR :=_output
RPM_BUILD_DIR :=$(OUTPUT_DIR)/rpmbuild
ISO_DIR :=$(OUTPUT_DIR)/image-builder
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin
FROM_SOURCE :=false
CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
ARCH :=$(shell uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/")
IPTABLES :=nft
PULLSECRET :=~/.pull-secret.json
AUTHORIZED_KEYS :=$(PROJECT_DIR)/authorized_keys
IMAGE_BUILDER_ARGS := -pull_secret_file $(PULLSECRET)
ifneq ("$(wildcard $(AUTHORIZED_KEYS))","")
	IMAGE_BUILDER_ARGS := $(IMAGE_BUILDER_ARGS) -authorized_keys_file $(AUTHORIZED_KEYS)
endif

# restrict included verify-* targets to only process project files
GO_PACKAGES=$(go list ./cmd/... ./pkg/...)

# Build to a place we can ignore
GO_BUILD_BINDIR :=$(OUTPUT_DIR)/bin

ifeq ($(DEBUG),true)
	# throw all the debug info in!
	LD_FLAGS =
	GC_FLAGS =-gcflags 'all=-N -l'
else
	# strip everything we can
	LD_FLAGS =-w -s
	GC_FLAGS =
endif


include Makefile.kube_git.var
GO_LD_FLAGS := $(GC_FLAGS) -ldflags " \
                   -X k8s.io/component-base/version.gitMajor=$(KUBE_GIT_MAJOR) \
                   -X k8s.io/component-base/version.gitMinor=$(KUBE_GIT_MINOR) \
                   -X k8s.io/component-base/version.gitVersion=$(KUBE_GIT_VERSION) \
                   -X k8s.io/component-base/version.gitCommit=$(KUBE_GIT_COMMIT) \
                   -X k8s.io/component-base/version.gitTreeState=$(KUBE_GIT_TREE_STATE) \
                   -X k8s.io/component-base/version.buildDate=$(BIN_TIMESTAMP) \
                   -X k8s.io/client-go/pkg/version.gitMajor=$(KUBE_GIT_MAJOR) \
                   -X k8s.io/client-go/pkg/version.gitMinor=$(KUBE_GIT_MINOR) \
                   -X k8s.io/client-go/pkg/version.gitVersion=$(KUBE_GIT_VERSION) \
                   -X k8s.io/client-go/pkg/version.gitCommit=$(KUBE_GIT_COMMIT) \
                   -X k8s.io/client-go/pkg/version.gitTreeState=$(KUBE_GIT_TREE_STATE) \
                   -X k8s.io/client-go/pkg/version.buildDate=$(BIN_TIMESTAMP) \
                   -X github.com/openshift/microshift/pkg/version.majorFromGit=$(MAJOR) \
                   -X github.com/openshift/microshift/pkg/version.minorFromGit=$(MINOR) \
                   -X github.com/openshift/microshift/pkg/version.versionFromGit=$(EMBEDDED_GIT_TAG) \
                   -X github.com/openshift/microshift/pkg/version.commitFromGit=$(EMBEDDED_GIT_COMMIT) \
                   -X github.com/openshift/microshift/pkg/version.gitTreeState=$(EMBEDDED_GIT_TREE_STATE) \
                   -X github.com/openshift/microshift/pkg/version.buildDate=$(BIN_TIMESTAMP) \
                   $(LD_FLAGS)"

debug:
	@echo FLAGS:"$(GO_LD_FLAGS)"
	@echo TAG:"$(SOURCE_GIT_TAG)"
	@echo SOURCE_GIT_TAG:"$(SOURCE_GIT_TAG)"

# These tags make sure we can statically link and avoid shared dependencies
GO_BUILD_FLAGS :=-tags 'include_gcs include_oss containers_image_openpgp gssapi providerless netgo osusergo'

# Set variables for test-unit target
GO_TEST_FLAGS=$(GO_BUILD_FLAGS)
GO_TEST_PACKAGES=./cmd/... ./pkg/...

all: microshift etcd

# target "build:" defined in vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk
# Disable CGO when building microshift binary
build: export CGO_ENABLED=0

microshift: build

.PHONY: etcd
export GO_BUILD_FLAGS 
etcd:
	GO_LD_FLAGS="$(GC_FLAGS) -ldflags \"\
                   -X main.majorFromGit=$(MAJOR) \
                   -X main.minorFromGit=$(MINOR) \
                   -X main.versionFromGit=$(EMBEDDED_GIT_TAG) \
                   -X main.commitFromGit=$(EMBEDDED_GIT_COMMIT) \
                   -X main.gitTreeState=$(EMBEDDED_GIT_TREE_STATE) \
                   -X main.buildDate=$(BIN_TIMESTAMP) \
					$(LD_FLAGS)\"" \
		$(MAKE) -C etcd

.PHONY: verify verify-images verify-assets
verify: verify-images verify-assets

verify-images:
	./scripts/verify_images.sh

verify-assets:
	./scripts/auto-rebase/presubmit.py

.PHONY: verify-go verify-golangci verify-govulncheck
verify-go: verify-golangci verify-govulncheck

verify-golangci:
	./scripts/fetch_tools.sh golangci-lint && \
	./_output/bin/golangci-lint run --verbose

verify-govulncheck:
	@if ! command -v govulncheck &>/dev/null; then \
		go install golang.org/x/vuln/cmd/govulncheck@latest ; \
	fi
	govulncheck ./...

.PHONY: verify-sh
verify-sh:
	./scripts/fetch_tools.sh shellcheck && \
	./_output/bin/shellcheck $$(find . -type d \( -path ./_output -o -path ./vendor -o -path ./assets -o -path ./etcd/vendor \) -prune -o -name '*.sh' -print)

.PHONY: verify-py
verify-py:
	@if ! command -v pylint &>/dev/null; then \
		pip install pylint ; \
	fi
	pylint $$(find . -type d \( -path ./_output -o -path ./vendor -o -path ./assets -o -path ./etcd/vendor \) -prune -o -name '*.py' -print)

###############################
# post install validate       #
###############################

##@ Download utilities

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

# download-tool will curl any file $2 and install it to $1.
define download-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)" ;\
curl -sSLo "$(1)" "$(2)" ;\
chmod a+x "$(1)" ;\
}
endef

# Execute kuttl health checks against infra pods
.PHONY: validate-cluster
validate-cluster:
	cd validate-microshift && ./kuttl-test.sh 

##@ Download utilities

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

###############################
# host build targets          #
###############################

_build_local:
	@mkdir -p "$(CROSS_BUILD_BINDIR)/$(GOOS)_$(GOARCH)"
	+@GOOS=$(GOOS) GOARCH=$(GOARCH) $(MAKE) --no-print-directory build \
		GO_BUILD_PACKAGES:=./cmd/microshift \
		GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/$(GOOS)_$(GOARCH)
	+@GOOS=$(GOOS) GOARCH=$(GOARCH) \
		GO_LD_FLAGS="$(GC_FLAGS) -ldflags \"\
                   -X main.majorFromGit=$(MAJOR) \
                   -X main.minorFromGit=$(MINOR) \
                   -X main.versionFromGit=$(EMBEDDED_GIT_TAG) \
                   -X main.commitFromGit=$(EMBEDDED_GIT_COMMIT) \
                   -X main.gitTreeState=$(EMBEDDED_GIT_TREE_STATE) \
                   -X main.buildDate=$(BIN_TIMESTAMP) \
					$(LD_FLAGS)\"" \
		$(MAKE) -C etcd --no-print-directory build \
			GO_BUILD_PACKAGES:=./cmd/microshift-etcd \
			GO_BUILD_BINDIR:=../$(CROSS_BUILD_BINDIR)/$(GOOS)_$(GOARCH)

cross-build-linux-amd64:
	+$(MAKE) _build_local GOOS=linux GOARCH=amd64
.PHONY: cross-build-linux-amd64

cross-build-linux-arm64:
	+$(MAKE) _build_local GOOS=linux GOARCH=arm64
.PHONY: cross-build-linux-arm64

cross-build: cross-build-linux-amd64 cross-build-linux-arm64
.PHONY: cross-build

RPM_RELEASE := 1
rpm:
	MICROSHIFT_VERSION=${MICROSHIFT_VERSION} \
	RPM_RELEASE="${RPM_RELEASE}" \
	SOURCE_GIT_TAG=${SOURCE_GIT_TAG} \
	SOURCE_GIT_COMMIT=${SOURCE_GIT_COMMIT} \
	SOURCE_GIT_TREE_STATE=${SOURCE_GIT_TREE_STATE} \
	./packaging/rpm/make-rpm.sh rpm local
.PHONY: rpm

srpm:
	MICROSHIFT_VERSION=${MICROSHIFT_VERSION} \
	RPM_RELEASE="${RPM_RELEASE}" \
	SOURCE_GIT_TAG=${SOURCE_GIT_TAG} \
	SOURCE_GIT_COMMIT=${SOURCE_GIT_COMMIT} \
	SOURCE_GIT_TREE_STATE=${SOURCE_GIT_TREE_STATE} \
	./packaging/rpm/make-rpm.sh srpm local
.PHONY: srpm

image-build-configure:
	./scripts/image-builder/configure.sh
.PHONY: image-build-configure

image-build-iso: rpm 
	./scripts/image-builder/build.sh $(IMAGE_BUILDER_ARGS)
.PHONY: image-build-iso

iso: image-build-configure image-build-iso
.PHONY: iso

rpm-podman:
	RPM_BUILDER_IMAGE_TAG="rhel-8-release-golang-1.19-openshift-4.13"; \
	podman build \
		--volume /etc/pki/entitlement/:/etc/pki/entitlement \
		--build-arg TAG=$$RPM_BUILDER_IMAGE_TAG \
		--tag microshift-builder:$$RPM_BUILDER_IMAGE_TAG - < ./packaging/images/Containerfile.rpm-builder ; \
	podman run \
		--rm -ti \
		--volume $$(pwd):/opt/microshift \
		--volume $$(go env GOCACHE):/go/.cache \
		--env TARGET_ARCH=$(TARGET_ARCH) \
		microshift-builder:$$RPM_BUILDER_IMAGE_TAG \
		bash -ilc 'cd /opt/microshift && make rpm & pid=$$! ; trap "pkill $${pid}" INT ; wait $${pid}'
.PHONY: rpm-podman

###############################
# dev targets                 #
###############################

clean-cross-build:
	if [ -d '$(CROSS_BUILD_BINDIR)' ]; then $(RM) -rf '$(CROSS_BUILD_BINDIR)'; fi
	if [ -d '$(OUTPUT_DIR)/staging' ]; then $(RM) -rf '$(OUTPUT_DIR)/staging'; fi
	if [ -d '$(RPM_BUILD_DIR)' ]; then $(RM) -rf '$(RPM_BUILD_DIR)'; fi
	if [ -d '$(ISO_DIR)' ]; then $(RM) -rf '$(ISO_DIR)'; fi
	if [ -d '$(OUTPUT_DIR)' ]; then rmdir --ignore-fail-on-non-empty '$(OUTPUT_DIR)'; fi
.PHONY: clean-cross-build

clean: clean-cross-build
.PHONY: clean

licensecheck: microshift bin/lichen
	bin/lichen -c .lichen.yaml microshift

bin:
	mkdir -p $@

bin/lichen: bin vendor/modules.txt
	GOBIN=$(realpath ./bin) go install github.com/uw-labs/lichen@latest

vendor:
	go mod vendor
	for p in $(sort $(wildcard scripts/auto-rebase/rebase_patches/*.patch)); do \
		echo "Applying patch $$p"; \
		git mailinfo /dev/null /dev/stderr 2<&1- < $$p | git apply --reject || exit 1; \
	done
.PHONY: vendor
