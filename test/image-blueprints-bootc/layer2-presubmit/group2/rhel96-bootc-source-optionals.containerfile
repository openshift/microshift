FROM localhost/rhel96-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift optional packages and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y \
        {{ range (env.Getenv "MICROSHIFT_OPTIONAL_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "SOURCE_VERSION" }}" \
        {{ end -}}
        {{ if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) -}}
        {{ range (env.Getenv "MICROSHIFT_X86_64_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "SOURCE_VERSION" }}" \
        {{ end -}}
        {{ end -}}
        && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

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
