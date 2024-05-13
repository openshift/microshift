FROM registry.redhat.io/rhel9/rhel-bootc:9.4

# Build arguments
ARG REPO_CONFIG_SCRIPT=/tmp/microshift_repo_config.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the repository configuration script
COPY --chmod=755 ./bootc-sources/microshift_repo_config.sh ${REPO_CONFIG_SCRIPT}

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Implement workarounds necessary for the successful MicroShift operation
# - The /var/run directory must be a symbolic link to /run, which is not
#   the case in bootc images. This cases problems with services like DBus
#   and NetworkManager. DBus creates its socket in /run/dbus directory, which
#   is accessed by NetworkManager at /var/run/dbus.
RUN [ ! -L /var/run ] && rm -rf /var/run && ln -s /run /var/

# Add the following repositories and print their contents:
# - MicroShift local RPM repository
# - OpenShift previous minor version RHEL repositories for MicroShift dependencies
# Install MicroShift, few helper utilities and cleanup
RUN ${REPO_CONFIG_SCRIPT} -create ${USHIFT_RPM_REPO_PATH} && \
    awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo && \
    dnf install -y vi firewalld microshift && \
    systemctl enable microshift && \
    ${REPO_CONFIG_SCRIPT} -delete && \
    rm -f ${REPO_CONFIG_SCRIPT} && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

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
