BUILD_LOCAL=1 # 0 == false
BUILD_CFG="./build/Dockerfile"
BUILD_TAG="ushift-build"
SRC_ROOT="$(shell pwd)"
BIN="./_output/bin/ushift"

CTR_CMD=$(or $(shell which podman), $(shell which docker))

all: build

.PHONY: build_local
build_local:
	go build -v -mod vendor -o _output/bin/ushift cmd/main.go

.PHONY: init
init:
	$(CTR_CMD) build -t $(BUILD_TAG) -f $(BUILD_CFG) ./build

.PHONY: build_ctr
build_ctr: init
	$(CTR_CMD) run -v $(SRC_ROOT):/opt/app-root/src/github.com/microshift $(BUILD_TAG)

.PHONY: build
ifeq ($(BUILD_LOCAL), 0)
build: build_local
else
build: build_ctr
endif

clean:
	rm -f _output/bin/ushift
