FROM localhost/cos9-bootc-source:latest

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
