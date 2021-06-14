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

# Function to get system architecture
get_arch() {
    ARCH=$(uname -m)
}

# If RHEL, use subscription manager to register
register_subs() {
    subscription-manager register --auto-attach
    subscription-manager repos --enable=rhocp-4.7-for-rhel-8-x86_64-rpms
}

# Apply SElinux policies (Permissive mode at the moment)
apply_selinux_policy() {
    setenforce 0
}

# Install dependencies
install_dependencies() {
    dnf install -y \
    policycoreutils-python-utils \
    conntrack \
    iptables-services
}

# Install CRI-O depending on the distro
install_crio() {
    if [ "$DISTRO" == "fedora" ]; then
        dnf module -y enable cri-o:1.20
        dnf install -y cri-o cri-tools
    fi
    if [ "$DISTRO" == "rhel" ]; then
        dnf install cri-o cri-tools -y
    else [ "$DISTRO" == "centos" ]
        VERSION=1.20
        OS=CentOS_8_Stream
        curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/devel:kubic:libcontainers:stable.repo
        curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable:cri-o:$VERSION/$OS/devel:kubic:libcontainers:stable:cri-o:$VERSION.repo
        dnf install -y cri-o cri-tools
    fi
}


# CRI-O config to match Microshift networking values
crio_conf() {
    sed -i 's/10.85.0.0\/16/10.42.0.0\/24/' /etc/cni/net.d/100-crio-bridge.conf

     if [ "$DISTRO" == "rhel" ]; then
        sed -i 's/\/usr\/libexec\/crio\/conmon/\/usr\/bin\/conmon/' /etc/crio/crio.conf 
     fi
}

# Start CRI-O
verify_crio() {
    systemctl enable crio
    systemctl start crio

}

# Download and install kubectl
get_kubectl() {
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x ./kubectl
    mv ./kubectl /usr/bin/kubectl

}

# Download and install microshift
get_microshift() {
    if [ $ARCH = "x86_64" ]; then
        curl -O https://github.com/redhat-et/microshift/releases/download/v0.2/microshift-linux-amd64 -o microshift
        curl https://github.com/redhat-et/microshift/releases/download/v0.2/release.sha256 
    fi

    SHA=$(sha256sum microshift)
    if [[ $SHA != $(cat release.sha256) ]]; then echo "SHA256 checksum failed" && exit 1; fi

    chmod +x microshift
    mv microshift /usr/bin/

    cat << EOF > /usr/lib/systemd/system/microshift.service
[Unit]
Description=Microshift

[Service]
WorkingDirectory=/usr/bin/
ExecStart=microshift run
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

    systemctl enable microshift.service --now

}


# Script execution
get_distro
get_arch
apply_selinux_policy
if [ $DISTRO = "rhel" ]; then
    register_subs
fi
install_dependencies
install_crio
crio_conf
verify_crio
get_kubectl



