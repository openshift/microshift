FROM localhost/cos9-bootc-source:latest

# Build arguments
ARG USHIFT_RPM_REPO_NAME=microshift-local
ARG USHIFT_RPM_REPO_PATH=/tmp/$USHIFT_RPM_REPO_NAME

# Copy the MicroShift repository contents
COPY $USHIFT_RPM_REPO_NAME $USHIFT_RPM_REPO_PATH

# MicroShift local RPM repository
RUN printf "\
[microshift-local]\n\
name=MicroShift Local Repository\n\
baseurl=%s\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" "${USHIFT_RPM_REPO_PATH}" > "/etc/yum.repos.d/microshift-local.repo"

# OpenShift Mirror beta RPM repository for MicroShift dependencies
# Only the released previous minor version is guaranteed to be available
RUN printf "\
[openshift-mirror-beta]\n\
name=OpenShift Mirror Beta Repository\n\
baseurl=https://mirror.openshift.com/pub/openshift-v4/{{ .Env.UNAME_M }}/dependencies/rpms/4.{{ .Env.PREVIOUS_MINOR_VERSION }}-el9-beta/\n\
enabled=1\n\
gpgcheck=0\n\
skip_if_unavailable=0\n" > "/etc/yum.repos.d/openshift-mirror-beta.repo"

# Install MicroShift optional RPMs
RUN dnf install -y microshift-olm microshift-multus && \
    rm -rf $USHIFT_RPM_REPO_PATH && \
    rm -f /etc/yum.repos.d/microshift*.repo && \
    rm -f /etc/yum.repos.d/openshift*.repo && \
    dnf clean all
