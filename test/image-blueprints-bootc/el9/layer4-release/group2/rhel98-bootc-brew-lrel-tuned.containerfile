# {{- if env.Getenv "BREW_LREL_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel98-test-agent:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-brew
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/

# Install MicroShift with low-latency support and cleanup.
# Optional components are excluded to reduce etcd pressure during startup
# on CPU-constrained tuned VMs (TuneD isolates most CPUs from system services).
RUN dnf repoinfo --enabled && \
    dnf install -y firewalld systemd-resolved \
        {{ range (env.Getenv "MICROSHIFT_MANDATORY_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "BREW_LREL_RELEASE_VERSION" }}" \
        {{ end -}}
        "microshift-low-latency-{{ env.Getenv "BREW_LREL_RELEASE_VERSION" }}" \
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

# Enable microshift-tuned service
RUN systemctl enable microshift-tuned

# Configure MicroShift for low-latency workloads
RUN cat > /etc/microshift/config.yaml <<'EOF'
kubelet:
  cpuManagerPolicy: static
  cpuManagerPolicyOptions:
    full-pcpus-only: "true"
  cpuManagerReconcilePeriod: 5s
  memoryManagerPolicy: Static
  topologyManagerPolicy: single-numa-node
  reservedSystemCPUs: 0-1
  reservedMemory:
  - limits:
      memory: 1100Mi
    numaNode: 0
  kubeReserved:
    memory: 500Mi
  systemReserved:
    memory: 500Mi
  evictionHard:
    imagefs.available: 15%
    memory.available: 100Mi
    nodefs.available: 10%
    nodefs.inodesFree: 5%
  evictionPressureTransitionPeriod: 0s
EOF

# Configure tuned baseline variables
RUN cat > /etc/tuned/microshift-baseline-variables.conf <<'EOF'
# Isolated cores should be complementary to kubelet's reserved CPUs.
# Isolated and reserved CPUs should contain all online CPUs.
# Core #3 is for testing offlining hence skipped.
isolated_cores=2,4-5
hugepages_size=2M
hugepages=10
additional_args=test1=on test2=true dummy
offline_cpu_set=3
EOF

# Configure MicroShift tuned profile
RUN cat > /etc/microshift/tuned.yaml <<'EOF'
profile: microshift-baseline
reboot_after_apply: True
EOF
# {{- end -}}
