# Include openshift build-machinery-go libraries
include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

SRC_ROOT :=$(shell pwd)

BUILD_CFG :=./images/build/Dockerfile
IMAGE_TAG :=quay.io/microshift/microshift
OUTPUT_DIR :=_output
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin

CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))

GO_EXT_LD_FLAGS :=-extldflags '-static'
GO_LD_EXTRAFLAGS :=-X k8s.io/component-base/version.gitMajor=1 \
                   -X k8s.io/component-base/version.gitMinor=20 \
                   -X k8s.io/component-base/version.gitVersion=v1.20.1 \
                   -X k8s.io/component-base/version.gitCommit=5feb30e1bd3620 \
                   -X k8s.io/component-base/version.gitTreeState=clean \
                   -X k8s.io/component-base/version.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
                   -X k8s.io/client-go/pkg/version.gitMajor=1 \
                   -X k8s.io/client-go/pkg/version.gitMinor=20 \
                   -X k8s.io/client-go/pkg/version.gitVersion=v1.20.1 \
                   -X k8s.io/client-go/pkg/version.gitCommit=5feb30e1bd3620 \
                   -X k8s.io/client-go/pkg/version.gitTreeState=clean \
                   -X k8s.io/client-go/pkg/version.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
                   $(GO_EXT_LD_FLAGS) \
                   -s -w

# These tags make sure we can statically link and avoid shared dependencies
GO_BUILD_FLAGS :=-tags 'include_gcs include_oss containers_image_openpgp gssapi providerless netgo osusergo'

# targets "all:" and "build:" defined in vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk

microshift: build-containerized-cross-build-linux-amd64
.PHONY: microshift

update: update-generated-completions
.PHONY: update

generate-versioninfo:
	SOURCE_GIT_TAG=$(SOURCE_GIT_TAG) hack/generate-versioninfo.sh
.PHONY: generate-versioninfo

###############################
# host build targets          #
###############################

_build_local:
	+@GOOS=linux GOARCH=amd64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/microshift GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_amd64

cross-build-linux-amd64:
	$(MAKE) _build_local GOOS=linux GOARCH=amd64
.PHONY: cross-build-linux-amd64

cross-build-linux-arm64:
	$(MAKE) _build_local GOOS=linux GOARCH=arm64
.PHONY: cross-build-linux-arm64

cross-build: cross-build-linux-amd64 cross-build-linux-arm64
.PHONY: cross-build

###############################
# containerized build targets #
###############################
_do_containerized_build:
	$(CTR_CMD) build -t quay.io/microshift/microshift:$(TAG) \
		-f "$(SRC_ROOT)"/images/build/Dockerfile \
		--build-arg ARCH=$(ARCH) \
		--build-arg MAKE_TARGET="cross-build-linux-amd64" \
		.
.PHONY: _do_containerized_build

build-containerized-cross-build-linux-amd64:
	$(MAKE) _do_containerized_build TAG=LINUX_AMD ARCH=amd64
	$(CTR_CMD) cp $(shell $(CTR_CMD) create --rm microshift):/usr/bin/microshift $(CROSS_BUILD_BINDIR)/linux_amd64/microshift
.PHONY: build-containerized-cross-build-linux-amd64

build-containerized-cross-build-linux-arm64:
	$(MAKE) _do_containerized_build TAG=LINUX_ARM ARCH=arm64
.PHONY: build-containerized-cross-build-linux-arm64

build-containerized-cross-build:
	$(MAKE) build-containerized-cross-build-linux-amd64 build-containerized-cross-build-linux-arm64
.PHONY: build-containerized-cross-build

vendor:
	./hack/vendoring.sh
.PHONY: vendor

clean-cross-build:
	$(RM) -r '$(CROSS_BUILD_BINDIR)'
	if [ -d '$(OUTPUT_DIR)' ]; then rmdir --ignore-fail-on-non-empty '$(OUTPUT_DIR)'; fi
.PHONY: clean-cross-build

clean: clean-cross-build
.PHONY: clean