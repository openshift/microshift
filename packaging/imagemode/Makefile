#
# The following variables can be overriden from the command line
# using NAME=value make arguments
#
PULL_SECRET ?= $(HOME)/.pull-secret.json
BASE_IMAGE_URL ?= registry.redhat.io/rhel9/rhel-bootc
BASE_IMAGE_TAG ?= 9.4
DNF_OPTIONS ?=

#
# Define the main targets
#
.PHONY: all
all:
	@echo "make [rhocp | repourl | repobase | <build_ver> | run | clean]"
	@echo "   rhocp:       build a MicroShift bootc image using 'rhocp' repository packages"
	@echo "                with versions specified as 'USHIFT_VER=value'"
	@echo "   repourl:     build a MicroShift bootc image using custom repository URLs"
	@echo "                specified as 'USHIFT_URL=value' and 'OCPDEP_URL=value'"
	@echo "   repobase:    build a MicroShift bootc image using preconfigured repositories"
	@echo "                from the base image specified as 'BASE_IMAGE_URL=value' and"
	@echo "                'BASE_IMAGE_TAG=value'. The produced image version should also"
	@echo "                be provided as 'IMAGE_VER=value' in this case."
	@echo "   <build_ver>: build a MicroShift bootc image of a specific version from the"
	@echo "                available predefined configurations listed below"
	@echo "   run:         run the 'localhost/microshift-\$${IMAGE_VER}' bootc image version"
	@echo "                specified as 'IMAGE_VER=value'"
	@echo "   stop:        stop all running 'microshift-*' containers"
	@echo "   clean:       delete all 'localhost/microshift-*' container images"
	@echo ""
	@echo "Available build versions:"
	@$(MAKE) -qp | awk -F':' '/^4.[^\t]*:([^=]|$$)/ {print "   " $$1}' | sort -u

.PHONY: rhocp
rhocp:
ifndef USHIFT_VER
	$(error Run 'make rhocp USHIFT_VER=value')
endif
	IMAGE_VER=$$(sudo dnf repoquery microshift \
		--quiet \
		--queryformat '%{version}' \
		--repo "rhocp-${USHIFT_VER}-for-rhel-9-$$(uname -m)-rpms" \
		--latest-limit 1) ; \
	IMAGE_NAME="microshift-$${IMAGE_VER}" ; \
	sudo podman build \
		--no-cache \
		--authfile "${PULL_SECRET}" \
		--build-arg BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
		--build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
		--build-arg DNF_OPTIONS="${DNF_OPTIONS}" \
		--build-arg USHIFT_VER="$${USHIFT_VER}" \
		-t "$${IMAGE_NAME}" \
		-f Containerfile.rhocp

.PHONY: repourl
repourl:
ifndef USHIFT_URL
	$(error Run 'make repourl USHIFT_URL=value OCPDEP_URL=value')
endif
ifndef OCPDEP_URL
	$(error Run 'make repourl USHIFT_URL=value OCPDEP_URL=value')
endif
	IMAGE_VER=$$(sudo dnf repoquery microshift \
		--quiet \
		--queryformat '%{version}' \
		--disablerepo '*' \
		--repofrompath "this,${USHIFT_URL}" \
		--latest-limit 1 | sed 's/~/-/g') ; \
	IMAGE_NAME="microshift-$${IMAGE_VER}" ; \
	sudo podman build \
		--no-cache \
		--authfile "${PULL_SECRET}" \
		--build-arg BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
		--build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
		--build-arg DNF_OPTIONS="${DNF_OPTIONS}" \
		--build-arg USHIFT_VER="$${IMAGE_VER}" \
		--build-arg USHIFT_URL="$${USHIFT_URL}" \
		--build-arg OCPDEP_URL="$${OCPDEP_URL}" \
		-t "$${IMAGE_NAME}" \
		-f Containerfile.repourl

.PHONY: repobase
repobase:
ifndef BASE_IMAGE_URL
	$(error Run 'make repobase BASE_IMAGE_URL=value BASE_IMAGE_TAG=value IMAGE_VER=value')
endif
ifndef BASE_IMAGE_TAG
	$(error Run 'make repobase BASE_IMAGE_URL=value BASE_IMAGE_TAG=value IMAGE_VER=value')
endif
ifndef IMAGE_VER
	$(error Run 'make repobase BASE_IMAGE_URL=value BASE_IMAGE_TAG=value IMAGE_VER=value')
endif
	IMAGE_NAME="microshift-$${IMAGE_VER}" ; \
	sudo podman build \
		--no-cache \
		--authfile "${PULL_SECRET}" \
		--build-arg BASE_IMAGE_URL="${BASE_IMAGE_URL}" \
		--build-arg BASE_IMAGE_TAG="${BASE_IMAGE_TAG}" \
		--build-arg DNF_OPTIONS="${DNF_OPTIONS}" \
		-t "$${IMAGE_NAME}" \
		-f Containerfile.repobase

.PHONY: run
run:
ifndef IMAGE_VER
	$(error Run 'make run IMAGE_VER=value')
endif
	IMAGE_NAME="microshift-${IMAGE_VER}" ; \
	sudo modprobe openvswitch ; \
	CONT_ID=$$(sudo podman run --rm -d --privileged \
		-v "${PULL_SECRET}":/etc/crio/openshift-pull-secret:ro \
		-v /var/lib/containers/storage:/var/lib/containers/storage \
		--replace \
		--name "$${IMAGE_NAME}" \
	    "localhost/$${IMAGE_NAME}") ; \
	echo "sudo podman exec -it $${CONT_ID} bash"

.PHONY: stop
stop:
	CONT_IDS=$$(sudo podman ps --format "{{.Names}}" | grep '^microshift-' || true) ; \
	if [ -n "$${CONT_IDS}" ] ; then sudo podman stop $${CONT_IDS} ; fi

.PHONY: clean
clean:
	IMAGES=$$(sudo podman images --format "{{.Repository}}" | grep ^localhost/microshift- || true) ; \
	if [ -n "$${IMAGES}" ] ; then sudo podman rmi -f $${IMAGES} ; fi

#
# Predefined container image builds
#
.PHONY: 4.16-el94
4.16-el94:
	@$(MAKE) rhocp USHIFT_VER=4.16

.PHONY: 4.17-el94
4.17-el94:
	@$(MAKE) rhocp USHIFT_VER=4.17

.PHONY: 4.18-ec-cos9
4.18-ec-cos9:
	@$(MAKE) repourl \
		BASE_IMAGE_URL=quay.io/centos-bootc/centos-bootc \
		BASE_IMAGE_TAG=stream9 \
		DNF_OPTIONS="--allowerasing --nobest" \
		USHIFT_URL=https://mirror.openshift.com/pub/openshift-v4/$$(uname -m)/microshift/ocp-dev-preview/latest-4.18/el9/os/ \
		OCPDEP_URL=https://mirror.openshift.com/pub/openshift-v4/$$(uname -m)/dependencies/rpms/4.18-el9-beta/ \

.PHONY: 4.18-ec-el94
4.18-ec-el94:
	@$(MAKE) repourl \
		USHIFT_URL=https://mirror.openshift.com/pub/openshift-v4/$$(uname -m)/microshift/ocp-dev-preview/latest-4.18/el9/os/ \
		OCPDEP_URL=https://mirror.openshift.com/pub/openshift-v4/$$(uname -m)/dependencies/rpms/4.18-el9-beta/

#
# ISO builds are named 'microshift-<make_target>' (the '.iso' suffix is part of
# the make target name) and stored at the ${BIB_OUTBASE} directory.
# Each build uses 'localhost/microshift-<make-target>' container image as input.
#
# BIB_IMAGE_URL ?= registry.redhat.io/rhel9/bootc-image-builder
# BIB_IMAGE_TAG ?= latest
# BIB_OUTBASE ?= $(HOME)/imagemode
# .PHONY: 4.16-el94.iso
# 4.16-el94.iso: 4.16-el94
# 	IMAGE_NAME="microshift-$$(echo $@ | sed 's/\.iso$$//')" ; \
# 	ISO_FNAME="microshift-$@" ; \
# 	BIB_OUTDIR="${BIB_OUTBASE}/$@" ; \
# 	sudo podman pull \
# 		--authfile "${PULL_SECRET}" \
# 		"${BIB_IMAGE_URL}:${BIB_IMAGE_TAG}" && \
# 	sudo mkdir -p "$${BIB_OUTDIR}" && \
# 	sudo podman run \
# 		--rm -i --privileged \
# 		--pull=never \
# 		--security-opt "label=type:unconfined_t" \
# 		-v "$${BIB_OUTDIR}:/output" \
# 		-v "/var/lib/containers/storage:/var/lib/containers/storage" \
# 		"${BIB_IMAGE_URL}:${BIB_IMAGE_TAG}" \
# 		--type anaconda-iso \
# 		--local \
# 		"localhost/$${IMAGE_NAME_URL}" && \
# 	sudo mv -f "$${BIB_OUTDIR}/bootiso/install.iso" "${BIB_OUTBASE}/$${ISO_FNAME}"
