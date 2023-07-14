REPO := $(shell git rev-parse --show-toplevel)
include $(REPO)/Makefile

NEXT_MINOR = $(shell expr $(MINOR) + 1)

.PHONY: fake-next-minor fake-next-minor-rpm
fake-next-minor:
	@echo $(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor

fake-next-minor-rpm:
	MICROSHIFT_VERSION="$(MAJOR).$(NEXT_MINOR).$(PATCH)_fake_next_minor" \
		$(MAKE) -C $(REPO) rpm
