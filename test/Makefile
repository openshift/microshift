REPO := $(shell git rev-parse --show-toplevel)
include $(REPO)/Makefile

RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-next-minor/

NEXT_MINOR = $(shell expr $(MINOR) + 1)

.PHONY: fake-next-minor-rpm
fake-next-minor-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor" \
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
