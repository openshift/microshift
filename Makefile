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

all: build

.PHONY: build_local
build_local:
	 GOOS=linux GOARCH=amd64 go build $(STATIC_OPTS) -mod vendor  -o _output/bin/microshift cmd/main.go

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

