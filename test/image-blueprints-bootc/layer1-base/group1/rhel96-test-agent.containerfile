# Note:
# This is a workaround for the problem described in https://issues.redhat.com/browse/USHIFT-5584.
# Revert to the rhel-9.6 base image when it is in GA.
#
# FROM registry.stage.redhat.io/rhel9-eus/rhel-9.6-bootc:9.6
FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

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
