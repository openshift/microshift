# {{- if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) }}
# Currently, RHOAI is only available for x86_64
FROM localhost/rhel96-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift optional packages and cleanup.
# qemu-guest-agent is used in offline tests
# vim-common for xxd utility
RUN dnf repoinfo --enabled && \
    dnf install -y \
        "microshift-release-info-{{ .Env.SOURCE_VERSION }}" \
        "microshift-ai-model-serving-{{ .Env.SOURCE_VERSION }}" \
        "microshift-ai-model-serving-release-info-{{ .Env.SOURCE_VERSION }}" \
        vim-common \
        qemu-guest-agent && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    systemctl enable qemu-guest-agent && \
    dnf clean all

# Override the default qemu-ga service configuration on the guest to allow all RPCs.
# BLOCK_RPCS is a deny-list of qemu-guest-agent RPCs to block, such as file read/write, process execution, etc. By
# default, all RPCs are blocked.
# FREEZE_HOOK_PATHNAME is the dir-path containing hook scripts. This override does not change the default value.
# These scripts are executed before or after guest image snapshot ops ("freezing" and "thawing" in libvirt speak).
# No hooks are installed by default.
RUN printf '\
\# This is a systemd environment file, not a shell script.\n\
\# It provides settings for "/lib/systemd/system/qemu-guest-agent.service".\n\
BLOCK_RPCS=\n\
FSFREEZE_HOOK_PATHNAME=/etc/qemu-ga/fsfreeze-hook\n' > /etc/sysconfig/qemu-ga

# Embed images based on contents of release-info RPMs:
# - First RUN embeds MicroShift core images
# - Second RUN embeds AI Model Serving images
#
# Note:
# - Two RUN commands create separate layer so they do not grow too big
COPY --chmod=755 ./bootc-images/microshift-copy-images.sh /usr/bin/microshift-copy-images
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    images="$(jq -r ".images[]" /usr/share/microshift/release/release-"$(uname -m)".json)" ; \
    IMAGE_PULL_LIST="${images}" /usr/bin/microshift-copy-images -pull

# Images are filtered:
# - 'kserve*'' are required for the kserve components
# - 'ovms-image' is needed for OpenVino Model Server
# Images of remaining ServingRuntimes are omitted due to their size (they vary between 2 and 5 GiB).
# Embedding them would unnecessarily increase size of the bootc image, size of the ISO installer,
# required VM disk size, and would greatly increase build and VM creations times.
#
# Additionally, OCI image with resnet50 model is embedded.
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    images="$(jq -r '.images | with_entries(select(.key | startswith("kserve") or startswith("ovms-image"))) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -m)".json)" ; \
    images="${images} quay.io/microshift/ai-testing-model:ovms-resnet50" ; \
    IMAGE_PULL_LIST="${images}" /usr/bin/microshift-copy-images -pull

# Install a systemd drop-in unit to address the problem with image upgrades
# overwriting the container images in additional store. The workaround is to
# copy the images from the pre-loaded to the main container storage directory.
# In this case, it is not necessary to update /etc/containers/storage.conf with
# the additional store path.
# See https://issues.redhat.com/browse/RHEL-75827
RUN mkdir -p /etc/systemd/system/microshift.service.d
COPY --chmod=644 ./bootc-images/microshift-copy-images.conf /etc/systemd/system/microshift.service.d/microshift-copy-images.conf

# Create test data
COPY --chmod=755 ./bootc-images/ai-model-serving-test-data.sh /tmp/ai-model-serving-test-data.sh
RUN /tmp/ai-model-serving-test-data.sh && rm -f /tmp/ai-model-serving-test-data.sh

# {{- end }}
