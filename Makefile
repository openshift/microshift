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

RELEASE_BASE := 4.12.0
RELEASE_PRE := ${RELEASE_BASE}-0.microshift

# Overload SOURCE_GIT_TAG value set in vendor/github.com/openshift/build-machinery-go/make/lib/golang.mk
# because since it doesn't work with our version scheme.
SOURCE_GIT_TAG :=$(shell git describe --tags --abbrev=7 --match '$(RELEASE_PRE)*' 2>/dev/null || echo '${RELEASE_PRE}-${TIMESTAMP}-untagged')

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
	GC_FLAGS =-gcflags "all=-N -l"
else
	# strip everything we can
	LD_FLAGS =-w -s
	GC_FLAGS =
endif


GO_LD_FLAGS := $(GC_FLAGS) -ldflags "-X k8s.io/component-base/version.gitMajor=1 \
                   -X k8s.io/component-base/version.gitMajor=1 \
                   -X k8s.io/component-base/version.gitMinor=24 \
                   -X k8s.io/component-base/version.gitVersion=v1.24.0 \
                   -X k8s.io/component-base/version.gitCommit=07c9eb7 \
                   -X k8s.io/component-base/version.gitTreeState=clean \
                   -X k8s.io/component-base/version.buildDate=$(BIN_TIMESTAMP) \
                   -X k8s.io/client-go/pkg/version.gitMajor=1 \
                   -X k8s.io/client-go/pkg/version.gitMinor=24 \
                   -X k8s.io/client-go/pkg/version.gitVersion=v1.24.0 \
                   -X k8s.io/client-go/pkg/version.gitCommit=07c9eb7 \
                   -X k8s.io/client-go/pkg/version.gitTreeState=clean \
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

# targets "all:" and "build:" defined in vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk
# Disable CGO when building microshift binary
all: export CGO_ENABLED=0

build: export CGO_ENABLED=0

microshift: build

.PHONY: verify-images
verify: verify-images
verify-images:
	./scripts/verify_images.sh

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

cross-build-linux-amd64:
	+$(MAKE) _build_local GOOS=linux GOARCH=amd64
.PHONY: cross-build-linux-amd64

cross-build-linux-arm64:
	+$(MAKE) _build_local GOOS=linux GOARCH=arm64
.PHONY: cross-build-linux-arm64

cross-build: cross-build-linux-amd64 cross-build-linux-arm64
.PHONY: cross-build

rpm:
	RELEASE_BASE=${RELEASE_BASE} \
	RELEASE_PRE=${RELEASE_PRE} \
	SOURCE_GIT_TAG=${SOURCE_GIT_TAG} \
	SOURCE_GIT_COMMIT=${SOURCE_GIT_COMMIT} \
	SOURCE_GIT_TREE_STATE=${SOURCE_GIT_TREE_STATE} \
	./packaging/rpm/make-rpm.sh rpm local
.PHONY: rpm

srpm:
	RELEASE_BASE=${RELEASE_BASE} \
	RELEASE_PRE=${RELEASE_PRE} \
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
	for p in $(wildcard scripts/auto-rebase/rebase_patches/*.patch); do \
		echo "Applying patch $$p"; \
		git mailinfo /dev/null /dev/stderr 2<&1- < $$p | git apply; \
	done
.PHONY: vendor
