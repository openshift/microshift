ARCH ?= $$(uname -m)
BASE_IMAGE_URL ?= quay.io/centos/centos
BASE_IMAGE_TAG ?= stream9
DEST_IMAGE_REPO ?= quay.io/microshift/microshift-ci

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
			--tag "${DEST_IMAGE_REPO}":"$${tag}"-"${ARCH}" \
			--file "$${file}"|| exit 1; \
	done

publish:
	for file in Containerfile.ci* ; do \
		tag="$$(awk -F. '{print $$NF}' <<< "$${file}")" ; \
		podman push \
			"${DEST_IMAGE_REPO}":"$${tag}"-"${ARCH}" || exit 1; \
	done
