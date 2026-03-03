# Detect go environment to build tools into env specific directories.
# This is helpful for those running tools locally and within containers across different OS/architechture combinations.
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
TOOLS_DIR=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
OUTPUT_DIR=$(TOOLS_DIR)/_output/bin/$(GOOS)/$(GOARCH)

# The CODEGEN_VERSION is set to the commit hash of the most recent commit that touched
# the codegen folder, vendor folder or go.mod and go.sum files.
# This should contain any file that may change the output of the build for codegen.
# Append dirty if the current working tree is dirty for any of the relevant files.
CODEGEN_VERSION=$(shell git log -n 1 --pretty=format:%H -- codegen vendor go.mod go.sum; git diff HEAD --quiet codegen vendor go.mod go.sum || echo -dirty)

# The VENDOR_VERSION is set to the commit hash of the most recent commit that touched
# the vendor folder or go.mod and go.sum files.
# This should contain any file that may change the output of the build for vendor based tooling.
# Append dirty if the current working tree is dirty for any of the relevant files.
VENDOR_VERSION=$(shell git log -n 1 --pretty=format:%H -- vendor go.mod go.sum; git diff HEAD --quiet vendor go.mod go.sum || echo -dirty)

# Tools builds all tools.
tools: codegen controller-gen deepcopy-gen go-to-protobuf openapi-gen protoc-gen-gogo yq

clean:
	rm $(OUTPUT_DIR)/*

.PHONY: run-codegen
run-codegen: codegen
	$(OUTPUT_DIR)/codegen --base-dir $(BASE_DIR) --api-group-versions $(API_GROUP_VERSIONS) --required-feature-sets $(OPENSHIFT_REQUIRED_FEATURESETS)

PULL_SECRET?=${PULL_SECRET:-$(HOME)/pull-secret.txt}

.PHONY: publish-kubebuilder-tools
publish-kubebuilder-tools:
	@# This should only need to be run once per combination of args.
	@# Make sure to update the `-version` to the full semver (x.y.z) of the Kube version that the release is based on (check o/k).
	@# The payload should be any CI build (not nightly/EC) published in the CI registry that matches the defined version.
	@# Pull secret should be the standard pull secret used for spinning up OpenShift clusters, if you do not have one, follow https://docs.google.com/document/d/1ez2jrjiIQJChobfSu2ISJ5uM_-3HGouamL_xIa2_1W4/edit#heading=h.me93l6citmsq
	@# This also expects you to be already authenticated with the OpenShift GCE devel project on Google Cloud, if you are not,
	@# follow the 'Create the Service Account and grant permissions' and 'Generate an access key' @# steps at https://docs.google.com/document/d/1qm37EKkjgoPtjW4909UClzvsjQO5VSpPUvFO_hW_PEg
	go run ./publish-kubebuilder-tools/main.go -version v1.34.1 -output-dir /tmp/kubebuilder-tools -payload=registry.ci.openshift.org/ocp/release:4.21.0-0.ci-2025-10-28-031919 -pull-secret $(PULL_SECRET) -index-file ../envtest-releases.yaml

#############################################
#
# BEGIN: Shortcuts to fully qualified targets
#
#############################################

.PHONY:codegen
codegen: $(OUTPUT_DIR)/codegen

.PHONY:controller-gen
controller-gen: $(OUTPUT_DIR)/controller-gen

.PHONY:deepcopy-gen
deepcopy-gen: $(OUTPUT_DIR)/deepcopy-gen

.PHONY:go-to-protobuf
go-to-protobuf: $(OUTPUT_DIR)/go-to-protobuf

.PHONY:golangci-lint
golangci-lint: $(OUTPUT_DIR)/golangci-lint

.PHONY:kube-api-linter
kube-api-linter: $(OUTPUT_DIR)/kube-api-linter.so

.PHONY:openapi-gen
openapi-gen: $(OUTPUT_DIR)/openapi-gen

.PHONY:prerelease-lifecycle-gen
prerelease-lifecycle-gen: $(OUTPUT_DIR)/prerelease-lifecycle-gen

.PHONY:protoc-gen-gogo
protoc-gen-gogo: $(OUTPUT_DIR)/protoc-gen-gogo

.PHONY:vendor-version
vendor-version: $(OUTPUT_DIR)/vendor-version

.PHONY:yq
yq: $(OUTPUT_DIR)/yq

###########################################
#
# End: Shortcuts to fully qualified targets
#
###########################################

#####################################
#
# BEGIN: Fully qualified tool targets
#
#####################################

# Because the codegen target relies on a phony, it becomes a phony itself.
# We must check the presence of the file before rebuilding.
$(OUTPUT_DIR)/codegen: check-codegen-version
	@ if [ ! -f $(OUTPUT_DIR)/codegen ]; then \
		echo "Building codegen version ${CODEGEN_VERSION}"; \
		go build -mod=vendor -o $(OUTPUT_DIR)/codegen -ldflags="-X main.version=${CODEGEN_VERSION}" ./codegen/cmd; \
	fi

$(OUTPUT_DIR)/controller-gen: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/controller-gen ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen

$(OUTPUT_DIR)/deepcopy-gen: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/deepcopy-gen ./vendor/k8s.io/code-generator/cmd/deepcopy-gen

$(OUTPUT_DIR)/go-to-protobuf: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/go-to-protobuf ./vendor/k8s.io/code-generator/cmd/go-to-protobuf

$(OUTPUT_DIR)/golangci-lint: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/golangci-lint ./vendor/github.com/golangci/golangci-lint/v2/cmd/golangci-lint

$(OUTPUT_DIR)/kube-api-linter.so: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -buildmode=plugin -o $(OUTPUT_DIR)/kube-api-linter.so ./vendor/sigs.k8s.io/kube-api-linter/pkg/plugin
	ln -fs $(OUTPUT_DIR)/kube-api-linter.so $(TOOLS_DIR)/_output/bin/kube-api-linter.so

$(OUTPUT_DIR)/openapi-gen: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/openapi-gen ./vendor/k8s.io/kube-openapi/cmd/openapi-gen

$(OUTPUT_DIR)/protoc-gen-gogo: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/protoc-gen-gogo ./vendor/k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo

$(OUTPUT_DIR)/yq: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/yq ./vendor/github.com/mikefarah/yq/v4

$(OUTPUT_DIR)/prerelease-lifecycle-gen: $(OUTPUT_DIR)/vendor-version
	go build -mod=vendor -o $(OUTPUT_DIR)/prerelease-lifecycle-gen ./vendor/k8s.io/code-generator/cmd/prerelease-lifecycle-gen

###################################
#
# END: Fully qualified tool targets
#
###################################

#################################
#
# BEGIN: version checking targets
#
#################################

# If the vendor version has changed, remove all compiled utils.
# This ensures that when dependencies are updated, the utils are rebuilt.
.PHONY: check-vendor-version
check-vendor-version:
	@ if [ -f $(OUTPUT_DIR)/vendor-version ] && [ "$(VENDOR_VERSION)" != "$$(cat $(OUTPUT_DIR)/vendor-version)" ]; then \
		echo "Tools vendor version mismatch, removing old utils"; \
		rm $(OUTPUT_DIR)/*; \
	fi

# This writes the vendor version to disk so that we can track it over time.
# If the version ever differs from what is written to disk we know the utils need to be rebuilt.
$(OUTPUT_DIR)/vendor-version: check-vendor-version
	@ if [ ! -f $(OUTPUT_DIR)/vendor-version ]; then \
		echo "Writing tools vendor version ${VENDOR_VERSION}"; \
		mkdir -p $(OUTPUT_DIR); \
		echo "${VENDOR_VERSION}" > $(OUTPUT_DIR)/vendor-version; \
	fi


# If the codegen utility is built, and the version doesn't match the latest version,
# Remove it so that it'll be rebuilt before being executed again.
.PHONY: check-codegen-version
check-codegen-version: vendor-version
	@ if [ -f $(OUTPUT_DIR)/codegen ] && [ "$(CODEGEN_VERSION)" != "$$($(OUTPUT_DIR)/codegen --version)" ]; then \
		echo "codegen version mismatch, removing old codegen"; \
		rm $(OUTPUT_DIR)/codegen; \
	fi

###############################
#
# END: version checking targets
#
###############################
