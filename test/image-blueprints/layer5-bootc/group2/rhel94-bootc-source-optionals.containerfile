FROM localhost/rhel94-bootc-source:latest

# Build arguments
ARG REPO_CONFIG_SCRIPT=/tmp/microshift_repo_config.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the repository configuration script
COPY --chmod=755 ./bootc-images/microshift_repo_config.sh ${REPO_CONFIG_SCRIPT}

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Add the following repositories and print their contents:
# - MicroShift local RPM repository
# - OpenShift previous minor version RHEL repositories for MicroShift dependencies
# Install MicroShift optional packages and cleanup.
RUN ${REPO_CONFIG_SCRIPT} -create ${USHIFT_RPM_REPO_PATH} && \
    dnf install -y systemd-resolved  microshift-olm microshift-multus && \
    ${REPO_CONFIG_SCRIPT} -delete && \
    rm -f ${REPO_CONFIG_SCRIPT} && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
