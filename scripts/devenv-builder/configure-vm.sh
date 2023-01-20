#!/bin/bash
#
# This script automates the VM configuration steps described in the "MicroShift Development Environment on RHEL 8" document.
# See https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md
#
set -eo pipefail

BUILD_AND_INSTALL=true

function usage() {
    echo "Usage: $(basename $0) [--no-build] <openshift-pull-secret-file>"
    echo "  --no-build   Do not build MicroShift code and install MicroShift RPMs"

    [ ! -z "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

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
if ! sudo subscription-manager status >& /dev/null ; then
   sudo subscription-manager register
fi

# Create Development Virtual Machine > Configuring VM
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#configuring-vm
echo -e 'microshift\tALL=(ALL)\tNOPASSWD: ALL' | sudo tee /etc/sudoers.d/microshift
sudo dnf clean all -y
sudo dnf update -y
sudo dnf install -y git cockpit make golang selinux-policy-devel rpm-build bash-completion
sudo systemctl enable --now cockpit.socket

YQ_URL=https://github.com/mikefarah/yq/releases/download/v4.26.1/yq_linux_$(go env GOARCH)
YQ_HASH_amd64=9e35b817e7cdc358c1fcd8498f3872db169c3303b61645cc1faf972990f37582
YQ_HASH_arm64=8966f9698a9bc321eae6745ffc5129b5e1b509017d3f710ee0eccec4f5568766
yq_hash="YQ_HASH_$(go env GOARCH)"
echo -n "${!yq_hash} -" > /tmp/sum.txt
if ! (curl -Ls "${YQ_URL}" | tee /tmp/yq | sha256sum -c /tmp/sum.txt &>/dev/null); then
    echo "ERROR: Expected file at ${YQ_URL} to have checksum ${!yq_hash} but instead got $(sha256sum </tmp/yq | cut -d' ' -f1)"
    exit 1
fi
chmod +x /tmp/yq && sudo cp /tmp/yq /usr/bin/yq

if $BUILD_AND_INSTALL ; then
    # Build MicroShift
    # https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#build-microshift
    if [ ! -e ~/microshift ] ; then
        git clone https://github.com/openshift/microshift.git ~/microshift
    fi
    cd ~/microshift

    # Build MicroShift > RPM Packages
    # https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#rpm-packages
    make clean
    make rpm
    make srpm
fi

# Run MicroShift Executable > Runtime Prerequisites
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#runtime-prerequisites
sudo subscription-manager config --rhsm.manage_repos=1
sudo subscription-manager repos \
    --enable rhocp-4.12-for-rhel-8-$(uname -i)-rpms \
    --enable fast-datapath-for-rhel-8-$(uname -i)-rpms
if $BUILD_AND_INSTALL ; then
    sudo dnf localinstall -y ~/microshift/_output/rpmbuild/RPMS/*/*.rpm

    sudo cp -f ${OCP_PULL_SECRET} /etc/crio/openshift-pull-secret
    sudo chmod 600                /etc/crio/openshift-pull-secret
fi

# Run MicroShift Executable > Installing Clients
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#installing-clients
sudo dnf install -y openshift-clients

if $BUILD_AND_INSTALL ; then
    # Run MicroShift Executable > Configuring MicroShift > Firewalld
    # https://github.com/openshift/microshift/blob/main/docs/howto_firewall.md#firewalld
    sudo dnf install -y firewalld
    sudo systemctl enable firewalld --now
    sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
    sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
    sudo firewall-cmd --reload

    # Run MicroShift Executable > Configuring MicroShift
    # https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#configuring-microshift
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
