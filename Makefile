all: build
.PHONY: all

DO_LOCAL:=1 # default false
DO_STATIC:=1 # default false
BUILD_CFG:=./build/Dockerfile
BUILD_TAG:=microshift-build
SRC_ROOT:=$(shell pwd)
BIN:=./_output/bin/microshift

CTR_CMD:=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
CACHE_VOL=go_cache

STATIC_OPTS=
ifeq ($(DO_STATIC), 0)
STATIC_OPTS=--ldflags '-extldflags "-static"'
endif

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

.PHONY: build_local
build_local:
	 GOOS=linux GOARCH=amd64 go build $(STATIC_OPTS) $(GO_LD_FLAGS) -tags ${TAGS} -mod vendor  -o _output/bin/microshift cmd/main.go

.PHONY: .init
.init:
	# docker will ignore volume create calls if the volume name already exists, but podman will fail, so ignore errors
	-$(CTR_CMD) volume create --label name=microshift-build $(CACHE_VOL)
	$(CTR_CMD) build -t $(BUILD_TAG) -f $(BUILD_CFG) ./build

.PHONY: build_ctr
build_ctr: .init
	$(CTR_CMD) run -v $(CACHE_VOL):/mnt/cache -v $(SRC_ROOT):/opt/app-root/src/github.com/microshift:z $(BUILD_TAG) DO_STATIC=$(DO_STATIC)

.PHONY: build
ifeq ($(DO_LOCAL), 0)
build: build_local
else
build: build_ctr
endif

.PHONY: vendor
vendor:
	./hack/vendoring.sh

.PHONY: clean
clean:
	rm -f _output/bin/microshift
ifdef CTR_CMD
	$(CTR_CMD) system prune --filter label=name=microshift-build -f
endif

