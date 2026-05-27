FROM registry.redhat.io/rhel9-eus/rhel-9.6-bootc:9.6

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Lock the OS release version and enable EUS repositories to prevent
# the installation of packages belonging to a newer OS release
# hadolint ignore=SC1091
RUN source /etc/os-release && \
    echo "${VERSION_ID}" > /etc/dnf/vars/releasever && \
    dnf config-manager --set-disabled '*' && \
    dnf config-manager --set-enabled \
        "rhel-9-for-$(uname -m)-baseos-eus-rpms" \
        "rhel-9-for-$(uname -m)-appstream-eus-rpms"

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo /etc/yum.repos.d/

# Print repository configuration contents.
# Install test agent and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y microshift-test-agent && \
    systemctl enable microshift-test-agent && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
