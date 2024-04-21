FROM registry.redhat.io/rhel9-beta/rhel-bootc:9.4

# Build arguments
ARG PREVIOUS_MINOR_VERSION=15
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
ADD $USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# MicroShift local RPM repository
RUN printf "\
[microshift-local]\n\
name=MicroShift Local Repository\n\
baseurl=$USHIFT_RPM_REPO_PATH\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" > "/etc/yum.repos.d/microshift-local.repo"

# OpenShift Mirror beta RPM repository for MicroShift dependencies
# Only the released previous minor version is guaranteed to be available
RUN printf "\
[openshift-mirror-beta]\n\
name=OpenShift Mirror Beta Repository\n\
baseurl=https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/4.${PREVIOUS_MINOR_VERSION}-el9-beta/\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" > "/etc/yum.repos.d/openshift-mirror-beta.repo"

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
    rm -f /etc/yum.repos.d/{microshift,openshift}*.repo && \
    dnf clean all

# Extend Greenboot wait timeout to 10m for MicroShift to be ready
RUN printf '\nMICROSHIFT_WAIT_TIMEOUT_SEC=600\n' >> /etc/greenboot/greenboot.conf
