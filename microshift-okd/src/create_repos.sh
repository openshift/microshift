#/bin/bash


repo_path=$1

USHIFT_LOCAL_REPO_FILE=/etc/yum.repos.d/microshift-local.repo
OCP_MIRROR_REPO_FILE=/etc/yum.repos.d/openshift-mirror-beta.repo
CENTOS_NFV_SIG_REPO_FILE=/etc/yum.repos.d/microshift-sig-nfv.repo

    cat > "${USHIFT_LOCAL_REPO_FILE}" <<EOF
[microshift-local]
name=MicroShift Local Repository
baseurl=${repo_path}
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF

    cat > "${OCP_MIRROR_REPO_FILE}" <<EOF
[openshift-mirror-beta]
name=OpenShift Mirror Beta Repository
baseurl=https://mirror.openshift.com/pub/openshift-v4/amd64/dependencies/rpms/4.17-el9-beta/
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF

    cat > "${CENTOS_NFV_SIG_REPO_FILE}" <<EOF
[nfv-sig]
name=CentOS Stream 9 - SIG NFV
baseurl=http://mirror.stream.centos.org/SIGs/9-stream/nfv/x86_64/openvswitch-2/
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF


