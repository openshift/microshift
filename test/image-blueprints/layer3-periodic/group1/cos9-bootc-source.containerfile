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

# The openvswitch3.1 dependency
RUN cat > "/etc/yum.repos.d/openvswitch2-rpms.repo" <<EOF
[sig-nfv]
name=CentOS Stream 9 - SIG NFV
baseurl=http://mirror.stream.centos.org/SIGs/9-stream/nfv/\$basearch/openvswitch-2/
gpgcheck=1
enabled=1
skip_if_unavailable=0
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-SIG-NFV
EOF

# The openshift-client dependency from OpenShift Mirror
# Only the released previous minor version is guaranteed to be available
# Note: Do not remove the trailing slash in the URL or some downloads
# may fail depending on a region
# TODO: Make the y-ver dynamic
RUN bash -c '\
    OCPVERSION=4.15 ; \
    OCC_SRC="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/${OCPVERSION}-el9-beta/" ; \
    OCC_LOC="$(mktemp -d /tmp/openshift-client-XXXXXXXX)" ; \
    dnf install -y wget &&  \
    wget -r -nd -l 1 -A "openshift-clients-${OCPVERSION}*.rpm" -P "${OCC_LOC}" "${OCC_SRC}" && \
    ls -l "${OCC_LOC}"/ && \
    dnf localinstall -y "${OCC_LOC}"/*.rpm && \
    rm -rf "${OCC_LOC}" \
'

# Make sure /opt is not a symlink
# See https://github.com/CentOS/centos-bootc-dev/pull/27
#RUN rm -rf /opt && \
#    mkdir -p /opt/cni

# The crio dependency from OKD copr repository
# Install the NFV package certificate
# Install and enable MicroShift
RUN bash -c '\
    dnf copr enable -y @OKD/okd "centos-stream-9-$(uname -m)" && \
    dnf install -y centos-release-nfv-common && \
    dnf install -y microshift && \
    systemctl enable microshift.service \
'

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

# Cleanup the copied repository
RUN rm -rf $MICROSHIFT_REPO_PATH
