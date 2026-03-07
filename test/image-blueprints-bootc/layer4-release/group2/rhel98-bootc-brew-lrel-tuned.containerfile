# {{- if env.Getenv "BREW_LREL_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel98-bootc-brew-lrel-optional:latest

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
