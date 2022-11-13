#!/bin/bash
#
# This script automates the VM configuration steps described in the "MicroShift Development Environment on RHEL 8" document.
# See https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md
#
set -eo pipefail

function usage() {
    echo "Usage: $(basename $0) <openshift-pull-secret-file>"
    [ ! -z "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

if [ $# -ne 1 ]; then
    usage "Wrong number of arguments"
fi

OCP_PULL_SECRET=$(realpath $1)
[ ! -f "${OCP_PULL_SECRET}" ] && usage "OpenShift pull secret ${OCP_PULL_SECRET} does not exist or is not a regular file."

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

# Build MicroShift
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#build-microshift
if [ ! -e ~/microshift ] ; then
    git clone https://github.com/openshift/microshift.git ~/microshift
fi
cd ~/microshift

# Build MicroShift > RPM Packages
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#rpm-packages
make rpm
make srpm

# Run MicroShift Executable > Runtime Prerequisites
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#runtime-prerequisites
sudo tee /etc/yum.repos.d/rhocp-4.12-el8-beta-$(uname -i)-rpms.repo >/dev/null <<EOF
[rhocp-4.12-el8-beta-$(uname -i)-rpms]
name=Beta rhocp-4.12 RPMs for RHEL8
baseurl=https://mirror.openshift.com/pub/openshift-v4/\$basearch/dependencies/rpms/4.12-el8-beta/
enabled=1
gpgcheck=0
skip_if_unavailable=1
EOF

sudo subscription-manager repos \
    --enable fast-datapath-for-rhel-8-$(uname -i)-rpms
#    --enable rhocp-4.12-for-rhel-8-$(uname -i)-rpms \
sudo dnf localinstall -y ~/microshift/_output/rpmbuild/RPMS/*/*.rpm

sudo cp -f ${OCP_PULL_SECRET} /etc/crio/openshift-pull-secret
sudo chmod 600                /etc/crio/openshift-pull-secret

# Run MicroShift Executable > Installing Clients
# https://github.com/openshift/microshift/blob/main/docs/devenv_rhel8.md#installing-clients
sudo dnf install -y openshift-clients

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
echo "watch sudo $(which oc) --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig get pods -A"
echo "echo 1 | /usr/bin/cleanup-all-microshift-data"
echo ""
echo "Done"
