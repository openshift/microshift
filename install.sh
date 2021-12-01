#!/usr/bin/env bash
set -e -o pipefail

# Usage:
# ./install.sh

# ENV VARS
# CONFIG_ENV_ONLY=true will short-circuit the installation immediately after the host is configured, and before
# the microshift release is downloaded.  This is to allow developers and CI to configure a host environment for testing
# non-release microshift runtimes.
CONFIG_ENV_ONLY=${CONFIG_ENV_ONLY:=false}

# Only get the version number if installing a release version
[ $CONFIG_ENV_ONLY = false ] && \
  VERSION=$(curl -s https://api.github.com/repos/redhat-et/microshift/releases | grep tag_name | head -n 1 | cut -d '"' -f 4)

# Function to get Linux distribution
get_distro() {
    DISTRO=$(egrep '^(ID)=' /etc/os-release| sed 's/"//g' | cut -f2 -d"=")
    if [[ $DISTRO != @(rhel|fedora|centos|ubuntu) ]]
    then
      echo "This Linux distro is not supported by the install script"
      exit 1
    fi
}

# Function to get system architecture
get_arch() {
    ARCH=$(uname -m | sed "s/x86_64/amd64/" | sed "s/aarch64/arm64/")
    if [[ $ARCH != @(amd64|arm64) ]]
    then
        printf "arch %s unsupported" "$ARCH" >&2
        exit 1
    fi
}

# Function to get OS version
get_os_version() {
    OS_VERSION=$(egrep '^(VERSION_ID)=' /etc/os-release | sed 's/"//g' | cut -f2 -d"=")
}

# Function to check system prerequisites
pre-check-installation(){
    mem_threshold='1024'
    disk_threshold='2048'
    numCPU_threshold='2'
    
    numCPU=$(nproc --all)
    if [ $numCPU -lt $numCPU_threshold ]; then
        echo "Warning: Pre-Install check number of CPUs cores less than recommended number: $numCPU_threshold"
        #uncomment this line to exit on error. By now informative only
        #exit 1
    fi

    mem_free=$(free -m | grep "Mem" | awk '{print $4+$6}')
    if [ $mem_free -lt $mem_threshold ]
    then
        echo "Warning: Pre-Install check MEM usage less than recommended number: $mem_threshold"
        #uncomment this line to exit on error. By now informative only
        #exit 1
    fi

    disk_free=$(df -m | grep /$ | grep -v -E '(tmp|boot)' | awk '{print $4}')
    if [ $disk_free -lt $disk_threshold ]
    then
        echo "Warning: Pre-Install check DISK usage less than recommended number: $disk_threshold"
        #uncomment this line to exit on error. By now informative only
        #exit 1
    fi
}

# If RHEL, use subscription-manager to register
register_subs() {
    set +e +o pipefail
    REPO="rhocp-4.7-for-rhel-8-x86_64-rpms"
    # Check subscription status and register if not
    STATUS=$(sudo subscription-manager status | awk '/Overall Status/ { print $3 }')
    if [[ $STATUS != "Current" ]]
    then
        sudo subscription-manager register --auto-attach < /dev/tty
        POOL=$(sudo subscription-manager list --available --matches '*OpenShift' | grep Pool | head -n1 | awk -F: '{print $2}' | tr -d ' ')
        sudo subscription-manager attach --pool $POOL
        sudo subscription-manager config --rhsm.manage_repos=1
    fi
    set -e -o pipefail
    # Check if already subscribed to the proper repository
    if ! sudo subscription-manager repos --list-enabled | grep -q ${REPO}
    then
        sudo subscription-manager repos --enable=${REPO}
    fi
}

# Apply SElinux policies
build_selinux_policy() {
    ## Workaround until packaged as RPM
    sudo dnf -y install selinux-policy-devel
    curl -L -o /tmp/microshift.fc https://raw.githubusercontent.com/redhat-et/microshift/main/packaging/selinux/microshift.fc
    curl -L -o /tmp/microshift.te https://raw.githubusercontent.com/redhat-et/microshift/main/packaging/selinux/microshift.te
    make -f /usr/share/selinux/devel/Makefile -C /tmp
    sudo dnf -y remove selinux-policy-devel
    if [ "$DISTRO" != "ubuntu" ]; then
        sudo mkdir -p /var/run/flannel
        sudo mkdir -p /var/run/kubelet
        sudo mkdir -p /var/lib/kubelet/pods
        sudo mkdir -p /var/run/secrets/kubernetes.io/serviceaccount
        sudo mkdir -p /var/hpvolumes
        sudo semodule -i /tmp/microshift.pp
        sudo restorecon -v /var/hpvolumes
	sudo restorecon -vR /var/lib/kubelet/pods
    fi
}

# Install dependencies
install_dependencies() {
    if [ "$DISTRO" = "ubuntu" ]; then
        sudo apt-get install -y \
            policycoreutils-python-utils \
            conntrack \
            firewalld
    else
        sudo dnf install -y \
            policycoreutils-python-utils \
            conntrack \
            firewalld
    fi
}

# Establish Iptables rules
establish_firewall () {
    sudo systemctl enable firewalld --now
    sudo firewall-cmd --zone=public --permanent --add-port=6443/tcp
    sudo firewall-cmd --zone=public --permanent --add-port=30000-32767/tcp
    sudo firewall-cmd --zone=public --permanent --add-port=2379-2380/tcp
    sudo firewall-cmd --zone=public --add-masquerade --permanent
    sudo firewall-cmd --zone=public --add-port=80/tcp --permanent
    sudo firewall-cmd --zone=public --add-port=443/tcp --permanent
    sudo firewall-cmd --zone=public --add-port=10250/tcp --permanent
    sudo firewall-cmd --zone=public --add-port=10251/tcp --permanent
    sudo firewall-cmd --permanent --zone=trusted --add-source=10.42.0.0/16
    sudo firewall-cmd --reload
}


# Install CRI-O depending on the distro
install_crio() {
    CRIOVERSION=1.21

    case $DISTRO in
      "fedora")
        sudo dnf module -y enable cri-o:$CRIOVERSION
        sudo dnf install -y cri-o cri-tools
      ;;
      "rhel")
        sudo dnf install cri-o cri-tools -y
      ;;
      "centos")
        OS=CentOS_8_Stream
        sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/devel:kubic:libcontainers:stable.repo
        sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable:cri-o:$CRIOVERSION.repo https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable:cri-o:$CRIOVERSION/$OS/devel:kubic:libcontainers:stable:cri-o:$CRIOVERSION.repo
        sudo dnf install -y cri-o cri-tools
      ;;
      "ubuntu")
        OS=xUbuntu_$OS_VERSION
        KEYRINGS_DIR=/usr/share/keyrings
        echo "deb [signed-by=$KEYRINGS_DIR/libcontainers-archive-keyring.gpg] https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/ /" | sudo tee /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list > /dev/null
        echo "deb [signed-by=$KEYRINGS_DIR/libcontainers-crio-archive-keyring.gpg] http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/$CRIOVERSION/$OS/ /" | sudo tee /etc/apt/sources.list.d/devel:kubic:libcontainers:stable:cri-o:$CRIOVERSION.list > /dev/null

        sudo mkdir -p $KEYRINGS_DIR
        curl -L https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/$OS/Release.key | sudo gpg --dearmor -o $KEYRINGS_DIR/libcontainers-archive-keyring.gpg
        curl -L https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/$CRIOVERSION/$OS/Release.key | sudo gpg --dearmor -o $KEYRINGS_DIR/libcontainers-crio-archive-keyring.gpg

        sudo apt-get update -y
        # Vagrant Ubuntu VMs don't provide containernetworking-plugins by default
        sudo apt-get install -y cri-o cri-o-runc cri-tools containernetworking-plugins
      ;;
    esac
}


# CRI-O config to match MicroShift networking values
crio_conf() {
    sudo sh -c 'cat << EOF > /etc/cni/net.d/100-crio-bridge.conf
{
    "cniVersion": "0.4.0",
    "name": "crio",
    "type": "bridge",
    "bridge": "cni0",
    "isGateway": true,
    "ipMasq": true,
    "hairpinMode": true,
    "ipam": {
        "type": "host-local",
        "routes": [
            { "dst": "0.0.0.0/0" }
        ],
        "ranges": [
            [{ "subnet": "10.42.0.0/24" }]
        ]
    }
}
EOF'
    
     if [ "$DISTRO" == "rhel" ]; then
        sudo sed -i 's|/usr/libexec/crio/conmon|/usr/bin/conmon|' /etc/crio/crio.conf 
     fi
}

# Start CRI-O
verify_crio() {
    sudo systemctl enable crio
    sudo systemctl restart crio
}

# Download and install kubectl
get_kubectl() {
    curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/$ARCH/kubectl"
    sudo chmod +x ./kubectl
    sudo mv ./kubectl /usr/local/bin/kubectl
}

# Download and install microshift
get_microshift() {
    curl -LO https://github.com/redhat-et/microshift/releases/download/$VERSION/microshift-linux-$ARCH
    curl -LO https://github.com/redhat-et/microshift/releases/download/$VERSION/release.sha256

    BIN_SHA="$(sha256sum microshift-linux-$ARCH | awk '{print $1}')"
    KNOWN_SHA="$(grep "microshift-linux-$ARCH" release.sha256 | awk '{print $1}')"

    if [[ "$BIN_SHA" != "$KNOWN_SHA" ]]; then 
        echo "SHA256 checksum failed" && exit 1
    fi

    sudo chmod +x microshift-linux-$ARCH
    sudo mv microshift-linux-$ARCH /usr/local/bin/microshift

    cat << EOF | sudo tee /usr/lib/systemd/system/microshift.service
[Unit]
Description=MicroShift
After=crio.service

[Service]
WorkingDirectory=/usr/local/bin/
ExecStart=microshift run
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

    if [ "$DISTRO" = "ubuntu" ] && [ "$OS_VERSION" = "18.04" ]; then
        sudo sed -i 's|^ExecStart=microshift|ExecStart=/usr/local/bin/microshift|' /usr/lib/systemd/system/microshift.service
    fi
    if [ "$DISTRO" != "ubuntu" ]; then
        sudo restorecon -v /usr/local/bin/microshift
    fi
    sudo systemctl enable microshift.service --now
}

# Locate kubeadmin configuration to default kubeconfig location
prepare_kubeconfig() {
    mkdir -p $HOME/.kube
    if [ -f $HOME/.kube/config ]; then
        mv $HOME/.kube/config $HOME/.kube/config.orig
    fi
    sudo KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig:$HOME/.kube/config.orig  /usr/local/bin/kubectl config view --flatten > $HOME/.kube/config
}

# validation checks for deployment 
validation_check(){
    echo $HOSTNAME | grep -P '(?=^.{1,254}$)(^(?>(?!\d+\.)[a-zA-Z0-9_\-]{1,63}\.?)+(?:[a-zA-Z]{2,})$)' && echo "Correct"
    if [ $? != 0 ];
    then
        echo "======================================================================"
        echo "!!! WARNING !!!"
        echo "The hostname $HOSTNAME does not follow FQDN, which might cause problems while operating the cluster."
        echo "See: https://github.com/redhat-et/microshift/issues/176"
        echo
        echo "If you face a problem or want to avoid them, please update your hostname and try again."
        echo "Example: 'sudo hostnamectl set-hostname $HOSTNAME.example.com'"
        echo "======================================================================"
    else
        echo "$HOSTNAME is a valid machine name continuing installation"
    fi
}

# Script execution
get_distro
get_arch
get_os_version
pre-check-installation
if [ "$DISTRO" = "rhel" ]; then
    register_subs
fi
validation_check
install_dependencies
establish_firewall
install_crio
crio_conf
verify_crio
get_kubectl
if [ "$DISTRO" != "ubuntu" ]; then
    build_selinux_policy
fi

[ "$CONFIG_ENV_ONLY" = true ] && { echo "Env config complete" && exit 0 ; }
get_microshift

until sudo test -f /var/lib/microshift/resources/kubeadmin/kubeconfig
do
     sleep 2
done
prepare_kubeconfig
