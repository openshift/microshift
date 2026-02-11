REPO := $(shell git rev-parse --show-toplevel)
include $(REPO)/Makefile

NEXT_MINOR = $(shell expr $(MINOR) + 1)

.PHONY: fake-next-minor-rpm
fake-next-minor-rpm: RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-next-minor/
fake-next-minor-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor" \
	RPMBUILD_DIR=$(RPMBUILD_DIR) \
		$(MAKE) -C $(REPO) rpm

# build-base-branch
# By default, BASE_BRANCH is derived from the microshift/Makefile.version.$(uname -i).var string.
# The BASE_BRANCH is pulled from the remote repo.  This is because OpenShift-CI fetches only PR commit tree
# and the main:HEAD commit tree. This means that we can't rely on the local source when testing PRs
# against non-main branches. The base-branch is used for the initial microshift layer deployed in some rpm-ostree scenarios.
.PHONY: build-base-branch
build-base-branch: BASE_BRANCH ?=release-$(MAJOR).$(MINOR)
build-base-branch: TMP_REPO !=mktemp -d
build-base-branch:
	if ! [ -d "$(REPO)/_output" ]; then mkdir $(REPO)/_output; fi
	git clone --single-branch --branch $(BASE_BRANCH) https://github.com/openshift/microshift.git $(TMP_REPO)
	MICROSHIFT_VERSION="$(MAJOR).$(MINOR).$(PATCH)_base_branch" \
		$(MAKE) -C $(TMP_REPO) rpm
	mv $(TMP_REPO)/_output/rpmbuild $(REPO)/_output/rpmbuild-base
	rm -rf "$(TMP_REPO)"

.PHONY: robotidy
robotidy:
	cd .. && make verify-rf

####################################################
# Testing Framework commands for local development #
####################################################

#
# Configuration commands
#
.PHONY: test-config-composer
test-config-composer:
	./bin/manage_composer_config.sh cleanup
	./bin/manage_composer_config.sh create
	./bin/manage_composer_config.sh create-workers

.PHONY: test-config-hypervisor
test-config-hypervisor:
	./bin/manage_hypervisor_config.sh cleanup
	./bin/manage_hypervisor_config.sh create

.PHONY: test-cache-download
test-cache-download:
	bash -euo pipefail -c '\
		source ./bin/common.sh ; \
		export AWS_BUCKET_NAME=microshift-build-cache-us-west-2 ; \
	 	./bin/manage_build_cache.sh download -b "$${SCENARIO_BUILD_BRANCH}" -t "$${SCENARIO_BUILD_TAG}" ; \
	'

# find_layer: finds a layer directory matching the token
# Usage: $(call find_layer,base_dir,token)
# Example: $(call find_layer,image-blueprints,base) returns ./image-blueprints/layer1-base
define find_layer
$(shell find $(1) -maxdepth 1 -type d -name '*$(2)' -print)
endef

#
# Build commands
#
.PHONY: test-build-registry-clean
test-build-registry-clean:
	./bin/manage_composer_config.sh cleanup
	rm -rf "$(REPO)/_output/test-images/mirror-registry"
	./bin/manage_webserver.sh start

PHONY: test-build-ostree-presubmit
test-build-ostree-presubmit:
	./bin/build_images.sh -l $(call find_layer,image-blueprints,base) $(ARGS)
	./bin/build_images.sh -l $(call find_layer,image-blueprints,presubmit) $(ARGS)

.PHONY: test-build-ostree-periodic
test-build-ostree-periodic:
	./bin/build_images.sh -l $(call find_layer,image-blueprints,periodic) $(ARGS)

.PHONY: test-build-ostree-release
test-build-ostree-release:
	./bin/build_images.sh -l $(call find_layer,image-blueprints,release) $(ARGS)

.PHONY: test-build-bootc-presubmit
test-build-bootc-presubmit:
	./bin/build_bootc_images.sh -l $(call find_layer,image-blueprints-bootc,base) $(ARGS)
	./bin/build_bootc_images.sh -l $(call find_layer,image-blueprints-bootc,presubmit) $(ARGS)

.PHONY: test-build-bootc-periodic
test-build-bootc-periodic:
	./bin/build_bootc_images.sh -l $(call find_layer,image-blueprints-bootc,periodic) $(ARGS)

.PHONY: test-build-bootc-release
test-build-bootc-release:
	./bin/build_bootc_images.sh -l $(call find_layer,image-blueprints-bootc,release) $(ARGS)

.PHONY: test-build-bootc-upstream
test-build-bootc-upstream:
	./bin/build_bootc_images.sh -l $(call find_layer,image-blueprints-bootc,upstream) $(ARGS)
