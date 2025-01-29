UNAME ?= $$(uname -m)
BASE_IMAGE_URL ?= quay.io/centos/centos
BASE_IMAGE_TAG ?= stream9
QUAY_IMAGE_URL ?= quay.io/microshift/microshift-ci

.PHONY: all build publish
all:
	@echo "Usage: make <build | publish>"
	@echo "  build:     Build images locally"
	@echo "  publish:   Publish images at quay.io/microshift"

build:
	for file in Containerfile.ci* ; do \
		tag="$$(awk -F. '{print $$NF}' <<< "$${file}")" ; \
		podman build \
			--build-arg BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
			--build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
			--build-arg DNF=dnf \
			--tag "${QUAY_IMAGE_URL}":"$${tag}"-"${UNAME}" \
			--file "$${file}"|| exit 1; \
	done

publish:
	for file in Containerfile.ci* ; do \
		tag="$$(awk -F. '{print $$NF}' <<< "$${file}")" ; \
		podman push \
			"${QUAY_IMAGE_URL}":"$${tag}"-"${UNAME}" || exit 1; \
	done
