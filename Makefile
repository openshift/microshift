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

RELEASE_BASE := 4.8.0
RELEASE_PRE := ${RELEASE_BASE}-0.microshift

# Overload SOURCE_GIT_TAG value set in vendor/github.com/openshift/build-machinery-go/make/lib/golang.mk
# because since it doesn't work with our version scheme.
SOURCE_GIT_TAG :=$(shell git describe --tags --abbrev=7 --match '$(RELEASE_PRE)*' || echo '4.8.0-0.microshift-unknown')

EMBEDDED_GIT_TAG ?= ${SOURCE_GIT_TAG}
EMBEDDED_GIT_COMMIT ?= ${SOURCE_GIT_COMMIT}
EMBEDDED_GIT_TREE_STATE ?= ${SOURCE_GIT_TREE_STATE}


SRC_ROOT :=$(shell pwd)

IMAGE_REPO :=quay.io/microshift/microshift
IMAGE_REPO_AIO :=quay.io/microshift/microshift-aio
OUTPUT_DIR :=_output
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin
FROM_SOURCE :=false
CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
ARCH :=$(shell uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/")
IPTABLES :=nft

# restrict included verify-* targets to only process project files
GO_PACKAGES=$(go list ./cmd/... ./pkg/...)

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
                   -X k8s.io/component-base/version.gitMinor=21 \
                   -X k8s.io/component-base/version.gitVersion=v1.21.0 \
                   -X k8s.io/component-base/version.gitCommit=c3b9e07a \
                   -X k8s.io/component-base/version.gitTreeState=clean \
                   -X k8s.io/component-base/version.buildDate=$(BIN_TIMESTAMP) \
                   -X k8s.io/client-go/pkg/version.gitMajor=1 \
                   -X k8s.io/client-go/pkg/version.gitMinor=21 \
                   -X k8s.io/client-go/pkg/version.gitVersion=v1.21.1 \
                   -X k8s.io/client-go/pkg/version.gitCommit=b09a9ce3 \
                   -X k8s.io/client-go/pkg/version.gitTreeState=clean \
                   -X k8s.io/client-go/pkg/version.buildDate=$(BIN_TIMESTAMP) \
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

# targets "all:" and "build:" defined in vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk
microshift: build-containerized-cross-build-linux-amd64
.PHONY: microshift

microshift-aio: build-containerized-all-in-one-amd64
.PHONY: microshift-aio

update-bindata:
	./scripts/bindata.sh
.PHONY: update-bindata

update: update-bindata
.PHONY: update

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
	BUILD=rpm \
	SOURCE_GIT_COMMIT=${SOURCE_GIT_COMMIT} \
	SOURCE_GIT_TREE_STATE=${SOURCE_GIT_TREE_STATE} RELEASE_BASE=${RELEASE_BASE}  \
	RELEASE_PRE=${RELEASE_PRE} ./packaging/rpm/make-rpm.sh local
.PHONY: rpm

srpm:
	BUILD=srpm \
	SOURCE_GIT_COMMIT=${SOURCE_GIT_COMMIT} \
	SOURCE_GIT_TREE_STATE=${SOURCE_GIT_TREE_STATE} RELEASE_BASE=${RELEASE_BASE}  \
	RELEASE_PRE=${RELEASE_PRE} ./packaging/rpm/make-rpm.sh local
.PHONY: srpm

###############################
# containerized build targets #
###############################
_build_containerized:
	@if [ -z '$(CTR_CMD)' ] ; then echo '!! ERROR: containerized builds require podman||docker CLI, none found $$PATH' >&2 && exit 1; fi
	echo BIN_TIMESTAMP==$(BIN_TIMESTAMP)
	$(CTR_CMD) build -t $(IMAGE_REPO):$(SOURCE_GIT_TAG)-linux-$(ARCH) \
		-f "$(SRC_ROOT)"/packaging/images/microshift/Dockerfile \
		--build-arg SOURCE_GIT_TAG=$(SOURCE_GIT_TAG) \
		--build-arg BIN_TIMESTAMP=$(BIN_TIMESTAMP) \
		--build-arg ARCH=$(ARCH) \
		--build-arg MAKE_TARGET="cross-build-linux-$(ARCH)" \
		--build-arg FROM_SOURCE=$(FROM_SOURCE) \
		--platform="linux/$(ARCH)" \
		.
.PHONY: _build_containerized

_build_containerized_aio:
	@if [ -z '$(CTR_CMD)' ] ; then echo '!! ERROR: containerized builds require podman||docker CLI, none found $$PATH' >&2 && exit 1; fi
	echo BIN_TIMESTAMP==$(BIN_TIMESTAMP)
	$(CTR_CMD) build -t $(IMAGE_REPO_AIO):$(SOURCE_GIT_TAG)-linux-$(IPTABLES)-$(ARCH) \
		-f "$(SRC_ROOT)"/packaging/images/microshift-aio/Dockerfile \
		--build-arg SOURCE_GIT_TAG=$(SOURCE_GIT_TAG) \
		--build-arg BIN_TIMESTAMP=$(BIN_TIMESTAMP) \
		--build-arg ARCH=$(ARCH) \
		--build-arg MAKE_TARGET="cross-build-linux-$(ARCH)" \
		--build-arg FROM_SOURCE=$(FROM_SOURCE) \
		--build-arg IPTABLES=$(IPTABLES) \
		--platform="linux/$(ARCH)" \
		.
.PHONY: _build_containerized_aio

build-containerized-cross-build-linux-amd64:
	+$(MAKE) _build_containerized ARCH=amd64
.PHONY: build-containerized-cross-build-linux-amd64

build-containerized-cross-build-linux-arm64:
	+$(MAKE) _build_containerized ARCH=arm64
.PHONY: build-containerized-cross-build-linux-arm64

build-containerized-cross-build:
	+$(MAKE) build-containerized-cross-build-linux-amd64
	+$(MAKE) build-containerized-cross-build-linux-arm64
.PHONY: build-containerized-cross-build

build-containerized-all-in-one-cross-build:
	+$(MAKE) build-containerized-all-in-one-amd64
	+$(MAKE) build-containerized-all-in-one-arm64
.PHONY: build-containerized-all-in-one-cross-build

build-containerized-all-in-one-amd64:
	+$(MAKE) _build_containerized_aio ARCH=amd64
.PHONY: build-containerized-all-in-one

build-containerized-all-in-one-arm64:
	+$(MAKE) _build_containerized_aio ARCH=arm64
.PHONY: build-containerized-all-in-one

build-containerized-all-in-one-iptables-arm64:
	+$(MAKE) _build_containerized_aio ARCH=arm64 IPTABLES=iptables
.PHONY: build-containerized-all-in-one-iptables-arm64

###############################
# dev targets                 #
###############################

clean-cross-build:
	$(RM) -r '$(CROSS_BUILD_BINDIR)'
	$(RM) -rf $(OUTPUT_DIR)/staging
	if [ -d '$(OUTPUT_DIR)' ]; then rmdir --ignore-fail-on-non-empty '$(OUTPUT_DIR)'; fi
.PHONY: clean-cross-build

clean: clean-cross-build
.PHONY: clean

release: SOURCE_GIT_TAG=$(RELEASE_PRE)-$(TIMESTAMP)
release:
	./scripts/release.sh --token $(TOKEN) --version $(SOURCE_GIT_TAG)
.PHONY: release

release-nightly:
	./scripts/release.sh --nightly --version $(SOURCE_GIT_TAG)
.PHONY: release-nightly
