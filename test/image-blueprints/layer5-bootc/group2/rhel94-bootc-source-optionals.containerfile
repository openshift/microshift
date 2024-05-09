FROM localhost/rhel94-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the repository configuration files
COPY ./bootc-sources/microshift-local.repo /etc/yum.repos.d/
COPY ./bootc-sources/openshift-fast-datapath.repo /etc/yum.repos.d/
COPY ./bootc-sources/openshift-rhocp.repo /etc/yum.repos.d/

# Copy the MicroShift repository contents
COPY ./rpm-repos/$USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Fix the local repositories to point to the right path
# Fix the remote repositories to use local SSL client key and certificate
# Print the resulting repository files for troubleshooting purposes
RUN sslkey=$(find /etc/pki/entitlement-host/ -type f -name "*-key.pem" -print -quit) && \
    sslcrt=$(find /etc/pki/entitlement-host/ -type f -name "*.pem" ! -name "*-key.pem" -print -quit) && \
    sed -i "s;REPLACE_USHIFT_RPM_REPO_PATH;$USHIFT_RPM_REPO_PATH;g" /etc/yum.repos.d/microshift*.repo && \
    sed -i "s;REPLACE_SSL_CLIENT_KEY;${sslkey};g" /etc/yum.repos.d/openshift*.repo && \
    sed -i "s;REPLACE_SSL_CLIENT_CRT;${sslcrt};g" /etc/yum.repos.d/openshift*.repo && \
    awk 'FNR==1 {print "=== " FILENAME " ==="} {print}' /etc/yum.repos.d/*.repo 

# Install MicroShift optional packages and cleanup
RUN dnf install -y microshift-olm microshift-multus && \
    rm -f /etc/yum.repos.d/microshift*.repo && \
    rm -f /etc/yum.repos.d/openshift*.repo && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    dnf clean all
