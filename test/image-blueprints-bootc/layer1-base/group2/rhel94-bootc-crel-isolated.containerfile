# {{- if env.Getenv "CURRENT_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel94-test-agent:latest

# Copy repository configuration
COPY ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-crel.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y firewalld "microshift-{{ .Env.CURRENT_RELEASE_VERSION }}" "microshift-release-info-{{ .Env.CURRENT_RELEASE_VERSION }}" && \
    systemctl enable microshift && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    dnf clean all

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

# Configure firewall
RUN firewall-offline-cmd --zone=public --add-port=22/tcp && \
    firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16 && \
    firewall-offline-cmd --zone=trusted --add-source=169.254.169.1 && \
    firewall-offline-cmd --zone=trusted --add-source=fd01::/48 && \
    firewall-offline-cmd --zone=public --add-port=80/tcp && \
    firewall-offline-cmd --zone=public --add-port=443/tcp && \
    firewall-offline-cmd --zone=public --add-port=5353/udp && \
    firewall-offline-cmd --zone=public --add-port=6443/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/udp    
# {{- end -}}
