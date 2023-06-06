include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

OUTPUT_DIR := ../_output
GO_BUILD_BINDIR := $(OUTPUT_DIR)/bin

# Enable CGO when building microshift binary for access to local libraries.
# Use an environment variable to allow CI to disable when cross-compiling.
export CGO_ENABLED ?= 1

vendor:
	go mod vendor
.PHONY: vendor
