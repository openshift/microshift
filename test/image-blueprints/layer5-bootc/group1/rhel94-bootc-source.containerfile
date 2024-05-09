FROM registry.redhat.io/rhel9/rhel-bootc:9.4

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the repository configuration files
COPY ./bootc-sources/microshift-local.repo /etc/yum.repos.d/
COPY ./bootc-sources/openshift-fast-datapath.repo /etc/yum.repos.d/
COPY ./bootc-sources/openshift-rhocp.repo /etc/yum.repos.d/

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Fix the local repositories to point to the right path
# Fix the remote repositories to use local SSL client key and certificate
# Print the resulting repository files for troubleshooting purposes
RUN sslkey=$(find /etc/pki/entitlement-host/ -type f -name "*-key.pem" -print -quit) && \
    sslcrt=$(find /etc/pki/entitlement-host/ -type f -name "*.pem" ! -name "*-key.pem" -print -quit) && \
    sed -i "s;REPLACE_USHIFT_RPM_REPO_PATH;$USHIFT_RPM_REPO_PATH;g" /etc/yum.repos.d/microshift*.repo && \
    sed -i "s;REPLACE_SSL_CLIENT_KEY;${sslkey};g" /etc/yum.repos.d/openshift*.repo && \
    sed -i "s;REPLACE_SSL_CLIENT_CRT;${sslcrt};g" /etc/yum.repos.d/openshift*.repo && \
    awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo 

# Implement workarounds necessary for the successful MicroShift operation
# - The /var/run directory must be a symbolic link to /run, which is not
#   the case in bootc images. This cases problems with services like DBus
#   and NetworkManager. DBus creates its socket in /run/dbus directory, which
#   is accessed by NetworkManager at /var/run/dbus.
RUN [ ! -L /var/run ] && rm -rf /var/run && ln -s /run /var/

# Install MicroShift, few helper utilities and cleanup
RUN dnf install -y vi firewalld microshift && \
    systemctl enable microshift && \
    rm -f /etc/yum.repos.d/microshift*.repo && \
    rm -f /etc/yum.repos.d/openshift*.repo && \
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

# Extend Greenboot wait timeout to 10m for MicroShift to be ready
RUN printf '\nMICROSHIFT_WAIT_TIMEOUT_SEC=600\n' >> /etc/greenboot/greenboot.conf
