#!/bin/sh
set -e

# Usage:
# ENV_VAR= ./install.sh
#
# Environment variables:
#
# - ROLES="controlplane,node"
#   Option by default, but controlplane role can be run standalone.

# Function to get Linux distribution
get_distro() {
    DISTRO=$(egrep '^(ID)=' /etc/os-release| sed 's/"//g' | cut -f2 -d"=")
}

# If RHEL, use subscription manager to register
register_subs() {
    subscription-manager register --auto-attach
    subscription-manager repos --enable=rhocp-4.7-for-rhel-8-x86_64-rpms
}

apply_selinux_policy() {
    setenforce 0
}

install_dependencies() {
    dnf install -y \
#    policycoreutils-python-utils \
    iptables-services
}

install_crio() {
    if [ "$DISTRO" == "fedora" ]; then
        dnf module enable cri-o:1.20 -y
        dnf install cri-o cri-tools -y
    else [ "$DISTRO" == "rhel" ]
        dnf install cri-o cri-tools -y
    fi
}


crio_conf() {
    sed -i 's/10.85.0.0\/16/10.42.0.0\/24/' /etc/cni/net.d/100-crio-bridge.conf

     if [ "$DISTRO" == "rhel" ]; then
        sed -i 's/\/usr\/libexec\/crio\/conmon/\/usr\/bin\/conmon/' /etc/crio/crio.conf 
     fi
}

verify_crio() {
    systemctl enable crio
    systemctl start crio

}

get_kubectl() {
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x ./kubectl
    mv ./kubectl /usr/bin/kubectl

}

get_distro
apply_selinux_policy
if [ $DISTRO = "rhel" ]; then
    register_subs
fi
install_dependencies
install_crio
#crio_conf
verify_crio
get_kubectl



