#!/bin/bash
set -exo pipefail

sudo dnf install -y git osbuild-composer composer-cli cockpit-composer bash-completion podman genisoimage createrepo syslinux
sudo systemctl enable osbuild-composer.socket --now
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd -q --add-service=cockpit
sudo firewall-cmd -q --add-service=cockpit --permanent
