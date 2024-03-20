#FROM registry-proxy.engineering.redhat.com/rh-osbs/rhel9-rhel_bootc:rhel-9.4
FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments with default values
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

# OpenShift Mirror beta RPM repository
# Only the released previous minor version is guaranteed to be available
# TODO: make repository y-ver dynamic
RUN cat > "/etc/yum.repos.d/openshift-mirror-beta.repo" <<EOF
[openshift-mirror-beta]
name=OpenShift Mirror Beta Repository
baseurl=https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rpms/4.15-el9-beta/
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF

RUN dnf install -y vi microshift && \
    systemctl enable microshift.service && \
    rm -rf $MICROSHIFT_REPO_PATH && \
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
