FROM localhost/rhel94-bootc-crel:latest

# Copy repository configuration
COPY ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-crel.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y "microshift-olm-{{ .Env.CURRENT_RELEASE_VERSION }}" "microshift-multus-{{ .Env.CURRENT_RELEASE_VERSION }}" && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    dnf clean all
