include ./vendor/github.com/openshift/build-machinery-go/make/golang.mk
include ./vendor/github.com/openshift/build-machinery-go/make/targets/openshift/deps.mk

OUTPUT_DIR := ../_output
GO_BUILD_BINDIR := $(OUTPUT_DIR)/bin

all: export CGO_ENABLED=0

build: export CGO_ENABLED=0

vendor:
	go mod vendor
.PHONY: vendor
