# Export shell defined to support Ubuntu
export SHELL := $(shell which bash)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

# Include openshift build-machinery-go libraries
include $(PROJECT_DIR)/vendor/github.com/openshift/build-machinery-go/make/golang.mk
include $(PROJECT_DIR)/vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

# TIMESTAMP is defined here, and only here, and propagated through out the build flow.  This ensures that every artifact
# (binary version and image tag) all have the exact same build timestamp.  Because kubectl/oc expect
# a timestamp composed with ':'s we must adjust the string so that it is still compliant with image tag format.
export BIN_TIMESTAMP ?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
export TIMESTAMP ?=$(shell echo $(BIN_TIMESTAMP) | tr -d ':' | tr 'T' '-' | tr -d 'Z')
SOURCE_GIT_COMMIT_TIMESTAMP ?= $(shell TZ=UTC0 git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")

include $(PROJECT_DIR)/Makefile.version.$(shell uname -m).var
MICROSHIFT_VERSION ?= $(subst -clean,,$(shell echo '${OCP_VERSION}-${SOURCE_GIT_COMMIT_TIMESTAMP}-${SOURCE_GIT_COMMIT}-${SOURCE_GIT_TREE_STATE}'))

# Overload SOURCE_GIT_TAG value set in vendor/github.com/openshift/build-machinery-go/make/lib/golang.mk
# because since it doesn't work with our version scheme.
SOURCE_GIT_TAG := ${MICROSHIFT_VERSION}
EMBEDDED_GIT_TAG ?= ${SOURCE_GIT_TAG}
EMBEDDED_GIT_COMMIT ?= ${SOURCE_GIT_COMMIT}
EMBEDDED_GIT_TREE_STATE ?= ${SOURCE_GIT_TREE_STATE}
MAJOR := $(shell echo $(SOURCE_GIT_TAG) | awk -F'[._~-]' '{print $$1}')
MINOR := $(shell echo $(SOURCE_GIT_TAG) | awk -F'[._~-]' '{print $$2}')
PATCH := $(shell echo $(SOURCE_GIT_TAG) | awk -F'[._~-]' '{print $$3}')

SRC_ROOT :=$(shell pwd)

OUTPUT_DIR :=_output
RPM_BUILD_DIR :=$(OUTPUT_DIR)/rpmbuild
ISO_DIR :=$(OUTPUT_DIR)/image-builder
CROSS_BUILD_BINDIR :=$(OUTPUT_DIR)/bin
FROM_SOURCE :=false
CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
ARCH :=$(shell uname -m |sed -e "s/x86_64/amd64/" |sed -e "s/aarch64/arm64/")
# Image builder arguments can be overriden from the environment:
# - PULLSECRET: path to a pull secret file
# - AUTHORIZED_KEYS: path to an SSH authorized keys file
# - IMAGE_BUILDER_ARGS: any build.sh script options
PULLSECRET ?= ~/.pull-secret.json
BASE_IMAGE_BUILDER_ARGS := -pull_secret_file $(PULLSECRET)
ifdef IMAGE_BUILDER_ARGS
    BASE_IMAGE_BUILDER_ARGS += $(IMAGE_BUILDER_ARGS)
endif
AUTHORIZED_KEYS ?= $(HOME)/.ssh/authorized_keys
ifneq ("$(wildcard $(AUTHORIZED_KEYS))","")
    BASE_IMAGE_BUILDER_ARGS += -authorized_keys_file $(AUTHORIZED_KEYS)
endif
IMAGE_BUILDER_ARGS = $(BASE_IMAGE_BUILDER_ARGS)

# restrict included verify-* targets to only process project files
GO_PACKAGES=$(go list ./cmd/... ./pkg/...)

# Build to a place we can ignore
GO_BUILD_BINDIR :=$(OUTPUT_DIR)/bin

GO_CACHE :=$(shell go env GOCACHE)

ifeq ($(DEBUG),true)
	# throw all the debug info in!
	LD_FLAGS =
	GC_FLAGS =-gcflags 'all=-N -l'
else
	# strip everything we can
	LD_FLAGS =-w -s
	GC_FLAGS =
endif


include $(PROJECT_DIR)/Makefile.kube_git.var
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
                   -X github.com/openshift/microshift/pkg/version.patchFromGit=$(PATCH) \
                   -X github.com/openshift/microshift/pkg/version.versionFromGit=$(EMBEDDED_GIT_TAG) \
                   -X github.com/openshift/microshift/pkg/version.commitFromGit=$(EMBEDDED_GIT_COMMIT) \
                   -X github.com/openshift/microshift/pkg/version.gitTreeState=$(EMBEDDED_GIT_TREE_STATE) \
                   -X github.com/openshift/microshift/pkg/version.buildDate=$(BIN_TIMESTAMP) \
                   $(LD_FLAGS)"

debug:
	@echo FLAGS:"$(GO_LD_FLAGS)"
	@echo TAG:"$(SOURCE_GIT_TAG)"
	@echo SOURCE_GIT_TAG:"$(SOURCE_GIT_TAG)"
	@echo MAJOR:"$(MAJOR)"
	@echo MINOR:"$(MINOR)"
	@echo PATCH:"$(PATCH)"

GO_BUILD_FLAGS :=-tags 'include_gcs include_oss containers_image_openpgp gssapi providerless netcgo osusergo strictfipsruntime'

# Set variables for test-unit target
GO_TEST_FLAGS=$(GO_BUILD_FLAGS)
GO_TEST_PACKAGES=./cmd/... ./pkg/...

# Enable CGO when building microshift binary for access to local libraries.
# Use an environment variable to allow CI to disable when cross-compiling.
export CGO_ENABLED ?= 1

# Specify OCP build tools image tag when building rpm with podman
RPM_BUILDER_IMAGE_TAG := rhel-9-golang-latest-openshift-4.16

all: generate-config microshift etcd

microshift: build

# A target for local developer workflows. This assumes configure-vm.sh
# was already run to build and install the RPM.
.PHONY: install
install: build
	sudo systemctl stop microshift
	sudo cp $(OUTPUT_DIR)/bin/microshift* /usr/bin/
	sudo systemctl start microshift

.PHONY: etcd
export GO_BUILD_FLAGS
etcd:
	GO_LD_FLAGS="$(GC_FLAGS) -ldflags \"\
                   -X main.majorFromGit=$(MAJOR) \
                   -X main.minorFromGit=$(MINOR) \
                   -X main.patchFromGit=$(PATCH) \
                   -X main.versionFromGit=$(EMBEDDED_GIT_TAG) \
                   -X main.commitFromGit=$(EMBEDDED_GIT_COMMIT) \
                   -X main.gitTreeState=$(EMBEDDED_GIT_TREE_STATE) \
                   -X main.buildDate=$(BIN_TIMESTAMP) \
					$(LD_FLAGS)\"" \
		$(MAKE) -C etcd

# Default verify target for developers
.PHONY: verify
verify: verify-fast

# Fast verification checks that developers can/should run locally
.PHONY: verify-fast
verify-fast: verify-go verify-assets verify-sh verify-py verify-config verify-rf

# Full verification checks that should run in CI
.PHONY: verify-ci
verify-ci: verify-fast verify-images verify-licenses verify-containers

.PHONY: verify-images
verify-images:
	./scripts/verify/verify-images.sh

.PHONY: verify-licenses
verify-licenses: microshift etcd
	./scripts/verify/verify-licenses.sh

.PHONY: verify-assets
verify-assets:
	./scripts/verify/verify-assets.sh

.PHONY: verify-go
verify-go: verify-gofmt verify-golangci

.PHONY: verify-golangci
verify-golangci:
	./scripts/fetch_tools.sh golangci-lint && \
	./_output/bin/golangci-lint run --verbose --timeout 20m0s

.PHONY: verify-sh
verify-sh:
	./scripts/verify/verify-shell.sh

.PHONY: verify-py
verify-py:
	./scripts/verify/verify-py.sh

.PHONY: verify-rf
verify-rf:
	./scripts/verify/verify-rf.sh

.PHONY: verify-containers
verify-containers:
	./scripts/fetch_tools.sh hadolint && \
	./_output/bin/hadolint $$(find . -iname 'Containerfile*' -o -iname 'Dockerfile*'| grep -v "vendor\|_output\|origin")

# Vulnerability check is not run in any default verify target
# It should be run explicitly before the release to track and fix known vulnerabilities
# Note: Errors are ignored to allow listing vulnerabilities from all the dependencies
.PHONY: verify-govulncheck
verify-govulncheck: microshift etcd
	./scripts/fetch_tools.sh govulncheck
	-./_output/bin/govulncheck -mode=binary ./_output/bin/microshift
	-./_output/bin/govulncheck -mode=binary ./_output/bin/microshift-etcd

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


# check if FIPS supported
ifeq (, $(shell go doc goexperiment.Flags | grep -i strictfipsruntime))
	GOEXPERIMENT =
else
	GOEXPERIMENT = "strictfipsruntime"
endif


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
					GOEXPERIMENT=${GOEXPERIMENT} \
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

image-build-commit: rpm
	./scripts/image-builder/build.sh $(IMAGE_BUILDER_ARGS) -build_edge_commit

commit: image-build-configure image-build-commit
.PHONY: commit

rpm-podman:
	podman build \
		--volume /etc/pki/entitlement/:/etc/pki/entitlement \
		--build-arg TAG=$(RPM_BUILDER_IMAGE_TAG) \
		--authfile $(PULLSECRET) \
		--tag microshift-builder:$(RPM_BUILDER_IMAGE_TAG) - < ./packaging/images/Containerfile.rpm-builder ; \
	podman run \
		--rm -i \
		--volume $$(pwd):/opt/microshift:z \
		--env TARGET_ARCH=$(TARGET_ARCH) \
		microshift-builder:$(RPM_BUILDER_IMAGE_TAG) \
		bash -ilc 'cd /opt/microshift && make rpm & pid=$$! ; trap "pkill $${pid}" INT ; wait $${pid}'
.PHONY: rpm-podman

###############################
# dev targets                 #
###############################

clean:
	if [ -d '$(OUTPUT_DIR)' ]; then rm -rf '$(OUTPUT_DIR)'; fi
.PHONY: clean

vendor:
	go mod vendor
	for p in $(sort $(wildcard scripts/auto-rebase/rebase_patches/*.patch)); do \
		echo "Applying patch $$p"; \
		git mailinfo /dev/null /dev/stderr 2<&1- < $$p | git apply --reject || exit 1; \
	done
.PHONY: vendor

# Update the etcd dependencies, including especially MicroShift itself.
vendor-etcd:
	$(MAKE) -C etcd vendor
.PHONY: vendor-etcd

# There should be no modified files in the etcd/vendor directory after
# running `make vendor-etcd`.
.PHONY: verify-vendor-etcd
verify: verify-vendor-etcd
verify-vendor-etcd: vendor-etcd
	./scripts/verify/verify-vendor-etcd.sh

# Use helper `go generate script` to dynamically config information into packaging info as well as documentation.
.PHONY: generate-config verify-config
generate-config:
	./scripts/fetch_tools.sh controller-gen && \
	go generate -mod vendor ./pkg/config

verify-config: generate-config
	./scripts/verify/verify-config.sh

# Run all of the end to end tests
.PHONY: e2e
e2e:
	./test/run.sh

.PHONY: rf-fmt
rf-fmt:
	@./test/format.sh