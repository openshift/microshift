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
    dnf install -y "microshift-olm-{{ .Env.SOURCE_VERSION }}" \
        "microshift-multus-{{ .Env.SOURCE_VERSION }}" \
        "microshift-gateway-api-{{ .Env.SOURCE_VERSION }}" \
        "microshift-observability-{{ .Env.SOURCE_VERSION }}" && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
