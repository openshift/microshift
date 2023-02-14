#!/bin/bash
#
# This script automates the VM configuration steps described in the "MicroShift Development Environment" document.
# See https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md
#
set -eo pipefail

BUILD_AND_INSTALL=true
RHEL_SUBSCRIPTION=false

function usage() {
    echo "Usage: $(basename $0) [--no-build] <openshift-pull-secret-file>"
    echo "  --no-build   Do not build MicroShift code and install MicroShift RPMs"

    [ ! -z "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

# Only RHEL requires a subscription
if grep -q 'Red Hat Enterprise Linux' /etc/redhat-release ; then
    RHEL_SUBSCRIPTION=true
fi

if [ $# -ne 1 ] && [ $# -ne 2 ]; then
    usage "Wrong number of arguments"
fi
if [ $# -eq 2 ] ; then
    [ "$1" != "--no-build" ] && usage "Wrong command line argument: $1"
    BUILD_AND_INSTALL=false
    shift
fi

OCP_PULL_SECRET=$1
[ ! -e "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret file '${OCP_PULL_SECRET}' does not exist"
OCP_PULL_SECRET=$(realpath "${OCP_PULL_SECRET}")
[ ! -f "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret '${OCP_PULL_SECRET}' is not a regular file"

if [ "$(whoami)" != "microshift" ] ; then
    echo "This script should be run from 'microshift' user account"
    exit 1
fi

# Check the subscription status and register if necessary
if [ $RHEL_SUBSCRIPTION = true ] ; then
    if ! sudo subscription-manager status >& /dev/null ; then
        sudo subscription-manager register
    fi
fi

# Create Development Virtual Machine > Configuring VM
# https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#configuring-vm
echo -e 'microshift\tALL=(ALL)\tNOPASSWD: ALL' | sudo tee /etc/sudoers.d/microshift
sudo dnf clean all -y
sudo dnf update -y
sudo dnf install -y git cockpit make golang jq selinux-policy-devel rpm-build bash-completion
sudo systemctl enable --now cockpit.socket

# Install go1.19
# This is installed into different location (/usr/local/bin/go) from dnf installed Go (/usr/bin/go) so it doesn't conflict
# /usr/local/bin is before /usr/bin in $PATH so newer one is picked up
GO_VER=1.19.4
GO_ARCH=$([ "$(uname -i)" == "x86_64" ] && echo "amd64" || echo "arm64")
curl -L -o "go${GO_VER}.linux-${GO_ARCH}.tar.gz" "https://go.dev/dl/go${GO_VER}.linux-${GO_ARCH}.tar.gz" &&
    sudo rm -rf "/usr/local/go${GO_VER}" && \
    sudo mkdir -p "/usr/local/go${GO_VER}" && \
    sudo tar -C "/usr/local/go${GO_VER}" -xzf "go${GO_VER}.linux-${GO_ARCH}.tar.gz" --strip-components 1 && \
    sudo rm -rfv /usr/local/bin/{go,gofmt} && \
    sudo ln --symbolic /usr/local/go${GO_VER}/bin/{go,gofmt} /usr/local/bin/ && \
    rm -rfv "go${GO_VER}.linux-${GO_ARCH}.tar.gz"

if [ $BUILD_AND_INSTALL = true ] ; then
    # Build MicroShift
    # https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#build-microshift
    if [ ! -e ~/microshift ] ; then
        git clone https://github.com/openshift/microshift.git ~/microshift
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
if [ $RHEL_SUBSCRIPTION = true ] ; then
    OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)

    sudo subscription-manager config --rhsm.manage_repos=1
    sudo subscription-manager repos \
        --enable rhocp-4.12-for-rhel-${OSVERSION}-$(uname -i)-rpms \
        --enable fast-datapath-for-rhel-${OSVERSION}-$(uname -i)-rpms
else
    sudo dnf install -y centos-release-nfv-common
    sudo dnf copr enable -y @OKD/okd centos-stream-9-$(uname -i)
    sudo tee /etc/yum.repos.d/openvswitch2-$(uname -i)-rpms.repo >/dev/null <<EOF
[sig-nfv]
name=CentOS Stream 9 - SIG NFV
baseurl=http://mirror.stream.centos.org/SIGs/9-stream/nfv/\$basearch/openvswitch-2/
gpgcheck=1
enabled=1
skip_if_unavailable=0
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-SIG-NFV
EOF
fi

if [ $BUILD_AND_INSTALL = true ] ; then
    sudo dnf localinstall -y ~/microshift/_output/rpmbuild/RPMS/*/*.rpm

    sudo cp -f ${OCP_PULL_SECRET} /etc/crio/openshift-pull-secret
    sudo chmod 600                /etc/crio/openshift-pull-secret
fi

# Run MicroShift Executable > Installing Clients
# https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#installing-clients
if [ $RHEL_SUBSCRIPTION = true ] ; then
    sudo dnf install -y openshift-clients
else
    OCC_REM=https://mirror.openshift.com/pub/openshift-v4/$(uname -i)/clients/ocp-dev-preview/latest-4.13/openshift-client-linux.tar.gz
    OCC_LOC=/tmp/openshift-client-linux.tar.gz

    curl -s ${OCC_REM} --output ${OCC_LOC}
    sudo tar zxf ${OCC_LOC} -C /usr/bin
    rm -f ${OCC_LOC}
fi

if [ $BUILD_AND_INSTALL = true ] ; then
    # Run MicroShift Executable > Configuring MicroShift > Firewalld
    # https://github.com/openshift/microshift/blob/main/docs/howto_firewall.md#firewalld
    sudo dnf install -y firewalld
    sudo systemctl enable firewalld --now
    sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
    sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
    sudo firewall-cmd --reload

    # Run MicroShift Executable > Configuring MicroShift
    # https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md#configuring-microshift
    sudo systemctl enable crio
    sudo systemctl start microshift

    echo ""
    echo "The configuration phase completed. Run the following commands to:"
    echo " - Wait until all MicroShift pods are running"
    echo " - Clean up MicroShift service configuration"
    echo ""
    echo "watch sudo \$(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A"
    echo "echo 1 | /usr/bin/cleanup-all-microshift-data"
fi

echo ""
echo "Done"
