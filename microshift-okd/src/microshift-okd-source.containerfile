FROM quay.io/centos-bootc/centos-bootc:stream9

# Build arguments
ARG REPO_CONFIG_SCRIPT=/tmp/create_repos.sh
ARG OKD_CONFIG_SCRIPT=/tmp/configure.sh
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/${USHIFT_RPM_REPO_NAME}

# Copy the repository configuration script
COPY --chmod=755 ./create_repos.sh ${REPO_CONFIG_SCRIPT}
COPY --chmod=755 ./configure.sh ${OKD_CONFIG_SCRIPT}

# Copy the MicroShift repository contents
# COPY /home/microshift/microshift/_output/rpmbuild/ $USHIFT_RPM_REPO_PATH
COPY rpmbuild/ $USHIFT_RPM_REPO_PATH

# Add the following repositories and print their contents:
# - MicroShift local RPM repository
# - OpenShift Mirror Beta previous minor version repository for MicroShift dependencies
# Install MicroShift optional packages and cleanup
RUN ${REPO_CONFIG_SCRIPT} ${USHIFT_RPM_REPO_PATH} && \
    dnf install -y microshift-olm microshift-multus && \
    ${REPO_CONFIG_SCRIPT} -delete && \
    rm -f ${REPO_CONFIG_SCRIPT} && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all

RUN ${OKD_CONFIG_SCRIPT} && rm -rf ${OKD_CONFIG_SCRIPT}