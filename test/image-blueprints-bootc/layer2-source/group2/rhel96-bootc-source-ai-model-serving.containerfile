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
        "microshift-ai-model-serving-{{ .Env.SOURCE_VERSION }}" \
        "microshift-release-info-{{ .Env.SOURCE_VERSION }}" \
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

# Embed images based on contents of release-info RPMs.
# First RUN embeds MicroShift core images.
# Second RUN embeds AI Model Serving images.
# Two RUNs to create separate layer so they don't grow too big.
#
# Note:
# - Retries work around sporadic "cannot set user namespace" podman error.
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    images="$(jq -r '.images | .[]' /usr/share/microshift/release/release-"$(uname -m)".json)" ; \
    for img in ${images} ; do \
        for i in 1 2 3; do \
            GOMAXPROCS=8 podman pull \
                --authfile /run/secrets/pull-secret.json \
                --root /var/lib/containers/storage-preloaded \
                "docker://${img}" && break; \
            if [ $i -eq 3 ] ; then \
                echo "ERROR: Failed to pull ${img} image after 3 attempts"; \
                exit 1; \
            fi; \
            echo "Attempt $i failed. Retrying in 5 seconds..." && sleep 5; \
        done ; \
    done

# Images are filtered:
# - `kserve*` are required for the kserve components
# - `ovms-image` is needed for OpenVino Model Server
# Images of remaining ServingRuntimes are omitted due to their size (they vary between 2 and 5 GiB).
# Embedding them would unnecessarily increase size of the bootc image, size of the ISO installer,
# required VM's disk size, and would greatly increase build and VM creations times.
#
# Additionally, OCI image with resnet50 model is embedded.
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    images="$(jq -r '.images | with_entries(select(.key | startswith("kserve") or startswith("ovms-image"))) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -m)".json)" ; \
    images="${images} quay.io/microshift/ai-testing-model:ovms-resnet50" ; \
    for img in ${images} ; do \
        for i in 1 2 3; do \
            GOMAXPROCS=8 podman pull \
                --authfile /run/secrets/pull-secret.json \
                --root /var/lib/containers/storage-preloaded \
                "docker://${img}" && break; \
            if [ $i -eq 3 ] ; then \
                echo "ERROR: Failed to pull ${img} image after 3 attempts"; \
                exit 1; \
            fi; \
            echo "Attempt $i failed. Retrying in 5 seconds..." && sleep 5; \
        done ; \
    done


# Edit the container storage configuration file to include the new path
RUN sed -i '/^additionalimagestores.*/a\   "/var/lib/containers/storage-preloaded",' /etc/containers/storage.conf

# Apply a workaround to set the SELinux context on the new storage directory and
# also restore 'NET_BIND_SERVICE' capability that is currently lost when including
# images in the container.
#
# Note: This requires setting the additional image stores path to a read-write
# location on the file system. The images will still be treated as read-only by
# the container subsystem.
# See https://github.com/ostreedev/ostree-rs-ext/issues/654
COPY --chmod=755 ./bootc-images/microshift-imagestore-config.sh /usr/bin/microshift-imagestore-config
COPY --chmod=644 ./bootc-images/microshift-imagestore-config.service /etc/systemd/system/microshift-imagestore-config.service
RUN systemctl enable microshift-imagestore-config.service

# Create test data
COPY --chmod=755 ./bootc-images/ai-model-serving-test-data.sh /tmp/ai-model-serving-test-data.sh
RUN /tmp/ai-model-serving-test-data.sh

# {{- end }}
