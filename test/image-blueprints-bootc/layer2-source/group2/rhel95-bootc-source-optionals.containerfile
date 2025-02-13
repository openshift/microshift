FROM localhost/rhel95-bootc-source:latest

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
# Make sure that systemd-resolved is installed as it is used in tests.
RUN dnf repoinfo --enabled && \
    dnf install -y "microshift-olm-{{ .Env.SOURCE_VERSION }}" \
        "microshift-multus-{{ .Env.SOURCE_VERSION }}" \
        "microshift-gateway-api-{{ .Env.SOURCE_VERSION }}" \
        systemd-resolved && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
