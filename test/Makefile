REPO := $(shell git rev-parse --show-toplevel)
include $(REPO)/Makefile

NEXT_MINOR = $(shell expr $(MINOR) + 1)
YPLUS2_MINOR = $(shell expr $(MINOR) + 2)

.PHONY: fake-y1-rpm
fake-y1-rpm: RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-y1/
fake-y1-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_y1" \
	RPMBUILD_DIR=$(RPMBUILD_DIR) \
		$(MAKE) -C $(REPO) rpm

.PHONY: fake-y2-rpm
fake-y2-rpm: RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-y2/
fake-y2-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(YPLUS2_MINOR).$(PATCH)_fake_y2" \
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
	cd $(TMP_REPO) && $(MAKE) rpm
	mv $(TMP_REPO)/_output/rpmbuild $(REPO)/_output/rpmbuild-base
	rm -rf "$(TMP_REPO)"

.PHONY: robotidy
robotidy:
	cd .. && make verify-rf
