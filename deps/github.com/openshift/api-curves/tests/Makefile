# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
# Expected a full x.y.z semantic version.
# DO NOT UPDATE THIS until you have published downstream builds for the kubebuilder tools, and checked
# that there is an equivalent upstream version as well. Check upstream at https://storage.googleapis.com/kubebuilder-tools.
# Publish downstream with `make -C tools publish-kubebuilder-tools`, see help text there for updating flags before publshing.
ENVTEST_K8S_VERSION = 1.32.1

# In case of emergency, use these to provide separate upstream/downstream K8s versions for testing.
ENVTEST_K8S_DOWNSTREAM_VERSION = ${ENVTEST_K8S_VERSION}
ENVTEST_K8S_UPSTREAM_VERSION = ${ENVTEST_K8S_VERSION}

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
ENVTEST = go run ${PROJECT_DIR}/vendor/sigs.k8s.io/controller-runtime/tools/setup-envtest

.PHONY: integration
integration: verify-vendor test

.PHONY: test
test: test-downstream

.PHONY: test-upstream
test-upstream: ## Run only the tests. Use upstream Kubebuilder artifacts. Note, if downstream KAS carries changes, this may not pass.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_UPSTREAM_VERSION) -p path --bin-dir $(PROJECT_DIR)/bin)" ./hack/test.sh

.PHONY: test-downstream
test-downstream: ## Run only the tests. Use downstream Kubebuilder artifacts published in a GCS bucket in openshift-gce-devel.
	@# This should be the same as upstream, except it uses the --index option to set to our build index, not upstream's.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_DOWNSTREAM_VERSION) -p path --bin-dir $(PROJECT_DIR)/bin --index https://raw.githubusercontent.com/openshift/api/master/envtest-releases.yaml)" ./hack/test.sh

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: verify-vendor
verify-vendor: vendor
	git diff --exit-code vendor
