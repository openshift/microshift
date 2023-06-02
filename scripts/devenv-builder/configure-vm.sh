#!/bin/bash
#
# This script automates the VM configuration steps described in the "MicroShift Development Environment" document.
# See https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md
#
set -eo pipefail

BUILD_AND_RUN=true
INSTALL_BUILD_DEPS=true
FORCE_FIREWALL=false
RHEL_SUBSCRIPTION=false
SET_RHEL_RELEASE=true

start=$(date +%s)

function usage() {
    echo "Usage: $(basename "$0") [--no-build] [--no-build-deps] [--force-firewall] [--no-set-release-version] <openshift-pull-secret-file>"
    echo ""
    echo "  --no-build                Do not build, install and start MicroShift"
    echo "  --no-build-deps           Do not install dependencies for building binaries and RPMs (implies --no-build)"
    echo "  --force-firewall          Install and configure firewalld regardless of other options"
    echo "  --no-set-release-version  Do NOT set the release subscription to the current release version"

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

while [ $# -gt 1 ]; do
    case "$1" in
    --no-build)
        BUILD_AND_RUN=false
        shift
        ;;
    --no-build-deps)
        INSTALL_BUILD_DEPS=false
        BUILD_AND_RUN=false
        shift
        ;;
    --force-firewall)
        FORCE_FIREWALL=true
        shift
        ;;
    --no-set-release-version)
        SET_RHEL_RELEASE=false
        shift
        ;;
    *) usage ;;
    esac
done

if [ $# -ne 1 ]; then
    usage "Wrong number of arguments"
fi

# Only RHEL requires a subscription
if grep -q 'Red Hat Enterprise Linux' /etc/redhat-release; then
    RHEL_SUBSCRIPTION=true
fi

OCP_PULL_SECRET=$1
[ ! -e "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret file '${OCP_PULL_SECRET}' does not exist"
OCP_PULL_SECRET=$(realpath "${OCP_PULL_SECRET}")
[ ! -f "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret '${OCP_PULL_SECRET}' is not a regular file"

# Create Development Virtual Machine > Configuring VM
# https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#configuring-vm
echo -e "${USER}\tALL=(ALL)\tNOPASSWD: ALL" | sudo tee "/etc/sudoers.d/${USER}" 

# Check the subscription status and register if necessary
if ${RHEL_SUBSCRIPTION}; then
    if ! sudo subscription-manager status >&/dev/null; then
        sudo subscription-manager register
    fi

    if ${SET_RHEL_RELEASE}; then
        # https://access.redhat.com/solutions/238533
        source /etc/os-release
        sudo subscription-manager release --set ${VERSION_ID}
        sudo subscription-manager release --show
        sudo dnf clean all -y
    fi
fi

if ${INSTALL_BUILD_DEPS} || ${BUILD_AND_RUN}; then
    sudo dnf clean all -y
    sudo dnf update -y
    sudo dnf install -y git cockpit make golang jq selinux-policy-devel rpm-build jq bash-completion
    sudo systemctl enable --now cockpit.socket
fi

if ${BUILD_AND_RUN}; then
    # Build MicroShift
    # https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#build-microshift
    if [ ! -e ~/microshift ]; then
        git clone https://github.com/openshift/microshift.git -b release-4.13 ~/microshift
    fi
    cd ~/microshift

    # Build MicroShift > RPM Packages
    # https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#rpm-packages
    make clean
    make rpm
    make srpm
fi

# Run MicroShift Executable > Runtime Prerequisites
# https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#runtime-prerequisites
if ${RHEL_SUBSCRIPTION}; then
    OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)
    sudo subscription-manager config --rhsm.manage_repos=1
    sudo subscription-manager repos \
        --enable "rhocp-4.13-for-rhel-${OSVERSION}-$(uname -m)-rpms" \
        --enable "fast-datapath-for-rhel-${OSVERSION}-$(uname -m)-rpms"
else
    sudo dnf install -y centos-release-nfv-common
    sudo dnf copr enable -y @OKD/okd "centos-stream-9-$(uname -m)"
    sudo tee "/etc/yum.repos.d/openvswitch2-$(uname -m)-rpms.repo" >/dev/null <<EOF
[sig-nfv]
name=CentOS Stream 9 - SIG NFV
baseurl=http://mirror.stream.centos.org/SIGs/9-stream/nfv/\$basearch/openvswitch-2/
gpgcheck=1
enabled=1
skip_if_unavailable=0
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-SIG-NFV
EOF
fi

if ${BUILD_AND_RUN}; then
    sudo dnf localinstall -y ~/microshift/_output/rpmbuild/RPMS/*/*.rpm
fi

if [ ! -e "/etc/crio/openshift-pull-secret" ]; then
    sudo mkdir -p /etc/crio/
    sudo cp -f "${OCP_PULL_SECRET}" /etc/crio/openshift-pull-secret
    sudo chmod 600 /etc/crio/openshift-pull-secret
fi

# Run MicroShift Executable > Installing Clients
# https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#installing-clients
if ${RHEL_SUBSCRIPTION}; then
    sudo dnf install -y openshift-clients 
else
    OCC_REM=https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/clients/ocp-dev-preview/latest-4.13/openshift-client-linux.tar.gz
    OCC_LOC=$(mktemp /tmp/openshift-client-linux-XXXXX.tar.gz)

    curl -s "${OCC_REM}" --output "${OCC_LOC}"
    sudo tar zxf "${OCC_LOC}" -C /usr/bin
    rm -f "${OCC_LOC}"
fi

if ${BUILD_AND_RUN} || ${FORCE_FIREWALL}; then
    # Run MicroShift Executable > Configuring MicroShift > Firewalld
    # https://github.com/openshift/microshift/blob/main/docs/howto_firewall.md#firewalld
    sudo dnf install -y firewalld
    sudo systemctl enable firewalld --now 
    sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16 
    sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1 
    sudo firewall-cmd --permanent --zone=public --add-port=80/tcp 
    sudo firewall-cmd --permanent --zone=public --add-port=443/tcp 
    sudo firewall-cmd --permanent --zone=public --add-port=5353/udp 
    sudo firewall-cmd --permanent --zone=public --add-port=30000-32767/tcp 
    sudo firewall-cmd --permanent --zone=public --add-port=30000-32767/udp 
    sudo firewall-cmd --permanent --zone=public --add-port=6443/tcp 
    sudo firewall-cmd --permanent --zone=public --add-service=mdns 
    sudo firewall-cmd --reload 
fi

if ${BUILD_AND_RUN}; then
    # Run MicroShift Executable > Configuring MicroShift
    # https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#configuring-microshift
    sudo systemctl enable crio 
    sudo systemctl start microshift

    echo ""
    echo "The configuration phase completed. Run the following commands to:"
    echo " - Wait until all MicroShift pods are running"
    echo "      watch sudo \$(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A"
    echo ""
    echo " - Get MicroShift logs"
    echo "      sudo journalctl -u microshift"
    echo ""
    echo " - Get microshift-etcd logs"
    echo "      sudo journalctl -u microshift-etcd.scope"
    echo ""
    echo " - Clean up MicroShift service configuration"
    echo "      echo 1 | sudo /usr/bin/microshift-cleanup-data --all"

fi

end="$(date +%s)"
duration_total_seconds=$((end - start))
duration_minutes=$((duration_total_seconds / 60))
duration_seconds=$((duration_total_seconds % 60))

echo ""
echo "Done in ${duration_minutes}m ${duration_seconds}s"
