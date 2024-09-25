FROM registry.redhat.io/rhel9/rhel-bootc:9.4

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/microshift-local.repo ./bootc-images/fast-datapath-rhel9.repo ./bootc-images/rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift, test agent and cleanup.
RUN awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo && \
    dnf install -y firewalld "microshift-{{ .Env.SOURCE_VERSION }}" "microshift-release-info-{{ .Env.SOURCE_VERSION }}" microshift-test-agent && \
    systemctl enable microshift microshift-test-agent && \
    rm -f "/etc/yum.repos.d/{microshift-local.repo,fast-datapath-rhel9.repo,rhel-y.repo}" && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

# Extend Greenboot wait timeout to 10m for MicroShift to be ready.
# Greenboot configuration must come after RPM install to avoid
# /etc/greenboot/greenboot.conf overwrite by the package update.
RUN printf "\nMICROSHIFT_WAIT_TIMEOUT_SEC=600\n" >> /etc/greenboot/greenboot.conf

# Configure firewall
RUN firewall-offline-cmd --zone=public --add-port=22/tcp && \
    firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16 && \
    firewall-offline-cmd --zone=trusted --add-source=169.254.169.1 && \
    firewall-offline-cmd --zone=public --add-port=80/tcp && \
    firewall-offline-cmd --zone=public --add-port=443/tcp && \
    firewall-offline-cmd --zone=public --add-port=5353/udp && \
    firewall-offline-cmd --zone=public --add-port=6443/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/udp
