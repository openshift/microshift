# {{- if env.Getenv "BREW_LREL_RELEASE_VERSION" "" -}}
# Note: This comment makes templating add a new line before the code
FROM localhost/rhel102-bootc-brew-lrel-optional:latest

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

# AI Model Serving is x86_64-only and is exercised exclusively by the
# ai-model-serving-online-fips scenario, which boots this image. Install it here
# instead of in the shared lrel-optional base so non-AI scenarios are not bloated.
# {{- if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) }}
ARG USHIFT_RPM_REPO_NAME=microshift-brew
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH
COPY ./bootc-images/$USHIFT_RPM_REPO_NAME.repo ./bootc-images/microshift-fast-datapath-rhel9.repo ./bootc-images/microshift-rhocp-y.repo \
    /etc/yum.repos.d/
RUN dnf install -y \
        {{ range (env.Getenv "MICROSHIFT_X86_64_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "BREW_LREL_RELEASE_VERSION" }}" \
        {{ end -}}
        && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
# {{- end }}
# {{- end -}}
