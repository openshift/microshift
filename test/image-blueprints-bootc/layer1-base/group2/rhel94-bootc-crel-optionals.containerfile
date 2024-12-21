FROM localhost/rhel94-test-agent:latest

# Copy repository configuration
COPY ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-crel.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift with optional components and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y firewalld \
        "microshift-{{ .Env.CURRENT_RELEASE_VERSION }}" "microshift-release-info-{{ .Env.CURRENT_RELEASE_VERSION }}" \
        "microshift-olm-{{ .Env.CURRENT_RELEASE_VERSION }}" "microshift-multus-{{ .Env.CURRENT_RELEASE_VERSION }}"  && \
    systemctl enable microshift && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    dnf clean all

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
