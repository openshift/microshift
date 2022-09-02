#!/bin/bash
set -exo pipefail

sudo dnf install -y git osbuild-composer composer-cli \
    cockpit-composer bash-completion podman genisoimage \
    createrepo syslinux yum-utils selinux-policy-devel jq wget lorax
sudo systemctl enable osbuild-composer.socket --now
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd -q --add-service=cockpit
sudo firewall-cmd -q --add-service=cockpit --permanent

# The mock utility comes from the EPEL repository
sudo dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm
sudo dnf install -y mock 
sudo usermod -a -G mock $(whoami)
