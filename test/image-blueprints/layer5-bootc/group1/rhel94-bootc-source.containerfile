FROM registry.redhat.io/rhel9/rhel-bootc:9.4

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY $USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Add the following repositories, using host entitlements for SSL keys:
# - MicroShift local RPM repository
# - Fast Datapath repository for MicroShift dependencies
# - OpenShift previous minor version repository for MicroShift dependencies
RUN sslkey=$(find /etc/pki/entitlement-host/ -type f -name "*-key.pem" -print -quit) && \
    sslcrt=$(find /etc/pki/entitlement-host/ -type f -name "*.pem" ! -name "*-key.pem" -print -quit) && \
    printf "\
[microshift-local]\n\
name=MicroShift Local Repository\n\
baseurl=%s\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" "${USHIFT_RPM_REPO_PATH}" > "/etc/yum.repos.d/microshift-local.repo" && \
    printf "\
[openshift-fast-datapath]\n\
name=Fast Datapath for RHEL 9\n\
baseurl=https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Env.UNAME_M }}/fast-datapath/os\n\
enabled=1\n\
gpgcheck=1\n\
gpgkey = file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release\n\
sslverify=1\n\
sslcacert = /etc/rhsm/ca/redhat-uep.pem\n\
sslclientkey = %s\n\
sslclientcert = %s\n\
skip_if_unavailable=0\n" "${sslkey}" "${sslcrt}" > "/etc/yum.repos.d/openshift-fast-datapath.repo" && \
    printf "\
[openshift-rhocp]\n\
name=OpenShift Dependencies RHEL 9\n\
baseurl=https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Env.UNAME_M }}/rhocp/4.{{ .Env.PREVIOUS_MINOR_VERSION }}/os\n\
enabled=1\n\
gpgcheck=1\n\
gpgkey = file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release\n\
sslverify=1\n\
sslcacert = /etc/rhsm/ca/redhat-uep.pem\n\
sslclientkey = %s\n\
sslclientcert = %s\n\
skip_if_unavailable=0\n" "${sslkey}" "${sslcrt}" > "/etc/yum.repos.d/openshift-rhocp.repo"

# Implement workarounds necessary for the successful MicroShift operation
# - The /var/run directory must be a symbolic link to /run, which is not
#   the case in bootc images. This cases problems with services like DBus
#   and NetworkManager. DBus creates its socket in /run/dbus directory, which
#   is accessed by NetworkManager at /var/run/dbus.
RUN [ ! -L /var/run ] && rm -rf /var/run && ln -s /run /var/

# Install MicroShift, few helper utilities and cleanup
RUN dnf install -y vi firewalld microshift && \
    systemctl enable microshift && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    rm -f /etc/yum.repos.d/microshift*.repo && \
    rm -f /etc/yum.repos.d/openshift*.repo && \
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
