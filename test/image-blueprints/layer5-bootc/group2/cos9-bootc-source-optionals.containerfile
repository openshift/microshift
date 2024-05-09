FROM localhost/cos9-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the repository configuration files
COPY ./bootc-sources/microshift-local.repo /etc/yum.repos.d/
COPY ./bootc-sources/openshift-mirror-beta.repo /etc/yum.repos.d/

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Fix the local repositories to point to the right path
# Print the resulting repository files for troubleshooting purposes
RUN sed -i "s;REPLACE_USHIFT_RPM_REPO_PATH;$USHIFT_RPM_REPO_PATH;g" /etc/yum.repos.d/microshift*.repo && \
    awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo 

# Install MicroShift optional packages and cleanup
RUN dnf install -y microshift-olm microshift-multus && \
    rm -f /etc/yum.repos.d/microshift*.repo && \
    rm -f /etc/yum.repos.d/openshift*.repo && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
