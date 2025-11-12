FROM localhost/rhel100-bootc-source:latest

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
    dnf install -y \
        {{ range (env.Getenv "MICROSHIFT_OPTIONAL_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "SOURCE_VERSION" }}" \
        {{ end -}}
        {{ if and (env.Getenv "UNAME_M" "") (eq "x86_64" .Env.UNAME_M) -}}
        {{ range (env.Getenv "MICROSHIFT_X86_64_RPMS" | strings.Split " ") -}}
        "{{ . }}-{{ env.Getenv "SOURCE_VERSION" }}" \
        {{ end -}}
        {{ end -}}
        && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

# Prepare system for testing Generic Device Plugin
COPY --chmod=755 ./bootc-images/build-serialsim.sh /tmp/build-serialsim.sh
RUN /tmp/build-serialsim.sh && rm -f /tmp/build-serialsim.sh
