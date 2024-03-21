FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG PREVIOUS_MINOR_VERSION=15
ARG MICROSHIFT_REPO_NAME=microshift-local
ARG MICROSHIFT_REPO_PATH=/tmp/$MICROSHIFT_REPO_NAME

# Copy the MicroShift repository contents
ADD $MICROSHIFT_REPO_NAME $MICROSHIFT_REPO_PATH

# MicroShift local RPM repository
RUN cat > "/etc/yum.repos.d/microshift-local.repo" <<EOF
[microshift-local]
name=MicroShift Local Repository
baseurl=$MICROSHIFT_REPO_PATH
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF

# OpenShift Mirror beta RPM repository for MicroShift dependencies
# Only the released previous minor version is guaranteed to be available
RUN cat > "/etc/yum.repos.d/openshift-mirror-beta.repo" <<EOF
[openshift-mirror-beta]
name=OpenShift Mirror Beta Repository
baseurl=https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rpms/4.${PREVIOUS_MINOR_VERSION}-el9-beta/
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF

# Implement workarounds necessary for the successful MicroShift operation
# - OVN-k copies an executable to /opt/cni/bin directory on the host, which is
#   read-only on bootc images. Create the /opt/cni/bin -> /var/opt/cni/bin
#   symbolic link as a workaround.
# - The /var/run directory must be a symbolic link to /run, which is not
#   the case in bootc images. This cases problems with services like DBus 
#   and NetworkManager. DBus creates its socket in /run/dbus directory, which
#   is accessed by NetworkManager at /var/run/dbus.
RUN mkdir -p /opt/cni && mkdir -p /var/opt/cni/bin && ln -s /var/opt/cni/bin /opt/cni/ ; \
    [ ! -L /var/run ] && rm -rf /var/run && ln -s /run /var/

# Install MicroShift, few helper utilities and cleanup
RUN dnf install -y vi firewalld microshift && \
    systemctl enable microshift && \
    rm -rf $MICROSHIFT_REPO_PATH && \
    rm -f /etc/yum.repos.d/{microshift,openshift}*.repo && \
    dnf clean all

# Configure crun runtime for crio, required on CentOS 9
RUN mkdir -p /etc/crio/crio.conf.d/ && \
    cat > /etc/crio/crio.conf.d/microshift-crun.conf <<EOF
[crio.runtime.runtimes.crun]
runtime_path = ""
runtime_type = "oci"
runtime_root = "/run/crun"
runtime_config_path = ""
monitor_path = ""
monitor_cgroup = "system.slice"
monitor_exec_cgroup = ""
privileged_without_host_devices = false
EOF
