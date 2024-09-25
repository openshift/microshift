FROM localhost/rhel94-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-fake-next-minor
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Copy repository configuration
COPY ./bootc-images/microshift-fake-next-minor.repo ./bootc-images/fast-datapath-rhel9.repo ./bootc-images/rhocp-y.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift, test agent and cleanup.
RUN awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo && \
    dnf install -y "microshift-4.{{ .Env.FAKE_NEXT_MINOR_VERSION }}.*" && \
    rm -vf /etc/yum.repos.d/microshift-fake-next-minor.repo /etc/yum.repos.d/fast-datapath-rhel9.repo /etc/yum.repos.d/rhocp-y.repo && \
    rm -rvf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
