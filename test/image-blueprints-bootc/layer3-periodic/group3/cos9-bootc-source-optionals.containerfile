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
# CentOS' package retention policy is different from RHEL's.
# Install 'kernel-devel-matched' even if that means downgrading the kernel:
# it's possible that most recent CentOS kernel doesn't have the devels available yet.
# hadolint ignore=DL3003
RUN dnf install -y git make kernel-devel-matched python3-pyserial && \
    dnf clean all && \
    git clone https://github.com/pmtk/serialsim.git /tmp/serialsim && \
    cd /tmp/serialsim && \
    KERNEL_VER_ARCH="$(rpm -q --qf "%{VERSION}-%{RELEASE}" kernel).$(uname -m)"; \
    make KERNEL="${KERNEL_VER_ARCH}" all install && \
    rm -rf /tmp/serialsim
