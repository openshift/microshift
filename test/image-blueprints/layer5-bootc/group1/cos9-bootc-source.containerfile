FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY $USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Add the following repositories:
# - MicroShift local RPM repository
# - OpenShift Mirror Beta previous minor version repository for MicroShift dependencies
RUN printf "\
[microshift-local]\n\
name=MicroShift Local Repository\n\
baseurl=%s\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" "${USHIFT_RPM_REPO_PATH}" > "/etc/yum.repos.d/microshift-local.repo" && \
    printf "\
[openshift-mirror-beta]\n\
name=OpenShift Mirror Beta Repository\n\
baseurl=https://mirror.openshift.com/pub/openshift-v4/{{ .Env.UNAME_M }}/dependencies/rpms/4.{{ .Env.PREVIOUS_MINOR_VERSION }}-el9-beta/\n\
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

# Configure crun runtime for crio, required on CentOS 9
RUN mkdir -p /etc/crio/crio.conf.d/ && \
    printf '\
[crio.runtime.runtimes.crun]\n\
runtime_path = ""\n\
runtime_type = "oci"\n\
runtime_root = "/run/crun"\n\
runtime_config_path = ""\n\
monitor_path = ""\n\
monitor_cgroup = "system.slice"\n\
monitor_exec_cgroup = ""\n\
privileged_without_host_devices = false\n' > "/etc/crio/crio.conf.d/microshift-crun.conf"

# Extend Greenboot wait timeout to 10m for MicroShift to be ready
RUN printf '\nMICROSHIFT_WAIT_TIMEOUT_SEC=600\n' >> /etc/greenboot/greenboot.conf
