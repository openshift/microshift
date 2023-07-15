REPO := $(shell git rev-parse --show-toplevel)
include $(REPO)/Makefile

RPMBUILD_DIR := $(REPO)/_output/rpmbuild-fake-next-minor/

NEXT_MINOR = $(shell expr $(MINOR) + 1)

.PHONY: fake-next-minor fake-next-minor-rpm
fake-next-minor:
	@echo $(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor

fake-next-minor-rpm:
	rm -rf $(RPMBUILD_DIR)
	MICROSHIFT_VERSION="$(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor" \
	RPMBUILD_DIR=$(RPMBUILD_DIR) \
		$(MAKE) -C $(REPO) rpm
