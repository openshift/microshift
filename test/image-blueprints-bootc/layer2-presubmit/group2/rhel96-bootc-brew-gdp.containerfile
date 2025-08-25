FROM localhost/rhel96-bootc-brew-ec-with-optional:latest

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
