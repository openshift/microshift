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
    dnf install -y "microshift-olm-{{ .Env.SOURCE_VERSION }}" \
        "microshift-multus-{{ .Env.SOURCE_VERSION }}" \
# {{- if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) }}
# Currently, RHOAI is only available for x86_64
        "microshift-ai-model-serving-{{ .Env.SOURCE_VERSION }}" \
        "microshift-ai-model-serving-release-info-{{ .Env.SOURCE_VERSION }}" \
# {{- end }}
        "microshift-gateway-api-{{ .Env.SOURCE_VERSION }}" && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
