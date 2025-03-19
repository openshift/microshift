FROM localhost/rhel96-bootc-source:latest

# The qemu-guest-agent is used in offline tests
RUN dnf install -y qemu-guest-agent && \
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

# Embed images based on contents of release-info RPMs
COPY --chmod=755 ./bootc-images/microshift-copy-images.sh /usr/bin/microshift-copy-images
RUN --mount=type=secret,id=pullsecret,dst=/run/secrets/pull-secret.json \
    images="$(jq -r ".images[]" /usr/share/microshift/release/release-"$(uname -m)".json)" ; \
    images="${images} quay.io/microshift/busybox:1.36" ; \
    IMAGE_PULL_LIST="${images}" /usr/bin/microshift-copy-images -pull

# Install a systemd drop-in unit to address the problem with image upgrades
# overwriting the container images in additional store. The workaround is to
# copy the images from the pre-loaded to the main container storage directory.
# In this case, it is not necessary to update /etc/containers/storage.conf with
# the additional store path.
# See https://issues.redhat.com/browse/RHEL-75827
RUN mkdir -p /etc/systemd/system/microshift.service.d
COPY --chmod=644 ./bootc-images/microshift-copy-images.conf /etc/systemd/system/microshift.service.d/microshift-copy-images.conf
