# {{- if env.Getenv "BREW_RC_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel96-bootc-brew-rc-with-optional:latest

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
# {{- end -}}
