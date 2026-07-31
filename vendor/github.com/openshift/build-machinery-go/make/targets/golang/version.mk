# verify-golang-versions — ensure Go versions are consistent across build sources.
#
# OpenShift repos declare a Go version in up to three places: go.mod,
# Dockerfile (builder image tag), and .ci-operator.yaml (CI build root).
# When these drift apart, builds can silently use the wrong Go version or
# fail in hard-to-diagnose ways. In particular, if go.mod declares a version
# higher than the CI builder, the build fails because GOTOOLCHAIN=local
# prevents Go from downloading a newer toolchain. This target catches
# that drift at verify time by extracting the Go MAJOR.MINOR from each source
# and comparing them.
#
# Rules:
#   1. All CI sources (Dockerfile, .ci-operator.yaml) must declare the same Go version.
#   2. go.mod may declare a version <= the CI version (Go is backward-compatible).
#   3. go.mod must NOT declare a version higher than the CI builder.
#   4. Every extracted version must be a valid MAJOR.MINOR number.
#
# Usage:
#   $(call verify-golang-versions,Dockerfile.rhel7)

include $(addprefix $(dir $(lastword $(MAKEFILE_LIST))), \
	../../lib/golang.mk \
	../../lib/tmp.mk \
)

.empty-golang-versions-files:
	@rm -f "$(PERMANENT_TMP)/golang-versions" "$(PERMANENT_TMP)/named-golang-versions"
.PHONY: .empty-golang-versions-files

verify-golang-versions:
	@if [ -f "$(PERMANENT_TMP)/golang-versions" ]; then \
		GOMOD_VER=""; \
		CI_VER=""; \
		if [ -f "$(PERMANENT_TMP)/named-golang-versions" ]; then \
			GOMOD_VER=$$(grep '^go\.mod:' "$(PERMANENT_TMP)/named-golang-versions" | sed 's/go\.mod: *//'); \
			CI_VER=$$(grep -v '^go\.mod:' "$(PERMANENT_TMP)/named-golang-versions" | sed 's/^[^:]*: *//' | sort | uniq); \
		fi; \
		CI_COUNT=$$(echo "$${CI_VER}" | grep -c . 2>/dev/null || :); \
		if [ "$${CI_COUNT}" -gt 1 ]; then \
			echo "Golang version mismatch:"; \
			cat "$(PERMANENT_TMP)/named-golang-versions" | sort | sed 's/^/- /'; \
			false; \
		elif [ -n "$${GOMOD_VER}" ] && [ -n "$${CI_VER}" ]; then \
			GOMOD_MAJOR=$$(echo "$${GOMOD_VER}" | cut -d. -f1); \
			GOMOD_MINOR=$$(echo "$${GOMOD_VER}" | cut -d. -f2); \
			CI_MAJOR=$$(echo "$${CI_VER}" | cut -d. -f1); \
			CI_MINOR=$$(echo "$${CI_VER}" | cut -d. -f2); \
			if [ "$${GOMOD_MAJOR}" -gt "$${CI_MAJOR}" ] 2>/dev/null || \
			   { [ "$${GOMOD_MAJOR}" -eq "$${CI_MAJOR}" ] 2>/dev/null && [ "$${GOMOD_MINOR}" -gt "$${CI_MINOR}" ] 2>/dev/null; }; then \
				echo "Golang version mismatch:"; \
				cat "$(PERMANENT_TMP)/named-golang-versions" | sort | sed 's/^/- /'; \
				false; \
			fi; \
		fi; \
	fi
.PHONY: verify-golang-versions

# $1 - filename (symbolic, used as postfix in Makefile target)
# $2 - golang version
define verify-golang-version-reference-internal
verify-golang-versions-$(1): .empty-golang-versions-files
verify-golang-versions-$(1):
	@mkdir -p "$(PERMANENT_TMP)"
	@if ! echo "$(2)" | grep -qxE '[0-9]+\.[0-9]+'; then \
		echo "Error: could not extract a valid golang version from $(1) (got '$(2)')"; \
		false; \
	fi
	@echo "$(1): $(2)" >> "$(PERMANENT_TMP)/named-golang-versions"
	@echo "$(2)" >> "$(PERMANENT_TMP)/golang-versions"
.PHONY: verify-golang-versions-$(1)

verify-golang-versions: verify-golang-versions-$(1)
endef

# $1 - filename (symbolic, used as postfix in Makefile target)
# $2 - golang version
define verify-golang-version-reference
$(eval $(call verify-golang-version-reference-internal,$(1),$(2)))
endef

# $1 - Dockerfile filename (symbolic, used as postfix in Makefile target)
define verify-Dockerfile-builder-golang-version
$(call verify-golang-version-reference,$(1),$(shell grep "AS builder" "$(1)" | sed 's/.*golang-\([[:digit:]][[:digit:]]*.[[:digit:]][[:digit:]]*\).*/\1/'))
endef

define verify-go-mod-golang-version
$(call verify-golang-version-reference,go.mod,$(shell grep -e 'go [[:digit:]]*\.[[:digit:]]*' go.mod 2>/dev/null | sed 's/go \([[:digit:]][[:digit:]]*.[[:digit:]][[:digit:]]*\).*/\1/'))
endef

define verify-buildroot-golang-version
$(call verify-golang-version-reference,.ci-operator.yaml,$(shell grep -e 'tag: .*golang-[[:digit:]]*\.[[:digit:]]' .ci-operator.yaml 2>/dev/null | sed 's/.*golang-\([[:digit:]][[:digit:]]*.[[:digit:]][[:digit:]]*\).*/\1/'))
endef

# $1 - optional Dockerfile filename (symbolic, used as postfix in Makefile target)
define verify-golang-versions
$(if $(1),$(call verify-Dockerfile-builder-golang-version,$(1))) \
$(if $(wildcard ./.ci-operator.yaml),$(if $(shell grep 'build_root_image:' .ci-operator.yaml 2>/dev/null),$(call verify-buildroot-golang-version))) \
$(if $(wildcard ./go.mod),$(call verify-go-mod-golang-version))
endef
