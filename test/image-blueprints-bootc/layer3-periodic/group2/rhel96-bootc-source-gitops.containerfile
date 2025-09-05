FROM localhost/rhel96-bootc-source:latest

# Copy repository configuration
COPY ./bootc-images/microshift-gitops.repo \
    /etc/yum.repos.d/

# Print repository configuration contents.
# Install MicroShift and cleanup.
RUN dnf install -y microshift-gitops && \
    rm -vf /etc/yum.repos.d/microshift-*.repo && \
    dnf clean all
