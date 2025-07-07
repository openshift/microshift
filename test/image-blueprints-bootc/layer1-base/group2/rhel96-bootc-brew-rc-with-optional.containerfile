# {{- if env.Getenv "BREW_RC_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel96-test-agent:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-brew
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift, test agent and cleanup.
RUN dnf repoinfo --enabled && \
    dnf install -y firewalld systemd-resolved \
        "microshift-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-greenboot-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-networking-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-selinux-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-olm-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-multus-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-gateway-api-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-observability-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-release-info-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-olm-release-info-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-multus-release-info-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        # {{- if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) }}
        "microshift-ai-model-serving-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        "microshift-ai-model-serving-release-info-{{ .Env.BREW_RC_RELEASE_VERSION }}" \
        # {{- end }}
        "microshift-gateway-api-release-info-{{ .Env.BREW_RC_RELEASE_VERSION }}" && \
    systemctl enable microshift microshift-test-agent && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

# Configure firewall
RUN firewall-offline-cmd --zone=public --add-port=22/tcp && \
    firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16 && \
    firewall-offline-cmd --zone=trusted --add-source=169.254.169.1 && \
    firewall-offline-cmd --zone=trusted --add-source=fd01::/48 && \
    firewall-offline-cmd --zone=public --add-port=80/tcp && \
    firewall-offline-cmd --zone=public --add-port=443/tcp && \
    firewall-offline-cmd --zone=public --add-port=5353/udp && \
    firewall-offline-cmd --zone=public --add-port=6443/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/tcp && \
    firewall-offline-cmd --zone=public --add-port=30000-32767/udp
# {{- end -}}
