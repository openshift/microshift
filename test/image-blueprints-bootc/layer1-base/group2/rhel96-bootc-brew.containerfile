# {{- if env.Getenv "BREW_VERSION" "" -}}
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
        {{ range (env.Getenv "MICROSHIFT_MANDATORY_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "BREW_VERSION" }}" \
        {{ end -}}
        && \
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
