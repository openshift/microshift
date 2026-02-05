FROM registry.stage.redhat.io/rhel9/rhel-bootc:9.8

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Configure the RPM repositories (no EUS in staging environment)
COPY --chmod=755 ./bootc-images/rpm-repo-config.sh /tmp/rpm-repo-config.sh
RUN /tmp/rpm-repo-config.sh && rm -f /tmp/rpm-repo-config.sh

# Configure the RHEL mirror RPM repositories (for use in the staging environment)
ARG RHEL_MIRROR_REPO_NAME=rhel98-mirror.repo
ARG RHEL_MIRROR_REPO_PATH=/etc/yum.repos.d/$RHEL_MIRROR_REPO_NAME
COPY --chmod=644 ./bootc-images/$RHEL_MIRROR_REPO_NAME $RHEL_MIRROR_REPO_PATH

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo /etc/yum.repos.d/

# Install the test agent and cleanup
RUN dnf install -y microshift-test-agent && \
    systemctl enable microshift-test-agent && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
