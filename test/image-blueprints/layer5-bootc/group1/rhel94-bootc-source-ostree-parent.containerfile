FROM localhost/rhel94-ostree-microshift-previous-minor:latest

# Build arguments
ARG REPO_CONFIG_SCRIPT=/tmp/microshift_repo_config.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Work around the missing ostree-container functionality:
# - Enable librhsm which enables host subscriptions to work in containers
# - Image has /opt symlink pointing to a non-existent /var/opt, so create the latter
RUN ln -sr /run/secrets/etc-pki-entitlement /etc/pki/entitlement-host && \
    ln -sr /run/secrets/rhsm /etc/rhsm-host && \
    mkdir -p /var/opt

# Copy the repository configuration script
COPY --chmod=755 ./bootc-images/microshift_repo_config.sh ${REPO_CONFIG_SCRIPT}

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Add the following repositories and print their contents:
# - MicroShift local RPM repository
# - OpenShift previous minor version RHEL repositories for MicroShift dependencies
# Install MicroShift, test agent, a few helper utilities and cleanup
#
# Allow erasing because update may require a new version of openvswitch and
# major versions of openvswitch are separate packages that obsolete each other.
RUN ${REPO_CONFIG_SCRIPT} -create ${USHIFT_RPM_REPO_PATH} && \
    awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo && \
    dnf install -y --allowerasing firewalld systemd-resolved microshift microshift-test-agent && \
    systemctl enable microshift microshift-test-agent && \
    ${REPO_CONFIG_SCRIPT} -delete && \
    rm -f ${REPO_CONFIG_SCRIPT} && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
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
