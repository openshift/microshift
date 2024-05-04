FROM localhost/rhel94-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY $USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# Add the following repositories, using host entitlements for SSL keys:
# - MicroShift local RPM repository
# - Fast Datapath repository for MicroShift dependencies
# - OpenShift previous minor version repository for MicroShift dependencies
RUN sslkey=$(find /etc/pki/entitlement-host/ -type f -name "*-key.pem" -print -quit) && \
    sslcrt=$(find /etc/pki/entitlement-host/ -type f -name "*.pem" ! -name "*-key.pem" -print -quit) && \
    printf "\
[microshift-local]\n\
name=MicroShift Local Repository\n\
baseurl=%s\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" "${USHIFT_RPM_REPO_PATH}" > "/etc/yum.repos.d/microshift-local.repo" && \
    printf "\
[openshift-fast-datapath]\n\
name=Fast Datapath for RHEL 9\n\
baseurl=https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Env.UNAME_M }}/fast-datapath/os\n\
enabled=1\n\
gpgcheck=1\n\
gpgkey = file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release\n\
sslverify=1\n\
sslcacert = /etc/rhsm/ca/redhat-uep.pem\n\
sslclientkey = %s\n\
sslclientcert = %s\n\
skip_if_unavailable=0\n" "${sslkey}" "${sslcrt}" > "/etc/yum.repos.d/openshift-fast-datapath.repo" && \
    printf "\
[openshift-rhocp]\n\
name=OpenShift Dependencies RHEL 9\n\
baseurl=https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Env.UNAME_M }}/rhocp/4.{{ .Env.PREVIOUS_MINOR_VERSION }}/os\n\
enabled=1\n\
gpgcheck=1\n\
gpgkey = file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release\n\
sslverify=1\n\
sslcacert = /etc/rhsm/ca/redhat-uep.pem\n\
sslclientkey = %s\n\
sslclientcert = %s\n\
skip_if_unavailable=0\n" "${sslkey}" "${sslcrt}" > "/etc/yum.repos.d/openshift-rhocp.repo"

# Install MicroShift optional RPMs
RUN dnf install -y microshift-olm microshift-multus && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    rm -f /etc/yum.repos.d/microshift*.repo && \
    rm -f /etc/yum.repos.d/openshift*.repo && \
    dnf clean all
