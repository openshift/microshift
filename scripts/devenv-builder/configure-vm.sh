#!/bin/bash
#
# This script automates the VM configuration steps described in the "MicroShift Development Environment on RHEL 8" document.
# See https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md
#
set -eo pipefail

function usage() {
    echo "Usage: $(basename $0) <openshift-pull-secret-file>"
    exit 1
}

if [ $# -ne 1 ] ; then
    usage
fi

OCP_PULL_SECRET=$(realpath $1)
[ ! -e "${OCP_PULL_SECRET}" ] && usage

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
sudo dnf update -y
sudo dnf install -y git cockpit make golang selinux-policy-devel rpm-build bash-completion
sudo systemctl enable --now cockpit.socket

# Build MicroShift
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#build-microshift
if [ ! -e ~/microshift ] ; then 
    git clone https://github.com/openshift/microshift.git ~/microshift
fi
cd ~/microshift

# Build MicroShift > Executable
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#executable
make

# Build MicroShift > RPM Packages
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#rpm-packages
make rpm 
make srpm

# Run MicroShift Executable > Installing Clients
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#installing-clients
OC_ARCHIVE=/tmp/openshift-client-linux.tar.gz
curl -o ${OC_ARCHIVE} -O https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/clients/ocp/stable/openshift-client-linux.tar.gz
sudo tar -xf ${OC_ARCHIVE} -C /usr/local/bin oc kubectl
rm -f ${OC_ARCHIVE}

# Run MicroShift Executable > Runtime Prerequisites
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#runtime-prerequisites
sudo subscription-manager repos --enable rhocp-4.10-for-rhel-8-$(uname -i)-rpms
sudo dnf install -y cri-o cri-tools
sudo systemctl enable crio --now

sudo cp -f ${OCP_PULL_SECRET} /etc/crio/openshift-pull-secret
sudo chmod 600                /etc/crio/openshift-pull-secret

# Run MicroShift Executable > Configuring MicroShift > Firewalld
# https://github.com/openshift/microshift/blob/main/docs/howto_firewall.md#firewalld
sudo dnf install -y firewalld
sudo systemctl enable firewalld --now
sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16 
sudo firewall-cmd --permanent --zone=trusted --add-source=169.254.169.1
sudo firewall-cmd --reload

# Run MicroShift Executable > Configuring MicroShift
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#configuring-microshift
sudo subscription-manager repos --enable fast-datapath-for-rhel-8-$(uname -i)-rpms
sudo dnf install -y ~/microshift/packaging/rpm/_rpmbuild/RPMS/*/*.rpm

sudo systemctl enable microshift --now

echo ""
echo "The configuration phase completed. Run the following commands to:"
echo " - Wait until all MicroShift pods are running"
echo " - Clean up MicroShift service configuration"
echo ""
echo "watch sudo $(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A"
echo "echo 1 | /usr/bin/cleanup-all-microshift-data"
echo ""
echo "Done"
