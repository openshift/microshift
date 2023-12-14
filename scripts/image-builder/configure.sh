#!/bin/bash
set -exo pipefail

OSVERSION=$(awk -F: '{print $5}' /etc/system-release-cpe)

# Necessary for embedding container images
if [ ! -e /etc/osbuild-worker/pull-secret.json ] ; then
    sudo mkdir -p /etc/osbuild-worker
    sudo ln -s /etc/crio/openshift-pull-secret /etc/osbuild-worker/pull-secret.json
    sudo tee /etc/osbuild-worker/osbuild-worker.toml &>/dev/null <<EOF
[containers]
auth_file_path = "/etc/osbuild-worker/pull-secret.json"
EOF
fi

# osbuild from COPR to install version that:
# 1. can build rhel-93 images
# 2. doesn't have issues with unexpected mirror's RPM verification
#    (like the one in 9.3 beta at the moment of writing this comment)
sudo dnf copr enable -y @osbuild/osbuild epel-9-"$(uname -m)"
sudo dnf copr enable -y @osbuild/osbuild-composer rhel-9-"$(uname -m)"

sudo dnf install -y \
    osbuild-composer-96-1.20231213092324044071.main.10.g991293a89.el9 \
    osbuild-101-1.20231213161521815551.main.24.g05c0fd31.el9 \
    git composer-cli ostree rpm-ostree \
    cockpit-composer bash-completion podman runc genisoimage \
    createrepo yum-utils selinux-policy-devel jq wget lorax rpm-build \
    containernetworking-plugins expect

sudo systemctl enable osbuild-composer.socket --now
sudo systemctl enable cockpit.socket --now
sudo firewall-cmd --add-service=cockpit --permanent

# The mock utility comes from the EPEL repository
sudo dnf install -y "https://dl.fedoraproject.org/pub/epel/epel-release-latest-${OSVERSION}.noarch.rpm"
sudo dnf install -y mock nginx tomcli parallel
sudo usermod -a -G mock "$(whoami)"

# Verify umask and home directory permissions
TEST_FILE=$(mktemp /tmp/configure-perm-test.XXXXX)

touch "${TEST_FILE}.file"
mkdir "${TEST_FILE}.dir"
HOME_PERM=$(stat -c 0%a ~)
FILE_PERM=$(stat -c 0%a "${TEST_FILE}.file")
DIR_PERM=$(stat -c 0%a "${TEST_FILE}.dir")

# Set the Correct Permissions for osbuild-composer
[ "${HOME_PERM}" -lt 0711 ]  && chmod go+x ~

if [ "${FILE_PERM}" -lt 0644 ] || [ "${DIR_PERM}" -lt 0711 ] ; then
    echo "Check ${TEST_FILE}.dir permissions. The umask setting must allow execute to group/others"
    echo "Check ${TEST_FILE}.file permissions. The umask setting must allow read to group/others"
    exit 1
fi

rm -rf "${TEST_FILE}"*
