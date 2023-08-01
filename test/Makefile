REPO := $(shell git rev-parse --show-toplevel)
include $(REPO)/Makefile

NEXT_MINOR = $(shell expr $(MINOR) + 1)
YPLUS2_MINOR = $(shell expr $(MINOR) + 2)

.PHONY: fake-next-minor-rpm
fake-next-minor-rpm: RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-next-minor/
fake-next-minor-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor" \
	RPMBUILD_DIR=$(RPMBUILD_DIR) \
		$(MAKE) -C $(REPO) rpm

.PHONY: fake-yplus2-minor-rpm
fake-yplus2-minor-rpm: RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-yplus2-minor/
fake-yplus2-minor-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(YPLUS2_MINOR).$(PATCH)_fake_YPLUS2_minor" \
	RPMBUILD_DIR=$(RPMBUILD_DIR) \
		$(MAKE) -C $(REPO) rpm

.PHONY: build-base-branch
build-base-branch: BASE_BRANCH :=release-$(MAJOR).$(MINOR)
build-base-branch: TMP_REPO !=mktemp -d
build-base-branch:
	if ! [ -d "$(REPO)/_output" ]; then mkdir $(REPO)/_output; fi
	git clone --no-hardlinks --single-branch --branch $(BASE_BRANCH) "$(REPO)" "$(TMP_REPO)"
	cd $(TMP_REPO) && $(MAKE) rpm
	mv $(TMP_REPO)/_output/rpmbuild $(REPO)/_output/rpmbuild-base
	rm -rf "$(TMP_REPO)"

.PHONY: robotidy
robotidy:
	cd .. && make verify-rf
