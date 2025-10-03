# {{- if env.Getenv "BREW_Y0_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel96-test-agent:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-brew
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift, test agent and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y firewalld systemd-resolved \
        {{ range (env.Getenv "MICROSHIFT_MANDATORY_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "BREW_Y0_RELEASE_VERSION" }}" \
        {{ end -}}
        {{ range (env.Getenv "MICROSHIFT_OPTIONAL_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "BREW_Y0_RELEASE_VERSION" }}" \
        {{ end -}}
        {{ if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) -}}
        {{ range (env.Getenv "MICROSHIFT_X86_64_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "BREW_Y0_RELEASE_VERSION" }}" \
        {{ end -}}
        {{ end -}}
        && \
    systemctl enable microshift microshift-test-agent && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
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
    firewall-offline-cmd --zone=public --add-port=8889/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/udp

# Prepare system for testing Generic Device Plugin.
# Upgrade the kernel to keep the same procedure with RHEL and CentOS.
# CentOS requires upgrade because of a different package retention policy
# which means that the kernel in the base bootc image might no longer
# be available in the repositories.
# hadolint ignore=DL3003
RUN dnf upgrade kernel -y && \
KERNEL_VER=$(rpm -q --qf "%{VERSION}-%{RELEASE}" kernel); \
KERNEL_VER_ARCH="${KERNEL_VER}.$(uname -m)"; \
dnf install -y git make "kernel-devel-${KERNEL_VER}" python3-pyserial && \
dnf clean all && \
git clone https://github.com/pmtk/serialsim.git /tmp/serialsim && \
cd /tmp/serialsim && \
make KERNEL="${KERNEL_VER_ARCH}" all install && \
rm -rf /tmp/serialsim

# {{- end -}}
