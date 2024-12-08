FROM localhost/rhel95-bootc-source:latest

# Add fips=1 kernel argument
# See https://containers.github.io/bootc/building/kernel-arguments.html
RUN cat > /usr/lib/bootc/kargs.d/01-fips.toml <<'EOF'
kargs = ["fips=1"]
match-architectures = ["x86_64"]
EOF

# Enable the FIPS crypto policy
RUN dnf install -y crypto-policies-scripts && \
    update-crypto-policies --no-reload --set FIPS && \
    dnf clean all
