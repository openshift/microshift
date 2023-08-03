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

.PHONY: local-main-rpm
local-main-rpm: TMP_REPO !=mktemp -d
local-main-rpm: RPMBUILD_DIR :=$(REPO)/_output/rpmbuild-main/
local-main-rpm: COMMIT !=git rev-parse main
local-main-rpm:
	git clone --single-branch --branch main --no-hardlinks "$(REPO)" "$(TMP_REPO)"
	cd "$(TMP_REPO)" && \
		MICROSHIFT_VERSION="$(MAJOR).$(MINOR).$(PATCH)-main-$(COMMIT)" \
		RPMBUILD_DIR="$(RPMBUILD_DIR)"\
		$(MAKE) rpm
	rm -rf "$(TMP_REPO)"

.PHONY: robotidy
robotidy:
	cd .. && make verify-rf
