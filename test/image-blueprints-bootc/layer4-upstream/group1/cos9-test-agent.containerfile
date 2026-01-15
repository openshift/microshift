FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Configure the RPM repositories (no EUS on CentOS)
COPY --chmod=755 ./bootc-images/rpm-repo-config.sh /tmp/rpm-repo-config.sh
RUN /tmp/rpm-repo-config.sh && rm -f /tmp/rpm-repo-config.sh

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
