FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-centos9-nfv.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y firewalld "microshift-{{ .Env.SOURCE_VERSION }}" "microshift-release-info-{{ .Env.SOURCE_VERSION }}" && \
    systemctl enable microshift && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
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
