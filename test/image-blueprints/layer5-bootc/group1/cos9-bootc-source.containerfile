FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG REPO_CONFIG_SCRIPT=/tmp/microshift_repo_config.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the repository configuration script
COPY --chmod=755 ./bootc-images/microshift_repo_config.sh ${REPO_CONFIG_SCRIPT}

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Add the following repositories and print their contents:
# - MicroShift local RPM repository
# - OpenShift Mirror Beta previous minor version repository for MicroShift dependencies
# Install MicroShift, few helper utilities and cleanup
RUN ${REPO_CONFIG_SCRIPT} -create ${USHIFT_RPM_REPO_PATH} && \
    dnf install -y firewalld systemd-resolved microshift && \
    systemctl enable microshift && \
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
