ARCH ?= $$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')
BASE_IMAGE_URL ?= quay.io/centos/centos
BASE_IMAGE_TAG ?= stream9
DEST_IMAGE_REPO ?= quay.io/microshift/microshift-ci
RHIT_CERT_FILE ?= /etc/pki/ca-trust/source/anchors/2015-RH-IT-Root-CA.pem

.PHONY: all build publish manifest
all:
	@echo "Usage: make <build | publish | manifest>"
	@echo "  build:     Build images locally"
	@echo "  publish:   Publish images at quay.io/microshift"
	@echo "  manifest:  Create a multi-arch manifest of the published images"

build:
	for file in Containerfile.ci* ; do \
		name="$$(awk -F. '{print $$NF}' <<< "$${file}")" ; \
		if [ ! -f "${RHIT_CERT_FILE}" ] ; then \
			echo "ERROR: Red Hat IT certificate file is missing" ; \
			exit 1 ; \
		fi ; \
		crt="$$(cat "${RHIT_CERT_FILE}")" ; \
		podman build \
			--platform "linux/${ARCH}" \
			--build-arg BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
			--build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
			--build-arg BREW_CERT="$${crt}" \
			--tag "${DEST_IMAGE_REPO}":"$${name}"-"${ARCH}" \
			--file "$${file}"|| exit 1; \
	done

publish:
	@echo Enter your credentials for 'quay.io/microshift' if asked ; \
	podman login quay.io/microshift || exit 1; \
	for file in Containerfile.ci* ; do \
		name="$$(awk -F. '{print $$NF}' <<< "$${file}")" ; \
		podman push \
			"${DEST_IMAGE_REPO}":"$${name}"-"${ARCH}" || exit 1; \
	done

# The local platform-specific images are deleted before manifest creation
# to avoid them being included into the manifest, which happens implicitly
# for the current platform. The resulting manifest must contain one image
# per each platform.
manifest:
	@echo Enter your credentials for 'quay.io/microshift' if asked ; \
	podman login quay.io/microshift || exit 1; \
	for file in Containerfile.ci* ; do \
		name="$$(awk -F. '{print $$NF}' <<< "$${file}")" ; \
		podman rmi -f "${DEST_IMAGE_REPO}":"$${name}-amd64" "${DEST_IMAGE_REPO}":"$${name}-arm64" ; \
		podman manifest create -a "localhost/$${name}:latest" \
			"${DEST_IMAGE_REPO}":"$${name}-amd64" \
			"${DEST_IMAGE_REPO}":"$${name}-arm64" || exit 1; \
		podman manifest push \
			"localhost/$${name}:latest" \
			"${DEST_IMAGE_REPO}":"$${name}" || exit 1; \
		podman pull "${DEST_IMAGE_REPO}":"$${name}" || exit 1; \
	done
