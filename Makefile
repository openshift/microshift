# Include openshift build-machinery-go libraries
include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

BUILD_CFG :=./images/build/Dockerfile
BUILD_IMAGE :=microshift-builder
SRC_ROOT :=$(shell pwd)
DEST_ROOT :=/opt/app-root/src/github.com/redhat-et/microshift

CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
CACHE_VOL =go_cache

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

OUTPUT_DIR :=$(shell pwd)/_output
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin

# targets "all:" and "build:" defined in vendor/github.com/openshift/build-machinery-go/make/targets/golang/build.mk

microshift: build-containerized-microshift
.PHONY: microshift

update: update-generated-completions
.PHONY: update

generate-versioninfo:
	SOURCE_GIT_TAG=$(SOURCE_GIT_TAG) hack/generate-versioninfo.sh
.PHONY: generate-versioninfo

cross-build-linux-amd64:
	+@GOOS=linux GOARCH=amd64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/microshift GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_amd64
.PHONY: cross-build-linux-amd64

cross-build-linux-arm64:
	+@GOOS=linux GOARCH=arm64 $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/microshift GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_arm64
.PHONY: cross-build-linux-arm64

cross-build-linux-ppc64le:
	+@GOOS=linux GOARCH=ppc64le $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/microshift GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_ppc64le
.PHONY: cross-build-linux-ppc64le

cross-build-linux-s390x:
	+@GOOS=linux GOARCH=s390x $(MAKE) --no-print-directory build GO_BUILD_PACKAGES:=./cmd/microshift GO_BUILD_BINDIR:=$(CROSS_BUILD_BINDIR)/linux_s390x
.PHONY: cross-build-linux-s390x

# cross-build-windows-amd64 excluded.  current git tag scheme breaks version injection, should be vX.Y.Z
cross-build: cross-build-linux-amd64 cross-build-linux-arm64 cross-build-linux-ppc64le cross-build-linux-s390x
.PHONY: cross-build

# Containerized build targets
.PHONY: .init
.init:
	# docker will ignore volume create calls if the volume name already exists, but podman will fail, so ignore errors
	-$(CTR_CMD) volume create --label name=microshift-build $(CACHE_VOL)
	$(CTR_CMD) build -t $(BUILD_IMAGE) -f $(BUILD_CFG) ./images

.PHONY: build-containerized-microshift
build-containerized-microshift:
	$(CTR_CMD) build -t microshift -f $(BUILD_CFG) .
	mkdir -p $(CROSS_BUILD_BINDIR)/linux_amd64
	$(CTR_CMD) cp $(shell $(CTR_CMD) create --rm microshift):/usr/bin/microshift $(CROSS_BUILD_BINDIR)/linux_amd64/microshift

.PHONY: build-containerized-cross-build-linux-amd64
build-containerized-cross-build-linux-amd64: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/redhat-et/microshift:z $(BUILD_IMAGE) cross-build-linux-amd64

.PHONY: build-containerized-cross-build-linux-arm64
build-containerized-cross-build-linux-arm64: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/redhat-et/microshift:z $(BUILD_IMAGE) cross-build-linux-arm64

.PHONY: build-containerized-cross-build-linux-ppc64le
build-containerized-cross-build-linux-ppc64le: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/redhat-et/microshift:z $(BUILD_IMAGE) cross-build-linux-ppc64le

.PHONY: build-containerized-cross-build-linux-s390x
build-containerized-cross-build-linux-s390x: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/redhat-et/microshift:z $(BUILD_IMAGE) cross-build-linux-s390x

.PHONY: build-containerized-cross-build
build-containerized-cross-build: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/redhat-et/microshift:z $(BUILD_IMAGE) cross-build

.PHONY: vendor
vendor:
	./hack/vendoring.sh

clean-cross-build:
	$(RM) -r '$(CROSS_BUILD_BINDIR)'
	$(RM) cmd/microshift/microshift.syso
	if [ -d '$(OUTPUT_DIR)' ]; then rmdir --ignore-fail-on-non-empty '$(OUTPUT_DIR)'; fi
.PHONY: clean-cross-build

clean: clean-cross-build
