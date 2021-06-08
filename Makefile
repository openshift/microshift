# Include the library makefile
include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

DO_LOCAL:=1 # default false
BUILD_CFG:=./images/Dockerfile
BUILD_TAG:=microshift-build
SRC_ROOT:=$(shell pwd)

CTR_CMD:=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
CACHE_VOL=go_cache

TAGS="providerless"

SOURCE_GIT_TAG ?=$(shell git describe --long --tags --abbrev=7 --match 'v[0-9]*' || echo 'v0.0.0-unknown')
SOURCE_GIT_COMMIT ?=$(shell git rev-parse --short "HEAD^{commit}" 2>/dev/null)
SOURCE_GIT_TREE_STATE ?=$(shell ( ( [ ! -d ".git/" ] || git diff --quiet ) && echo 'clean' ) || echo 'dirty')
GO_LD_EXTRAFLAGS :=-X k8s.io/component-base/version.gitMajor="1" \
                   -X k8s.io/component-base/version.gitMinor="20" \
                   -X k8s.io/component-base/version.gitVersion="v1.20.1" \
                   -X k8s.io/component-base/version.gitCommit="5feb30e1bd3620" \
                   -X k8s.io/component-base/version.gitTreeState="clean" \
                   -X k8s.io/component-base/version.buildDate="$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')" \
                   -X k8s.io/client-go/pkg/version.gitMajor="1" \
                   -X k8s.io/client-go/pkg/version.gitMinor="20" \
                   -X k8s.io/client-go/pkg/version.gitVersion="v1.20.1" \
                   -X k8s.io/client-go/pkg/version.gitCommit="5feb30e1bd3620" \
                   -X k8s.io/client-go/pkg/version.gitTreeState="clean" \
                   -X k8s.io/client-go/pkg/version.buildDate="$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')" \
                   -X github.com/openshift/microshift/pkg/version.majorFromGit="4" \
                   -X github.com/openshift/microshift/pkg/version.minorFromGit="7" \
                   -X github.com/openshift/microshift/pkg/version.versionFromGit="$(SOURCE_GIT_TAG)" \
                   -X github.com/openshift/microshift/pkg/version.commitFromGit="$(SOURCE_GIT_COMMIT)" \
                   -X github.com/openshift/microshift/pkg/version.gitTreeState="$(SOURCE_GIT_TREE_STATE)" \
                   -X github.com/openshift/microshift/pkg/version.buildDate="$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')"
GO_LD_FLAGS ?=-ldflags "$(GO_LD_EXTRAFLAGS)"

.PHONY: .init
.init:
	# docker will ignore volume create calls if the volume name already exists, but podman will fail, so ignore errors
	-$(CTR_CMD) volume create --label name=microshift-build $(CACHE_VOL)
	$(CTR_CMD) build -t $(BUILD_TAG) -f $(BUILD_CFG) ./images

.PHONY: build-containerized
build-containerized: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/microshift:z $(BUILD_TAG)

.PHONY: vendor
vendor:
	./hack/vendoring.sh

