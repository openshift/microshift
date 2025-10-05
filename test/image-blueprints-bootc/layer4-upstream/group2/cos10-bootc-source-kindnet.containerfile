FROM localhost/cos9-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-centos9-nfv.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift optional packages and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y "microshift-kindnet-{{ env.Getenv "SOURCE_VERSION" }}" && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

# Kindnet depends on containernetworking-plugins, which is not available in the
# CentOS 10 repositories. Install it from the package GitHub release page.
RUN /bin/bash <<'EOF'
    set -euo pipefail
    set -x
    # Set the package version and name
    CNP_VER=v1.8.0
    CNP_PKG="cni-plugins-linux-amd64-${CNP_VER}.tgz"
    [ "$(uname -m)" = "aarch64" ] && CNP_PKG="cni-plugins-linux-arm64-${CNP_VER}.tgz"

    # Download the package
    curl -sSL --retry 5 -o "/tmp/${CNP_PKG}" \
        "https://github.com/containernetworking/plugins/releases/download/${CNP_VER}/${CNP_PKG}"

    # Extract the package into the CNI plugins directory as defined 
    # in the crio.conf.d/13-microshift-kindnet.conf file.
    mkdir -p /usr/libexec/cni
    tar zxvf "/tmp/${CNP_PKG}" -C /usr/libexec/cni && \

    # Clean up
    rm -f "/tmp/${CNP_PKG}"
EOF
